package redshift

import (
	"database/sql"

	"github.com/k1LoW/tbls/drivers/postgres"
)

type Redshift struct {
	postgres.Postgres
}

// New return new Redshift
func New(db *sql.DB) *Redshift {
	p := postgres.New(db)
	p.EnableRsMode()
	return &Redshift{*p}
}
