package mysql

import (
	"database/sql"
	"fmt"
	"github.com/k1LoW/tbls/schema"
	"regexp"
	"strings"
)

var reFK = regexp.MustCompile(`FOREIGN KEY \((.+)\) REFERENCES ([^\s]+)\s?\((.+)\)`)

// Mysql struct
type Mysql struct{}

// Analyze MySQL database schema
func (m *Mysql) Analyze(db *sql.DB, s *schema.Schema) error {
	// tables and comments
	tableRows, err := db.Query(`
SELECT table_name, table_type, table_comment FROM information_schema.tables WHERE table_schema = ?;`, s.Name)
	defer tableRows.Close()
	if err != nil {
		return err
	}

	relations := []*schema.Relation{}

	tables := []*schema.Table{}
	for tableRows.Next() {
		var (
			tableName    string
			tableType    string
			tableComment string
		)
		err := tableRows.Scan(&tableName, &tableType, &tableComment)
		if err != nil {
			return err
		}
		table := &schema.Table{
			Name:    tableName,
			Type:    tableType,
			Comment: tableComment,
		}

		// table definition
		if tableType == "BASE TABLE" {
			tableDefRows, err := db.Query(fmt.Sprintf("SHOW CREATE TABLE %s", tableName))
			defer tableDefRows.Close()
			if err != nil {
				return err
			}
			for tableDefRows.Next() {
				var (
					tableName string
					tableDef  string
				)
				err := tableDefRows.Scan(&tableName, &tableDef)
				if err != nil {
					return err
				}
				table.Def = tableDef
			}
		}

		// view definition
		if tableType == "VIEW" {
			viewDefRows, err := db.Query(`
SELECT view_definition FROM information_schema.views
WHERE table_schema = ?
AND table_name = ?;
		`, s.Name, tableName)
			defer viewDefRows.Close()
			if err != nil {
				return err
			}
			for viewDefRows.Next() {
				var tableDef string
				err := viewDefRows.Scan(&tableDef)
				if err != nil {
					return err
				}
				table.Def = fmt.Sprintf("CREATE VIEW %s AS (%s)", tableName, tableDef)
			}
		}

		// indexes
		indexRows, err := db.Query(`
SELECT
(CASE WHEN s.index_name='PRIMARY' AND s.non_unique=0 THEN 'PRIMARY KEY'
      WHEN s.index_name!='PRIMARY' AND s.non_unique=0 THEN 'UNIQUE KEY'
      WHEN s.non_unique=1 THEN 'KEY'
      ELSE null
  END) AS key_type,
s.index_name, GROUP_CONCAT(s.column_name ORDER BY s.seq_in_index SEPARATOR ', '), s.index_type
FROM information_schema.statistics AS s
LEFT JOIN information_schema.columns AS c ON s.column_name = c.column_name
WHERE s.table_name = c.table_name
AND s.table_schema = ?
AND s.table_name = ?
GROUP BY key_type, s.table_name, s.index_name, s.index_type`, s.Name, tableName)
		defer indexRows.Close()
		if err != nil {
			return err
		}

		indexes := []*schema.Index{}
		for indexRows.Next() {
			var (
				indexKeyType    string
				indexName       string
				indexColumnName string
				indexType       string
				indexDef        string
			)
			err = indexRows.Scan(&indexKeyType, &indexName, &indexColumnName, &indexType)
			if err != nil {
				return err
			}

			if indexKeyType == "PRIMARY KEY" {
				indexDef = fmt.Sprintf("%s (%s) USING %s", indexKeyType, indexColumnName, indexType)
			} else {
				indexDef = fmt.Sprintf("%s %s (%s) USING %s", indexKeyType, indexName, indexColumnName, indexType)
			}

			index := &schema.Index{
				Name: indexName,
				Def:  indexDef,
			}
			indexes = append(indexes, index)
		}
		table.Indexes = indexes

		// constraints
		constraintRows, err := db.Query(`
SELECT
  kcu.constraint_name,
  sub.costraint_type,
  GROUP_CONCAT(kcu.column_name ORDER BY kcu.ordinal_position, position_in_unique_constraint SEPARATOR ', ') AS column_name,
  kcu.referenced_table_name,
  GROUP_CONCAT(kcu.referenced_column_name ORDER BY kcu.ordinal_position, position_in_unique_constraint SEPARATOR ', ') AS referenced_column_name
FROM information_schema.key_column_usage AS kcu
LEFT JOIN information_schema.columns AS c ON kcu.column_name = c.column_name
LEFT JOIN
  (
   SELECT
   kcu.constraint_name,
   kcu.column_name,
    (CASE WHEN c.column_key='PRI' THEN 'PRIMARY KEY'
        WHEN c.column_key='UNI' THEN 'UNIQUE'
        WHEN c.column_key='MUL' AND kcu.referenced_table_name IS NULL THEN 'UNIQUE'
        WHEN c.column_key='MUL' AND kcu.referenced_table_name IS NOT NULL THEN 'FOREIGN KEY'
        ELSE null
   END) AS costraint_type
   FROM information_schema.key_column_usage AS kcu
   LEFT JOIN information_schema.columns AS c ON kcu.column_name = c.column_name
   WHERE
   kcu.table_name = c.table_name
   AND kcu.table_name = ?
   AND kcu.ordinal_position = 1
  ) AS sub
ON kcu.constraint_name = sub.constraint_name
WHERE kcu.table_name = c.table_name
AND kcu.constraint_schema= ?
AND kcu.table_name = ?
GROUP BY kcu.constraint_name, sub.costraint_type, kcu.referenced_table_name`, tableName, s.Name, tableName)
		defer constraintRows.Close()
		if err != nil {
			return err
		}

		constraints := []*schema.Constraint{}
		for constraintRows.Next() {
			var (
				constraintName          string
				constraintType          string
				constraintColumnName    string
				constraintRefTableName  sql.NullString
				constraintRefColumnName sql.NullString
				constraintDef           string
			)
			err = constraintRows.Scan(&constraintName, &constraintType, &constraintColumnName, &constraintRefTableName, &constraintRefColumnName)
			if err != nil {
				return err
			}
			switch constraintType {
			case "PRIMARY KEY":
				constraintDef = fmt.Sprintf("PRIMARY KEY (%s)", constraintColumnName)
			case "UNIQUE":
				constraintDef = fmt.Sprintf("UNIQUE KEY %s (%s)", constraintName, constraintColumnName)
			case "FOREIGN KEY":
				constraintDef = fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)", constraintColumnName, constraintRefTableName.String, constraintRefColumnName.String)
				relation := &schema.Relation{
					Table: table,
					Def:   constraintDef,
				}
				relations = append(relations, relation)
			}

			constraint := &schema.Constraint{
				Name: constraintName,
				Type: constraintType,
				Def:  constraintDef,
			}

			constraints = append(constraints, constraint)
		}
		table.Constraints = constraints

		// columns and comments
		columnRows, err := db.Query(`
SELECT column_name, column_default, is_nullable, column_type, column_comment
FROM information_schema.columns
WHERE table_schema = ? AND table_name = ? ORDER BY ordinal_position`, s.Name, tableName)
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

func convertColumnNullable(str string) bool {
	if str == "NO" {
		return false
	}
	return true
}
