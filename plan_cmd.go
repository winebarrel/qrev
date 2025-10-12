package qrev

import (
	"database/sql"
	"fmt"
	"path/filepath"

	sq "github.com/Masterminds/squirrel"
	"github.com/fatih/color"
	"github.com/winebarrel/qrev/util"
)

type PlanCmd struct {
	Path       string `arg:"" default:"*.sql" help:"Path of SQL files to run."`
	IfModified bool   `xor:"status" help:"Run if file has modified"`
	ForceRerun bool   `xor:"status" help:"Rerun any failed SQL files."`
}

func (cmd *PlanCmd) Run(options *Options) error {
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
		head, err := t.Head()

		if err != nil {
			return fmt.Errorf("failed to read: %s: %w", t.Path, err)
		}

		if head == "" {
			return fmt.Errorf("file is empty: %s", t.Path)
		}

		name := t.Name

		if t.Rerun {
			name = color.YellowString(name + "*")
		}

		fmt.Fprintln(options.Output, name, head)
	}

	return nil
}

type planOptions struct {
	ifModified bool
	forceRerun bool
}

func plan(db *sql.DB, files []*util.File, options *planOptions) ([]*util.File, error) {
	target := []*util.File{}

	for _, f := range files {
		where := sq.And{
			sq.Eq{"filename": f.Name},
		}

		var status Status
		var hash string
		sel := sq.Select("status", "hash").From(historyTable).Where(where)
		err := sel.RunWith(db).QueryRow().Scan(&status, &hash)

		if err != nil {
			if err == sql.ErrNoRows {
				target = append(target, f)
				continue
			}

			return nil, fmt.Errorf("failed to fetch SQL history: %w", err)
		}

		if options.ifModified {
			if status == StatusFail && hash != f.Hash {
				f.Rerun = true
				target = append(target, f)
			}
		} else if options.forceRerun {
			if status == StatusFail {
				f.Rerun = true
				target = append(target, f)
			}
		}
	}

	return target, nil
}
