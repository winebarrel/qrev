package qrev

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/winebarrel/qrev/driver"
	"github.com/winebarrel/qrev/util"
)

type ApplyCmd struct {
	Path       string `arg:"" default:"*.sql" help:"Path of SQL files to run."`
	IfModified bool   `xor:"status" help:"Run if file has modified"`
	ForceRerun bool   `xor:"status" help:"Rerun any failed SQL files."`
	BeforeSQL  string `help:"SQL statements to execute before applying."`
	Exclude    string `env:"QREV_EXCLUDE" help:"Glob for filenames to exclude from SQL files."`
	Exclusive  bool   `xor:"exclusive" env:"QREV_EXCLUSIVE" help:"Fail if another exclusive apply is running on the database."`
	// A pointer because 0 is a valid value (wait without limit).
	ExclusiveWait *UnsignedDuration `xor:"exclusive" env:"QREV_EXCLUSIVE_WAIT" placeholder:"DURATION" help:"Like --exclusive, but wait up to DURATION for the other apply to finish (0 waits without limit)."`
}

func (cmd *ApplyCmd) Run(options *Options) error {
	paths, err := filepath.Glob(cmd.Path)

	if err != nil {
		return err
	}

	if cmd.Exclude != "" {
		newPaths := []string{}

		for _, f := range paths {
			if m, _ := filepath.Match(cmd.Exclude, f); !m {
				newPaths = append(newPaths, f)
			}
		}

		paths = newPaths
	}

	if len(paths) == 0 {
		return fmt.Errorf("target file not found: %s", cmd.Path)
	}

	files, err := util.PathsToFiles(paths)

	if err != nil {
		return err
	}

	db, err := options.Driver.Open()

	if err != nil {
		return err
	}

	defer db.Close()

	// Take the lock before the history is read, so the plan below cannot come
	// from a state another apply is still changing.
	if cmd.Exclusive || cmd.ExclusiveWait != nil {
		lock, err := cmd.lock(options)

		if err != nil {
			return err
		}

		defer lock.Close()
	}

	targets, err := plan(db, files, &planOptions{
		ifModified: cmd.IfModified,
		forceRerun: cmd.ForceRerun,
	})

	if err != nil {
		return err
	}

	if len(targets) == 0 {
		fmt.Fprintln(options.Output, "No SQL file to run")
		return nil
	}

	for _, t := range targets {
		err := apply(db, t, cmd.BeforeSQL, options)

		if err != nil {
			return err
		}
	}

	return nil
}

func (cmd *ApplyCmd) lock(options *Options) (io.Closer, error) {
	var wait *time.Duration

	if cmd.ExclusiveWait != nil {
		d := time.Duration(*cmd.ExclusiveWait)
		wait = &d
	}

	lock, err := options.Driver.Lock(context.Background(), &driver.LockOptions{
		Wait:   wait,
		Output: options.Output,
	})

	if err != nil {
		if wait == nil && errors.Is(err, driver.ErrLocked) {
			return nil, fmt.Errorf("%w (--exclusive-wait waits for it)", err)
		}

		return nil, err
	}

	return lock, nil
}

type sqlErr struct {
	error
}

func apply(db *sql.DB, f *util.File, beforeSQL string, options *Options) error {
	q, err := f.Read()

	if err != nil {
		return fmt.Errorf("failed to read: %s: %w", f.Path, err)
	}

	now := time.Now()
	var dur time.Duration

	err = util.WithTx(db, options.Timeout, func(ctx context.Context, tx *sql.Tx) error {
		if beforeSQL != "" {
			_, err := tx.ExecContext(ctx, beforeSQL)

			if err != nil {
				return fmt.Errorf("failed to execute before-SQL: %w", err)
			}
		}

		start := time.Now()
		_, err := tx.ExecContext(ctx, q)
		dur = time.Since(start)

		if err != nil {
			return &sqlErr{error: err}
		}

		return upsertStatus(ctx, tx, f, now, dur, StatusDone, "")
	})

	if err != nil {
		var se *sqlErr

		if !errors.As(err, &se) {
			return err
		}

		updStatErr := util.WithTx(db, options.Timeout, func(ctx context.Context, tx *sql.Tx) error {
			return upsertStatus(ctx, tx, f, now, -1, StatusFail, se.Error())
		})

		if updStatErr != nil {
			return errors.Join(err, updStatErr)
		}

		fmt.Fprintln(options.Output, StatusFail.Color(), f.Name, util.HeadContent(q))
		fmt.Fprintln(options.Output, util.FormatError(se.Error()))
		return errors.New("SQL fails")
	}

	fmt.Fprintln(options.Output, StatusDone.Color(), f.Name, dur, util.HeadContent(q))
	return nil
}

func upsertStatus(ctx context.Context, tx *sql.Tx, f *util.File, now time.Time, dur time.Duration, status Status, lastError string) error {
	del := sq.Delete(historyTable).Where(sq.Eq{"filename": f.Name})
	_, err := del.RunWith(tx).ExecContext(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete history: %w", err)
	}

	ins := sq.Insert(historyTable).Columns("filename", "hash", "executed_at", "execution_time", "status", "last_error").
		Values(f.Name, f.Hash, now.UTC().Format(time.RFC3339), int(dur.Milliseconds()), status, lastError)
	_, err = ins.RunWith(tx).ExecContext(ctx)

	if err != nil {
		return fmt.Errorf("failed to insert history: %w", err)
	}

	return nil
}
