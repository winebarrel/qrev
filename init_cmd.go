package qrev

import (
	_ "embed"
	"fmt"
)

const (
	historyTable = "qrev_history"
)

//go:embed sql/create_table.sql
var createTableSQL string

//go:embed sql/create_index.sql
var createIndexSQL string

type InitCmd struct {
}

func (cmd *InitCmd) Run(options *Options) error {
	db, err := options.Driver.Open()

	if err != nil {
		return err
	}

	defer db.Close()

	_, err = db.Exec(createTableSQL)

	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	_, err = db.Exec(createIndexSQL)

	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	fmt.Fprintf(options.Output, "%s table has been created\n", historyTable)

	return nil
}
