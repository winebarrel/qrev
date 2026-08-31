package qrev_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev"
)

func TestUnsignedDuration_UnmarshalText(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var d qrev.UnsignedDuration

	require.NoError(d.UnmarshalText([]byte("1m30s")))
	assert.Equal(90*time.Second, time.Duration(d))

	require.NoError(d.UnmarshalText([]byte("0")))
	assert.Equal(time.Duration(0), time.Duration(d))

	assert.ErrorContains(d.UnmarshalText([]byte("-1s")), "must not be negative: -1s")
	assert.ErrorContains(d.UnmarshalText([]byte("soon")), "invalid duration")
}
