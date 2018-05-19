package postgres

import (
	"database/sql"
	"fmt"
	"github.com/k1LoW/tbls/schema"
)

func Analize(db *sql.DB, s *schema.Schema) error {
	tableRows, err := db.Query(`
SELECT table_name, table_type
FROM information_schema.tables
WHERE table_schema != 'pg_catalog' AND table_schema != 'information_schema'
`)
	if err != nil {
		return err
	}
	defer tableRows.Close()

	tables := []*schema.Table{}
	for tableRows.Next() {
		var tableName string
		var tableType string
		err := tableRows.Scan(&tableName, &tableType)
		if err != nil {
			return err
		}
		table := &schema.Table{
			Name: tableName,
			Type: tableType,
		}

		var columnName string
		var columnDefault sql.NullString
		var isNullable string
		var dataType string
		var udtName string
		var characterMaximumLength sql.NullInt64

		columnRows, err := db.Query(`
SELECT column_name, column_default, is_nullable, data_type, udt_name, character_maximum_length
FROM information_schema.columns
WHERE table_name = $1`, tableName)
		if err != nil {
			return err
		}
		defer columnRows.Close()

		columns := []*schema.Column{}
		for columnRows.Next() {
			err = columnRows.Scan(&columnName, &columnDefault, &isNullable, &dataType, &udtName, &characterMaximumLength)
			if err != nil {
				return err
			}
			column := &schema.Column{
				Name:    columnName,
				Type:    colmunDateType(dataType, udtName, characterMaximumLength),
				NotNull: columnNotNull(isNullable),
				Default: columnDefault,
			}
			columns = append(columns, column)
		}
		table.Columns = columns

		tables = append(tables, table)
	}

	s.Tables = tables
	return nil
}

// colmunDateType ...
func colmunDateType(dataType string, udtName string, characterMaximumLength sql.NullInt64) string {
	switch dataType {
	case "USER-DEFINED":
		return udtName
	case "character varying":
		return fmt.Sprintf("varchar(%d)", characterMaximumLength.Int64)
	default:
		return dataType
	}
}

func columnNotNull(str string) bool {
	if str == "NO" {
		return true
	}
	return false
}
