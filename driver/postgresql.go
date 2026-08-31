package driver

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/winebarrel/qrev/rds"
)

type PostgreSQL struct {
	DSN     string
	IAMAuth bool
}

func (dri *PostgreSQL) Open() (*sql.DB, error) {
	opts := []stdlib.OptionOpenDB{}
	pgcfg, err := pgx.ParseConfig(dri.DSN)

	if err != nil {
		return nil, err
	}

	if dri.IAMAuth {
		host, err := rds.ResolveCNAME(pgcfg.Host)

		if err != nil {
			return nil, err
		}

		endpoint := fmt.Sprintf("%s:%d", host, pgcfg.Port)
		user := pgcfg.User

		opts = append(opts, stdlib.OptionBeforeConnect(func(ctx context.Context, cc *pgx.ConnConfig) error {
			token, err := rds.BuildIAMAuthToken(ctx, endpoint, user)

			if err != nil {
				return err
			}

			cc.Password = token
			return nil
		}))
	}

	connector := stdlib.GetConnector(*pgcfg, opts...)

	return sql.OpenDB(connector), nil
}

// pgLockClassID is the first key of the advisory lock ("qrev" in ASCII). The
// second key is hashtext(current_database()), which scopes the cluster-wide
// advisory lock namespace to one database.
const pgLockClassID = 0x71726576

// Lock takes a session-level advisory lock: it outlives the transaction each
// SQL file runs in, and ends with the session. The wait is bounded by a context
// deadline rather than lock_timeout, which would leak into that session.
func (dri *PostgreSQL) Lock(ctx context.Context, options *LockOptions) (io.Closer, error) {
	lock, err := openLockSession(ctx, dri)

	if err != nil {
		return nil, err
	}

	var acquired bool
	err = lock.conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1, hashtext(current_database()))", pgLockClassID).Scan(&acquired)

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
	waitCtx := ctx

	if *options.Wait > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, *options.Wait)
		defer cancel()
	}

	_, err = lock.conn.ExecContext(waitCtx, "SELECT pg_advisory_lock($1, hashtext(current_database()))", pgLockClassID)

	if err != nil {
		lock.Close()

		if waitCtx.Err() != nil && ctx.Err() == nil {
			return nil, fmt.Errorf("%w: gave up after %s", ErrLocked, *options.Wait)
		}

		return nil, err
	}

	return lock, nil
}
