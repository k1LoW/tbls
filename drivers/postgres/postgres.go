package postgres

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

var reFK = regexp.MustCompile(`FOREIGN KEY \((.+)\) REFERENCES ([^\s]+)\s?\((.+)\)`)
var defaultSchemaName = "public"

// Postgres struct
type Postgres struct{}

// Analyze PostgreSQL database schema
func (p *Postgres) Analyze(db *sql.DB, s *schema.Schema) error {

	// tables
	tableRows, err := db.Query(`
SELECT DISTINCT cls.oid AS oid, cls.relname AS table_name, tbl.table_type AS table_type, tbl.table_schema AS table_schema
FROM pg_catalog.pg_class cls
INNER JOIN pg_namespace ns ON cls.relnamespace = ns.oid
INNER JOIN (SELECT table_name, table_type, table_schema
FROM information_schema.tables
WHERE table_schema != 'pg_catalog' AND table_schema != 'information_schema'
AND table_catalog = $1) tbl ON cls.relname = tbl.table_name
ORDER BY oid`, s.Name)
	defer tableRows.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	relations := []*schema.Relation{}

	tables := []*schema.Table{}
	for tableRows.Next() {
		var (
			tableOid    string
			tableName   string
			tableType   string
			tableSchema string
		)
		err := tableRows.Scan(&tableOid, &tableName, &tableType, &tableSchema)
		if err != nil {
			return errors.WithStack(err)
		}

		name := tableName
		if tableSchema != defaultSchemaName {
			name = fmt.Sprintf("%s.%s", tableSchema, tableName)
		}

		table := &schema.Table{
			Name: name,
			Type: tableType,
		}

		// table comment
		tableCommentRows, err := db.Query(`
SELECT pd.description as comment
FROM pg_stat_user_tables AS ps, pg_description AS pd
WHERE ps.relid=pd.objoid
AND pd.objsubid=0
AND ps.relname = $1
AND ps.schemaname = $2`, tableName, tableSchema)
		defer tableCommentRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

		for tableCommentRows.Next() {
			var tableComment string
			err = tableCommentRows.Scan(&tableComment)
			if err != nil {
				return errors.WithStack(err)
			}
			table.Comment = tableComment
		}

		// view definition
		if tableType == "VIEW" {
			viewDefRows, err := db.Query(`
SELECT view_definition FROM information_schema.views
WHERE table_catalog = $1
AND table_name = $2
AND table_schema = $3;
		`, s.Name, tableName, tableSchema)
			defer viewDefRows.Close()
			if err != nil {
				return errors.WithStack(err)
			}
			for viewDefRows.Next() {
				var tableDef string
				err := viewDefRows.Scan(&tableDef)
				if err != nil {
					return errors.WithStack(err)
				}
				table.Def = fmt.Sprintf("CREATE VIEW %s AS (\n%s\n)", tableName, strings.TrimRight(tableDef, ";"))
			}
		}

		// indexes
		indexRows, err := db.Query(`
SELECT
i.relname AS indexname,
pg_get_indexdef(i.oid) AS indexdef
FROM ((((pg_index x
JOIN pg_class c ON ((c.oid = x.indrelid)))
JOIN pg_class i ON ((i.oid = x.indexrelid)))
LEFT JOIN pg_namespace n ON ((n.oid = c.relnamespace))))
WHERE ((c.relkind = ANY (ARRAY['r'::"char", 'm'::"char"])) AND (i.relkind = 'i'::"char"))
AND c.relname = $1
AND n.nspname = $2
ORDER BY x.indexrelid
`, tableName, tableSchema)
		defer indexRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

		indexes := []*schema.Index{}
		for indexRows.Next() {
			var (
				indexName string
				indexDef  string
			)
			err = indexRows.Scan(&indexName, &indexDef)
			if err != nil {
				return errors.WithStack(err)
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
  pc.conname AS name,
  (CASE WHEN contype='t' THEN pg_get_triggerdef((SELECT oid FROM pg_trigger WHERE tgconstraint = pc.oid LIMIT 1))
        ELSE pg_get_constraintdef(pc.oid)
   END) AS def,
  contype AS type
FROM pg_constraint AS pc
LEFT JOIN pg_stat_user_tables AS ps ON ps.relid = pc.conrelid
WHERE ps.relname = $1
AND ps.schemaname = $2
ORDER BY pc.conrelid`, tableName, tableSchema)
		defer constraintRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

		constraints := []*schema.Constraint{}
		for constraintRows.Next() {
			var (
				constraintName string
				constraintDef  string
				constraintType string
			)
			err = constraintRows.Scan(&constraintName, &constraintDef, &constraintType)
			if err != nil {
				return errors.WithStack(err)
			}
			constraint := &schema.Constraint{
				Name: constraintName,
				Type: convertConstraintType(constraintType),
				Def:  constraintDef,
			}
			if constraintType == "f" {
				relation := &schema.Relation{
					Table: table,
					Def:   constraintDef,
				}
				relations = append(relations, relation)
			}
			constraints = append(constraints, constraint)
		}
		table.Constraints = constraints

		// triggers
		triggerRows, err := db.Query(`
SELECT tgname, pg_get_triggerdef(pt.oid)
FROM pg_trigger AS pt
LEFT JOIN pg_stat_user_tables AS ps ON ps.relid = pt.tgrelid
WHERE pt.tgisinternal = false
AND ps.relname = $1
AND ps.schemaname = $2
ORDER BY pt.tgrelid
`, tableName, tableSchema)
		defer triggerRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

		triggers := []*schema.Trigger{}
		for triggerRows.Next() {
			var (
				triggerName string
				triggerDef  string
			)
			err = triggerRows.Scan(&triggerName, &triggerDef)
			if err != nil {
				return errors.WithStack(err)
			}
			trigger := &schema.Trigger{
				Name: triggerName,
				Def:  triggerDef,
			}
			triggers = append(triggers, trigger)
		}
		table.Triggers = triggers

		// columns comments
		columnCommentRows, err := db.Query(`
SELECT pa.attname AS column_name, pd.description AS comment
FROM pg_stat_all_tables AS ps ,pg_description AS pd ,pg_attribute AS pa
WHERE ps.relid=pd.objoid
AND pd.objsubid != 0
AND pd.objoid=pa.attrelid
AND pd.objsubid=pa.attnum
AND ps.relname = $1
AND ps.schemaname = $2`, tableName, tableSchema)
		defer columnCommentRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

		columnComments := make(map[string]string)
		for columnCommentRows.Next() {
			var (
				columnName    string
				columnComment string
			)
			err = columnCommentRows.Scan(&columnName, &columnComment)
			if err != nil {
				return errors.WithStack(err)
			}
			columnComments[columnName] = columnComment
		}

		// columns
		columnRows, err := db.Query(`
SELECT column_name, column_default, is_nullable, data_type, udt_name, character_maximum_length
FROM information_schema.columns
WHERE table_name = $1
AND table_schema = $2
ORDER BY ordinal_position
`, tableName, tableSchema)
		defer columnRows.Close()
		if err != nil {
			return errors.WithStack(err)
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
				return errors.WithStack(err)
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

func convertConstraintType(t string) string {
	switch t {
	case "p":
		return "PRIMARY KEY"
	case "u":
		return "UNIQUE"
	case "f":
		return "FOREIGN KEY"
	case "c":
		return "CHECK"
	case "t":
		return "TRIGGER"
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
