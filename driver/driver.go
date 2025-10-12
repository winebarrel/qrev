package driver

import (
	"database/sql"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
)

type Driver interface {
	Open() (*sql.DB, error)
}

func New(dsn string, iamAuth bool) (Driver, error) {
	if _, err := mysql.ParseDSN(dsn); err == nil {
		return &MySQL{DSN: dsn, IAMAuth: iamAuth}, nil
	} else if _, err := pgx.ParseConfig(dsn); err == nil {
		sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
		return &PostgreSQL{DSN: dsn, IAMAuth: iamAuth}, nil
	} else if strings.HasPrefix(dsn, "file:") {
		return &SQLite{DSN: dsn}, nil
	}

	return nil, fmt.Errorf("fail to detect DB driver from DSN: %s", dsn)
}
