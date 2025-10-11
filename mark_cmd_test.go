package qrev_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev"
)

func TestMarkCmd_NoFilename(t *testing.T) {
	assert := assert.New(t)

	var buf bytes.Buffer
	options := &qrev.Options{Driver: testDB(t), Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.MarkCmd{Name: "not-exist.sql"}
	err := cmd.Run(options)

	assert.ErrorContains(err, "filename not found in SQL history")
}

func TestMarkCmd_MarkSkip(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"

	dri := testDB(t, init)
	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.MarkCmd{Name: "20251012-delete-old-data.sql", Status: qrev.StatusSkip}
	err := cmd.Run(options)

	require.NoError(err)
	assert.Equal("skip 20251012-delete-old-data.sql\n", buf.String())

	assert.Equal([]string{
		"20251010-init-table.sql 123abc456 2025-10-10T12:23:00Z 1 done ",
		"20251011-update-data.sql c123ab567 2025-10-10T11:20:00Z 2 skip ",
		"20251012-delete-old-data.sql bc123a678 2025-10-10T12:25:00Z 3 skip error:\ntest.go:10\n",
	}, testDumpDB(t, dri))
}

func TestMarkCmd_MarkSkip_Err(t *testing.T) {
	assert := assert.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"

	dri := testDB(t, init)
	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.MarkCmd{Name: "20251011-update-data.sql", Status: qrev.StatusSkip}
	err := cmd.Run(options)

	assert.ErrorContains(err, "skip cannot be changed to skip")
}

func TestMarkCmd_MarkSkip_Noop(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"

	dri := testDB(t, init)
	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.MarkCmd{Name: "20251012-delete-old-data.sql", Status: qrev.StatusSkip, Noop: true}
	err := cmd.Run(options)

	require.NoError(err)
	assert.Equal("skip 20251012-delete-old-data.sql (noop)\n", buf.String())

	assert.Equal([]string{
		"20251010-init-table.sql 123abc456 2025-10-10T12:23:00Z 1 done ",
		"20251011-update-data.sql c123ab567 2025-10-10T11:20:00Z 2 skip ",
		"20251012-delete-old-data.sql bc123a678 2025-10-10T12:25:00Z 3 fail error:\ntest.go:10\n",
	}, testDumpDB(t, dri))
}

func TestMarkCmd_MarkFail_Err(t *testing.T) {
	assert := assert.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"

	dri := testDB(t, init)
	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.MarkCmd{Name: "20251012-delete-old-data.sql", Status: qrev.StatusFail}
	err := cmd.Run(options)

	assert.ErrorContains(err, "fail cannot be changed to fail")
}

func TestMarkCmd_MarkFail(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"

	dri := testDB(t, init)
	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.MarkCmd{Name: "20251011-update-data.sql", Status: qrev.StatusFail}
	err := cmd.Run(options)

	require.NoError(err)
	assert.Equal("fail 20251011-update-data.sql\n", buf.String())

	assert.Equal([]string{
		"20251010-init-table.sql 123abc456 2025-10-10T12:23:00Z 1 done ",
		"20251011-update-data.sql c123ab567 2025-10-10T11:20:00Z 2 fail ",
		"20251012-delete-old-data.sql bc123a678 2025-10-10T12:25:00Z 3 fail error:\ntest.go:10\n",
	}, testDumpDB(t, dri))
}

func TestMarkCmd_MarkFail_Noop(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"

	dri := testDB(t, init)
	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.MarkCmd{Name: "20251011-update-data.sql", Status: qrev.StatusFail, Noop: true}
	err := cmd.Run(options)

	require.NoError(err)
	assert.Equal("fail 20251011-update-data.sql (noop)\n", buf.String())

	assert.Equal([]string{
		"20251010-init-table.sql 123abc456 2025-10-10T12:23:00Z 1 done ",
		"20251011-update-data.sql c123ab567 2025-10-10T11:20:00Z 2 skip ",
		"20251012-delete-old-data.sql bc123a678 2025-10-10T12:25:00Z 3 fail error:\ntest.go:10\n",
	}, testDumpDB(t, dri))
}

func TestMarkCmd_Timeout(t *testing.T) {
	assert := assert.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"

	dri := testDB(t, init)
	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 0 * time.Minute}

	cmd := &qrev.MarkCmd{Name: "20251012-delete-old-data.sql", Status: qrev.StatusSkip}
	err := cmd.Run(options)

	assert.ErrorContains(err, "context deadline exceeded")

	assert.Equal([]string{
		"20251010-init-table.sql 123abc456 2025-10-10T12:23:00Z 1 done ",
		"20251011-update-data.sql c123ab567 2025-10-10T11:20:00Z 2 skip ",
		"20251012-delete-old-data.sql bc123a678 2025-10-10T12:25:00Z 3 fail error:\ntest.go:10\n",
	}, testDumpDB(t, dri))
}
