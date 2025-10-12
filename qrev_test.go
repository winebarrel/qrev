package qrev_test

import (
	"os"
	"testing"

	"github.com/fatih/color"
)

func TestMain(m *testing.M) {
	color.NoColor = true
	os.Setenv("TZ", "Asia/Tokyo") // UTC+9
	m.Run()
}
