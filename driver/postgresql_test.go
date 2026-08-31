package driver_test

import (
	"path/filepath"
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

func TestPostgreSQL_Open_Err(t *testing.T) {
	assert := assert.New(t)

	dri := &driver.PostgreSQL{DSN: "postgres://%zz"}
	_, err := dri.Open()

	assert.ErrorContains(err, "cannot parse")
}

func TestPostgreSQL_Open_IAMAuth(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	dri := &driver.PostgreSQL{DSN: "postgres://root@db.abc123.us-east-1.rds.amazonaws.com:5432/qrev", IAMAuth: true}
	db, err := dri.Open()

	require.NoError(err)
	assert.NoError(db.Close())
}

func TestPostgreSQL_Open_IAMAuth_Err(t *testing.T) {
	assert := assert.New(t)

	// "a..b" has an empty label, which the resolver rejects without a query.
	dri := &driver.PostgreSQL{DSN: "postgres://root@a..b:5432/qrev", IAMAuth: true}
	_, err := dri.Open()

	assert.ErrorContains(err, "no such host")
}

func TestPostgreSQL_Open_IAMAuth_Token(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_REGION", "us-east-1")

	dri := &driver.PostgreSQL{DSN: "postgres://root@" + testIAMAuthDSNHost + ":5432/qrev", IAMAuth: true}
	db, err := dri.Open()
	require.NoError(err)
	defer db.Close()

	// The token is built before the connection is dialed, so the dial failing
	// still means it was built.
	assert.ErrorContains(db.Ping(), "no such host")
}

func TestPostgreSQL_Open_IAMAuth_TokenErr(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	t.Setenv("AWS_CONFIG_FILE", filepath.Join(t.TempDir(), "not-exist"))
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(t.TempDir(), "not-exist"))
	t.Setenv("AWS_PROFILE", "not-exist")

	dri := &driver.PostgreSQL{DSN: "postgres://root@" + testIAMAuthDSNHost + ":5432/qrev", IAMAuth: true}
	db, err := dri.Open()
	require.NoError(err)
	defer db.Close()

	assert.ErrorContains(db.Ping(), "failed to get shared config profile, not-exist")
}
