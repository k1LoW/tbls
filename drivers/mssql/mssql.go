package mssql

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

var defaultSchemaName = "dbo"
var typeFk = "FOREIGN KEY"
var typeCheck = "CHECK"
var reSystemNamed = regexp.MustCompile(`_[^_]+$`)

// Mssql struct
type Mssql struct {
	db *sql.DB
}

type relationLink struct {
	table         string
	columns       []string
	parentTable   string
	parentColumns []string
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
	links := []relationLink{}

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

		// constraints
		constraints := []*schema.Constraint{}
		/// key constraints
		keyRows, err := m.db.Query(`
SELECT
  c.name,
  i.type_desc,
  i.is_unique,
  i.is_primary_key,
  i.is_unique_constraint,
  STRING_AGG(COL_NAME(ic.object_id, ic.column_id), ', ') WITHIN GROUP ( ORDER BY ic.key_ordinal ),
  c.is_system_named
FROM sys.key_constraints AS c
LEFT JOIN sys.indexes AS i ON i.object_id = c.parent_object_id AND i.index_id = c.unique_index_id
INNER JOIN sys.index_columns AS ic
ON i.object_id = ic.object_id AND i.index_id = ic.index_id
WHERE i.object_id = OBJECT_ID($1)
GROUP BY c.name, i.index_id, i.type_desc, i.is_unique, i.is_primary_key, i.is_unique_constraint, c.is_system_named
ORDER BY i.index_id
`, fmt.Sprintf("%s.%s", tableSchema, tableName))
		defer keyRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}
		for keyRows.Next() {
			var (
				indexName               string
				indexClusterType        string
				indexIsUnique           bool
				indexIsPrimaryKey       bool
				indexIsUniqueConstraint bool
				indexColumnName         sql.NullString
				indexIsSystemNamed      bool
			)
			err = keyRows.Scan(&indexName, &indexClusterType, &indexIsUnique, &indexIsPrimaryKey, &indexIsUniqueConstraint, &indexColumnName, &indexIsSystemNamed)
			if err != nil {
				return errors.WithStack(err)
			}
			indexType := "-"
			indexDef := []string{
				indexClusterType,
			}
			if indexIsUnique {
				indexDef = append(indexDef, "unique")
			}
			if indexIsPrimaryKey {
				indexType = "PRIMARY KEY"
				indexDef = append(indexDef, "part of a PRIMARY KEY constraint")
			}
			if indexIsUniqueConstraint {
				indexType = "UNIQUE"
				indexDef = append(indexDef, "part of a UNIQUE constraint")
			}
			indexDef = append(indexDef, fmt.Sprintf("[ %s ]", indexColumnName.String))

			constraint := &schema.Constraint{
				Name:    convertSystemNamed(indexName, indexIsSystemNamed),
				Type:    indexType,
				Def:     strings.Join(indexDef, ", "),
				Table:   &table.Name,
				Columns: strings.Split(indexColumnName.String, ", "),
			}
			constraints = append(constraints, constraint)
		}

		/// foreign_keys
		fkRows, err := m.db.Query(`
SELECT
  f.name AS f_name,
  object_name(f.parent_object_id) AS table_name,
  object_name(f.referenced_object_id) AS parent_table_name,
  STRING_AGG(COL_NAME(fc.parent_object_id, fc.parent_column_id), ', ') AS column_names,
  STRING_AGG(COL_NAME(fc.referenced_object_id, fc.referenced_column_id), ', ') AS parent_column_names,
  update_referential_action_desc,
  delete_referential_action_desc,
  f.is_system_named
FROM sys.foreign_keys AS f
LEFT JOIN sys.foreign_key_columns AS fc ON f.object_id = fc.constraint_object_id
WHERE f.parent_object_id = OBJECT_ID($1)
GROUP BY f.name, f.parent_object_id, f.referenced_object_id, delete_referential_action_desc, update_referential_action_desc, f.is_system_named
`, fmt.Sprintf("%s.%s", tableSchema, tableName))
		defer fkRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}
		for fkRows.Next() {
			var (
				fkName              string
				fkTableName         string
				fkParentTableName   string
				fkColumnNames       string
				fkParentColumnNames string
				fkUpdateAction      string
				fkDeleteAction      string
				fkIsSystemNamed     bool
			)
			err = fkRows.Scan(&fkName, &fkTableName, &fkParentTableName, &fkColumnNames, &fkParentColumnNames, &fkUpdateAction, &fkDeleteAction, &fkIsSystemNamed)
			if err != nil {
				return errors.WithStack(err)
			}
			fkDef := fmt.Sprintf("FOREIGN KEY(%s) REFERENCES %s(%s) ON UPDATE %s ON DELETE %s", fkColumnNames, fkParentTableName, fkParentColumnNames, fkUpdateAction, fkDeleteAction)
			constraint := &schema.Constraint{
				Name:             convertSystemNamed(fkName, fkIsSystemNamed),
				Type:             typeFk,
				Def:              fkDef,
				Table:            &table.Name,
				Columns:          strings.Split(fkColumnNames, ", "),
				ReferenceTable:   &fkParentTableName,
				ReferenceColumns: strings.Split(fkParentColumnNames, ", "),
			}
			links = append(links, relationLink{
				table:         table.Name,
				columns:       strings.Split(fkColumnNames, ", "),
				parentTable:   fkParentTableName,
				parentColumns: strings.Split(fkParentColumnNames, ", "),
			})

			constraints = append(constraints, constraint)
		}

		/// check_constraints
		checkRows, err := m.db.Query(`
SELECT name, definition, is_system_named FROM sys.check_constraints
WHERE parent_object_id = OBJECT_ID($1)
`, fmt.Sprintf("%s.%s", tableSchema, tableName))
		defer checkRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}
		for checkRows.Next() {
			var (
				checkName          string
				checkDef           string
				checkIsSystemNamed bool
			)
			err = checkRows.Scan(&checkName, &checkDef, &checkIsSystemNamed)
			if err != nil {
				return errors.WithStack(err)
			}
			constraint := &schema.Constraint{
				Name:  convertSystemNamed(checkName, checkIsSystemNamed),
				Type:  typeCheck,
				Def:   fmt.Sprintf("CHECK%s", checkDef),
				Table: &table.Name,
			}
			constraints = append(constraints, constraint)
		}

		table.Constraints = constraints

		// indexes
		indexRows, err := m.db.Query(`
SELECT
  i.name AS index_name,
  i.type_desc,
  i.is_unique,
  i.is_primary_key,
  i.is_unique_constraint,
  STRING_AGG(COL_NAME(ic.object_id, ic.column_id), ', ') WITHIN GROUP ( ORDER BY ic.key_ordinal ),
  c.is_system_named
FROM sys.indexes AS i
INNER JOIN sys.index_columns AS ic
ON i.object_id = ic.object_id AND i.index_id = ic.index_id
LEFT JOIN sys.key_constraints AS c
ON i.object_id = c.parent_object_id AND i.index_id = c.unique_index_id
WHERE i.object_id = OBJECT_ID($1)
GROUP BY i.name, i.index_id, i.type_desc, i.is_unique, i.is_primary_key, i.is_unique_constraint, c.is_system_named
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
				indexIsSytemNamed       sql.NullBool
			)
			err = indexRows.Scan(&indexName, &indexType, &indexIsUnique, &indexIsPrimaryKey, &indexIsUniqueConstraint, &indexColumnName, &indexIsSytemNamed)
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
				Name:    convertSystemNamed(indexName, indexIsSytemNamed.Bool),
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

	// relations
	relations := []*schema.Relation{}
	for _, l := range links {
		r := &schema.Relation{}
		table, err := s.FindTableByName(l.table)
		if err != nil {
			return err
		}
		r.Table = table
		for _, c := range l.columns {
			column, err := table.FindColumnByName(c)
			if err != nil {
				return err
			}
			r.Columns = append(r.Columns, column)
			column.ParentRelations = append(column.ParentRelations, r)
		}
		parentTable, err := s.FindTableByName(l.parentTable)
		if err != nil {
			return err
		}
		r.ParentTable = parentTable
		for _, c := range l.parentColumns {
			column, err := parentTable.FindColumnByName(c)
			if err != nil {
				return err
			}
			r.ParentColumns = append(r.ParentColumns, column)
			column.ChildRelations = append(column.ChildRelations, r)
		}
		relations = append(relations, r)
	}

	s.Relations = relations

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

func convertSystemNamed(name string, isSytemNamed bool) string {
	if isSytemNamed {
		return reSystemNamed.ReplaceAllString(name, "*")
	}
	return name
}
