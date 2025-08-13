package datasource

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/k1LoW/tbls/drivers/databricks"
	"github.com/k1LoW/tbls/schema"
)

// AnalyzeDatabricks analyze `databricks://`
func AnalyzeDatabricks(urlstr string) (_ *schema.Schema, err error) {
	s := &schema.Schema{}

	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}

	catalog := u.Query().Get("catalog")
	if catalog == "" {
		return nil, errors.New("no catalog name in the connection string")
	}
	schemaName := u.Query().Get("schema")
	if schemaName == "" {
		return nil, errors.New("no schema name in the connection string")
	}

	s.Name = fmt.Sprintf("%s.%s", catalog, schemaName)

	dsnStr := strings.TrimPrefix(urlstr, "databricks://")

	db, err := sql.Open("databricks", dsnStr)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	driver := databricks.New(db)
	if err := driver.Analyze(s); err != nil {
		return nil, err
	}

	return s, nil
}
