package mssql

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/k1LoW/tbls/ddl"
	"github.com/k1LoW/tbls/dict"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

var defaultSchemaName = "dbo"
var typeFk = schema.TypeFK
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

// New ...
func New(db *sql.DB) *Mssql {
	return &Mssql{
		db: db,
	}
}

func (m *Mssql) Analyze(s *schema.Schema) error {
	d, err := m.Info()
	if err != nil {
		return errors.WithStack(err)
	}
	s.Driver = d

	// tables and comments
	tableRows, err := m.db.Query(`
SELECT schema_name(schema_id) AS table_schema, o.name, o.object_id, o.type, cast(e.value as NVARCHAR(MAX)) AS table_comment
FROM sys.objects AS o
LEFT JOIN sys.extended_properties AS e ON
e.major_id = o.object_id AND e.name = 'MS_Description' AND e.minor_id = 0
WHERE type IN ('U', 'V')  ORDER BY OBJECT_ID
`)
	if err != nil {
		return errors.WithStack(err)
	}
	defer tableRows.Close()

	tables := []*schema.Table{}
	links := []relationLink{}

	for tableRows.Next() {
		var (
			tableSchema  string
			tableName    string
			tableOid     string
			tableType    string
			tableComment sql.NullString
		)
		err := tableRows.Scan(&tableSchema, &tableName, &tableOid, &tableType, &tableComment)
		if err != nil {
			return errors.WithStack(err)
		}
		tableType = convertTableType(tableType)

		name := tableName
		if tableSchema != defaultSchemaName {
			name = fmt.Sprintf("%s.%s", tableSchema, tableName)
		}

		table := &schema.Table{
			Name:    name,
			Type:    tableType,
			Comment: tableComment.String,
		}

		// view definition
		if tableType == "VIEW" {
			viewDefRows, err := m.db.Query(`
SELECT definition FROM sys.sql_modules WHERE object_id = @p1
`, tableOid)
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
				table.Def = tableDef.String
			}
		}

		// columns and comments
		columnRows, err := m.db.Query(`
SELECT
  c.name,
  t.name AS type,
  c.max_length,
  c.is_nullable,
  c.is_identity,
  object_definition(c.default_object_id),
  CAST(e.value AS NVARCHAR(MAX)) AS column_comment
FROM sys.columns AS c
LEFT JOIN sys.types AS t ON c.system_type_id = t.system_type_id
LEFT JOIN sys.extended_properties AS e ON
e.major_id = c.object_id AND e.name = 'MS_Description' AND e.minor_id = c.column_id
WHERE c.object_id = @p1
and t.name != 'sysname'
ORDER BY c.column_id
`, tableOid)
		if err != nil {
			return errors.WithStack(err)
		}
		defer columnRows.Close()

		columns := []*schema.Column{}
		for columnRows.Next() {
			var (
				columnName    string
				dataType      string
				maxLength     int
				isNullable    bool
				isIdentity    bool
				columnDefault sql.NullString
				columnComment sql.NullString
			)
			err = columnRows.Scan(&columnName, &dataType, &maxLength, &isNullable, &isIdentity, &columnDefault, &columnComment)
			if err != nil {
				return errors.WithStack(err)
			}
			column := &schema.Column{
				Name:     columnName,
				Type:     convertColumnType(dataType, maxLength),
				Nullable: isNullable,
				Default:  columnDefault,
				Comment:  columnComment.String,
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
WHERE i.object_id = object_id(@p1)
GROUP BY c.name, i.index_id, i.type_desc, i.is_unique, i.is_primary_key, i.is_unique_constraint, c.is_system_named
ORDER BY i.index_id
`, fmt.Sprintf("%s.%s", tableSchema, tableName))
		if err != nil {
			return errors.WithStack(err)
		}
		defer keyRows.Close()
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
WHERE f.parent_object_id = object_id(@p1)
GROUP BY f.name, f.parent_object_id, f.referenced_object_id, delete_referential_action_desc, update_referential_action_desc, f.is_system_named
`, fmt.Sprintf("%s.%s", tableSchema, tableName))
		if err != nil {
			return errors.WithStack(err)
		}
		defer fkRows.Close()
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
			fkDef := fmt.Sprintf("FOREIGN KEY(%s) REFERENCES %s(%s) ON UPDATE %s ON DELETE %s", fkColumnNames, fkParentTableName, fkParentColumnNames, fkUpdateAction, fkDeleteAction) // #nosec
			constraint := &schema.Constraint{
				Name:              convertSystemNamed(fkName, fkIsSystemNamed),
				Type:              typeFk,
				Def:               fkDef,
				Table:             &table.Name,
				Columns:           strings.Split(fkColumnNames, ", "),
				ReferencedTable:   &fkParentTableName,
				ReferencedColumns: strings.Split(fkParentColumnNames, ", "),
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
SELECT name, definition, is_system_named
FROM sys.check_constraints
WHERE parent_object_id = object_id(@p1)
`, fmt.Sprintf("%s.%s", tableSchema, tableName))
		if err != nil {
			return errors.WithStack(err)
		}
		defer checkRows.Close()
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

		// triggers
		triggerRows, err := m.db.Query(`
SELECT name, definition
FROM sys.triggers AS t
INNER JOIN sys.sql_modules AS sm
ON sm.object_id = t.object_id
WHERE type = 'TR'
AND parent_id = object_id(@p1)
`, fmt.Sprintf("%s.%s", tableSchema, tableName))
		if err != nil {
			return errors.WithStack(err)
		}
		defer triggerRows.Close()

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
WHERE i.object_id = object_id(@p1)
GROUP BY i.name, i.index_id, i.type_desc, i.is_unique, i.is_primary_key, i.is_unique_constraint, c.is_system_named
ORDER BY i.index_id
`, fmt.Sprintf("%s.%s", tableSchema, tableName))
		if err != nil {
			return errors.WithStack(err)
		}
		defer indexRows.Close()
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

	functions, err := m.getFunctions()
	if err != nil {
		return err
	}
	s.Functions = functions

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

	// referenced tables of view
	for _, t := range s.Tables {
		if t.Type != "VIEW" {
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

const query = `select schema_name(obj.schema_id) as schema_name,
	obj.name as name,
	case type
		when 'FN' then 'SQL scalar function'
		when 'TF' then 'SQL table-valued-function'
		when 'IF' then 'SQL inline table-valued function'
		when 'P' then 'SQL Stored Procedure'
		when 'X' then 'Extended stored procedure'
	end as type,
	TYPE_NAME(ret.user_type_id) as return_type,
	substring(par.parameters, 0, len(par.parameters)) as parameters
from sys.objects obj
join sys.sql_modules mod
on mod.object_id = obj.object_id
cross apply (select p.name + ' ' + TYPE_NAME(p.user_type_id) + ', ' 
			from sys.parameters p
			where p.object_id = obj.object_id 
						and p.parameter_id != 0 
		 for xml path ('') ) par (parameters)
left join sys.parameters ret
	 on obj.object_id = ret.object_id
	 and ret.parameter_id = 0
where obj.type in ('FN', 'TF', 'IF', 'P', 'X')
order by schema_name,
	name;`

func (m *Mssql) getFunctions() ([]*schema.Function, error) {
	functions := []*schema.Function{}
	functionsResult, err := m.db.Query(query)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer functionsResult.Close()

	for functionsResult.Next() {
		var (
			schemaName string
			name       string
			typeValue  string
			returnType sql.NullString
			arguments  sql.NullString
		)
		err := functionsResult.Scan(&schemaName, &name, &typeValue, &returnType, &arguments)
		if err != nil {
			return functions, errors.WithStack(err)
		}
		function := &schema.Function{
			Name:       fullTableName(schemaName, name),
			Type:       typeValue,
			ReturnType: returnType.String,
			Arguments:  arguments.String,
		}

		functions = append(functions, function)
	}
	return functions, nil
}

func fullTableName(owner string, tableName string) string {
	return fmt.Sprintf("%s.%s", owner, tableName)
}

func (m *Mssql) Info() (*schema.Driver, error) {
	var v string
	row := m.db.QueryRow(`SELECT @@VERSION`)
	err := row.Scan(&v)
	if err != nil {
		return nil, err
	}

	dct := dict.New()
	dct.Merge(map[string]string{
		"Functions": "Stored procedures and functions",
	})

	d := &schema.Driver{
		Name:            "sqlserver",
		DatabaseVersion: v,
		Meta: &schema.DriverMeta{
			Dict: &dct,
		},
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

func convertColumnType(t string, maxLength int) string {
	switch t {
	case "varchar":
		var len string = strconv.Itoa(maxLength)
		if maxLength == -1 {
			len = "MAX"
		}
		return fmt.Sprintf("varchar(%s)", len)
	case "nvarchar":
		//nvarchar length is 2 byte, return character length
		var len string = strconv.Itoa(maxLength / 2)
		if maxLength == -1 {
			len = "MAX"
		}
		return fmt.Sprintf("nvarchar(%s)", len)
	case "varbinary":
		var len string = strconv.Itoa(maxLength)
		if maxLength == -1 {
			len = "MAX"
		}
		return fmt.Sprintf("varbinary(%s)", len)
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
