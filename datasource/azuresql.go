package datasource

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/k1LoW/errors"
	mssqlDriver "github.com/k1LoW/tbls/drivers/mssql"
	"github.com/k1LoW/tbls/schema"
	_ "github.com/microsoft/go-mssqldb/azuread" // registers "azuresql" driver
)

// prepareAzureSQLURL swaps the scheme to sqlserver:// (the only scheme
// go-mssqldb/azuread's msdsn.Parse URL-parses) and sets default query params.
// Returns the rewritten URL string and the database name.
func prepareAzureSQLURL(urlstr string) (string, string, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return "", "", err
	}
	q := u.Query()

	dbName := q.Get("database")
	if dbName == "" {
		return "", "", fmt.Errorf("no database name in azuresql connection string")
	}
	if q.Get("fedauth") == "" {
		q.Set("fedauth", "ActiveDirectoryServicePrincipal")
	}
	q.Set("encrypt", "true")
	u.RawQuery = q.Encode()
	u.Scheme = "sqlserver"

	return u.String(), dbName, nil
}

func AnalyzeAzureSQL(urlstr string) (_ *schema.Schema, err error) {
	defer func() { err = errors.WithStack(err) }()

	connURL, dbName, err := prepareAzureSQLURL(urlstr)
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
