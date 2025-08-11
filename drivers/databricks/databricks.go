package databricks

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/schema"

	// Import the Databricks driver for side effects (database/sql driver registration).
	_ "github.com/databricks/databricks-sql-go"
)

// Databricks struct.
type Databricks struct {
	db *sql.DB
}

// New return new Databricks.
func New(db *sql.DB) *Databricks {
	return &Databricks{
		db: db,
	}
}

// Analyze Databricks database schema.
func (dbx *Databricks) Analyze(s *schema.Schema) error {
	d, err := dbx.Info()
	if err != nil {
		return errors.WithStack(err)
	}
	s.Driver = d

	// Get current catalog and schema
	currentCatalog, currentSchema, err := dbx.getCurrentContext()
	if err != nil {
		return errors.WithStack(err)
	}

	// Set catalog and schema name
	s.Name = fmt.Sprintf("%s.%s", currentCatalog, currentSchema)

	// Get tables
	tables, err := dbx.getTables(currentCatalog, currentSchema)
	if err != nil {
		return errors.WithStack(err)
	}

	// Get columns for each table
	for _, table := range tables {
		columns, err := dbx.getColumns(currentCatalog, currentSchema, table.Name)
		if err != nil {
			return errors.WithStack(err)
		}
		table.Columns = columns

		// Get constraints for each table
		constraints, err := dbx.getConstraints(currentCatalog, currentSchema, table.Name)
		if err != nil {
			return errors.WithStack(err)
		}
		table.Constraints = constraints
	}

	s.Tables = tables

	// Get relations (foreign keys)
	relations, err := dbx.getRelations(currentCatalog, currentSchema, tables)
	if err != nil {
		return errors.WithStack(err)
	}
	s.Relations = relations

	return nil
}

// getCurrentContext gets the current catalog and schema from Databricks.
func (dbx *Databricks) getCurrentContext() (string, string, error) {
	var catalog, schema string

	// Get current catalog
	catRow := dbx.db.QueryRow(`SELECT current_catalog()`)
	if err := catRow.Scan(&catalog); err != nil {
		return "", "", errors.WithStack(err)
	}

	// Get current schema
	schemaRow := dbx.db.QueryRow(`SELECT current_schema()`)
	if err := schemaRow.Scan(&schema); err != nil {
		return "", "", errors.WithStack(err)
	}

	return catalog, schema, nil
}

// getTables retrieves all tables and views from the current schema.
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
			Type:    normalizeTableType(tableType),
			Comment: tableComment,
		}

		// Get view definition if it's a view
		if strings.ToUpper(tableType) == "VIEW" {
			viewDef, err := dbx.getViewDefinition(catalog, schemaName, tableName)
			if err != nil {
				// Don't fail if view definition is not available
				viewDef = ""
			}
			table.Def = viewDef
		}

		tables = append(tables, table)
	}

	return tables, nil
}

