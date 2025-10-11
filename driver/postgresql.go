package driver

import (
	"database/sql"
)

type PostgreSQL struct {
	DSN string
}

func (dri *PostgreSQL) Open() (*sql.DB, error) {
	return nil, nil // TODO:
}
