package qrev_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev"
)

func TestInitCmd_OK(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	dri := testDBWithoutTable(t)

	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.InitCmd{}
	err := cmd.Run(options)

	require.NoError(err)
	assert.Equal("qrev_history table has been created\n", buf.String())
	assert.Equal([]string{}, testDumpDB(t, dri))
}

func TestInitCmd_Err(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	dri := testDBWithoutTable(t)

	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.InitCmd{}
	err := cmd.Run(options)

	require.NoError(err)
	assert.Equal("qrev_history table has been created\n", buf.String())

	err = cmd.Run(options)
	assert.ErrorContains(err, "failed to create table: SQL logic error: table qrev_history already exists (1)")
}
