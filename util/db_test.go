package util_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev/util"
	_ "modernc.org/sqlite"
)

func TestWithTx_Commit(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	db, err := sql.Open("sqlite", "file::memory:")
	require.NoError(err)
	_, err = db.Exec("create table foo (id int primary key)")
	require.NoError(err)
	defer db.Close()

	err = util.WithTx(db, 10*time.Minute, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "insert into foo (id) values (100)")
		require.NoError(err)
		return nil
	})

	assert.NoError(err)
	var n int
	err = db.QueryRow("select count(*) from foo where id = 100").Scan(&n)
	require.NoError(err)
	assert.Equal(1, n)
}

func TestWithTx_Rollback(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	db, err := sql.Open("sqlite", "file::memory:")
	require.NoError(err)
	_, err = db.Exec("create table foo (id int primary key)")
	require.NoError(err)
	defer db.Close()

	err = util.WithTx(db, 10*time.Minute, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "insert into foo (id) values (100)")
		require.NoError(err)
		return errors.New("any error")
	})

	assert.ErrorContains(err, "any error")
	var n int
	err = db.QueryRow("select count(*) from foo where id = 100").Scan(&n)
	require.NoError(err)
	assert.Equal(0, n)
}

func TestWithTx_Timeout(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	db, err := sql.Open("sqlite", "file::memory:")
	require.NoError(err)
	_, err = db.Exec("create table foo (id int primary key)")
	require.NoError(err)
	defer db.Close()

	err = util.WithTx(db, 0, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "insert into foo (id) values (100)")
		require.NoError(err)
		return errors.New("any error")
	})

	assert.ErrorContains(err, "context deadline exceeded")
	var n int
	err = db.QueryRow("select count(*) from foo where id = 100").Scan(&n)
	require.NoError(err)
	assert.Equal(0, n)
}
