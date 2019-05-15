package redshift

import (
	"database/sql"

	"github.com/k1LoW/tbls/drivers/postgres"
)

type Redshift struct {
	postgres.Postgres
}

// NewRedshift return new Redshift
func NewRedshift(db *sql.DB) *Redshift {
	p := postgres.NewPostgres(db)
	p.EnableRsMode()
	return &Redshift{*p}
}
