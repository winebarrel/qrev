package qrev

import (
	"github.com/fatih/color"
)

type Status string

const (
	StatusDone Status = "done"
	StatusFail Status = "fail"
	StatusSkip Status = "skip"
)

func (status Status) Color() string {
	s := string(status)

	switch status {
	case StatusDone:
		s = color.GreenString(s)
	case StatusFail:
		s = color.RedString(s)
	case StatusSkip:
		s = color.YellowString(s)
	}

	return s
}
