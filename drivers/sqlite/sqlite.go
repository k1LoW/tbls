package sqlite

import (
	"database/sql"
	"fmt"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

// Sqlite struct
type Sqlite struct{}

// Analyze SQLite database schema
func (l *Sqlite) Analyze(db *sql.DB, s *schema.Schema) error {
	// tables
	tableRows, err := db.Query(`
SELECT name, type, sql
FROM sqlite_master
WHERE name != 'sqlite_sequence' AND (type = 'table' OR type = 'view');`)
	defer tableRows.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	tables := []*schema.Table{}
	for tableRows.Next() {
		var (
			tableName string
			tableType string
			tableDef  string
		)
		err := tableRows.Scan(&tableName, &tableType, &tableDef)
		if err != nil {
			return errors.WithStack(err)
		}

		table := &schema.Table{
			Name: tableName,
			Type: tableType,
			Def:  tableDef,
		}

		// indexes
		indexRows, err := db.Query(`
SELECT name, sql FROM sqlite_master WHERE type = 'index' AND tbl_name = ?;
`, tableName)
		defer indexRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

		indexes := []*schema.Index{}
		for indexRows.Next() {
			var (
				indexName string
				indexDef  sql.NullString
			)
			err = indexRows.Scan(&indexName, &indexDef)
			if err != nil {
				fmt.Printf("%s\n", tableName)

				return errors.WithStack(err)
			}
			index := &schema.Index{
				Name: indexName,
				Def:  indexDef.String,
			}
			indexes = append(indexes, index)
		}
		table.Indexes = indexes

		// constraints

		// triggers

		// columns
		columnRows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
		defer columnRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

		columns := []*schema.Column{}
		for columnRows.Next() {
			var (
				columnID      string
				columnName    string
				dataType      string
				columnNotNull string
				columnDefault sql.NullString
				columnPk      string
			)
			err = columnRows.Scan(&columnID, &columnName, &dataType, &columnNotNull, &columnDefault, &columnPk)
			if err != nil {
				return errors.WithStack(err)
			}
			column := &schema.Column{
				Name:     columnName,
				Type:     dataType,
				Nullable: convertColumnNullable(columnNotNull),
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

func convertColumnNullable(str string) bool {
	if str == "1" {
		return false
	}
	return true
}
