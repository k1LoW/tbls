package mssql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

// Mssql struct
type Mssql struct {
	db *sql.DB
}

// NewMssql ...
func NewMssql(db *sql.DB) *Mssql {
	return &Mssql{
		db: db,
	}
}

func (m *Mssql) Analyze(s *schema.Schema) error {
	// tables
	tableRows, err := m.db.Query(`
SELECT name, object_id, type FROM sys.objects WHERE type IN ('U', 'V');
`)
	defer tableRows.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	tables := []*schema.Table{}
	for tableRows.Next() {
		var (
			tableName string
			tableOid  string
			tableType string
		)
		err := tableRows.Scan(&tableName, &tableOid, &tableType)
		if err != nil {
			return errors.WithStack(err)
		}
		tableType = convertTableType(tableType)

		table := &schema.Table{
			Name: tableName,
			Type: tableType,
		}

		// view definition
		if tableType == "VIEW" {
			viewDefRows, err := m.db.Query(`
SELECT definition FROM sys.sql_modules WHERE object_id = $1
`, tableOid)
			defer viewDefRows.Close()
			if err != nil {
				return errors.WithStack(err)
			}
			for viewDefRows.Next() {
				var tableDef sql.NullString
				err := viewDefRows.Scan(&tableDef)
				if err != nil {
					return errors.WithStack(err)
				}
				table.Def = tableDef.String
			}
		}

		// columns
		columnRows, err := m.db.Query(`
SELECT c.name, t.name AS type, c.max_length, c.is_nullable, c.is_identity, object_definition(c.default_object_id) FROM sys.columns AS c
LEFT JOIN sys.types AS t ON c.system_type_id = t.system_type_id
WHERE c.object_id = $1
ORDER BY c.column_id
`, tableOid)
		defer columnRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

		columns := []*schema.Column{}
		for columnRows.Next() {
			var (
				columnName    string
				dataType      string
				maxLength     int
				isNullable    bool
				isIdentity    bool
				columnDefault sql.NullString
			)
			err = columnRows.Scan(&columnName, &dataType, &maxLength, &isNullable, &isIdentity, &columnDefault)
			if err != nil {
				return errors.WithStack(err)
			}
			column := &schema.Column{
				Name:     columnName,
				Type:     convertColmunType(dataType, maxLength),
				Nullable: isNullable,
				Default:  columnDefault,
			}
			columns = append(columns, column)
		}
		table.Columns = columns

		tables = append(tables, table)
	}

	s.Tables = tables

	return nil
}

func (m *Mssql) Info() (*schema.Driver, error) {
	var v string
	row := m.db.QueryRow(`SELECT @@VERSION`)
	row.Scan(&v)
	name := "mssql"
	d := &schema.Driver{
		Name:            name,
		DatabaseVersion: v,
	}
	return d, nil
}

func convertTableType(t string) string {
	switch strings.Trim(t, " ") {
	case "U":
		return "BASIC TABLE"
	case "V":
		return "VIEW"
	default:
		return t
	}
}

func convertColmunType(t string, maxLength int) string {
	switch t {
	case "varchar":
		return fmt.Sprintf("varchar(%d)", maxLength)
	default:
		return t
	}
}
