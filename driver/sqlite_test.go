package driver_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev/driver"
	_ "modernc.org/sqlite"
)

func TestSQLite(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	dri := &driver.SQLite{DSN: "file::memory:"}
	db, err := dri.Open()
	require.NoError(err)
	err = db.Ping()
	assert.NoError(err)
}
