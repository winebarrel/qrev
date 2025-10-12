package driver_test

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev/driver"
)

func TestAcc_MySQL(t *testing.T) {
	if !testAcc {
		t.Skip()
	}

	assert := assert.New(t)
	require := require.New(t)

	dri := &driver.MySQL{DSN: testDSN_MySQL}
	db, err := dri.Open()
	require.NoError(err)
	err = db.Ping()
	assert.NoError(err)
}
