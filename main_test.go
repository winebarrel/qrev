package qrev

import (
	"testing"

	"github.com/fatih/color"
)

func TestMain(m *testing.M) {
	color.NoColor = true
	m.Run()
}
