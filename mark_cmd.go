package qrev

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/winebarrel/qrev/util"
)

type MarkCmd struct {
	Status Status `arg:"" enum:"skip,fail" help:"Status to change."`
	Name   string `arg:"" required:"" help:"Filename in SQL history to mark."`
	Noop   bool   `short:"n" help:"No-op mode."`
}

func (cmd *MarkCmd) Run(options *Options) error {
	db, err := options.Driver.Open()

	if err != nil {
		return err
	}

	defer db.Close()

	// check status
	{
		var srcStatus Status

		switch cmd.Status {
		case StatusSkip:
			srcStatus = StatusFail
		case StatusFail:
			srcStatus = StatusSkip
		}

		sel := sq.Select("status").From(historyTable).
			Where(sq.And{sq.Eq{"filename": cmd.Name}})

		var status Status

		if err := sel.RunWith(db).QueryRow().Scan(&status); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("filename not found in SQL history")
			} else {
				return fmt.Errorf("failed to fetch SQL history: %w", err)
			}
		}

		if status != srcStatus {
			return fmt.Errorf("%s cannot be changed to %s", status.Color(), cmd.Status.Color())
		}
	}

	if !cmd.Noop {
		err := util.WithTx(db, options.Timeout, func(ctx context.Context, tx *sql.Tx) error {
			upd := sq.Update(historyTable).Set("status", cmd.Status).
				Where(sq.Eq{"filename": cmd.Name})

			if _, err := upd.RunWith(tx).ExecContext(ctx); err != nil {
				return fmt.Errorf("failed to mark %s: %w", cmd.Status, err)
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	msg := []any{cmd.Status.Color(), cmd.Name}

	if cmd.Noop {
		msg = append(msg, "(noop)")
	}

	fmt.Fprintln(options.Output, msg...)

	return nil
}
