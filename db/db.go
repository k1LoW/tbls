package db

import (
	"database/sql"
	"fmt"
	"github.com/k1LoW/tbls/drivers/postgres"
	"github.com/k1LoW/tbls/schema"
	"github.com/xo/dburl"
	"strings"
)

type Driver interface {
	Analize(*sql.DB, *schema.Schema) error
}

func Analyze(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}
	u, err := dburl.Parse(urlstr)
	if err != nil {
		return s, err
	}
	splitted := strings.Split(u.Short(), "/")
	s.Name = splitted[1]

	db, err := dburl.Open(urlstr)
	if err != nil {
		return s, err
	}
	defer db.Close()

	var driver Driver

	switch u.Driver {
	case "postgres":
		driver = new(postgres.Postgres)
	default:
		return s, fmt.Errorf("Error: %s", "unsupported driver")
	}
	err = driver.Analize(db, s)
	if err != nil {
		return s, err
	}
	return s, nil
}
