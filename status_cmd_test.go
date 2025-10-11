package qrev_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev"
)

func TestStatusCmd_Empty(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var buf bytes.Buffer
	options := &qrev.Options{Driver: testDB(t), Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.StatusCmd{}
	err := cmd.Run(options)

	require.NoError(err)
	assert.Equal("No SQL history\n", buf.String())
}

func TestStatusCmd_OK(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"

	var buf bytes.Buffer
	options := &qrev.Options{Driver: testDB(t, init), Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.StatusCmd{}
	err := cmd.Run(options)

	require.NoError(err)
	assert.Equal(`10 Oct 21:25 fail bc123a6 20251012-delete-old-data.sql
10 Oct 21:23 done 123abc4 20251010-init-table.sql
10 Oct 20:20 skip c123ab5 20251011-update-data.sql
`, buf.String())
}

func TestStatusCmd_WithCount(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"

	var buf bytes.Buffer
	options := &qrev.Options{Driver: testDB(t, init), Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.StatusCmd{Count: 2}
	err := cmd.Run(options)

	require.NoError(err)
	assert.Equal(`10 Oct 21:25 fail bc123a6 20251012-delete-old-data.sql
10 Oct 21:23 done 123abc4 20251010-init-table.sql
`, buf.String())
}

func TestStatusCmd_ShowError(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"

	var buf bytes.Buffer
	options := &qrev.Options{Driver: testDB(t, init), Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.StatusCmd{ShowError: true}
	err := cmd.Run(options)

	require.NoError(err)
	assert.Equal(`10 Oct 21:25 fail bc123a6 20251012-delete-old-data.sql
│ error:
│ test.go:10
10 Oct 21:23 done 123abc4 20251010-init-table.sql
10 Oct 20:20 skip c123ab5 20251011-update-data.sql
`, buf.String())
}

func TestStatusCmd_WithStatus(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"

	dri := testDB(t, init)

	tests := []struct {
		status   string
		expected string
	}{
		{status: "done", expected: "10 Oct 21:23 done 123abc4 20251010-init-table.sql\n"},
		{status: "skip", expected: "10 Oct 20:20 skip c123ab5 20251011-update-data.sql\n"},
		{status: "fail", expected: "10 Oct 21:25 fail bc123a6 20251012-delete-old-data.sql\n"},
	}

	for _, test := range tests {
		var buf bytes.Buffer
		options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}
		cmd := &qrev.StatusCmd{StatusOrFilename: &test.status}
		err := cmd.Run(options)

		require.NoError(err)
		assert.Equal(test.expected, buf.String())
	}
}

func TestStatusCmd_WithFilename(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"

	dri := testDB(t, init)

	tests := []struct {
		filename string
		expected string
	}{
		{filename: "20251010-init-table.sql", expected: "10 Oct 21:23 done 123abc4 20251010-init-table.sql\n"},
		{filename: "20251011-update-data.sql", expected: "10 Oct 20:20 skip c123ab5 20251011-update-data.sql\n"},
		{filename: "20251012-delete-old-data.sql", expected: "10 Oct 21:25 fail bc123a6 20251012-delete-old-data.sql\n"},
	}

	for _, test := range tests {
		var buf bytes.Buffer
		options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}
		cmd := &qrev.StatusCmd{StatusOrFilename: &test.filename}
		err := cmd.Run(options)

		require.NoError(err)
		assert.Equal(test.expected, buf.String())
	}
}
