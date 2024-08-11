package mysql

import (
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/aquasecurity/go-version/pkg/version"
	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/ddl"
	"github.com/k1LoW/tbls/dict"
	"github.com/k1LoW/tbls/drivers"
	"github.com/k1LoW/tbls/schema"
)

var reFK = regexp.MustCompile(`FOREIGN KEY \((.+)\) REFERENCES ([^\s\)]+)\s?\(([^\)]+)\)`)
var reAI = regexp.MustCompile(` AUTO_INCREMENT=[\d]+`)
var supportGeneratedColumn = true
var supportCheckConstraint = true

// Mysql struct
type Mysql struct {
	db        *sql.DB
	mariaMode bool

	// Show AUTO_INCREMENT with increment number
	showAutoIncrement bool

	// Hide the entire AUTO_INCREMENT clause
	hideAutoIncrement bool
}

func ShowAutoIcrrement() drivers.Option {
	return func(d drivers.Driver) error {
		switch d := d.(type) {
		case *Mysql:
			d.showAutoIncrement = true
		}
		return nil
	}
}

func HideAutoIcrrement() drivers.Option {
	return func(d drivers.Driver) error {
		switch d := d.(type) {
		case *Mysql:
			d.hideAutoIncrement = true
		}
		return nil
	}
}

