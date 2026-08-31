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

func TestInitCmd_OpenErr(t *testing.T) {
	assert := assert.New(t)

	var buf bytes.Buffer
	options := &qrev.Options{Driver: testBrokenDriver(), Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.InitCmd{}
	err := cmd.Run(options)

	assert.ErrorContains(err, "invalid DSN")
}

func TestInitCmd_IndexErr(t *testing.T) {
	assert := assert.New(t)

	// The index name is already taken, so the table is created but the index is
	// not.
	init := []string{
		"CREATE TABLE other (status VARCHAR(255))",
		"CREATE INDEX idx_qrev_history_status ON other (status)",
	}

	var buf bytes.Buffer
	options := &qrev.Options{Driver: testDBWithoutTable(t, init...), Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.InitCmd{}
	err := cmd.Run(options)

	assert.ErrorContains(err, "failed to create index:")
}
