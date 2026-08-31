package driver_test

import (
	"path/filepath"
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

func TestMySQL_Open_Err(t *testing.T) {
	assert := assert.New(t)

	dri := &driver.MySQL{DSN: "not a dsn"}
	_, err := dri.Open()

	assert.ErrorContains(err, "invalid DSN")
}

func TestMySQL_Open_IAMAuth(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	dri := &driver.MySQL{DSN: "root@tcp(db.abc123.us-east-1.rds.amazonaws.com:3306)/qrev", IAMAuth: true}
	db, err := dri.Open()

	require.NoError(err)
	assert.NoError(db.Close())
}

func TestMySQL_Open_IAMAuth_Err(t *testing.T) {
	assert := assert.New(t)

	dri := &driver.MySQL{DSN: "root@tcp(:3306)/qrev", IAMAuth: true}
	_, err := dri.Open()

	assert.ErrorContains(err, "no such host")
}

// testIAMAuthDSNHost ends in .rds.amazonaws.com so ResolveCNAME returns it
// without a lookup, and has an empty label so the dial that follows fails
// without one either.
const testIAMAuthDSNHost = "db..rds.amazonaws.com"

func TestMySQL_Open_IAMAuth_Token(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_REGION", "us-east-1")

	dri := &driver.MySQL{DSN: "root@tcp(" + testIAMAuthDSNHost + ":3306)/qrev", IAMAuth: true}
	db, err := dri.Open()
	require.NoError(err)
	defer db.Close()

	// The token is built before the connection is dialed, so the dial failing
	// still means it was built.
	assert.ErrorContains(db.Ping(), "no such host")
}

func TestMySQL_Open_IAMAuth_TokenErr(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	t.Setenv("AWS_CONFIG_FILE", filepath.Join(t.TempDir(), "not-exist"))
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(t.TempDir(), "not-exist"))
	t.Setenv("AWS_PROFILE", "not-exist")

	dri := &driver.MySQL{DSN: "root@tcp(" + testIAMAuthDSNHost + ":3306)/qrev", IAMAuth: true}
	db, err := dri.Open()
	require.NoError(err)
	defer db.Close()

	assert.ErrorContains(db.Ping(), "failed to get shared config profile, not-exist")
}
