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

	Apply  qrev.ApplyCmd  `cmd:"" help:"TODO"`
	Init   qrev.InitCmd   `cmd:"" help:"TODO"`
	Mark   qrev.MarkCmd   `cmd:"" help:"TODO"`
	Plan   qrev.PlanCmd   `cmd:"" help:"TODO"`
	Status qrev.StatusCmd `cmd:"" help:"TODO"`
}

func main() {
	kctx := kong.Parse(&cli, kong.Vars{"version": version})
	cli.Output = os.Stdout
	err := kctx.Run(&cli.Options)
	kctx.FatalIfErrorf(err)
}
