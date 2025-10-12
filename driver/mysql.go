package driver

import (
	"context"
	"database/sql"
	"strings"

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
