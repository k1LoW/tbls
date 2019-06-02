package mssql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

var defaultSchemaName = "dbo"

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
SELECT schema_name(schema_id) AS table_schema, name, object_id, type FROM sys.objects WHERE type IN ('U', 'V');
`)
	defer tableRows.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	tables := []*schema.Table{}
	for tableRows.Next() {
		var (
			tableSchema string
			tableName   string
			tableOid    string
			tableType   string
		)
		err := tableRows.Scan(&tableSchema, &tableName, &tableOid, &tableType)
		if err != nil {
			return errors.WithStack(err)
		}
		tableType = convertTableType(tableType)

		name := tableName
		if tableSchema != defaultSchemaName {
			name = fmt.Sprintf("%s.%s", tableSchema, tableName)
		}

		table := &schema.Table{
			Name: name,
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

		// indexes
		indexRows, err := m.db.Query(`
SELECT
  i.name AS index_name,
  i.type_desc,
  i.is_unique,
  i.is_primary_key,
  i.is_unique_constraint,
  STRING_AGG(COL_NAME(ic.object_id, ic.column_id), ', ') WITHIN GROUP ( ORDER BY ic.key_ordinal )
FROM sys.indexes AS i
INNER JOIN sys.index_columns AS ic
ON i.object_id = ic.object_id AND i.index_id = ic.index_id
WHERE i.object_id = OBJECT_ID($1)
GROUP BY i.name, i.index_id, i.type_desc, i.is_unique, i.is_primary_key, i.is_unique_constraint
ORDER BY i.index_id
`, fmt.Sprintf("%s.%s", tableSchema, tableName))
		defer indexRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}
		indexes := []*schema.Index{}
		for indexRows.Next() {
			var (
				indexName               string
				indexType               string
				indexIsUnique           bool
				indexIsPrimaryKey       bool
				indexIsUniqueConstraint bool
				indexColumnName         sql.NullString
			)
			err = indexRows.Scan(&indexName, &indexType, &indexIsUnique, &indexIsPrimaryKey, &indexIsUniqueConstraint, &indexColumnName)
			if err != nil {
				return errors.WithStack(err)
			}

			indexDef := []string{
				indexType,
			}
			if indexIsUnique {
				indexDef = append(indexDef, "unique")
			}
			if indexIsPrimaryKey {
				indexDef = append(indexDef, "part of a PRIMARY KEY constraint")
			}
			if indexIsUniqueConstraint {
				indexDef = append(indexDef, "part of a UNIQUE constraint")
			}
			indexDef = append(indexDef, fmt.Sprintf("[ %s ]", indexColumnName.String))

			index := &schema.Index{
				Name:    indexName,
				Def:     strings.Join(indexDef, ", "),
				Table:   &table.Name,
				Columns: strings.Split(indexColumnName.String, ", "),
			}

			indexes = append(indexes, index)
		}
		table.Indexes = indexes

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
