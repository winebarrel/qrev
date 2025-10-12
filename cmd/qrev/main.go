package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/winebarrel/qrev"
)

var version string

var cli struct {
	qrev.Options
	Version kong.VersionFlag

	Apply  qrev.ApplyCmd  `cmd:"" help:"Apply SQL files to the database and record their execution history."`
	Init   qrev.InitCmd   `cmd:"" help:"Initialize the SQL execution history table in the database."`
	Mark   qrev.MarkCmd   `cmd:"" help:"Manually mark the status of a SQL file (e.g., done, fail) in the history."`
	Plan   qrev.PlanCmd   `cmd:"" help:"Show the list of SQL files that are planned to be executed."`
	Status qrev.StatusCmd `cmd:"" help:"Display the execution status of SQL files, optionally filtered by status or filename."`
}

func main() {
	kctx := kong.Parse(&cli, kong.Vars{"version": version})
	cli.Output = os.Stdout
	err := kctx.Run(&cli.Options)
	kctx.FatalIfErrorf(err)
}
