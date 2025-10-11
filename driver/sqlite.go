package driver

import "database/sql"

type SQLite struct {
	DSN string
}

func (dri *SQLite) Open() (*sql.DB, error) {
	return sql.Open("sqlite", dri.DSN)
}
