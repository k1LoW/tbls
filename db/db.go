package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/k1LoW/tbls/drivers/mysql"
	"github.com/k1LoW/tbls/drivers/postgres"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
	"github.com/xo/dburl"
)

// Driver is the common interface for database driver
type Driver interface {
	Analyze(*sql.DB, *schema.Schema) error
}

// Analyze database
func Analyze(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}
	u, err := dburl.Parse(urlstr)
	if err != nil {
		return s, errors.WithStack(err)
	}
	splitted := strings.Split(u.Short(), "/")
	s.Name = splitted[1]

	db, err := dburl.Open(urlstr)
	defer db.Close()
	if err != nil {
		return s, errors.WithStack(err)
	}
	if err = db.Ping(); err != nil {
		return s, errors.WithStack(err)
	}

	var driver Driver

	switch u.Driver {
	case "postgres":
		driver = new(postgres.Postgres)
	case "mysql":
		driver = new(mysql.Mysql)
	default:
		return s, fmt.Errorf("Error: %s", "unsupported driver")
	}
	err = driver.Analyze(db, s)
	if err != nil {
		return s, errors.WithStack(err)
	}
	return s, nil
}
