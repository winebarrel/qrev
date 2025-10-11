package util

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func WithTx(db *sql.DB, timeout time.Duration, cb func(context.Context, *sql.Tx) error) error {
	ctx, cancel := context.WithCancel(context.Background())
	ctx, cancel = context.WithTimeout(ctx, timeout)

	// Trap SIGINT
	{
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)

		go func() {
			select {
			case <-ctx.Done():
				// Nothing to do
			case <-sigint:
				// Stop query on interrupt
				cancel()
				os.Exit(130)
			}
		}()
	}

	tx, err := db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	defer tx.Rollback() //nolint:errcheck

	err = cb(ctx, tx)

	if err == nil {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}
	}

	return err
}
