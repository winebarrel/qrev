package driver

import (
	"database/sql"
)

type MySQL struct {
	DSN string
}

func (dri *MySQL) Open() (*sql.DB, error) {
	return nil, nil // TODO:
}
