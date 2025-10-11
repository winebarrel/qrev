package qrev

import (
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/winebarrel/qrev/util"
)

type StatusCmd struct {
	StatusOrFilename *string `arg:"" optional:"" help:"Status or Filename to filter."`
	ShowError        bool    `help:"Show last error message."`
	Count            uint64  `short:"n" help:"Number of output lines."`
}

func (cmd *StatusCmd) Run(options *Options) error {
	db, err := options.Driver.Open()

	if err != nil {
		return err
	}

	defer db.Close()

	sel := sq.Select("filename", "hash", "executed_at", "status", "last_error").
		From(historyTable).OrderBy("executed_at DESC")

	if cmd.StatusOrFilename != nil {
		sel = sel.Where(sq.Or{
			sq.Eq{"status": *cmd.StatusOrFilename},
			sq.Eq{"filename": *cmd.StatusOrFilename},
		})
	}

	if cmd.Count >= 1 {
		sel = sel.Limit(cmd.Count)
	}

	rows, err := sel.RunWith(db).Query()

	if err != nil {
		return fmt.Errorf("failed to fetch SQL history: %w", err)
	}

	hasRows := false

	for rows.Next() {
		hasRows = true

		var (
			filename   string
			hash       string
			executedAt string
			status     Status
			lastError  string
		)

		if err := rows.Scan(&filename, &hash, &executedAt, &status, &lastError); err != nil {
			return fmt.Errorf("failed to scan row: %s", err)
		}

		t, _ := time.Parse(time.RFC3339, executedAt)
		fmt.Fprintln(options.Output,
			t.Local().Format("02 Jan 15:04"), status.Color(), hash[:7], filename)

		if cmd.ShowError && lastError != "" {
			fmt.Fprintln(options.Output, util.FormatError(lastError))
		}
	}

	if !hasRows {
		fmt.Fprintln(options.Output, "No SQL history")
	}

	return nil
}
