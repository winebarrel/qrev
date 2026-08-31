package rds_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev/rds"
)

func TestBuildIAMAuthToken(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_DEFAULT_REGION", "us-east-1")

	token, err := rds.BuildIAMAuthToken(t.Context(), "rds.example.com:12345", "root")

	require.NoError(err)
	assert.Contains(token, "rds.example.com:12345?Action=connect&DBUser=root&X-Amz-Algorithm=AWS4-HMAC-SHA256")
}

func TestResolveCNAME(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	host, err := rds.ResolveCNAME("test.winebarrel.jp")
	require.NoError(err)
	assert.Equal("example.com", host)
}

func TestResolveCNAME_RDSEndpoint(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// An RDS endpoint is already the name the token has to be signed for, so it
	// is returned without a lookup.
	host, err := rds.ResolveCNAME("db.abc123.us-east-1.rds.amazonaws.com")

	require.NoError(err)
	assert.Equal("db.abc123.us-east-1.rds.amazonaws.com", host)
}

func TestResolveCNAME_Err(t *testing.T) {
	assert := assert.New(t)

	_, err := rds.ResolveCNAME("")

	assert.ErrorContains(err, "no such host")
}

func TestBuildIAMAuthToken_Err(t *testing.T) {
	assert := assert.New(t)

	t.Setenv("AWS_CONFIG_FILE", filepath.Join(t.TempDir(), "not-exist"))
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(t.TempDir(), "not-exist"))
	t.Setenv("AWS_PROFILE", "not-exist")

	_, err := rds.BuildIAMAuthToken(t.Context(), "rds.example.com:12345", "root")

	assert.ErrorContains(err, "failed to get shared config profile, not-exist")
}
