package databricks

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/schema"

	_ "github.com/databricks/databricks-sql-go"
)

type Databricks struct {
	db *sql.DB
}

func New(db *sql.DB) *Databricks {
	return &Databricks{
		db: db,
	}
}

func (dbx *Databricks) Analyze(s *schema.Schema) error {
	d, err := dbx.Info()
	if err != nil {
		return errors.WithStack(err)
	}
	s.Driver = d

	currentCatalog, currentSchema, err := dbx.getCurrentContext()
	if err != nil {
		return errors.WithStack(err)
	}

	s.Name = fmt.Sprintf("%s.%s", currentCatalog, currentSchema)

	tables, err := dbx.getTables(currentCatalog, currentSchema)
	if err != nil {
		return errors.WithStack(err)
	}

	columnsByTable, err := dbx.getAllColumns(currentCatalog, currentSchema)
	if err != nil {
		return errors.WithStack(err)
	}

	constraintsByTable, err := dbx.getAllConstraints(currentCatalog, currentSchema)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, table := range tables {
		if columns, exists := columnsByTable[table.Name]; exists {
			table.Columns = columns
		}

		if constraints, exists := constraintsByTable[table.Name]; exists {
			table.Constraints = constraints
		}
	}

	s.Tables = tables

	relations, err := dbx.getRelations(currentCatalog, currentSchema, tables)
	if err != nil {
		return errors.WithStack(err)
	}
	s.Relations = relations

	return nil
}

func (dbx *Databricks) getCurrentContext() (string, string, error) {
	var catalog, schema string

	catRow := dbx.db.QueryRow(`SELECT current_catalog()`)
	if err := catRow.Scan(&catalog); err != nil {
		return "", "", errors.WithStack(err)
	}

	schemaRow := dbx.db.QueryRow(`SELECT current_schema()`)
	if err := schemaRow.Scan(&schema); err != nil {
		return "", "", errors.WithStack(err)
	}

	return catalog, schema, nil
}

