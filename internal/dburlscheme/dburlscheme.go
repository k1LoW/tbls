// Package dburlscheme registers the dburl schemes tbls opens, so that every
// package connecting through dburl gets the same drivers and DSNs.
package dburlscheme

import (
	"net/url"
	"strings"

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

	// The redshift generator of v0.53.0 also joins a repeated query key with
	// commas, which lib/pq reads as one value, so keep the plain URL generator.
	rs := dburl.Unregister("redshift")
	genRedshift := dburl.GenFromURL("postgres://localhost:5439/")
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

	// dburl no longer accepts ch+http:// and ch+https://, and stopped adding
	// the native port 9000 to clickhouse://, so restore the generator tbls
	// connected with before.
	ch := dburl.Unregister("clickhouse")
	ch.Transport = dburl.TransportAny
	ch.Generator = genClickhouse
	dburl.Register(*ch)

	// dburl now writes snowflake://user:pass@<account>.snowflakecomputing.com/...,
	// which gosnowflake reads as user "snowflake" and password "//user:pass",
	// and it rejects URLs without a password, which externalbrowser and JWT
	// authentication rely on. Keep the scheme-less DSN tbls used before.
	sf := dburl.Unregister("snowflake")
	sf.Generator = genSnowflake
	dburl.Register(*sf)

	// dburl appends .database.windows.net to a host without a dot for
	// azuresql:// and for any URL with fedauth, which changes the server a
	// local or aliased host name points at, so keep the host as written.
	ms := dburl.Unregister("sqlserver")
	ms.Generator = genSqlserver
	dburl.Register(*ms)
}

func genSqlserver(u *dburl.URL) (string, string, error) {
	z := &url.URL{
		Scheme:   "sqlserver",
		Opaque:   u.Opaque,
		User:     u.User,
		Host:     u.Host,
		Path:     u.Path,
		RawQuery: u.RawQuery,
		Fragment: u.Fragment,
	}
	if z.Host == "" {
		z.Host = "localhost"
	}
	driver := "sqlserver"
	if strings.Contains(strings.ToLower(u.Scheme), "azuresql") || u.Query().Get("fedauth") != "" {
		driver = "azuresql"
	}
	v := strings.Split(strings.TrimPrefix(z.Path, "/"), "/")
	if n, q := len(v), z.Query(); !q.Has("database") && n != 0 && len(v[0]) != 0 {
		q.Set("database", v[n-1])
		z.Path, z.RawQuery = "/"+strings.Join(v[:n-1], "/"), q.Encode()
	}
	return z.String(), driver, nil
}

var (
	genClickhouseTCP   = dburl.GenFromURL("clickhouse://localhost:9000/")
	genClickhouseHTTP  = dburl.GenFromURL("http://localhost/")
	genClickhouseHTTPS = dburl.GenFromURL("https://localhost/")
)

func genClickhouse(u *dburl.URL) (string, string, error) {
	switch strings.ToLower(u.Transport) {
	case "", "tcp":
		return genClickhouseTCP(u)
	case "http":
		return genClickhouseHTTP(u)
	case "https":
		return genClickhouseHTTPS(u)
	}
	return "", "", dburl.ErrInvalidTransportProtocol
}

func genSnowflake(u *dburl.URL) (string, string, error) {
	host, port, dbname := u.Hostname(), u.Port(), strings.TrimPrefix(u.Path, "/")
	if host == "" {
		return "", "", dburl.ErrMissingHost
	}
	if port != "" {
		port = ":" + port
	}
	if u.User == nil {
		return "", "", dburl.ErrMissingUser
	}
	user := u.User.Username()
	if pass, _ := u.User.Password(); pass != "" {
		user += ":" + pass
	}
	dsn := user + "@" + host + port + "/" + dbname
	if q := u.Query().Encode(); q != "" {
		dsn += "?" + q
	}
	return dsn, "", nil
}
