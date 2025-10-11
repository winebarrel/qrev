package driver_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev/driver"
)

func TestNew_OK(t *testing.T) {
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
		dri, err := driver.New(test.dsn, false)
		require.NoError(err)
		assert.IsType(test.driver, dri)
	}
}

func TestNew_Err(t *testing.T) {
	assert := assert.New(t)
	_, err := driver.New("**invalid**", false)
	assert.ErrorContains(err, "fail to detect DB driver from DSN: **invalid**")
}
