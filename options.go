package qrev

import (
	"io"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/winebarrel/qrev/driver"
)

type Options struct {
	DSN     string        `short:"d" required:"" env:"QREV_DSN" help:"DSN for the database to connect to."`
	Driver  driver.Driver `kong:"-"`
	Timeout time.Duration `env:"QREV_TIMEOUT" default:"3m" help:"Transaction timeout duration."`
	IAMAuth bool          `negatable:"" env:"QREV_IAM_AUTH" help:"Use RDS IAM authentication."`
	Color   bool          `negatable:"" env:"QREV_COLOR" short:"C" help:"Colorize output."`
	Output  io.Writer     `kong:"-"`
}

func (options *Options) BeforeApply() error {
	options.Color = isatty.IsTerminal(os.Stdout.Fd())
	return nil
}

func (options *Options) AfterApply() error {
	dri, err := driver.New(options.DSN, options.IAMAuth)

	if err != nil {
		return err
	}

	options.Driver = dri
	color.NoColor = !options.Color

	return nil
}
