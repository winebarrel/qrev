package rds_test

import (
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
