// Package dburlscheme registers the dburl schemes tbls opens, so that every
// package connecting through dburl gets the same drivers and DSNs.
package dburlscheme

import (
	"github.com/xo/dburl"
)

func init() {
	// dburl opens postgres:// and redshift:// with pgx, but tbls registers lib/pq
	// (main.go), drivers/postgres reads arrays through pq.Array, and dsn.tls maps
	// onto lib/pq's ssl parameters, so both schemes keep opening lib/pq, which
	// registers itself as "postgres".
	pg := dburl.Unregister("postgres")
	pg.Generator = dburl.GenPq
	dburl.Register(*pg)

	rs := dburl.Unregister("redshift")
	genRedshift := rs.Generator
	rs.Generator = func(u *dburl.URL) (string, string, error) {
		dsn, _, err := genRedshift(u)
		return dsn, "postgres", err
	}
	dburl.Register(*rs)

	// dburl adds loc=Local, parseTime=true and sql_mode=ansi to every MySQL DSN
	// unless the URL sets them. sql_mode=ansi makes the server return definitions
	// with double-quoted identifiers, so drop the defaults and connect with what
	// the user wrote, as tbls always has.
	my := dburl.Unregister("mysql")
	my.Defaults = nil
	dburl.Register(*my)
}
