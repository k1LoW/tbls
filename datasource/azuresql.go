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

func parseAzureSQLDatabaseName(urlstr string) (string, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return "", err
	}
	dbName := u.Query().Get("database")
	if dbName == "" {
		return "", fmt.Errorf("no database name in azuresql connection string")
	}
	return dbName, nil
}

// buildAzureSQLConnStr converts azuresql:// URL to a key=value connection string
// because go-mssqldb/azuread only URL-parses sqlserver:// scheme, not azuresql://.
func buildAzureSQLConnStr(urlstr string) (string, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return "", err
	}
	q := u.Query()
	server := u.Host
	if server == "" {
		return "", fmt.Errorf("no server host in azuresql connection string")
	}
	database := q.Get("database")
	if database == "" {
		return "", fmt.Errorf("no database name in azuresql connection string")
	}
	fedauth := q.Get("fedauth")
	if fedauth == "" {
		fedauth = "ActiveDirectoryServicePrincipal"
	}
	userID := q.Get("user id")   // url.Values.Get decodes "user+id" key as "user id"
	password := q.Get("password")

	return fmt.Sprintf(
		"Server=%s;Database=%s;fedauth=%s;User ID=%s;Password=%s;Encrypt=true;TrustServerCertificate=false",
		server, database, fedauth, userID, password,
	), nil
}

func AnalyzeAzureSQL(urlstr string) (_ *schema.Schema, err error) {
	defer func() { err = errors.WithStack(err) }()

	dbName, err := parseAzureSQLDatabaseName(urlstr)
	if err != nil {
		return nil, err
	}

	connStr, err := buildAzureSQLConnStr(urlstr)
	if err != nil {
		return nil, err
	}

	s := &schema.Schema{Name: dbName}

	db, err := sql.Open("azuresql", connStr)
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
