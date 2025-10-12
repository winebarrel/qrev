package qrev_test

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev"
)

func TestApplyCmd_NoTarget(t *testing.T) {
	assert := assert.New(t)

	var buf bytes.Buffer
	options := &qrev.Options{Driver: testDB(t), Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.ApplyCmd{Path: "not-exist.sql"}
	err := cmd.Run(options)

	assert.ErrorContains(err, "target file not found: not-exist.sql")
}

func TestApplyCmd_NewFile(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"
	dri := testDB(t, init)

	tempDir := t.TempDir()
	t.Chdir(tempDir)
	os.WriteFile("20251010-init-table.sql", []byte("select 1"), 0400)
	os.WriteFile("20251011-update-data.sql", []byte("select 2"), 0400)
	os.WriteFile("20251012-delete-old-data.sql", []byte("select 3"), 0400)
	os.WriteFile("20251013-new.sql", []byte("select 4"), 0400)

	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.ApplyCmd{Path: "*.sql"}
	err := cmd.Run(options)

	require.NoError(err)
	assert.Equal(`done 20251013-new.sql select 4
`, buf.String())

	assert.Equal([]string{
		"20251010-init-table.sql 123abc456 done ",
		"20251011-update-data.sql c123ab567 skip ",
		"20251012-delete-old-data.sql bc123a678 fail error:\ntest.go:10\n",
		"20251013-new.sql dd17f0177066d50dbbdd656f896641550a240d6e1d162d6ca8410ee47aa57344 done ",
	}, testDumpDBWithoutTime(t, dri))
}

func TestApplyCmd_WithEditFile(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"
	dri := testDB(t, init)

	tempDir := t.TempDir()
	t.Chdir(tempDir)
	os.WriteFile("20251010-init-table.sql", []byte("select 1"), 0400)
	os.WriteFile("20251011-update-data.sql", []byte("select 2"), 0400)
	os.WriteFile("20251012-delete-old-data.sql", []byte("select 3"), 0400)
	os.WriteFile("20251013-new.sql", []byte("select 4"), 0400)

	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.ApplyCmd{Path: "*.sql", IfModified: true}
	err := cmd.Run(options)

	require.NoError(err)
	assert.Equal(`done 20251012-delete-old-data.sql select 3
done 20251013-new.sql select 4
`, buf.String())

	assert.Equal([]string{
		"20251010-init-table.sql 123abc456 done ",
		"20251011-update-data.sql c123ab567 skip ",
		"20251012-delete-old-data.sql 3efaf2f2e7527fc540b26c1517859b5446dff36e946f40b56f25f6941b170cc2 done ",
		"20251013-new.sql dd17f0177066d50dbbdd656f896641550a240d6e1d162d6ca8410ee47aa57344 done ",
	}, testDumpDBWithoutTime(t, dri))
}

func TestApplyCmd_WithoutEditFile(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', '3efaf2f2e7527fc540b26c1517859b5446dff36e946f40b56f25f6941b170cc2', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"
	dri := testDB(t, init)

	tempDir := t.TempDir()
	t.Chdir(tempDir)
	os.WriteFile("20251010-init-table.sql", []byte("select 1"), 0400)
	os.WriteFile("20251011-update-data.sql", []byte("select 2"), 0400)
	os.WriteFile("20251012-delete-old-data.sql", []byte("select 3"), 0400)
	os.WriteFile("20251013-new.sql", []byte("select 4"), 0400)

	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.ApplyCmd{Path: "*.sql", IfModified: true}
	err := cmd.Run(options)

	require.NoError(err)
	assert.Equal(`done 20251013-new.sql select 4
`, buf.String())

	assert.Equal([]string{
		"20251010-init-table.sql 123abc456 done ",
		"20251011-update-data.sql c123ab567 skip ",
		"20251012-delete-old-data.sql 3efaf2f2e7527fc540b26c1517859b5446dff36e946f40b56f25f6941b170cc2 fail error:\ntest.go:10\n",
		"20251013-new.sql dd17f0177066d50dbbdd656f896641550a240d6e1d162d6ca8410ee47aa57344 done ",
	}, testDumpDBWithoutTime(t, dri))
}

func TestApplyCmd_ForceRerun(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', '3efaf2f2e7527fc540b26c1517859b5446dff36e946f40b56f25f6941b170cc2', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"
	dri := testDB(t, init)

	tempDir := t.TempDir()
	t.Chdir(tempDir)
	os.WriteFile("20251010-init-table.sql", []byte("select 1"), 0400)
	os.WriteFile("20251011-update-data.sql", []byte("select 2"), 0400)
	os.WriteFile("20251012-delete-old-data.sql", []byte("select 3"), 0400)
	os.WriteFile("20251013-new.sql", []byte("select 4"), 0400)

	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.ApplyCmd{Path: "*.sql", ForceRerun: true}
	err := cmd.Run(options)

	require.NoError(err)
	assert.Equal(`done 20251012-delete-old-data.sql select 3
done 20251013-new.sql select 4
`, buf.String())

	assert.Equal([]string{
		"20251010-init-table.sql 123abc456 done ",
		"20251011-update-data.sql c123ab567 skip ",
		"20251012-delete-old-data.sql 3efaf2f2e7527fc540b26c1517859b5446dff36e946f40b56f25f6941b170cc2 done ",
		"20251013-new.sql dd17f0177066d50dbbdd656f896641550a240d6e1d162d6ca8410ee47aa57344 done ",
	}, testDumpDBWithoutTime(t, dri))
}

func TestApplyCmd_NoChanges(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"
	dri := testDB(t, init)

	tempDir := t.TempDir()
	t.Chdir(tempDir)
	os.WriteFile("20251010-init-table.sql", []byte("select 1"), 0400)
	os.WriteFile("20251011-update-data.sql", []byte("select 2"), 0400)
	os.WriteFile("20251012-delete-old-data.sql", []byte("select 3"), 0400)

	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.ApplyCmd{Path: "*.sql"}
	err := cmd.Run(options)

	require.NoError(err)
	assert.Equal(`No SQL file to run
`, buf.String())

	assert.Equal([]string{
		"20251010-init-table.sql 123abc456 2025-10-10T12:23:00Z 1 done ",
		"20251011-update-data.sql c123ab567 2025-10-10T11:20:00Z 2 skip ",
		"20251012-delete-old-data.sql bc123a678 2025-10-10T12:25:00Z 3 fail error:\ntest.go:10\n",
	}, testDumpDB(t, dri))
}

func TestApplyCmd_Fail(t *testing.T) {
	assert := assert.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"
	dri := testDB(t, init)

	tempDir := t.TempDir()
	t.Chdir(tempDir)
	os.WriteFile("20251010-init-table.sql", []byte("select 1"), 0400)
	os.WriteFile("20251011-update-data.sql", []byte("select 2"), 0400)
	os.WriteFile("20251012-delete-old-data.sql", []byte("select 3"), 0400)
	os.WriteFile("20251013-new.sql", []byte("**invalid**"), 0400)

	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.ApplyCmd{Path: "*.sql"}
	err := cmd.Run(options)

	assert.ErrorContains(err, "SQL fails")
	assert.Equal(`fail 20251013-new.sql **invalid**
│ SQL logic error: near "*": syntax error (1)
`, buf.String())

	assert.Equal([]string{
		"20251010-init-table.sql 123abc456 done ",
		"20251011-update-data.sql c123ab567 skip ",
		"20251012-delete-old-data.sql bc123a678 fail error:\ntest.go:10\n",
		"20251013-new.sql 331c30a26d727cf96e44fbc7a3b9dbf0ff124fc62f3aca763e2bb5d65719796b fail SQL logic error: near \"*\": syntax error (1)",
	}, testDumpDBWithoutTime(t, dri))
}

func TestApplyCmd_Fail_WithEditFile(t *testing.T) {
	assert := assert.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"
	dri := testDB(t, init)

	tempDir := t.TempDir()
	t.Chdir(tempDir)
	os.WriteFile("20251010-init-table.sql", []byte("select 1"), 0400)
	os.WriteFile("20251011-update-data.sql", []byte("select 2"), 0400)
	os.WriteFile("20251012-delete-old-data.sql", []byte("**invalid**"), 0400)
	os.WriteFile("20251013-new.sql", []byte("select 4"), 0400)

	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.ApplyCmd{Path: "*.sql", IfModified: true}
	err := cmd.Run(options)

	assert.ErrorContains(err, "SQL fails")
	assert.Equal(`fail 20251012-delete-old-data.sql **invalid**
│ SQL logic error: near "*": syntax error (1)
`, buf.String())

	assert.Equal([]string{
		"20251010-init-table.sql 123abc456 done ",
		"20251011-update-data.sql c123ab567 skip ",
		"20251012-delete-old-data.sql 331c30a26d727cf96e44fbc7a3b9dbf0ff124fc62f3aca763e2bb5d65719796b fail SQL logic error: near \"*\": syntax error (1)",
	}, testDumpDBWithoutTime(t, dri))
}

func TestApplyCmd_Fail_WithNewAndEditFile(t *testing.T) {
	assert := assert.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"
	dri := testDB(t, init)

	tempDir := t.TempDir()
	t.Chdir(tempDir)
	os.WriteFile("20251010-init-table.sql", []byte("select 1"), 0400)
	os.WriteFile("20251011-update-data.sql", []byte("select 2"), 0400)
	os.WriteFile("20251012-delete-old-data.sql", []byte("**invalid**"), 0400)
	os.WriteFile("20251009-new.sql", []byte("select 4"), 0400)

	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	cmd := &qrev.ApplyCmd{Path: "*.sql", IfModified: true}
	err := cmd.Run(options)

	assert.ErrorContains(err, "SQL fails")
	assert.Equal(`done 20251009-new.sql select 4
fail 20251012-delete-old-data.sql **invalid**
│ SQL logic error: near "*": syntax error (1)
`, buf.String())

	assert.Equal([]string{
		"20251009-new.sql dd17f0177066d50dbbdd656f896641550a240d6e1d162d6ca8410ee47aa57344 done ",
		"20251010-init-table.sql 123abc456 done ",
		"20251011-update-data.sql c123ab567 skip ",
		"20251012-delete-old-data.sql 331c30a26d727cf96e44fbc7a3b9dbf0ff124fc62f3aca763e2bb5d65719796b fail SQL logic error: near \"*\": syntax error (1)",
	}, testDumpDBWithoutTime(t, dri))
}

func TestApplyCmd_Timeout(t *testing.T) {
	assert := assert.New(t)
	//	require := require.New(t)

	init := "insert into qrev_history (filename, hash, executed_at, execution_time, status, last_error) values " +
		" ('20251010-init-table.sql',      '123abc456', '2025-10-10T12:23:00Z', 1, 'done', '')" +
		",('20251011-update-data.sql',     'c123ab567', '2025-10-10T11:20:00Z', 2, 'skip', '')" +
		",('20251012-delete-old-data.sql', 'bc123a678', '2025-10-10T12:25:00Z', 3, 'fail', 'error:' || char(10) || 'test.go:10' || char(10))"
	dri := testDB(t, init)

	tempDir := t.TempDir()
	t.Chdir(tempDir)
	os.WriteFile("20251010-init-table.sql", []byte("select 1"), 0400)
	os.WriteFile("20251011-update-data.sql", []byte("select 2"), 0400)
	os.WriteFile("20251012-delete-old-data.sql", []byte("select 3"), 0400)
	os.WriteFile("20251013-new.sql", []byte("select 4"), 0400)

	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 0}

	cmd := &qrev.ApplyCmd{Path: "*.sql"}
	err := cmd.Run(options)

	assert.ErrorContains(err, "context deadline exceeded")
	assert.Equal("", buf.String())

	assert.Equal([]string{
		"20251010-init-table.sql 123abc456 2025-10-10T12:23:00Z 1 done ",
		"20251011-update-data.sql c123ab567 2025-10-10T11:20:00Z 2 skip ",
		"20251012-delete-old-data.sql bc123a678 2025-10-10T12:25:00Z 3 fail error:\ntest.go:10\n",
	}, testDumpDB(t, dri))
}
