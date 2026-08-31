package driver_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev/driver"
)

func testLockDuration(d time.Duration) *time.Duration {
	return &d
}

func TestSQLiteLock(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	dri := &driver.SQLite{DSN: "file:" + filepath.Join(t.TempDir(), "test.db")}
	db, err := dri.Open()
	require.NoError(err)
	defer db.Close()
	_, err = db.Exec("CREATE TABLE t (id INTEGER)")
	require.NoError(err)

	lock, err := dri.Lock(context.Background(), &driver.LockOptions{})
	require.NoError(err)

	_, err = dri.Lock(context.Background(), &driver.LockOptions{})
	assert.ErrorIs(err, driver.ErrLocked)

	// The lock does not get in the way of the queries the run then executes.
	_, err = db.Exec("INSERT INTO t VALUES (1)")
	assert.NoError(err)

	require.NoError(lock.Close())

	lock, err = dri.Lock(context.Background(), &driver.LockOptions{})
	assert.NoError(err)
	lock.Close()
}

func TestSQLiteLock_WaitTimeout(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	dri := &driver.SQLite{DSN: "file:" + filepath.Join(t.TempDir(), "test.db")}
	db, err := dri.Open()
	require.NoError(err)
	defer db.Close()
	require.NoError(db.Ping())

	lock, err := dri.Lock(context.Background(), &driver.LockOptions{})
	require.NoError(err)
	defer lock.Close()

	var buf bytes.Buffer
	_, err = dri.Lock(context.Background(), &driver.LockOptions{
		Wait:   testLockDuration(300 * time.Millisecond),
		Output: &buf,
	})
	assert.ErrorIs(err, driver.ErrLocked)
	assert.ErrorContains(err, "gave up after 300ms")
	assert.Equal("Waiting for another qrev apply to finish\n", buf.String())
}

func TestSQLiteLock_WaitForHolder(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	dri := &driver.SQLite{DSN: "file:" + filepath.Join(t.TempDir(), "test.db")}
	db, err := dri.Open()
	require.NoError(err)
	defer db.Close()
	require.NoError(db.Ping())

	held, err := dri.Lock(context.Background(), &driver.LockOptions{})
	require.NoError(err)

	go func() {
		time.Sleep(200 * time.Millisecond)
		held.Close()
	}()

	lock, err := dri.Lock(context.Background(), &driver.LockOptions{
		Wait: testLockDuration(10 * time.Second),
	})
	require.NoError(err)
	assert.NoError(lock.Close())
}

func TestSQLiteLock_NotAFile(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	dri := &driver.SQLite{DSN: "file::memory:"}
	db, err := dri.Open()
	require.NoError(err)
	defer db.Close()

	_, err = dri.Lock(context.Background(), &driver.LockOptions{})
	assert.ErrorContains(err, "cannot lock a database that is not a file: file::memory:")
}

func TestAcc_MySQLLock(t *testing.T) {
	if !testAcc {
		t.Skip()
	}

	testAccLock(t, &driver.MySQL{DSN: testDSN_MySQL})
}

func TestAcc_PostgreSQLLock(t *testing.T) {
	if !testAcc {
		t.Skip()
	}

	testAccLock(t, &driver.PostgreSQL{DSN: testDSN_PostgreSQL})
}

func testAccLock(t *testing.T, dri driver.Driver) {
	t.Helper()
	assert := assert.New(t)
	require := require.New(t)

	db, err := dri.Open()
	require.NoError(err)
	defer db.Close()
	require.NoError(db.Ping())

	lock, err := dri.Lock(context.Background(), &driver.LockOptions{})
	require.NoError(err)

	// A second session on the same database is excluded...
	_, err = dri.Lock(context.Background(), &driver.LockOptions{})
	assert.ErrorIs(err, driver.ErrLocked)

	// ...and waiting for it gives up when the holder does not finish.
	var buf bytes.Buffer
	start := time.Now()
	_, err = dri.Lock(context.Background(), &driver.LockOptions{
		Wait:   testLockDuration(1 * time.Second),
		Output: &buf,
	})
	assert.ErrorIs(err, driver.ErrLocked)
	assert.ErrorContains(err, "gave up after 1s")
	assert.Equal("Waiting for another qrev apply to finish\n", buf.String())
	assert.GreaterOrEqual(time.Since(start), 1*time.Second)

	// The lock is session level: it holds no table lock and does not block a run
	// without the flag.
	var one int
	require.NoError(db.QueryRow("SELECT 1").Scan(&one))
	assert.Equal(1, one)

	// Closing the lock releases it.
	require.NoError(lock.Close())

	lock, err = dri.Lock(context.Background(), &driver.LockOptions{})
	assert.NoError(err)
	lock.Close()

	// A waiting run takes the lock as soon as the holder releases it.
	held, err := dri.Lock(context.Background(), &driver.LockOptions{})
	require.NoError(err)

	go func() {
		time.Sleep(300 * time.Millisecond)
		held.Close()
	}()

	lock, err = dri.Lock(context.Background(), &driver.LockOptions{
		Wait: testLockDuration(30 * time.Second),
	})
	require.NoError(err)
	assert.NoError(lock.Close())
}

