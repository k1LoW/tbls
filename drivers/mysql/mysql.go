package mysql

import (
	"database/sql"
	"github.com/k1LoW/tbls/schema"
)

// Mysql struct
type Mysql struct{}

// Analyze MySQL database schema
func (m *Mysql) Analyze(db *sql.DB, s *schema.Schema) error {
	// tables
	tableRows, err := db.Query(`
SELECT table_name, table_type FROM information_schema.tables WHERE table_schema = ?;`, s.Name)
	defer tableRows.Close()
	if err != nil {
		return err
	}
	tables := []*schema.Table{}
	for tableRows.Next() {
		var (
			tableName string
			tableType string
		)
		err := tableRows.Scan(&tableName, &tableType)
		if err != nil {
			return err
		}
		table := &schema.Table{
			Name: tableName,
			Type: tableType,
		}

		// columns and comments
		columnRows, err := db.Query(`
SELECT column_name, column_default, is_nullable, column_type, column_comment
FROM information_schema.columns
WHERE table_schema=? AND table_name=?`, s.Name, tableName)
		defer columnRows.Close()
		if err != nil {
			return err
		}
		columns := []*schema.Column{}
		for columnRows.Next() {
			var (
				columnName    string
				columnDefault sql.NullString
				isNullable    string
				columnType    string
				columnComment sql.NullString
			)
			err = columnRows.Scan(&columnName, &columnDefault, &isNullable, &columnType, &columnComment)
			if err != nil {
				return err
			}
			column := &schema.Column{
				Name:     columnName,
				Type:     columnType,
				Nullable: convertColumnNullable(isNullable),
				Default:  columnDefault,
				Comment:  columnComment.String,
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
	if str == "NO" {
		return false
	}
	return true
}
