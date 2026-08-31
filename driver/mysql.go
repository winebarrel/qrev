package driver

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/winebarrel/qrev/rds"
)

type MySQL struct {
	DSN     string
	IAMAuth bool
}

func (dri *MySQL) Open() (*sql.DB, error) {
	mycfg, err := mysql.ParseDSN(dri.DSN)

	if err != nil {
		return nil, err
	}

	if dri.IAMAuth {
		hostPort := strings.SplitN(mycfg.Addr, ":", 2)
		host, err := rds.ResolveCNAME(hostPort[0])

		if err != nil {
			return nil, err
		}

		port := hostPort[1]
		endpoint := host + ":" + port
		user := mycfg.User

		bc := func(ctx context.Context, mc *mysql.Config) error {
			token, err := rds.BuildIAMAuthToken(ctx, endpoint, user)

			if err != nil {
				return err
			}

			mc.Passwd = token
			return nil
		}

		err = mycfg.Apply(mysql.BeforeConnect(bc))

		if err != nil {
			return nil, err
		}

		mycfg.AllowCleartextPasswords = true

		if mycfg.TLSConfig == "" {
			mycfg.TLSConfig = "preferred"
		}
	}

	connector, err := mysql.NewConnector(mycfg)

	if err != nil {
		return nil, err
	}

	return sql.OpenDB(connector), nil
}

// mysqlLockNamePrefix starts the lock name. Named locks are server-wide, so the
// name carries a digest of the current database; the digest also keeps the name
// within the 64 character limit.
const mysqlLockNamePrefix = "qrev:"

// Lock takes a session-level named lock: it outlives the transaction each SQL
// file runs in, and ends with the session. The wait is handed to GET_LOCK
// rather than bounded by a context deadline, which would only close the client
// socket and leave the server waiting.
func (dri *MySQL) Lock(ctx context.Context, options *LockOptions) (io.Closer, error) {
	lock, err := openLockSession(ctx, dri)

	if err != nil {
		return nil, err
	}

	name, err := mysqlLockName(ctx, lock.conn)

	if err != nil {
		lock.Close()
		return nil, err
	}

	acquired, err := mysqlGetLock(ctx, lock.conn, name, 0)

	if err != nil {
		lock.Close()
		return nil, err
	}

	if acquired {
		return lock, nil
	}

	if options.Wait == nil {
		lock.Close()
		return nil, ErrLocked
	}

	options.notifyWaiting()

	// GET_LOCK waits without limit on a negative timeout.
	timeout := -1 * time.Second
	wait := *options.Wait

	if wait > 0 {
		// GET_LOCK truncates the timeout to whole seconds, so round up.
		timeout = wait.Truncate(time.Second)

		if timeout < wait {
			timeout += time.Second
		}

		wait = timeout
	}

	acquired, err = mysqlGetLock(ctx, lock.conn, name, timeout)

	if err != nil {
		lock.Close()
		return nil, err
	}

	if !acquired {
		lock.Close()
		return nil, fmt.Errorf("%w: gave up after %s", ErrLocked, wait)
	}

	return lock, nil
}

// mysqlLockName digests the database name in Go: MD5() no longer exists in
// recent MySQL releases.
func mysqlLockName(ctx context.Context, conn *sql.Conn) (string, error) {
	var database sql.NullString
	err := conn.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&database)

	if err != nil {
		return "", err
	}

	digest := sha256.Sum256([]byte(database.String))

	return mysqlLockNamePrefix + hex.EncodeToString(digest[:16]), nil
}

func mysqlGetLock(ctx context.Context, conn *sql.Conn, name string, timeout time.Duration) (bool, error) {
	// GET_LOCK returns NULL when the lock cannot be waited for, e.g. on a
	// deadlock.
	var acquired sql.NullInt64
	err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", name, int(timeout.Seconds())).Scan(&acquired)

	if err != nil {
		return false, err
	}

	if !acquired.Valid {
		return false, errors.New("failed to acquire the apply lock")
	}

	return acquired.Int64 == 1, nil
}
