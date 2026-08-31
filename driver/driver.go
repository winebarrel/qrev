package driver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
)

// ErrLocked is returned by Lock when another qrev run holds the lock.
var ErrLocked = errors.New("another qrev apply is running")

type LockOptions struct {
	// Wait is nil to fail on a held lock, otherwise how long to wait for it;
	// zero waits without limit.
	Wait *time.Duration
	// Output is where the waiting notice is written. May be nil.
	Output io.Writer
}

func (options *LockOptions) notifyWaiting() {
	if options.Output != nil {
		fmt.Fprintln(options.Output, "Waiting for another qrev apply to finish")
	}
}

type Driver interface {
	Open() (*sql.DB, error)
	// Lock makes qrev runs on the same database mutually exclusive. It is
	// released by closing the returned io.Closer, or when the process dies.
	Lock(ctx context.Context, options *LockOptions) (io.Closer, error)
}

// dbLock is a lock held on a database session. It is released by ending the
// session, not by returning the connection to the pool.
type dbLock struct {
	db   *sql.DB
	conn *sql.Conn
}

func (lock *dbLock) Close() error {
	return errors.Join(lock.conn.Close(), lock.db.Close())
}

// openLockSession opens a pool of its own, so that dbLock.Close ends the
// session instead of returning the connection to the pool qrev runs queries on.
func openLockSession(ctx context.Context, dri Driver) (*dbLock, error) {
	db, err := dri.Open()

	if err != nil {
		return nil, err
	}

	conn, err := db.Conn(ctx)

	if err != nil {
		db.Close()
		return nil, err
	}

	return &dbLock{db: db, conn: conn}, nil
}

func New(dsn string, iamAuth bool) (Driver, error) {
	if _, err := mysql.ParseDSN(dsn); err == nil {
		return &MySQL{DSN: dsn, IAMAuth: iamAuth}, nil
	} else if _, err := pgx.ParseConfig(dsn); err == nil {
		sq.StatementBuilder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
		return &PostgreSQL{DSN: dsn, IAMAuth: iamAuth}, nil
	} else if strings.HasPrefix(dsn, "file:") {
		return &SQLite{DSN: dsn}, nil
	}

	return nil, fmt.Errorf("fail to detect DB driver from DSN: %s", dsn)
}
