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
	defer schemaRows.Close()
	if err != nil {
		return errors.WithStack(err)
	}
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
	defer pathRows.Close()
	if err != nil {
		return errors.WithStack(err)
	}
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
        WHEN cls.relkind = 'f' THEN 'FOREIGN TABLE'
    END AS table_type,
    ns.nspname AS table_schema,
    descr.description AS table_comment
FROM pg_class cls
INNER JOIN pg_namespace ns ON cls.relnamespace = ns.oid
LEFT JOIN pg_description descr ON cls.oid = descr.objoid AND descr.objsubid = 0
WHERE ns.nspname NOT IN ('pg_catalog', 'information_schema')
AND cls.relkind IN ('r', 'p', 'v', 'f')
ORDER BY oid`)
	defer tableRows.Close()
	if err != nil {
		return errors.WithStack(err)
	}

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
			Name: name,
			Type: tableType,
			Comment: tableComment.String,
		}

		// view definition
		if tableType == "VIEW" {
			viewDefRows, err := p.db.Query(`SELECT pg_get_viewdef($1::oid);`, tableOid)
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
				table.Def = fmt.Sprintf("CREATE VIEW %s AS (\n%s\n)", tableName, strings.TrimRight(tableDef.String, ";"))
			}
		}

		// constraints
		constraintRows, err := p.db.Query(p.queryForConstraints(), tableOid)
		defer constraintRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

		constraints := []*schema.Constraint{}

		for constraintRows.Next() {
			var (
				constraintName                string
				constraintDef                 string
				constraintType                string
				constraintReferenceTable      sql.NullString
				constraintColumnName          sql.NullString
				constraintReferenceColumnName sql.NullString
			)
			err = constraintRows.Scan(&constraintName, &constraintDef, &constraintType, &constraintReferenceTable, &constraintColumnName, &constraintReferenceColumnName)
			if err != nil {
				return errors.WithStack(err)
			}
			rt := constraintReferenceTable.String
			constraint := &schema.Constraint{
				Name:             constraintName,
				Type:             convertConstraintType(constraintType),
				Def:              constraintDef,
				Table:            &table.Name,
				Columns:          strings.Split(constraintColumnName.String, ", "),
				ReferenceTable:   &rt,
				ReferenceColumns: strings.Split(constraintReferenceColumnName.String, ", "),
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
SELECT tgname, pg_get_triggerdef(oid)
FROM pg_trigger
WHERE tgisinternal = false
AND tgrelid = $1::oid
ORDER BY tgrelid
`, tableOid)
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

		// columns
		columnRows, err := p.db.Query(`
SELECT
    attr.attname AS column_name,
    pg_get_expr(def.adbin, def.adrelid) AS column_default,
    NOT (attr.attnotnull OR tp.typtype = 'd' AND tp.typnotnull) AS is_nullable,
    CASE
        WHEN attr.atttypid::regtype = ANY(ARRAY['character varying'::regtype, 'character varying[]'::regtype]) THEN
            REPLACE(format_type(attr.atttypid, attr.atttypmod), 'character varying', 'varchar')
        ELSE format_type(attr.atttypid, attr.atttypmod)
    END AS data_type,
    descr.description as comment
FROM pg_attribute attr
INNER JOIN pg_type tp ON attr.atttypid = tp.oid
LEFT JOIN pg_attrdef def ON attr.attrelid = def.adrelid AND attr.attnum = def.adnum
LEFT JOIN pg_description descr ON attr.attrelid = descr.objoid AND attr.attnum = descr.objsubid
WHERE
    attr.attnum > 0
AND NOT attr.attisdropped
AND attr.attrelid = $1::oid
ORDER BY attr.attnum;
`, tableOid)
		defer columnRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

		columns := []*schema.Column{}
		for columnRows.Next() {
			var (
				columnName             string
				columnDefault          sql.NullString
				isNullable             bool
				dataType               string
				columnComment          sql.NullString
			)
			err = columnRows.Scan(&columnName, &columnDefault, &isNullable, &dataType, &columnComment)
			if err != nil {
				return errors.WithStack(err)
			}
			column := &schema.Column{
				Name:     columnName,
				Type:     dataType,
				Nullable: isNullable,
				Default:  columnDefault,
				Comment:  columnComment.String,
			}
			columns = append(columns, column)
		}
		table.Columns = columns

		// indexes
		indexRows, err := p.db.Query(p.queryForIndexes(), tableOid)
		defer indexRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

		indexes := []*schema.Index{}
		for indexRows.Next() {
			var (
				indexName       string
				indexDef        string
				indexColumnName sql.NullString
			)
			err = indexRows.Scan(&indexName, &indexDef, &indexColumnName)
			if err != nil {
				return errors.WithStack(err)
			}
			index := &schema.Index{
				Name:    indexName,
				Def:     indexDef,
				Table:   &table.Name,
				Columns: strings.Split(indexColumnName.String, ", "),
			}

			indexes = append(indexes, index)
		}
		table.Indexes = indexes

		tables = append(tables, table)
	}

	s.Tables = tables

	// Relations
	for _, r := range relations {
		result := reFK.FindAllStringSubmatch(r.Def, -1)
		strColumns := []string{}
		for _, c := range strings.Split(result[0][1], ", ") {
			strColumns = append(strColumns, strings.Trim(c, `"`))
		}
		strParentTable := strings.Trim(result[0][2], `"`)
		strParentColumns := []string{}
		for _, c := range strings.Split(result[0][3], ", ") {
			strParentColumns = append(strParentColumns, strings.Trim(c, `"`))
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

	return nil
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

	d := &schema.Driver{
		Name:            name,
		DatabaseVersion: v,
		Meta:            &schema.DriverMeta{},
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
  conname, pg_get_constraintdef(oid) AS def, contype, NULL, NULL, NULL
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
  ARRAY_TO_STRING(ARRAY_AGG(attr.attname), ', '),
  ARRAY_TO_STRING(ARRAY_AGG(fattr.attname), ', ')
FROM pg_constraint AS cons
LEFT JOIN pg_trigger trig ON trig.tgconstraint = cons.oid AND NOT trig.tgisinternal
LEFT JOIN pg_class AS fcls ON cons.confrelid = fcls.oid
LEFT JOIN pg_attribute attr ON attr.attrelid = cons.conrelid
LEFT JOIN pg_attribute fattr ON fattr.attrelid = cons.confrelid
WHERE
	cons.conrelid = $1::oid
AND (cons.conkey IS NULL OR attr.attnum = ANY(cons.conkey))
AND (cons.confkey IS NULL OR fattr.attnum = ANY(cons.confkey))
GROUP BY cons.conindid, cons.conname, cons.contype, cons.oid, trig.oid, fcls.relname
ORDER BY cons.conindid, cons.conname`
}

func (p *Postgres) queryForIndexes() string {
	if p.rsMode {
		return `
SELECT
  cls.relname AS indexname,
  pg_get_indexdef(idx.indexrelid) AS indexdef,
  NULL
FROM pg_index idx
INNER JOIN pg_class cls ON idx.indexrelid = cls.oid
WHERE idx.indrelid = $1::oid
ORDER BY idx.indexrelid`
	}
	return `
SELECT
  cls.relname AS indexname,
  pg_get_indexdef(idx.indexrelid) AS indexdef,
  ARRAY_TO_STRING(ARRAY_AGG(attr.attname), ', ')
FROM pg_index idx
INNER JOIN pg_class cls ON idx.indexrelid = cls.oid
INNER JOIN pg_attribute attr on idx.indexrelid = attr.attrelid
WHERE idx.indrelid = $1::oid
GROUP BY cls.relname, idx.indexrelid
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
