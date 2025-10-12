package driver_test

import (
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev/driver"
)

func TestAcc_PostgreSQL(t *testing.T) {
	if !testAcc {
		t.Skip()
	}

	assert := assert.New(t)
	require := require.New(t)

	dri := &driver.PostgreSQL{DSN: testDSN_PostgreSQL}
	db, err := dri.Open()
	require.NoError(err)
	err = db.Ping()
	assert.NoError(err)
}
