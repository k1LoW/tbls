package postgres

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

var reFK = regexp.MustCompile(`FOREIGN KEY \((.+)\) REFERENCES ([^\s]+)\s?\((.+)\)`)
var defaultSchemaName = "public"

// Postgres struct
type Postgres struct {
	db     *sql.DB
	rsMode bool
}

type constraintLink struct {
	constraint     *schema.Constraint
	table          string
	referenceTable string
	conkey         string
	confkey        string
}

// NewPostgres return new Postgres
func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{
		db:     db,
		rsMode: false,
	}
}

// Analyze PostgreSQL database schema
func (p *Postgres) Analyze(s *schema.Schema) error {
	constraintLinks := []constraintLink{}

	// tables
	tableRows, err := p.db.Query(`
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
		tableCommentRows, err := p.db.Query(`
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
			viewDefRows, err := p.db.Query(`
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

		// constraints
		constraintRows, err := p.db.Query(p.queryForConstraints(), tableName, tableSchema)
		defer constraintRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

		constraints := []*schema.Constraint{}

		for constraintRows.Next() {
			var (
				constraintName           string
				constraintDef            string
				constraintType           string
				constraintReferenceTable sql.NullString
				constraintConkey         sql.NullString
				constraintConfkey        sql.NullString
			)
			err = constraintRows.Scan(&constraintName, &constraintDef, &constraintType, &constraintReferenceTable, &constraintConkey, &constraintConfkey)
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
			constraintLinks = append(constraintLinks, constraintLink{
				constraint:     constraint,
				table:          table.Name,
				referenceTable: constraintReferenceTable.String,
				conkey:         constraintConkey.String,
				confkey:        constraintConfkey.String,
			})
		}
		table.Constraints = constraints

		// triggers
		if !p.rsMode {
			triggerRows, err := p.db.Query(`
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
		}

		// columns comments
		columnCommentRows, err := p.db.Query(`
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
		columnRows, err := p.db.Query(`
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

		// indexes
		indexRows, err := p.db.Query(`
SELECT
  i.relname AS indexname,
  pg_get_indexdef(i.oid) AS indexdef,
  indkey::text
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
				indkey    string
			)
			err = indexRows.Scan(&indexName, &indexDef, &indkey)
			if err != nil {
				return errors.WithStack(err)
			}
			index := &schema.Index{
				Name:  indexName,
				Def:   indexDef,
				Table: &table.Name,
			}
			idxs := indkeyToInts(indkey)
			for _, idx := range idxs {
				index.Columns = append(index.Columns, table.Columns[idx-1].Name)
			}

			indexes = append(indexes, index)
		}
		table.Indexes = indexes

		tables = append(tables, table)
	}

	s.Tables = tables

	// Link Constraints
	for _, l := range constraintLinks {
		if l.table != "" {
			table, err := s.FindTableByName(l.table)
			if err != nil {
				return err
			}
			l.constraint.Table = &table.Name
			if l.conkey != "" {
				idxs := colkeyToInts(l.conkey)
				for _, idx := range idxs {
					l.constraint.Columns = append(l.constraint.Columns, table.Columns[idx-1].Name)
				}
			}
		}
		if l.referenceTable != "" {
			referenceTable, err := s.FindTableByName(l.referenceTable)
			if err != nil {
				return err
			}
			l.constraint.ReferenceTable = &referenceTable.Name
			if l.confkey != "" {
				idxs := colkeyToInts(l.confkey)
				for _, idx := range idxs {
					l.constraint.ReferenceColumns = append(l.constraint.ReferenceColumns, referenceTable.Columns[idx-1].Name)
				}
			}
		}
	}

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

// Info return schema.Driver
func (p *Postgres) Info() (*schema.Driver, error) {
	var v string
	row := p.db.QueryRow(`SELECT version();`)
	row.Scan(&v)

	name := "postgres"
	if p.rsMode {
		name = "redshift"
	}

	d := &schema.Driver{
		Name:            name,
		DatabaseVersion: v,
	}
	return d, nil
}

// EnableRsMode enable rsMode
func (p *Postgres) EnableRsMode() {
	p.rsMode = true
}

func (p *Postgres) queryForConstraints() string {
	if p.rsMode {
		return `
SELECT
  pc.conname AS name,
  pg_get_constraintdef(pc.oid) AS def,
  contype AS type,
  psf.relname,
  pc.conkey::text,
  pc.confkey::text
FROM pg_constraint AS pc
LEFT JOIN pg_stat_user_tables AS ps ON ps.relid = pc.conrelid
LEFT JOIN pg_stat_user_tables AS psf ON psf.relid = pc.confrelid
WHERE ps.relname = $1
AND ps.schemaname = $2
ORDER BY pc.conrelid, pc.conname`
	}
	return `
SELECT
  pc.conname AS name,
  (CASE WHEN contype='t' THEN pg_get_triggerdef((SELECT oid FROM pg_trigger WHERE tgconstraint = pc.oid LIMIT 1))
        ELSE pg_get_constraintdef(pc.oid)
   END) AS def,
  contype AS type,
  psf.relname,
  pc.conkey::text,
  pc.confkey::text
FROM pg_constraint AS pc
LEFT JOIN pg_stat_user_tables AS ps ON ps.relid = pc.conrelid
LEFT JOIN pg_stat_user_tables AS psf ON psf.relid = pc.confrelid
WHERE ps.relname = $1
AND ps.schemaname = $2
ORDER BY pc.conrelid, pc.conindid, pc.conname`
}

func colkeyToInts(colkey string) []int {
	ints := []int{}
	if colkey == "" {
		return ints
	}
	strs := strings.Split(strings.Trim(colkey, "{}"), ",")
	for _, s := range strs {
		i, _ := strconv.Atoi(s)
		ints = append(ints, i)
	}
	return ints
}

func indkeyToInts(indkey string) []int {
	ints := []int{}
	if indkey == "" {
		return ints
	}
	strs := strings.Split(indkey, " ")
	for _, s := range strs {
		i, _ := strconv.Atoi(s)
		ints = append(ints, i)
	}
	return ints
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