// New return new Mysql
func New(db *sql.DB, opts ...drivers.Option) (*Mysql, error) {
	m := &Mysql{
		db: db,
	}
	for _, opt := range opts {
		err := opt(m)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

// Analyze MySQL database schema
func (m *Mysql) Analyze(s *schema.Schema) error {
	d, err := m.Info()
	if err != nil {
		return errors.WithStack(err)
	}
	s.Driver = d

	if m.mariaMode {
		verGeneratedColumn, err := version.Parse("10.2")
		if err != nil {
			return err
		}
		verCheck, err := version.Parse("10.2.1")
		if err != nil {
			return err
		}
		splitted := strings.Split(s.Driver.DatabaseVersion, "-")
		v, err := version.Parse(splitted[0])
		if err != nil {
			return err
		}
		if v.LessThan(verGeneratedColumn) {
			supportGeneratedColumn = false
		}
		if v.LessThan(verCheck) {
			supportCheckConstraint = false
		}
	} else {
		verGeneratedColumn, err := version.Parse("5.7.6")
		if err != nil {
			return err
		}
		verCheck, err := version.Parse("8.0.16")
		if err != nil {
			return err
		}
		v, err := version.Parse(s.Driver.DatabaseVersion)
		if err != nil {
			return err
		}
		if v.LessThan(verGeneratedColumn) {
			supportGeneratedColumn = false
		}
		if v.LessThan(verCheck) {
			supportCheckConstraint = false
		}
	}

	// bulk get indexes
	indexRows, err := m.db.Query(`
SELECT
s.table_name,
(CASE WHEN s.index_name='PRIMARY' AND s.non_unique=0 THEN 'PRIMARY KEY'
      WHEN s.index_name!='PRIMARY' AND s.non_unique=0 THEN 'UNIQUE KEY'
      WHEN s.non_unique=1 THEN 'KEY'
      ELSE null
  END) AS key_type,
s.index_name, GROUP_CONCAT(s.column_name ORDER BY s.seq_in_index SEPARATOR ', '), s.index_type
FROM information_schema.statistics AS s
LEFT JOIN information_schema.columns AS c ON s.table_schema = c.table_schema AND s.table_name = c.table_name AND s.column_name = c.column_name
WHERE s.table_name = c.table_name
AND s.table_schema = ?
GROUP BY s.table_name, key_type, s.table_name, s.index_name, s.index_type`, s.Name)
	if err != nil {
		return errors.WithStack(err)
	}
	defer indexRows.Close()

	tableIndexes := map[string][]*schema.Index{}
	for indexRows.Next() {
		var (
			tableName       string
			indexKeyType    string
			indexName       string
			indexColumnName string
			indexType       string
			indexDef        string
		)
		err = indexRows.Scan(&tableName, &indexKeyType, &indexName, &indexColumnName, &indexType)
		if err != nil {
			return errors.WithStack(err)
		}

		if indexKeyType == "PRIMARY KEY" {
			indexDef = fmt.Sprintf("%s (%s) USING %s", indexKeyType, indexColumnName, indexType)
		} else {
			indexDef = fmt.Sprintf("%s %s (%s) USING %s", indexKeyType, indexName, indexColumnName, indexType)
		}

		index := &schema.Index{
			Name:    indexName,
			Def:     indexDef,
			Table:   &tableName,
			Columns: strings.Split(indexColumnName, ", "),
		}
		tableIndexes[tableName] = append(tableIndexes[tableName], index)
	}

	// bulk get triggers
	triggerRows, err := m.db.Query(`
SELECT
  event_object_table,
  trigger_name,
  action_timing,
  event_manipulation,
  event_object_table,
  action_orientation,
  action_statement
FROM information_schema.triggers
WHERE event_object_schema = ?
`, s.Name)
	if err != nil {
		return errors.WithStack(err)
	}
	defer triggerRows.Close()
	tableTriggers := map[string][]*schema.Trigger{}
	for triggerRows.Next() {
		var (
			tableName                string
			triggerName              string
			triggerActionTiming      string
			triggerEventManipulation string
			triggerEventObjectTable  string
			triggerActionOrientation string
			triggerActionStatement   string
			triggerDef               string
		)
		err = triggerRows.Scan(&tableName, &triggerName, &triggerActionTiming, &triggerEventManipulation, &triggerEventObjectTable, &triggerActionOrientation, &triggerActionStatement)
		if err != nil {
			return errors.WithStack(err)
		}
		triggerDef = fmt.Sprintf("CREATE TRIGGER %s %s %s ON %s\nFOR EACH %s\n%s", triggerName, triggerActionTiming, triggerEventManipulation, triggerEventObjectTable, triggerActionOrientation, triggerActionStatement)
		trigger := &schema.Trigger{
			Name: triggerName,
			Def:  triggerDef,
		}
		tableTriggers[tableName] = append(tableTriggers[tableName], trigger)
	}

	// bulk get columns and comments
	columnStmt := `
SELECT table_name, column_name, column_default, is_nullable, column_type, column_comment, extra, generation_expression
FROM information_schema.columns
WHERE table_schema = ? ORDER BY table_name, ordinal_position`
	if !supportGeneratedColumn {
		columnStmt = `
SELECT table_name, column_name, column_default, is_nullable, column_type, column_comment, extra
FROM information_schema.columns
WHERE table_schema = ? ORDER BY table_name, ordinal_position`
	}
	columnRows, err := m.db.Query(columnStmt, s.Name)
	if err != nil {
		return errors.WithStack(err)
	}
	defer columnRows.Close()
	tableColumns := map[string][]*schema.Column{}
	for columnRows.Next() {
		var (
			tableName      string
			columnName     string
			columnDefault  sql.NullString
			isNullable     string
			columnType     string
			columnComment  sql.NullString
			extra          sql.NullString
			generationExpr sql.NullString
		)
		if supportGeneratedColumn {
			err = columnRows.Scan(&tableName, &columnName, &columnDefault, &isNullable, &columnType, &columnComment, &extra, &generationExpr)
			if err != nil {
				return errors.WithStack(err)
			}
		} else {
			err = columnRows.Scan(&tableName, &columnName, &columnDefault, &isNullable, &columnType, &columnComment, &extra)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		extraDef := extra.String
		if generationExpr.String != "" {
			switch extraDef {
			case "VIRTUAL GENERATED":
				extraDef = fmt.Sprintf("GENERATED ALWAYS AS %s VIRTUAL", generationExpr.String)
			case "STORED GENERATED":
				extraDef = fmt.Sprintf("GENERATED ALWAYS AS %s STORED", generationExpr.String)
			default:
				extraDef = fmt.Sprintf("%s:%s", extraDef, generationExpr.String)
			}
		}
		column := &schema.Column{
			Name:     columnName,
			Type:     columnType,
			Nullable: convertColumnNullable(isNullable),
			Default:  columnDefault,
			Comment:  columnComment.String,
			ExtraDef: extraDef,
		}

		tableColumns[tableName] = append(tableColumns[tableName], column)
	}

	// tables and comments
	tableRows, err := m.db.Query(m.queryForTables(), s.Name)
	if err != nil {
		return errors.WithStack(err)
	}
	defer tableRows.Close()

	relations := []*schema.Relation{}

	tableMap := map[string]*schema.Table{}
	tableOrderMap := map[string]int{}
	tableOrder := 0
	tables := []*schema.Table{}
	for tableRows.Next() {
		var (
			tableName    string
			tableType    string
			tableComment string
		)
		err := tableRows.Scan(&tableName, &tableType, &tableComment)
		if err != nil {
			return errors.WithStack(err)
		}
		table := &schema.Table{
			Name:    tableName,
			Type:    tableType,
			Comment: tableComment,
		}

		// table definition
		if tableType == "BASE TABLE" {
			tableDefRows, err := m.db.Query(fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName))
			if err != nil {
				return errors.WithStack(err)
			}
			defer tableDefRows.Close()
			for tableDefRows.Next() {
				var (
					tableName string
					tableDef  string
				)
				err := tableDefRows.Scan(&tableName, &tableDef)
				if err != nil {
					return errors.WithStack(err)
				}

				switch {
				case m.showAutoIncrement:
					table.Def = tableDef
				case m.hideAutoIncrement:
					table.Def = reAI.ReplaceAllLiteralString(tableDef, "")
				default:
					table.Def = reAI.ReplaceAllLiteralString(tableDef, " AUTO_INCREMENT=[Redacted by tbls]")
				}
			}
		}

		// view definition
		if tableType == "VIEW" {
			viewDefRows, err := m.db.Query(`
SELECT view_definition FROM information_schema.views
WHERE table_schema = ?
AND table_name = ?;
		`, s.Name, tableName)
			if err != nil {
				return errors.WithStack(err)
			}
			defer viewDefRows.Close()
			for viewDefRows.Next() {
				var tableDef string
				err := viewDefRows.Scan(&tableDef)
				if err != nil {
					return errors.WithStack(err)
				}
				table.Def = fmt.Sprintf("CREATE VIEW %s AS (%s)", tableName, tableDef)
			}
		}

		// indexes
		table.Indexes = tableIndexes[table.Name]

		// triggers
		table.Triggers = tableTriggers[table.Name]

		// columns and comments
		table.Columns = tableColumns[table.Name]

		tables = append(tables, table)
		tableMap[table.Name] = table
		tableOrderMap[table.Name] = tableOrder
		tableOrder++
	}

	// bulk get constraints (PRIMARY KEY, UNIQUE, FOREIGN KEY)
	constraintRows, err := m.db.Query(`
SELECT
  kcu.table_name,
  kcu.constraint_name,
  sub.costraint_type,
  GROUP_CONCAT(kcu.column_name ORDER BY kcu.ordinal_position, position_in_unique_constraint SEPARATOR ', ') AS column_name,
  kcu.referenced_table_name,
  GROUP_CONCAT(kcu.referenced_column_name ORDER BY kcu.ordinal_position, position_in_unique_constraint SEPARATOR ', ') AS referenced_column_name
FROM information_schema.key_column_usage AS kcu
LEFT JOIN information_schema.columns AS c ON kcu.table_schema = c.table_schema AND kcu.table_name = c.table_name AND kcu.column_name = c.column_name
INNER JOIN
  (
   SELECT
   kcu.table_schema,
   kcu.table_name,
   kcu.constraint_name,
   kcu.column_name,
   kcu.referenced_table_name,
   (CASE WHEN kcu.referenced_table_name IS NOT NULL THEN 'FOREIGN KEY'
        WHEN c.column_key = 'PRI' AND kcu.constraint_name = 'PRIMARY' THEN 'PRIMARY KEY'
        WHEN c.column_key = 'PRI' AND kcu.constraint_name != 'PRIMARY' THEN 'UNIQUE'
        WHEN c.column_key = 'UNI' THEN 'UNIQUE'
        WHEN c.column_key = 'MUL' THEN 'UNIQUE'
        ELSE 'UNKNOWN'
   END) AS costraint_type
   FROM information_schema.key_column_usage AS kcu
   LEFT JOIN information_schema.columns AS c ON kcu.table_schema = c.table_schema AND kcu.table_name = c.table_name AND kcu.column_name = c.column_name
   WHERE kcu.ordinal_position = 1
  ) AS sub
ON kcu.constraint_name = sub.constraint_name
  AND kcu.table_schema = sub.table_schema
  AND kcu.table_name = sub.table_name
  AND (kcu.referenced_table_name = sub.referenced_table_name OR (kcu.referenced_table_name IS NULL AND sub.referenced_table_name IS NULL))
WHERE kcu.table_schema= ?
GROUP BY kcu.table_name, kcu.constraint_name, sub.costraint_type, kcu.referenced_table_name`, s.Name)
	if err != nil {
		return errors.WithStack(err)
	}
	defer constraintRows.Close()

	for constraintRows.Next() {
		var (
			tableName               string
			constraintName          string
			constraintType          string
			constraintColumnName    string
			constraintRefTableName  sql.NullString
			constraintRefColumnName sql.NullString
			constraintDef           string
		)
		err = constraintRows.Scan(&tableName, &constraintName, &constraintType, &constraintColumnName, &constraintRefTableName, &constraintRefColumnName)
		if err != nil {
			return errors.WithStack(err)
		}

		switch constraintType {
		case "PRIMARY KEY":
			constraintDef = fmt.Sprintf("PRIMARY KEY (%s)", constraintColumnName)
		case "UNIQUE":
			constraintDef = fmt.Sprintf("UNIQUE KEY %s (%s)", constraintName, constraintColumnName)
		case "FOREIGN KEY":
			constraintType = schema.TypeFK
			constraintDef = fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)", constraintColumnName, constraintRefTableName.String, constraintRefColumnName.String)
			relation := &schema.Relation{
				Table: tableMap[tableName],
				Def:   constraintDef,
			}
			relations = append(relations, relation)
		case "UNKNOWN":
			constraintDef = fmt.Sprintf("UNKNOWN CONSTRAINT (%s) (%s) (%s)", constraintColumnName, constraintRefTableName.String, constraintRefColumnName.String)
		}

		constraint := &schema.Constraint{
			Name:    constraintName,
			Type:    constraintType,
			Def:     constraintDef,
			Table:   &tableName,
			Columns: strings.Split(constraintColumnName, ", "),
		}
		if constraintRefTableName.String != "" {
			constraint.ReferencedTable = &constraintRefTableName.String
			constraint.ReferencedColumns = strings.Split(constraintRefColumnName.String, ", ")
		}

		tableMap[tableName].Constraints = append(tableMap[tableName].Constraints, constraint)
	}

	// bulk get constraints (CHECK)
	if supportCheckConstraint {
		constraintRows, err := m.db.Query(`
SELECT
  t.table_name,
  c.constraint_name,
  c.check_clause
FROM information_schema.check_constraints AS c
JOIN information_schema.table_constraints AS t ON c.constraint_schema = t.constraint_schema AND c.constraint_name = t.constraint_name
WHERE t.table_schema = ?
`, s.Name)
		if err != nil {
			return errors.WithStack(err)
		}
		defer constraintRows.Close()

		for constraintRows.Next() {
			var (
				tableName      string
				constraintName string
				checkClause    string
			)
			err = constraintRows.Scan(&tableName, &constraintName, &checkClause)
			if err != nil {
				return errors.WithStack(err)
			}
			constraint := &schema.Constraint{
				Name:  constraintName,
				Type:  "CHECK",
				Def:   fmt.Sprintf("CHECK (%s)", checkClause),
				Table: &tableName,
			}
			tableMap[tableName].Constraints = append(tableMap[tableName].Constraints, constraint)
		}
	}

	functions, err := m.getFunctions()
	if err != nil {
		return err
	}
	s.Functions = functions

	s.Tables = tables

	// Relations
	sort.SliceStable(relations, func(i, j int) bool {
		return tableOrderMap[relations[i].Table.Name] < tableOrderMap[relations[j].Table.Name]
	})
	for _, r := range relations {
		strColumns, strParentTable, strParentColumns, err := parseFK(r.Def)
		if err != nil {
			return err
		}
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

	// referenced tables of view
	for _, t := range s.Tables {
		if t.Type != "VIEW" {
			continue
		}
		for _, rts := range ddl.ParseReferencedTables(t.Def) {
			rt, err := s.FindTableByName(strings.TrimPrefix(rts, fmt.Sprintf("%s.", s.Name)))
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

const queryFunctions = `SELECT r.routine_schema as database_name,
r.routine_name,
r.routine_type AS type,
r.data_type AS return_type,
GROUP_CONCAT(CONCAT(p.parameter_name, ' ', p.data_type) SEPARATOR '; ') AS parameter
FROM information_schema.routines r
LEFT JOIN information_schema.parameters p
	 ON p.specific_schema = r.routine_schema
	 AND p.specific_name = r.specific_name
WHERE routine_schema NOT IN ('sys', 'information_schema', 'mysql', 'performance_schema')
GROUP BY r.routine_schema, r.routine_name, r.routine_type, r.data_type, r.routine_definition`

func (m *Mysql) getFunctions() ([]*schema.Function, error) {
	functions := []*schema.Function{}
	functionsResult, err := m.db.Query(queryFunctions)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer functionsResult.Close()

	for functionsResult.Next() {
		var (
			databaseName string
			name         string
			typeValue    string
			returnType   string
			arguments    sql.NullString
		)
		err := functionsResult.Scan(&databaseName, &name, &typeValue, &returnType, &arguments)
		if err != nil {
			return functions, errors.WithStack(err)
		}
		subroutine := &schema.Function{
			Name:       name,
			Type:       typeValue,
			ReturnType: returnType,
			Arguments:  arguments.String,
		}

		functions = append(functions, subroutine)
	}
	return functions, nil
}

// Info return schema.Driver
func (m *Mysql) Info() (*schema.Driver, error) {
	var v string
	row := m.db.QueryRow(`SELECT version();`)
	err := row.Scan(&v)
	if err != nil {
		return nil, err
	}

	name := "mysql"
	if m.mariaMode {
		name = "mariadb"
	}

	dct := dict.New()
	dct.Merge(map[string]string{
		"Functions": "Stored procedures and functions",
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

// EnableMariaMode enable mariaMode
func (m *Mysql) EnableMariaMode() {
	m.mariaMode = true
}

func (m *Mysql) queryForTables() string {
	if m.mariaMode {
		return `
SELECT table_name, table_type, table_comment FROM information_schema.tables WHERE table_schema = ? ORDER BY table_name;`
	}
	return `
SELECT table_name, table_type, table_comment FROM information_schema.tables WHERE table_schema = ?;`
}

func convertColumnNullable(str string) bool {
	return str != "NO"
}

func parseFK(def string) (_ []string, _ string, _ []string, err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	result := reFK.FindAllStringSubmatch(def, -1)
	if len(result) < 1 || len(result[0]) < 4 {
		return nil, "", nil, fmt.Errorf("can not parse foreign key: %s", def)
	}
	strColumns := strings.Split(result[0][1], ", ")
	strParentTable := strings.Trim(result[0][2], `"`)
	strParentColumns := strings.Split(result[0][3], ", ")
	return strColumns, strParentTable, strParentColumns, nil
}
