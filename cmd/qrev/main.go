package main

import (
	"github.com/alecthomas/kong"
	"github.com/winebarrel/qrev"
)

var version string

var cli struct {
	qrev.Options
	Version kong.VersionFlag

	Mark   qrev.MarkCmd   `cmd:"" help:"TODO"`
	Status qrev.StatusCmd `cmd:"" help:"TODO"`
}

func main() {
	kctx := kong.Parse(&cli, kong.Vars{"version": version})
	err := kctx.Run(&cli.Options)
	kctx.FatalIfErrorf(err)
}
