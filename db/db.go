package db

import (
	"fmt"
	"github.com/k1LoW/tbls/drivers/postgres"
	"github.com/k1LoW/tbls/schema"
	"github.com/xo/dburl"
	"strings"
)

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
	switch u.Driver {
	case "postgres":
		err := postgres.Analize(db, s)
		if err != nil {
			return s, err
		}
	default:
		return s, fmt.Errorf("Error: %s", "unsupported driver")
	}
	return s, nil
}
