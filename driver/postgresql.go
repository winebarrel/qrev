package driver

import (
	"context"
	"database/sql"
	"fmt"

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
