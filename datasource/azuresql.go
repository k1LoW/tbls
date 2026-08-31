package datasource

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/config"
	mssqlDriver "github.com/k1LoW/tbls/drivers/mssql"
	"github.com/k1LoW/tbls/schema"
	_ "github.com/microsoft/go-mssqldb/azuread" // registers "azuresql" driver
)

// prepareAzureSQLURL swaps the scheme to sqlserver:// (the only scheme
// go-mssqldb/azuread's msdsn.Parse URL-parses) and sets default query params.
// Returns the rewritten URL string and the database name.
func prepareAzureSQLURL(urlstr string, t *config.TLS) (string, string, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return "", "", err
	}
	q := u.Query()

	dbName := q.Get("database")
	if dbName == "" && u.Path != "" && u.Path != "/" {
		dbName = strings.TrimPrefix(u.Path, "/")
		q.Set("database", dbName)
	}
	u.Path = ""
	if dbName == "" {
		return "", "", fmt.Errorf("no database name in azuresql connection string")
	}

	if q.Get("fedauth") == "" {
		if (u.User != nil && u.User.Username() != "") || q.Get("user id") != "" {
			q.Set("fedauth", "ActiveDirectoryServicePrincipal")
		} else {
			q.Set("fedauth", "ActiveDirectoryDefault")
		}
	}
	// azuread shares the sqlserver TLS parameters, so dsn.tls is applied through
	// the same helper. It runs before the encrypt default below so that its
	// conflict checks see the value the user actually wrote.
	if t != nil {
		if err := validateTLSOptions(*t); err != nil {
			return "", "", err
		}
		if err := applySQLServerTLS(*t, q); err != nil {
			return "", "", err
		}
	}
	if q.Get("encrypt") == "" {
		q.Set("encrypt", "true")
	}
	u.RawQuery = q.Encode()
	u.Scheme = "sqlserver"

	return u.String(), dbName, nil
}

func AnalyzeAzureSQL(dsn config.DSN) (_ *schema.Schema, err error) {
	defer func() { err = errors.WithStack(err) }()

	connURL, dbName, err := prepareAzureSQLURL(dsn.URL, dsn.TLS)
	if err != nil {
		return nil, err
	}

	s := &schema.Schema{Name: dbName}

	db, err := sql.Open("azuresql", connURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		return nil, err
	}

	driver := mssqlDriver.New(db)
	if err := driver.Analyze(s); err != nil {
		return nil, err
	}
	return s, nil
}
