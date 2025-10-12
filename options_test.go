package qrev_test

import (
	"os"
	"testing"

	"github.com/creack/pty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev"
	"github.com/winebarrel/qrev/driver"
)

func Test_Options_BeforerApply_IsTTY(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	stdout := os.Stdout
	ptmx, _, err := pty.Open()
	require.NoError(err)
	os.Stdout = ptmx

	t.Cleanup(func() {
		os.Stdout = stdout
	})

	options := qrev.Options{}
	err = options.BeforeApply()
	require.NoError(err)

	assert.True(options.Color)
}

func Test_Options_BeforerApply_IsNotTTY(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	options := qrev.Options{}
	err := options.BeforeApply()
	require.NoError(err)
	assert.False(options.Color)
}

func Test_Options_AfterApply(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tests := []struct {
		dsn    string
		driver driver.Driver
	}{
		{dsn: "root@tcp(127.0.0.1:13306)/mysql", driver: &driver.MySQL{}},
		{dsn: "postgres://postgres@localhost:15432", driver: &driver.PostgreSQL{}},
		{dsn: "file::memory:", driver: &driver.SQLite{}},
	}

	for _, test := range tests {
		options := qrev.Options{DSN: test.dsn}
		err := options.AfterApply()
		require.NoError(err)
		assert.IsType(test.driver, options.Driver)
	}
}

func Test_Options_AfterApply_InvalidDSN(t *testing.T) {
	assert := assert.New(t)
	options := qrev.Options{DSN: "**invalid**"}
	err := options.AfterApply()
	assert.ErrorContains(err, "fail to detect DB driver from DSN: **invalid**")
}
