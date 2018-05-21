package postgres

import (
	"database/sql"
	"fmt"
	"github.com/k1LoW/tbls/schema"
	"regexp"
	"strings"
)

var reFK = regexp.MustCompile(`FOREIGN KEY \((.+)\) REFERENCES (.+)\((.+)\)`)

type Postgres struct{}

// Analize PostgreSQL database schema
func (p *Postgres) Analize(db *sql.DB, s *schema.Schema) error {

	// tables
	tableRows, err := db.Query(`
SELECT table_name, table_type
FROM information_schema.tables
WHERE table_schema != 'pg_catalog' AND table_schema != 'information_schema'
`)
	defer tableRows.Close()
	if err != nil {
		return err
	}

	relations := []*schema.Relation{}

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

		// table comment
		tableCommentRows, err := db.Query(`
SELECT pd.description as comment
FROM pg_stat_user_tables AS ps, pg_description AS pd
WHERE ps.relid=pd.objoid
AND pd.objsubid=0
AND ps.relname = $1`, tableName)
		defer tableCommentRows.Close()
		if err != nil {
			return err
		}

		for tableCommentRows.Next() {
			var tableComment string
			err = tableCommentRows.Scan(&tableComment)
			if err != nil {
				return err
			}
			table.Comment = tableComment
		}

		// indexes
		indexRows, err := db.Query(`
SELECT indexname, indexdef
FROM pg_indexes
WHERE schemaname != 'pg_catalog'
AND tablename = $1`, tableName)
		defer indexRows.Close()
		if err != nil {
			return err
		}

		indexes := []*schema.Index{}
		for indexRows.Next() {
			var (
				indexName string
				indexDef  string
			)
			err = indexRows.Scan(&indexName, &indexDef)
			if err != nil {
				return err
			}
			index := &schema.Index{
				Name: indexName,
				Def:  indexDef,
			}
			indexes = append(indexes, index)
		}
		table.Indexes = indexes

		// constraits
		constraitRows, err := db.Query(`
SELECT pc.conname AS name, pg_get_constraintdef(pc.oid) AS def, contype AS type
FROM pg_constraint AS pc
LEFT JOIN pg_stat_user_tables AS ps ON ps.relid = pc.conrelid
WHERE ps.relname = $1`, tableName)
		defer constraitRows.Close()
		if err != nil {
			return err
		}

		constraits := []*schema.Constrait{}
		for constraitRows.Next() {
			var (
				constraitName string
				constraitDef  string
				constraitType string
			)
			err = constraitRows.Scan(&constraitName, &constraitDef, &constraitType)
			if err != nil {
				return err
			}
			constrait := &schema.Constrait{
				Name: constraitName,
				Type: convertConstraitType(constraitType),
				Def:  constraitDef,
			}
			if constraitType == "f" {
				relation := &schema.Relation{
					Table: table,
					Def:   constraitDef,
				}
				relations = append(relations, relation)
			}
			constraits = append(constraits, constrait)
		}
		table.Constraits = constraits

		// columns comments
		columnCommentRows, err := db.Query(`
SELECT pa.attname AS column_name, pd.description AS comment
FROM pg_stat_all_tables AS ps ,pg_description AS pd ,pg_attribute AS pa
WHERE ps.relid=pd.objoid
AND pd.objsubid != 0
AND pd.objoid=pa.attrelid
AND pd.objsubid=pa.attnum
AND ps.relname = $1`, tableName)
		defer columnCommentRows.Close()
		if err != nil {
			return err
		}

		columnComments := make(map[string]string)
		for columnCommentRows.Next() {
			var (
				columnName    string
				columnComment string
			)
			err = columnCommentRows.Scan(&columnName, &columnComment)
			if err != nil {
				return err
			}
			columnComments[columnName] = columnComment
		}

		columnRows, err := db.Query(`
SELECT column_name, column_default, is_nullable, data_type, udt_name, character_maximum_length
FROM information_schema.columns
WHERE table_name = $1`, tableName)
		defer columnRows.Close()
		if err != nil {
			return err
		}

		columns := []*schema.Column{}
		for columnRows.Next() {
			var (
				columnName             string
				columnDefault          sql.NullString
				isNullable             string
				dataType               string
				udtName                string
				characterMaximumLength sql.NullInt64
			)
			err = columnRows.Scan(&columnName, &columnDefault, &isNullable, &dataType, &udtName, &characterMaximumLength)
			if err != nil {
				return err
			}
			column := &schema.Column{
				Name:     columnName,
				Type:     convertColmunType(dataType, udtName, characterMaximumLength),
				Nullable: convertColumnNullable(isNullable),
				Default:  columnDefault,
			}
			if comment, ok := columnComments[columnName]; ok {
				column.Comment = comment
			}
			columns = append(columns, column)
		}
		table.Columns = columns

		tables = append(tables, table)
	}

	s.Tables = tables

	// Relations
	for _, r := range relations {
		result := reFK.FindAllStringSubmatch(r.Def, -1)
		strColumns := strings.Split(result[0][1], ", ")
		strParentTable := result[0][2]
		strParentColumns := strings.Split(result[0][3], ", ")
		for _, c := range strColumns {
			column, err := r.Table.FindColumnByName(c)
			if err != nil {
				return err
			}
			r.Columns = append(r.Columns, column)
			column.ParentRelations = append(column.ParentRelations, r)
		}
		parentTable, err := s.FindTableByName(strParentTable)
		if err != nil {
			return err
		}
		r.ParentTable = parentTable
		for _, c := range strParentColumns {
			column, err := parentTable.FindColumnByName(c)
			if err != nil {
				return err
			}
			r.ParentColumns = append(r.ParentColumns, column)
			column.ChildRelations = append(column.ChildRelations, r)
		}
	}

	s.Relations = relations

	return nil
}

func convertColmunType(t string, udtName string, characterMaximumLength sql.NullInt64) string {
	switch t {
	case "USER-DEFINED":
		return udtName
	case "ARRAY":
		return "array"
	case "character varying":
		return fmt.Sprintf("varchar(%d)", characterMaximumLength.Int64)
	default:
		return t
	}
}

func convertConstraitType(t string) string {
	switch t {
	case "p":
		return "PRIMARY KEY"
	case "u":
		return "UNIQUE"
	case "f":
		return "FOREIGN KEY"
	default:
		return t
	}
}

func convertColumnNullable(str string) bool {
	if str == "NO" {
		return false
	}
	return true
}
