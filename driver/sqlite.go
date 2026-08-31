package driver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"syscall"
	"time"

	_ "modernc.org/sqlite"
)

const (
	// flock(2) has no timeout, so a bounded wait polls.
	sqliteLockPollInterval = 100 * time.Millisecond
	// The database file itself cannot be locked: SQLite flocks it too, so
	// holding it would fail this run's own queries with SQLITE_BUSY.
	sqliteLockFileSuffix = ".qrev-lock"
)

type SQLite struct {
	DSN string
}

func (dri *SQLite) Open() (*sql.DB, error) {
	return sql.Open("sqlite", dri.DSN)
}

// Lock takes an exclusive flock(2) on a sidecar of the database file. SQLite
// has no session-level lock to hold for the whole run, so the lock lives in the
// filesystem instead. The kernel drops it when the file is closed.
func (dri *SQLite) Lock(ctx context.Context, options *LockOptions) (io.Closer, error) {
	path, err := sqliteLockPath(dri.DSN)

	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0600)

	if err != nil {
		return nil, err
	}

	acquired, err := sqliteFlock(file)

	if err != nil {
		file.Close()
		return nil, err
	}

	if acquired {
		return file, nil
	}

	if options.Wait == nil {
		file.Close()
		return nil, ErrLocked
	}

	options.notifyWaiting()
	waitCtx := ctx

	if *options.Wait > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, *options.Wait)
		defer cancel()
	}

	ticker := time.NewTicker(sqliteLockPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-waitCtx.Done():
			file.Close()

			if ctx.Err() == nil {
				return nil, fmt.Errorf("%w: gave up after %s", ErrLocked, *options.Wait)
			}

			return nil, waitCtx.Err()
		case <-ticker.C:
			acquired, err := sqliteFlock(file)

			if err != nil {
				file.Close()
				return nil, err
			}

			if acquired {
				return file, nil
			}
		}
	}
}

func sqliteFlock(file *os.File) (bool, error) {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)

	if err == nil {
		return true, nil
	}

	if errors.Is(err, syscall.EWOULDBLOCK) {
		return false, nil
	}

	return false, err
}

// sqliteLockPath names the lock file for a "file:" DSN, which may be an opaque
// path or a file URL and may carry query parameters.
func sqliteLockPath(dsn string) (string, error) {
	u, err := url.Parse(dsn)

	if err != nil {
		return "", err
	}

	path := u.Opaque

	if path == "" {
		path = u.Path
	} else {
		path, err = url.PathUnescape(path)

		if err != nil {
			return "", err
		}
	}

	if path == "" || path == ":memory:" {
		return "", fmt.Errorf("cannot lock a database that is not a file: %s", dsn)
	}

	return path + sqliteLockFileSuffix, nil
}
