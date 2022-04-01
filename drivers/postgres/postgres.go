package postgres

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/aquasecurity/go-version/pkg/version"
	"github.com/k1LoW/tbls/ddl"
	"github.com/k1LoW/tbls/dict"
	"github.com/k1LoW/tbls/schema"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

var reFK = regexp.MustCompile(`FOREIGN KEY \((.+)\) REFERENCES ([^\s]+)\s?\((.+)\)`)
var reVersion = regexp.MustCompile(`([0-9]+(\.[0-9]+)*)`)

// Postgres struct
type Postgres struct {
	db     *sql.DB
	rsMode bool
}

// New return new Postgres
func New(db *sql.DB) *Postgres {
	return &Postgres{
		db:     db,
		rsMode: false,
	}
}

// Analyze PostgreSQL database schema
func (p *Postgres) Analyze(s *schema.Schema) error {
	d, err := p.Info()
	if err != nil {
		return errors.WithStack(err)
	}
	s.Driver = d

	// current schema
	var currentSchema string
	schemaRows, err := p.db.Query(`SELECT current_schema()`)
	if err != nil {
		return errors.WithStack(err)
	}
	defer schemaRows.Close()
	for schemaRows.Next() {
		err := schemaRows.Scan(&currentSchema)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	s.Driver.Meta.CurrentSchema = currentSchema

	// search_path
	var searchPaths string
	pathRows, err := p.db.Query(`SHOW search_path`)
	if err != nil {
		return errors.WithStack(err)
	}
	defer pathRows.Close()
	for pathRows.Next() {
		err := pathRows.Scan(&searchPaths)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	s.Driver.Meta.SearchPaths = strings.Split(searchPaths, ", ")

	fullTableNames := []string{}

	// tables
	tableRows, err := p.db.Query(`
SELECT
    cls.oid AS oid,
    cls.relname AS table_name,
    CASE
        WHEN cls.relkind IN ('r', 'p') THEN 'BASE TABLE'
        WHEN cls.relkind = 'v' THEN 'VIEW'
        WHEN cls.relkind = 'm' THEN 'MATERIALIZED VIEW'
        WHEN cls.relkind = 'f' THEN 'FOREIGN TABLE'
    END AS table_type,
    ns.nspname AS table_schema,
    descr.description AS table_comment
FROM pg_class AS cls
INNER JOIN pg_namespace AS ns ON cls.relnamespace = ns.oid
LEFT JOIN pg_description AS descr ON cls.oid = descr.objoid AND descr.objsubid = 0
WHERE ns.nspname NOT IN ('pg_catalog', 'information_schema')
AND cls.relkind IN ('r', 'p', 'v', 'f', 'm')
ORDER BY oid`)
	if err != nil {
		return errors.WithStack(err)
	}
	defer tableRows.Close()

	relations := []*schema.Relation{}

	tables := []*schema.Table{}
	for tableRows.Next() {
		var (
			tableOid     uint64
			tableName    string
			tableType    string
			tableSchema  string
			tableComment sql.NullString
		)
		err := tableRows.Scan(&tableOid, &tableName, &tableType, &tableSchema, &tableComment)
		if err != nil {
			return errors.WithStack(err)
		}

		name := fmt.Sprintf("%s.%s", tableSchema, tableName)

		fullTableNames = append(fullTableNames, name)

		table := &schema.Table{
			Name:    name,
			Type:    tableType,
			Comment: tableComment.String,
		}

		// (materialized) view definition
		if tableType == "VIEW" || tableType == "MATERIALIZED VIEW" {
			viewDefRows, err := p.db.Query(`SELECT pg_get_viewdef($1::oid);`, tableOid)
			if err != nil {
				return errors.WithStack(err)
			}
			defer viewDefRows.Close()
			for viewDefRows.Next() {
				var tableDef sql.NullString
				err := viewDefRows.Scan(&tableDef)
				if err != nil {
					return errors.WithStack(err)
				}
				table.Def = fmt.Sprintf("CREATE %s %s AS (\n%s\n)", tableType, tableName, strings.TrimRight(tableDef.String, ";"))
			}
		}

		// constraints
		constraintRows, err := p.db.Query(p.queryForConstraints(), tableOid)
		if err != nil {
			return errors.WithStack(err)
		}
		defer constraintRows.Close()

		constraints := []*schema.Constraint{}

		for constraintRows.Next() {
			var (
				constraintName                  string
				constraintDef                   string
				constraintType                  string
				constraintReferencedTable       sql.NullString
				constraintColumnNames           []sql.NullString
				constraintReferencedColumnNames []sql.NullString
				constraintComment               sql.NullString
			)
			err = constraintRows.Scan(&constraintName, &constraintDef, &constraintType, &constraintReferencedTable, pq.Array(&constraintColumnNames), pq.Array(&constraintReferencedColumnNames), &constraintComment)
			if err != nil {
				return errors.WithStack(err)
			}
			rt := constraintReferencedTable.String
			constraint := &schema.Constraint{
				Name:              constraintName,
				Type:              convertConstraintType(constraintType),
				Def:               constraintDef,
				Table:             &table.Name,
				Columns:           arrayRemoveNull(constraintColumnNames),
				ReferencedTable:   &rt,
				ReferencedColumns: arrayRemoveNull(constraintReferencedColumnNames),
				Comment:           constraintComment.String,
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
		if !p.rsMode {
			triggerRows, err := p.db.Query(`
SELECT tgname, pg_get_triggerdef(trig.oid), descr.description AS comment
FROM pg_trigger AS trig
LEFT JOIN pg_description AS descr ON trig.oid = descr.objoid
WHERE tgisinternal = false
AND tgrelid = $1::oid
ORDER BY tgrelid
`, tableOid)
			if err != nil {
				return errors.WithStack(err)
			}
			defer triggerRows.Close()

			triggers := []*schema.Trigger{}
			for triggerRows.Next() {
				var (
					triggerName    string
					triggerDef     string
					triggerComment sql.NullString
				)
				err = triggerRows.Scan(&triggerName, &triggerDef, &triggerComment)
				if err != nil {
					return errors.WithStack(err)
				}
				trigger := &schema.Trigger{
					Name:    triggerName,
					Def:     triggerDef,
					Comment: triggerComment.String,
				}
				triggers = append(triggers, trigger)
			}
			table.Triggers = triggers
		}

		// columns
		columnStmt, err := p.queryForColumns(s.Driver.DatabaseVersion)
		if err != nil {
			return errors.WithStack(err)
		}
		columnRows, err := p.db.Query(columnStmt, tableOid)
		if err != nil {
			return errors.WithStack(err)
		}
		defer columnRows.Close()

		columns := []*schema.Column{}
		for columnRows.Next() {
			var (
				columnName               string
				columnDefaultOrGenerated sql.NullString
				attrgenerated            sql.NullString
				isNullable               bool
				dataType                 string
				columnComment            sql.NullString
			)
			err = columnRows.Scan(&columnName, &columnDefaultOrGenerated, &attrgenerated, &isNullable, &dataType, &columnComment)
			if err != nil {
				return errors.WithStack(err)
			}
			column := &schema.Column{
				Name:     columnName,
				Type:     dataType,
				Nullable: isNullable,
				Comment:  columnComment.String,
			}
			switch attrgenerated.String {
			case "":
				column.Default = columnDefaultOrGenerated
			case "s":
				column.ExtraDef = fmt.Sprintf("GENERATED ALWAYS AS %s STORED", columnDefaultOrGenerated.String)
			default:
				return errors.Errorf("unsupported pg_attribute.attrgenerated '%s'", attrgenerated.String)
			}
			columns = append(columns, column)
		}
		table.Columns = columns

		// indexes
		indexRows, err := p.db.Query(p.queryForIndexes(), tableOid)
		if err != nil {
			return errors.WithStack(err)
		}
		defer indexRows.Close()

		indexes := []*schema.Index{}
		for indexRows.Next() {
			var (
				indexName        string
				indexDef         string
				indexColumnNames []sql.NullString
				indexComment     sql.NullString
			)
			err = indexRows.Scan(&indexName, &indexDef, pq.Array(&indexColumnNames), &indexComment)
			if err != nil {
				return errors.WithStack(err)
			}
			index := &schema.Index{
				Name:    indexName,
				Def:     indexDef,
				Table:   &table.Name,
				Columns: arrayRemoveNull(indexColumnNames),
				Comment: indexComment.String,
			}

			indexes = append(indexes, index)
		}
		table.Indexes = indexes

		tables = append(tables, table)
	}

	subroutines, err := p.getSubroutines()
	if err != nil {
		return err
	}
	s.Subroutines = subroutines

	s.Tables = tables

	// Relations
	for _, r := range relations {
		result := reFK.FindAllStringSubmatch(r.Def, -1)
		if len(result) < 1 || len(result[0]) < 4 {
			return errors.Errorf("can not parse foreign key: %s", r.Def)
		}
		strColumns := []string{}
		for _, c := range strings.Split(result[0][1], ", ") {
			strColumns = append(strColumns, strings.ReplaceAll(c, `"`, ""))
		}
		strParentTable := strings.ReplaceAll(result[0][2], `"`, "")
		strParentColumns := []string{}
		for _, c := range strings.Split(result[0][3], ", ") {
			strParentColumns = append(strParentColumns, strings.ReplaceAll(c, `"`, ""))
		}
		for _, c := range strColumns {
			column, err := r.Table.FindColumnByName(c)
			if err != nil {
				return err
			}
			r.Columns = append(r.Columns, column)
			column.ParentRelations = append(column.ParentRelations, r)
		}

		dn, err := detectFullTableName(strParentTable, s.Driver.Meta.SearchPaths, fullTableNames)
		if err != nil {
			return err
		}
		strParentTable = dn
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

	// referenced tables of view
	for _, t := range s.Tables {
		if t.Type != "VIEW" && t.Type != "MATERIALIZED VIEW" {
			continue
		}
		for _, rts := range ddl.ParseReferencedTables(t.Def) {
			rt, err := s.FindTableByName(rts)
			if err != nil {
				rt = &schema.Table{
					Name:     rts,
					External: true,
				}
			}
			t.ReferencedTables = append(t.ReferencedTables, rt)
		}
	}

	return nil
}

const querySubroutines95 = `select n.nspname as schema_name,
p.proname as specific_name,
TEXT 'FUNCTION',
t.typname as return_type,
pg_get_function_arguments(p.oid) as arguments
from pg_proc p
left join pg_namespace n on p.pronamespace = n.oid
left join pg_type t on t.oid = p.prorettype 
where n.nspname not in ('pg_catalog', 'information_schema')`

const querySubroutines = `select n.nspname as schema_name,
p.proname as specific_name,
case when p.prokind = 'p' then TEXT 'PROCEDURE' else case when p.prokind = 'f' then TEXT 'FUNCTION' else p.prokind end end,
t.typname as return_type,
pg_get_function_arguments(p.oid) as arguments
from pg_proc p
left join pg_namespace n on p.pronamespace = n.oid
left join pg_type t on t.oid = p.prorettype 
where n.nspname not in ('pg_catalog', 'information_schema')`

const queryStoredProcedureSupported = `SELECT column_name 
FROM information_schema.columns 
WHERE table_name='pg_proc' and column_name='prokind';`

func (p *Postgres) isProceduresSupported() (bool, error) {
	result, err := p.db.Query(queryStoredProcedureSupported)
	if err != nil {
		return false, errors.WithStack(err)
	}
	defer result.Close()

	for result.Next() {
		var (
			name sql.NullString
		)
		err := result.Scan(&name)
		if err != nil {
			return false, errors.WithStack(err)
		}
		return true, nil
	}
	return false, nil
}

func (p *Postgres) getSubroutines() ([]*schema.Subroutine, error) {
	var subroutines []*schema.Subroutine
	storedProcedureSupported, err := p.isProceduresSupported()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if storedProcedureSupported {
		subroutines, err = p.getSubroutinesByQuery(querySubroutines)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		subroutines, err = p.getSubroutinesByQuery(querySubroutines95)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return subroutines, nil
}

func (p *Postgres) getSubroutinesByQuery(query string) ([]*schema.Subroutine, error) {
	subroutines := []*schema.Subroutine{}
	subroutinesResult, err := p.db.Query(query)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer subroutinesResult.Close()

	for subroutinesResult.Next() {
		var (
			schemaName string
			name       string
			typeValue  string
			returnType string
			arguments  sql.NullString
		)
		err := subroutinesResult.Scan(&schemaName, &name, &typeValue, &returnType, &arguments)
		if err != nil {
			return subroutines, errors.WithStack(err)
		}
		subroutine := &schema.Subroutine{
			Name:       fullTableName(schemaName, name),
			Type:       typeValue,
			ReturnType: returnType,
			Arguments:  arguments.String,
		}

		subroutines = append(subroutines, subroutine)
	}
	return subroutines, nil
}

func fullTableName(owner string, tableName string) string {
	return fmt.Sprintf("%s.%s", owner, tableName)
}

// Info return schema.Driver
func (p *Postgres) Info() (*schema.Driver, error) {
	var v string
	row := p.db.QueryRow(`SELECT version();`)
	err := row.Scan(&v)
	if err != nil {
		return nil, err
	}

	name := "postgres"
	if p.rsMode {
		name = "redshift"
	}

	dct := dict.New()
	dct.Merge(map[string]string{
		"Subroutines": "Stored procedures and functions",
	})

	d := &schema.Driver{
		Name:            name,
		DatabaseVersion: v,
		Meta: &schema.DriverMeta{
			Dict: &dct,
		},
	}
	return d, nil
}

// EnableRsMode enable rsMode
func (p *Postgres) EnableRsMode() {
	p.rsMode = true
}

func (p *Postgres) queryForColumns(v string) (string, error) {
	verGeneratedColumn, err := version.Parse("12")
	if err != nil {
		return "", err
	}
	// v => PostgreSQL 9.5.24 on x86_64-pc-linux-gnu (Debian 9.5.24-1.pgdg90+1), compiled by gcc (Debian 6.3.0-18+deb9u1) 6.3.0 20170516, 64-bit
	matches := reVersion.FindStringSubmatch(v)
	if matches == nil || len(matches) < 2 {
		return "", errors.Errorf("malformed version: %s", v)
	}
	vv, err := version.Parse(matches[1])
	if err != nil {
		return "", err
	}
	if vv.LessThan(verGeneratedColumn) {
		return `
SELECT
    attr.attname AS column_name,
    pg_get_expr(def.adbin, def.adrelid) AS column_default,
    '' as dummy,
    NOT (attr.attnotnull OR tp.typtype = 'd' AND tp.typnotnull) AS is_nullable,
    CASE
        WHEN 'character varying'::regtype = ANY(ARRAY[attr.atttypid, tp.typelem]) THEN
            REPLACE(format_type(attr.atttypid, attr.atttypmod), 'character varying', 'varchar')
        ELSE format_type(attr.atttypid, attr.atttypmod)
    END AS data_type,
    descr.description AS comment
FROM pg_attribute AS attr
INNER JOIN pg_type AS tp ON attr.atttypid = tp.oid
LEFT JOIN pg_attrdef AS def ON attr.attrelid = def.adrelid AND attr.attnum = def.adnum AND attr.atthasdef
LEFT JOIN pg_description AS descr ON attr.attrelid = descr.objoid AND attr.attnum = descr.objsubid
WHERE
    attr.attnum > 0
AND NOT attr.attisdropped
AND attr.attrelid = $1::oid
ORDER BY attr.attnum;
`, nil
	} else {
		return `
SELECT
    attr.attname AS column_name,
    pg_get_expr(def.adbin, def.adrelid) AS column_default,
    attr.attgenerated,
    NOT (attr.attnotnull OR tp.typtype = 'd' AND tp.typnotnull) AS is_nullable,
    CASE
        WHEN 'character varying'::regtype = ANY(ARRAY[attr.atttypid, tp.typelem]) THEN
            REPLACE(format_type(attr.atttypid, attr.atttypmod), 'character varying', 'varchar')
        ELSE format_type(attr.atttypid, attr.atttypmod)
    END AS data_type,
    descr.description AS comment
FROM pg_attribute AS attr
INNER JOIN pg_type AS tp ON attr.atttypid = tp.oid
LEFT JOIN pg_attrdef AS def ON attr.attrelid = def.adrelid AND attr.attnum = def.adnum AND attr.atthasdef
LEFT JOIN pg_description AS descr ON attr.attrelid = descr.objoid AND attr.attnum = descr.objsubid
WHERE
    attr.attnum > 0
AND NOT attr.attisdropped
AND attr.attrelid = $1::oid
ORDER BY attr.attnum;
`, nil
	}
}

func (p *Postgres) queryForConstraints() string {
	if p.rsMode {
		return `
SELECT
  conname, pg_get_constraintdef(oid), contype, NULL, NULL, NULL, NULL
FROM pg_constraint
WHERE conrelid = $1::oid
ORDER BY conname`
	}
	return `
SELECT
  cons.conname AS name,
  CASE WHEN cons.contype = 't' THEN pg_get_triggerdef(trig.oid)
        ELSE pg_get_constraintdef(cons.oid)
  END AS def,
  cons.contype AS type,
  fcls.relname,
  ARRAY_AGG(attr.attname),
  ARRAY_AGG(fattr.attname),
  descr.description AS comment
FROM pg_constraint AS cons
LEFT JOIN pg_trigger AS trig ON trig.tgconstraint = cons.oid AND NOT trig.tgisinternal
LEFT JOIN pg_class AS fcls ON cons.confrelid = fcls.oid
LEFT JOIN pg_attribute AS attr ON attr.attrelid = cons.conrelid
LEFT JOIN pg_attribute AS fattr ON fattr.attrelid = cons.confrelid
LEFT JOIN pg_description AS descr ON cons.oid = descr.objoid
WHERE
	cons.conrelid = $1::oid
AND (cons.conkey IS NULL OR attr.attnum = ANY(cons.conkey))
AND (cons.confkey IS NULL OR fattr.attnum = ANY(cons.confkey))
GROUP BY cons.conindid, cons.conname, cons.contype, cons.oid, trig.oid, fcls.relname, descr.description
ORDER BY cons.conindid, cons.conname`
}

// arrayRemoveNull
func arrayRemoveNull(in []sql.NullString) []string {
	out := []string{}
	for _, i := range in {
		if i.Valid {
			out = append(out, i.String)
		}
	}
	return out
}

func (p *Postgres) queryForIndexes() string {
	if p.rsMode {
		return `
SELECT
  cls.relname AS indexname,
  pg_get_indexdef(idx.indexrelid) AS indexdef,
  NULL,
  NULL
FROM pg_index AS idx
INNER JOIN pg_class AS cls ON idx.indexrelid = cls.oid
WHERE idx.indrelid = $1::oid
ORDER BY idx.indexrelid`
	}
	return `
SELECT
  cls.relname AS indexname,
  pg_get_indexdef(idx.indexrelid) AS indexdef,
  ARRAY_AGG(attr.attname),
  descr.description AS comment
FROM pg_index AS idx
INNER JOIN pg_class AS cls ON idx.indexrelid = cls.oid
INNER JOIN pg_attribute AS attr ON idx.indexrelid = attr.attrelid
LEFT JOIN pg_description AS descr ON idx.indexrelid = descr.objoid
WHERE idx.indrelid = $1::oid
GROUP BY cls.relname, idx.indexrelid, descr.description
ORDER BY idx.indexrelid`
}

func detectFullTableName(name string, searchPaths, fullTableNames []string) (string, error) {
	if strings.Contains(name, ".") {
		return name, nil
	}
	fns := []string{}
	for _, n := range fullTableNames {
		if strings.HasSuffix(n, name) {
			for _, p := range searchPaths {
				// TODO: Support $user
				if n == fmt.Sprintf("%s.%s", p, name) {
					fns = append(fns, n)
				}
			}
		}
	}
	if len(fns) != 1 {
		return "", errors.Errorf("can not detect table name: %s", name)
	}
	return fns[0], nil
}

func convertConstraintType(t string) string {
	switch t {
	case "p":
		return "PRIMARY KEY"
	case "u":
		return "UNIQUE"
	case "f":
		return schema.TypeFK
	case "c":
		return "CHECK"
	case "t":
		return "TRIGGER"
	default:
		return t
	}
}
