package qrev

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/winebarrel/qrev/util"
)

type ApplyCmd struct {
	Path       string `arg:"" default:"*.sql" help:"Path of SQL files to run."`
	IfModified bool   `xor:"status" help:"Run if file has modified"`
	ForceRerun bool   `xor:"status" help:"Rerun any failed SQL files."`
}

func (cmd *ApplyCmd) Run(options *Options) error {
	paths, err := filepath.Glob(cmd.Path)

	if err != nil {
		return err
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
		err := apply(db, t, options)

		if err != nil {
			return err
		}
	}

	return nil
}

func apply(db *sql.DB, f *util.File, options *Options) error {
	q, err := f.Read()

	if err != nil {
		return fmt.Errorf("failed to read: %s: %w", f.Path, err)
	}

	now := time.Now()
	var targetSQLErr error

	err = util.WithTx(db, options.Timeout, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q)
		dur := time.Since(now)

		if err != nil {
			targetSQLErr = err
			return nil
		}

		return upsertStatus(ctx, tx, f, now, dur, StatusDone, "")
	})

	if err != nil {
		return err
	}

	if targetSQLErr != nil {
		err := util.WithTx(db, options.Timeout, func(ctx context.Context, tx *sql.Tx) error {
			return upsertStatus(ctx, tx, f, now, -1, StatusFail, targetSQLErr.Error())
		})

		if err != nil {
			return err
		}

		fmt.Fprintln(options.Output, StatusFail.Color(), f.Name, util.HeadContent(q))
		fmt.Fprintln(options.Output, util.FormatError(targetSQLErr.Error()))
		return errors.New("SQL fails")
	} else {
		fmt.Fprintln(options.Output, StatusDone.Color(), f.Name, util.HeadContent(q))
	}

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