func TestAcc_MySQLLock_SubSecondWait(t *testing.T) {
	if !testAcc {
		t.Skip()
	}

	assert := assert.New(t)
	require := require.New(t)

	dri := &driver.MySQL{DSN: testDSN_MySQL}
	held, err := dri.Lock(context.Background(), &driver.LockOptions{})
	require.NoError(err)
	defer held.Close()

	// GET_LOCK takes whole seconds, so a sub-second wait is rounded up rather
	// than turning into no wait at all.
	start := time.Now()
	_, err = dri.Lock(context.Background(), &driver.LockOptions{
		Wait: testLockDuration(500 * time.Millisecond),
	})
	assert.ErrorIs(err, driver.ErrLocked)
	assert.ErrorContains(err, "gave up after 1s")
	assert.GreaterOrEqual(time.Since(start), 1*time.Second)
}

func TestMySQLLock_ConnErr(t *testing.T) {
	assert := assert.New(t)

	// Port 1 is closed, so Open succeeds and taking the lock session fails.
	dri := &driver.MySQL{DSN: "root@tcp(127.0.0.1:1)/qrev"}
	_, err := dri.Lock(context.Background(), &driver.LockOptions{})

	assert.ErrorContains(err, "connection refused")
}

func TestPostgreSQLLock_ConnErr(t *testing.T) {
	assert := assert.New(t)

	dri := &driver.PostgreSQL{DSN: "postgres://postgres@127.0.0.1:1/qrev"}
	_, err := dri.Lock(context.Background(), &driver.LockOptions{})

	assert.ErrorContains(err, "connection refused")
}

func TestMySQLLock_OpenErr(t *testing.T) {
	assert := assert.New(t)

	dri := &driver.MySQL{DSN: "not a dsn"}
	_, err := dri.Lock(context.Background(), &driver.LockOptions{})

	assert.ErrorContains(err, "invalid DSN")
}

func TestSQLiteLock_OpenErr(t *testing.T) {
	assert := assert.New(t)

	dri := &driver.SQLite{DSN: "file:" + filepath.Join(t.TempDir(), "not-exist", "test.db")}
	_, err := dri.Lock(context.Background(), &driver.LockOptions{})

	assert.ErrorIs(err, os.ErrNotExist)
}

func TestSQLiteLock_BadDSN(t *testing.T) {
	assert := assert.New(t)

	dri := &driver.SQLite{DSN: "file:test\x7f.db"}
	_, err := dri.Lock(context.Background(), &driver.LockOptions{})

	assert.ErrorContains(err, "invalid control character in URL")
}

func TestSQLiteLock_WaitCanceled(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	dri := &driver.SQLite{DSN: "file:" + filepath.Join(t.TempDir(), "test.db")}
	held, err := dri.Lock(context.Background(), &driver.LockOptions{})
	require.NoError(err)
	defer held.Close()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	// Zero waits without limit, so only the caller's context ends the wait.
	_, err = dri.Lock(ctx, &driver.LockOptions{Wait: testLockDuration(0)})

	assert.ErrorIs(err, context.Canceled)
}

func TestSQLiteLock_BadEscape(t *testing.T) {
	assert := assert.New(t)

	dri := &driver.SQLite{DSN: "file:%zz"}
	_, err := dri.Lock(context.Background(), &driver.LockOptions{})

	assert.ErrorContains(err, `invalid URL escape "%zz"`)
}
