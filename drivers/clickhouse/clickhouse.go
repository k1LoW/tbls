package clickhouse

import (
	"database/sql"
	"fmt"
	"regexp"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/dict"
	"github.com/k1LoW/tbls/schema"
	"github.com/samber/lo"
)

var shadowTableRe = regexp.MustCompile(`^\.inner_id\.`)

// ClickHouse struct
type ClickHouse struct {
	db *sql.DB
}

// New return new Postgres
func New(db *sql.DB) *ClickHouse {
	return &ClickHouse{
		db: db,
	}
}

// Analyze PostgreSQL database schema
func (ch *ClickHouse) Analyze(s *schema.Schema) error {
	d, err := ch.Info()
	if err != nil {
		return errors.WithStack(err)
	}
	s.Driver = d

	// tables
	strTableDependencies := make(map[string][]string)
	tablePartitionKeys := make(map[string]*schema.Constraint)
	tableSortingKeys := make(map[string]*schema.Constraint)
	tablePrimaryKeys := make(map[string]*schema.Constraint)
	tableSamplingKeys := make(map[string]*schema.Constraint)
	var filtered []string

	tableRows, err := ch.db.Query(`
SELECT
    uuid,
    name,
    engine,
    partition_key,
    sorting_key,
    sampling_key,
    primary_key,
    create_table_query,
    comment,
    dependencies_database,
    dependencies_table
FROM system.tables
WHERE database = ?
`, s.Name)
	if err != nil {
		return errors.WithStack(err)
	}
	defer tableRows.Close()

	for tableRows.Next() {
		var (
			tableUuid                 string
			tableName                 string
			tableType                 string
			tablePartitionKey         string
			tableSortingKey           string
			tableSamplingKey          string
			tablePrimaryKey           string
			tableDef                  string
			tableComment              string
			tableDependenciesDatabase []string
			tableDependenciesTable    []string
		)
		err := tableRows.Scan(&tableUuid, &tableName, &tableType, &tablePartitionKey, &tableSortingKey, &tableSamplingKey, &tablePrimaryKey, &tableDef, &tableComment, &tableDependenciesDatabase, &tableDependenciesTable)
		if err != nil {
			return errors.WithStack(err)
		}

		table := &schema.Table{
			Name:    tableName,
			Type:    tableType,
			Def:     tableDef,
			Comment: tableComment,
		}

		if shadowTableRe.MatchString(tableName) {
			filtered = append(filtered, tableName)
			continue
		}

		s.Tables = append(s.Tables, table)

		if tablePartitionKey != "" {
			tablePartitionKeys[tableName] = &schema.Constraint{
				Name:  "partition key",
				Table: &tableName,
				Def:   fmt.Sprintf("PARTITION BY (%s)", tablePartitionKey),
				Type:  "PARTITION KEY",
			}
		}
		if tableSortingKey != "" {
			tableSortingKeys[tableName] = &schema.Constraint{
				Name:  "sorting key",
				Table: &tableName,
				Def:   fmt.Sprintf("ORDER BY (%s)", tableSortingKey),
				Type:  "SORTING KEY",
			}
		}
		if tablePrimaryKey != "" {
			tablePrimaryKeys[tableName] = &schema.Constraint{
				Name:  "primary key",
				Table: &tableName,
				Def:   fmt.Sprintf("PRIMARY KEY (%s)", tablePrimaryKey),
				Type:  "PRIMARY KEY",
			}
		}
		if tableSamplingKey != "" {
			tableSamplingKeys[tableName] = &schema.Constraint{
				Name:  "sampling key",
				Table: &tableName,
				Def:   fmt.Sprintf("SAMPLE BY (%s)", tableSamplingKey),
				Type:  "SAMPLING KEY",
			}
		}

		strTableDependencies[tableName] = tableDependenciesTable
	}

	// referenced tables (from materialized views)
	for tableName, dependencies := range strTableDependencies {
		targetTable, err := s.FindTableByName(tableName)
		if err != nil {
			return errors.WithStack(err)
		}

		for _, dependency := range dependencies {
			table, err := s.FindTableByName(dependency)
			if err != nil {
				return errors.WithStack(err)
			}

			table.ReferencedTables = append(table.ReferencedTables, targetTable)
		}
	}

	// columns
	columnRows, err := ch.db.Query(`
SELECT
    table,
    name,
    type,
    default_kind,
    default_expression,
    is_in_partition_key,
    is_in_sorting_key,
    is_in_primary_key,
    is_in_sampling_key,
    comment
FROM system.columns
WHERE database = ?
ORDER BY table
`, s.Name)
	if err != nil {
		return errors.WithStack(err)
	}
	defer columnRows.Close()
	for columnRows.Next() {
		var (
			colTable             string
			colName              string
			colType              string
			colDefaultKind       string
			colDefaultExpression string
			colIsInPartitionKey  bool
			colIsInSortingKey    bool
			colIsInPrimaryKey    bool
			colIsInSamplingKey   bool
			colComment           string
		)
		err := columnRows.Scan(&colTable, &colName, &colType, &colDefaultKind, &colDefaultExpression, &colIsInPartitionKey, &colIsInSortingKey, &colIsInPrimaryKey, &colIsInSamplingKey, &colComment)
		if err != nil {
			return errors.WithStack(err)
		}

		columnDefault := sql.NullString{}
		if colDefaultKind != "" && colDefaultExpression != "" {
			columnDefault.String = fmt.Sprintf("%s %s", colDefaultKind, colDefaultExpression)
			columnDefault.Valid = true
		}

		column := &schema.Column{
			Name:     colName,
			Type:     colType,
			Comment:  colComment,
			Default:  columnDefault,
			Nullable: false,
		}

		if lo.Contains(filtered, colTable) {
			continue
		}

		table, err := s.FindTableByName(colTable)
		if err != nil {
			return errors.WithStack(err)
		}
		table.Columns = append(table.Columns, column)

		if colIsInPartitionKey {
			if constraint, ok := tablePartitionKeys[colTable]; ok {
				constraint.Columns = append(constraint.Columns, colName)
			}
		}
		if colIsInSortingKey {
			if constraint, ok := tableSortingKeys[colTable]; ok {
				constraint.Columns = append(constraint.Columns, colName)
			}
		}
		if colIsInPrimaryKey {
			if constraint, ok := tablePrimaryKeys[colTable]; ok {
				constraint.Columns = append(constraint.Columns, colName)
			}
		}
		if colIsInSamplingKey {
			if constraint, ok := tableSamplingKeys[colTable]; ok {
				constraint.Columns = append(constraint.Columns, colName)
			}
		}
	}

	// special constraints (partition, sorting, primary, sampling keys)
	for tableName, constraint := range tablePartitionKeys {
		table, err := s.FindTableByName(tableName)
		if err != nil {
			return errors.WithStack(err)
		}
		table.Constraints = append(table.Constraints, constraint)
	}
	for tableName, constraint := range tableSortingKeys {
		table, err := s.FindTableByName(tableName)
		if err != nil {
			return errors.WithStack(err)
		}
		table.Constraints = append(table.Constraints, constraint)
	}
	for tableName, constraint := range tablePrimaryKeys {
		table, err := s.FindTableByName(tableName)
		if err != nil {
			return errors.WithStack(err)
		}
		table.Constraints = append(table.Constraints, constraint)
	}
	for tableName, constraint := range tableSamplingKeys {
		table, err := s.FindTableByName(tableName)
		if err != nil {
			return errors.WithStack(err)
		}
		table.Constraints = append(table.Constraints, constraint)
	}

	// indices
	indexRows, err := ch.db.Query(`
SELECT
    table,
    name,
    type_full,
    expr
FROM system.data_skipping_indices
WHERE database = ?
ORDER BY table
`, s.Name)
	if err != nil {
		return errors.WithStack(err)
	}
	defer indexRows.Close()
	for indexRows.Next() {
		var (
			idxTable      string
			idxName       string
			idxType       string
			idxExpression string
		)
		err := indexRows.Scan(&idxTable, &idxName, &idxType, &idxExpression)
		if err != nil {
			return errors.WithStack(err)
		}

		if lo.Contains(filtered, idxTable) {
			continue
		}

		table, err := s.FindTableByName(idxTable)
		if err != nil {
			return errors.WithStack(err)
		}

		index := &schema.Index{
			Name:    idxName,
			Def:     idxType,
			Table:   &idxTable,
			Columns: []string{idxExpression}, // TODO: parse expression and split accordingly
		}

		table.Indexes = append(table.Indexes, index)
	}

	// functions
	functionRows, err := ch.db.Query(`
SELECT
    name,
    create_query,
    arguments,
    returned_value
FROM system.functions
WHERE origin = 'SQLUserDefined'
`)
	if err != nil {
		return errors.WithStack(err)
	}
	defer functionRows.Close()

	for functionRows.Next() {
		var (
			funcName          string
			funcDef           string
			funcArguments     string
			funcReturnedValue string
		)
		err := functionRows.Scan(&funcName, &funcDef, &funcArguments, &funcReturnedValue)
		if err != nil {
			return errors.WithStack(err)
		}

		function := &schema.Function{
			Name:       funcName,
			Arguments:  funcArguments,
			ReturnType: funcReturnedValue,
		}

		s.Functions = append(s.Functions, function)
	}

	//relations := []*schema.Relation{}
	return nil
}

// Info return schema.Driver
func (ch *ClickHouse) Info() (*schema.Driver, error) {
	var v string
	row := ch.db.QueryRow(`SELECT version();`)
	err := row.Scan(&v)
	if err != nil {
		return nil, err
	}

	name := "clickhouse"

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