func (dbx *Databricks) getTables(catalog, schemaName string) ([]*schema.Table, error) {
	query := `
		SELECT 
			table_name, 
			table_type,
			COALESCE(comment, '') as table_comment
		FROM system.information_schema.tables 
		WHERE table_catalog = ? AND table_schema = ?
		ORDER BY table_name`

	rows, err := dbx.db.Query(query, catalog, schemaName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	var tables []*schema.Table
	for rows.Next() {
		var tableName, tableType, tableComment string
		if err := rows.Scan(&tableName, &tableType, &tableComment); err != nil {
			return nil, errors.WithStack(err)
		}

		table := &schema.Table{
			Name:    tableName,
			Type:    strings.ToUpper(tableType),
			Comment: tableComment,
		}

		if strings.ToUpper(tableType) == "VIEW" {
			viewDef, err := dbx.getViewDefinition(catalog, schemaName, tableName)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			table.Def = viewDef
		}

		tables = append(tables, table)
	}

	return tables, nil
}

func (dbx *Databricks) getAllColumns(catalog, schemaName string) (map[string][]*schema.Column, error) {
	query := `
		SELECT 
			table_name,
			column_name,
			data_type,
			is_nullable,
			column_default,
			COALESCE(comment, '') as column_comment
		FROM system.information_schema.columns 
		WHERE table_catalog = ? AND table_schema = ?
		    AND table_name IN (
		        SELECT table_name 
		        FROM system.information_schema.tables 
		        WHERE table_catalog = ? AND table_schema = ?
		    )
		ORDER BY table_name, ordinal_position`

	rows, err := dbx.db.Query(query, catalog, schemaName, catalog, schemaName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	columnsByTable := make(map[string][]*schema.Column)
	for rows.Next() {
		var tableName, columnName, dataType, isNullable string
		var columnDefault, columnComment sql.NullString

		if err := rows.Scan(&tableName, &columnName, &dataType, &isNullable, &columnDefault, &columnComment); err != nil {
			return nil, errors.WithStack(err)
		}

		column := &schema.Column{
			Name:     columnName,
			Type:     dataType,
			Nullable: strings.ToUpper(isNullable) == "YES",
			Default:  columnDefault,
			Comment:  columnComment.String,
		}

		columnsByTable[tableName] = append(columnsByTable[tableName], column)
	}

	return columnsByTable, nil
}

func (dbx *Databricks) getAllConstraints(catalog, schemaName string) (map[string][]*schema.Constraint, error) {
	query := `
		SELECT 
			tc.table_name,
			tc.constraint_name,
			tc.constraint_type,
			COALESCE(COLLECT_LIST(kcu.column_name), ARRAY()) as constraint_columns,
			COALESCE(MAX(kcu2.table_name), '') as referenced_table_name,
			COALESCE(COLLECT_LIST(kcu2.column_name), ARRAY()) as referenced_columns
		FROM system.information_schema.table_constraints tc
		LEFT JOIN system.information_schema.key_column_usage kcu
			ON tc.constraint_catalog = kcu.constraint_catalog
			AND tc.constraint_schema = kcu.constraint_schema
			AND tc.constraint_name = kcu.constraint_name
			AND tc.table_name = kcu.table_name
		LEFT JOIN system.information_schema.referential_constraints rc
			ON tc.constraint_catalog = rc.constraint_catalog
			AND tc.constraint_schema = rc.constraint_schema
			AND tc.constraint_name = rc.constraint_name
		LEFT JOIN system.information_schema.key_column_usage kcu2
			ON rc.unique_constraint_catalog = kcu2.constraint_catalog
			AND rc.unique_constraint_schema = kcu2.constraint_schema
			AND rc.unique_constraint_name = kcu2.constraint_name
			AND kcu.position_in_unique_constraint = kcu2.ordinal_position
		WHERE tc.table_catalog = ? AND tc.table_schema = ?
		    AND tc.table_name IN (
		        SELECT table_name 
		        FROM system.information_schema.tables 
		        WHERE table_catalog = ? AND table_schema = ?
		    )
		GROUP BY tc.table_name, tc.constraint_name, tc.constraint_type
		ORDER BY tc.table_name, tc.constraint_name`

	rows, err := dbx.db.Query(query, catalog, schemaName, catalog, schemaName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	constraintsByTable := make(map[string][]*schema.Constraint)
	for rows.Next() {
		var tableName, constraintName, constraintType, referencedTableName string
		var constraintColumnsStr, referencedColumnsStr string

		if err := rows.Scan(&tableName, &constraintName, &constraintType, &constraintColumnsStr, &referencedTableName, &referencedColumnsStr); err != nil {
			return nil, errors.WithStack(err)
		}

		constraintColumns := dbx.parseArrayString(constraintColumnsStr)
		referencedColumns := dbx.parseArrayString(referencedColumnsStr)

		def := dbx.buildConstraintDefinition(constraintType, constraintColumns, referencedTableName, referencedColumns)

		constraint := &schema.Constraint{
			Name:    constraintName,
			Type:    strings.ToUpper(constraintType),
			Table:   &tableName,
			Def:     def,
			Columns: constraintColumns,
		}

		if strings.ToUpper(constraintType) == "FOREIGN KEY" && referencedTableName != "" {
			constraint.ReferencedTable = &referencedTableName
			constraint.ReferencedColumns = referencedColumns
		}

		constraintsByTable[tableName] = append(constraintsByTable[tableName], constraint)
	}

	return constraintsByTable, nil
}

func (dbx *Databricks) parseArrayString(arrayStr string) []string {
	// Trim whitespace from input
	arrayStr = strings.TrimSpace(arrayStr)

	if arrayStr == "" || arrayStr == "[]" {
		return []string{}
	}

	// Remove brackets and split by comma
	arrayStr = strings.TrimPrefix(arrayStr, "[")
	arrayStr = strings.TrimSuffix(arrayStr, "]")

	if arrayStr == "" {
		return []string{}
	}

	// Split and clean up each element
	parts := strings.Split(arrayStr, ",")
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Remove quotes if present
		if len(part) >= 2 && part[0] == '"' && part[len(part)-1] == '"' {
			part = part[1 : len(part)-1]
		}
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

func (dbx *Databricks) buildConstraintDefinition(constraintType string, columns []string, referencedTable string, referencedColumns []string) string {
	if len(columns) == 0 {
		return ""
	}

	columnsStr := strings.Join(columns, ", ")

	if strings.ToUpper(constraintType) == "FOREIGN KEY" {
		if referencedTable != "" && len(referencedColumns) > 0 {
			referencedColumnsStr := strings.Join(referencedColumns, ", ")
			return fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)", columnsStr, referencedTable, referencedColumnsStr)
		}
		return fmt.Sprintf("FOREIGN KEY (%s)", columnsStr)
	}

	return fmt.Sprintf("%s (%s)", strings.ToUpper(constraintType), columnsStr)
}

func (dbx *Databricks) getRelations(catalog, schemaName string, tables []*schema.Table) ([]*schema.Relation, error) {
	query := `
		SELECT 
			rc.constraint_name,
			kcu1.table_name as table_name,
			kcu1.column_name as column_name,
			rc.unique_constraint_catalog,
			rc.unique_constraint_schema,
			rc.unique_constraint_name,
			kcu2.table_name as referenced_table_name,
			kcu2.column_name as referenced_column_name,
			kcu1.ordinal_position
		FROM system.information_schema.referential_constraints rc
		INNER JOIN system.information_schema.key_column_usage kcu1 
			ON rc.constraint_catalog = kcu1.constraint_catalog
			AND rc.constraint_schema = kcu1.constraint_schema
			AND rc.constraint_name = kcu1.constraint_name
		INNER JOIN system.information_schema.key_column_usage kcu2
			ON rc.unique_constraint_catalog = kcu2.constraint_catalog
			AND rc.unique_constraint_schema = kcu2.constraint_schema
			AND rc.unique_constraint_name = kcu2.constraint_name
			AND kcu1.position_in_unique_constraint = kcu2.ordinal_position
		WHERE rc.constraint_catalog = ?
			AND rc.constraint_schema = ?
		ORDER BY rc.constraint_name, kcu1.ordinal_position`

	rows, err := dbx.db.Query(query, catalog, schemaName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	relationMap := make(map[string]*schema.Relation)
	tableMap := make(map[string]*schema.Table)

	for _, table := range tables {
		tableMap[table.Name] = table
	}

	for rows.Next() {
		var constraintName, tableName, columnName string
		var refCatalog, refSchema, refConstraintName, refTableName, refColumnName string
		var ordinalPosition int

		if err := rows.Scan(&constraintName, &tableName, &columnName,
			&refCatalog, &refSchema, &refConstraintName, &refTableName, &refColumnName, &ordinalPosition); err != nil {
			continue
		}

		relation, exists := relationMap[constraintName]
		if !exists {
			relation = &schema.Relation{
				Table:       tableMap[tableName],
				ParentTable: tableMap[refTableName],
				Def:         fmt.Sprintf("FOREIGN KEY REFERENCES %s", refTableName),
			}
			relationMap[constraintName] = relation
		}

		if relation.Table != nil {
			if column, err := relation.Table.FindColumnByName(columnName); err == nil {
				relation.Columns = append(relation.Columns, column)
				column.ParentRelations = append(column.ParentRelations, relation)
			}
		}
		if relation.ParentTable != nil {
			if parentColumn, err := relation.ParentTable.FindColumnByName(refColumnName); err == nil {
				relation.ParentColumns = append(relation.ParentColumns, parentColumn)
				parentColumn.ChildRelations = append(parentColumn.ChildRelations, relation)
			}
		}
	}

	// Convert map to slice and filter out incomplete relations
	var relations []*schema.Relation
	for _, relation := range relationMap {
		// Only include relations that have both table and parent table with columns
		if relation.Table != nil && relation.ParentTable != nil &&
			len(relation.Columns) > 0 && len(relation.ParentColumns) > 0 {
			relations = append(relations, relation)
		}
	}

	sort.Slice(relations, func(i, j int) bool {
		return relations[i].Table.Name < relations[j].Table.Name
	})

	return relations, nil
}

func (dbx *Databricks) getViewDefinition(catalog, schemaName, viewName string) (string, error) {
	query := fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`.`%s`", catalog, schemaName, viewName)
	row := dbx.db.QueryRow(query)

	var createStatement string
	if err := row.Scan(&createStatement); err != nil {
		return "", err
	}

	return createStatement, nil
}

func (dbx *Databricks) Info() (*schema.Driver, error) {
	var v string
	row := dbx.db.QueryRow(`SELECT VERSION()`)
	if err := row.Scan(&v); err != nil {
		return nil, err
	}

	return &schema.Driver{
		Name:            "databricks",
		DatabaseVersion: v,
	}, nil
}