// getColumns retrieves column information for a specific table.
func (dbx *Databricks) getColumns(catalog, schemaName, tableName string) ([]*schema.Column, error) {
	query := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default,
			COALESCE(comment, '') as column_comment
		FROM system.information_schema.columns 
		WHERE table_catalog = ? AND table_schema = ? AND table_name = ?
		ORDER BY ordinal_position`

	rows, err := dbx.db.Query(query, catalog, schemaName, tableName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	var columns []*schema.Column
	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault, columnComment sql.NullString

		if err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault, &columnComment); err != nil {
			return nil, errors.WithStack(err)
		}

		column := &schema.Column{
			Name:     columnName,
			Type:     dataType,
			Nullable: strings.ToUpper(isNullable) == "YES",
			Default:  columnDefault,
			Comment:  columnComment.String,
		}

		columns = append(columns, column)
	}

	return columns, nil
}

// getConstraints retrieves constraints for a specific table.
func (dbx *Databricks) getConstraints(catalog, schemaName, tableName string) ([]*schema.Constraint, error) {
	query := `
		SELECT 
			constraint_name,
			constraint_type
		FROM system.information_schema.table_constraints 
		WHERE table_catalog = ? AND table_schema = ? AND table_name = ?
		ORDER BY constraint_name`

	rows, err := dbx.db.Query(query, catalog, schemaName, tableName)
	if err != nil {
		// If constraints query fails, return empty constraints
		return []*schema.Constraint{}, nil
	}
	defer rows.Close()

	var constraints []*schema.Constraint
	for rows.Next() {
		var constraintName, constraintType string
		if err := rows.Scan(&constraintName, &constraintType); err != nil {
			return nil, errors.WithStack(err)
		}

		// Get columns for this constraint
		columns, referencedTable, referencedColumns := dbx.getConstraintDetails(catalog, schemaName, tableName, constraintName, constraintType)

		// Build constraint definition
		def := dbx.buildConstraintDefinition(constraintType, columns, referencedTable, referencedColumns)

		constraint := &schema.Constraint{
			Name:              constraintName,
			Type:              normalizeConstraintType(constraintType),
			Table:             &tableName,
			Def:               def,
			Columns:           columns,
		}

		// For foreign key constraints, populate referenced table and columns
		if strings.ToUpper(constraintType) == "FOREIGN KEY" && referencedTable != "" {
			constraint.ReferencedTable = &referencedTable
			constraint.ReferencedColumns = referencedColumns
		}

		constraints = append(constraints, constraint)
	}

	return constraints, nil
}

// getConstraintDetails retrieves column and reference information for a constraint.
func (dbx *Databricks) getConstraintDetails(catalog, schemaName, tableName, constraintName, constraintType string) ([]string, string, []string) {
	// Get columns for this constraint from key_column_usage
	columnQuery := `
		SELECT 
			column_name
		FROM system.information_schema.key_column_usage
		WHERE table_catalog = ? AND table_schema = ? AND table_name = ? AND constraint_name = ?
		ORDER BY ordinal_position`

	columnRows, err := dbx.db.Query(columnQuery, catalog, schemaName, tableName, constraintName)
	if err != nil {
		return []string{}, "", []string{}
	}
	defer columnRows.Close()

	var columns []string
	for columnRows.Next() {
		var columnName string
		if err := columnRows.Scan(&columnName); err != nil {
			continue
		}
		columns = append(columns, columnName)
	}

	// For foreign keys, get referenced table and columns
	var referencedTable string
	var referencedColumns []string

	if strings.ToUpper(constraintType) == "FOREIGN KEY" {
		// Get referenced table and columns from referential_constraints + key_column_usage
		refQuery := `
			SELECT 
				kcu2.table_name as referenced_table_name,
				kcu2.column_name as referenced_column_name
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
				AND rc.constraint_name = ?
			ORDER BY kcu1.ordinal_position`

		refRows, err := dbx.db.Query(refQuery, catalog, schemaName, constraintName)
		if err == nil {
			defer refRows.Close()
			for refRows.Next() {
				var refTable, refColumn string
				if err := refRows.Scan(&refTable, &refColumn); err != nil {
					continue
				}
				if referencedTable == "" {
					referencedTable = refTable
				}
				referencedColumns = append(referencedColumns, refColumn)
			}
		}
	}

	return columns, referencedTable, referencedColumns
}

// buildConstraintDefinition creates the SQL definition string for a constraint.
func (dbx *Databricks) buildConstraintDefinition(constraintType string, columns []string, referencedTable string, referencedColumns []string) string {
	if len(columns) == 0 {
		return ""
	}

	columnsStr := strings.Join(columns, ", ")

	switch strings.ToUpper(constraintType) {
	case "PRIMARY KEY":
		return fmt.Sprintf("PRIMARY KEY (%s)", columnsStr)
	case "UNIQUE":
		return fmt.Sprintf("UNIQUE (%s)", columnsStr)
	case "FOREIGN KEY":
		if referencedTable != "" && len(referencedColumns) > 0 {
			referencedColumnsStr := strings.Join(referencedColumns, ", ")
			return fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)", columnsStr, referencedTable, referencedColumnsStr)
		}
		return fmt.Sprintf("FOREIGN KEY (%s)", columnsStr)
	case "CHECK":
		// For CHECK constraints, we can't easily get the check condition from information_schema
		// in Databricks, so we'll use a generic format
		return fmt.Sprintf("CHECK (%s)", columnsStr)
	default:
		return fmt.Sprintf("%s (%s)", constraintType, columnsStr)
	}
}

// getRelations retrieves foreign key relationships.
func (dbx *Databricks) getRelations(catalog, schemaName string, tables []*schema.Table) ([]*schema.Relation, error) {
	// Use REFERENTIAL_CONSTRAINTS and KEY_COLUMN_USAGE to get complete foreign key information
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
		// If foreign key queries fail, return empty relations
		return []*schema.Relation{}, nil
	}
	defer rows.Close()

	relationMap := make(map[string]*schema.Relation)
	tableMap := make(map[string]*schema.Table)

	// Create table lookup map
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

		// Get or create relation
		relation, exists := relationMap[constraintName]
		if !exists {
			relation = &schema.Relation{
				Table:       tableMap[tableName],
				ParentTable: tableMap[refTableName],
				Def:         fmt.Sprintf("FOREIGN KEY REFERENCES %s", refTableName),
			}
			relationMap[constraintName] = relation
		}

		// Add columns to relation
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

	// Sort relations for consistent output
	sort.Slice(relations, func(i, j int) bool {
		return relations[i].Table.Name < relations[j].Table.Name
	})

	return relations, nil
}

// getViewDefinition retrieves the SQL definition for a view.
func (dbx *Databricks) getViewDefinition(catalog, schemaName, viewName string) (string, error) {
	// Try to get view definition using SHOW CREATE TABLE
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

// normalizeTableType converts Databricks table types to standard types.
func normalizeTableType(tableType string) string {
	return strings.ToUpper(tableType)
	//{
	// case "BASE TABLE":
	// return "BASE TABLE"
	// case "VIEW":
	// return "VIEW"
	// case "MATERIALIZED VIEW":
	// return "MATERIALIZED VIEW"
	// default:
	// return tableType
	// }
}

// normalizeConstraintType converts constraint types to standard format.
func normalizeConstraintType(constraintType string) string {
	return strings.ToUpper(constraintType)
	//{
	// case "PRIMARY KEY":
	// return "PRIMARY KEY"
	// case "FOREIGN KEY":
	// return "FOREIGN KEY"
	// case "CHECK":
	// return "CHECK"
	// case "UNIQUE":
	// return "UNIQUE"
	// }
}
