package qrev

import (
	"os"

	"github.com/mattn/go-isatty"
	"github.com/winebarrel/qrev/driver"
)

type Options struct {
	DSN     string `short:"d" required:"" env:"QREV_DSN" help:"DSN for the database to connect to."`
	Driver  driver.Driver
	IAMAuth bool `negatable:"" env:"QREV_IAM_AUTH" help:"Use RDS IAM authentication."`
	Color   bool `negatable:"" env:"QREV_COLOR" short:"C" help:"Colorize output."`
}

func (options *Options) BeforeApply() error {
	options.Color = isatty.IsTerminal(os.Stdout.Fd())
	return nil
}

func (options *Options) AfterApply() error {
	dri, err := driver.New(options.DSN)

	if err != nil {
		return err
	}

	options.Driver = dri

	return nil
}
