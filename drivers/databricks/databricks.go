package databricks

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/dict"
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
func (d *Databricks) Analyze(s *schema.Schema) error {
	driver, err := d.Info()
	if err != nil {
		return errors.WithStack(err)
	}
	s.Driver = driver

	// Get current catalog and schema
	currentCatalog, currentSchema, err := d.getCurrentContext()
	if err != nil {
		return errors.WithStack(err)
	}

	// Set schema name
	s.Name = fmt.Sprintf("%s.%s", currentCatalog, currentSchema)

	// Get tables
	tables, err := d.getTables(currentCatalog, currentSchema)
	if err != nil {
		return errors.WithStack(err)
	}

	// Get columns for each table
	for _, table := range tables {
		columns, err := d.getColumns(currentCatalog, currentSchema, table.Name)
		if err != nil {
			return errors.WithStack(err)
		}
		table.Columns = columns

		// Get constraints for each table
		constraints, err := d.getConstraints(currentCatalog, currentSchema, table.Name)
		if err != nil {
			return errors.WithStack(err)
		}
		table.Constraints = constraints
	}

	s.Tables = tables

	// Get relations (foreign keys)
	relations, err := d.getRelations(currentCatalog, currentSchema, tables)
	if err != nil {
		return errors.WithStack(err)
	}
	s.Relations = relations

	return nil
}

// getCurrentContext gets the current catalog and schema from Databricks.
func (d *Databricks) getCurrentContext() (string, string, error) {
	var catalog, schema string
	
	// Get current catalog
	row := d.db.QueryRow(`SELECT current_catalog()`)
	if err := row.Scan(&catalog); err != nil {
		return "", "", errors.WithStack(err)
	}
	
	// Get current schema
	row = d.db.QueryRow(`SELECT current_schema()`)
	if err := row.Scan(&schema); err != nil {
		return "", "", errors.WithStack(err)
	}
	
	return catalog, schema, nil
}

// getTables retrieves all tables and views from the current schema.
func (d *Databricks) getTables(catalog, schemaName string) ([]*schema.Table, error) {
	query := `
		SELECT 
			table_name, 
			table_type,
			COALESCE(comment, '') as table_comment
		FROM system.information_schema.tables 
		WHERE table_catalog = ? AND table_schema = ?
		ORDER BY table_name`

	rows, err := d.db.Query(query, catalog, schemaName)
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
			viewDef, err := d.getViewDefinition(catalog, schemaName, tableName)
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
func (d *Databricks) getColumns(catalog, schemaName, tableName string) ([]*schema.Column, error) {
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

	rows, err := d.db.Query(query, catalog, schemaName, tableName)
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
func (d *Databricks) getConstraints(catalog, schemaName, tableName string) ([]*schema.Constraint, error) {
	query := `
		SELECT 
			constraint_name,
			constraint_type
		FROM system.information_schema.table_constraints 
		WHERE table_catalog = ? AND table_schema = ? AND table_name = ?
		ORDER BY constraint_name`

	rows, err := d.db.Query(query, catalog, schemaName, tableName)
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

		constraint := &schema.Constraint{
			Name:  constraintName,
			Type:  normalizeConstraintType(constraintType),
			Table: &tableName,
		}

		constraints = append(constraints, constraint)
	}

	return constraints, nil
}

// getRelations retrieves foreign key relationships.
func (d *Databricks) getRelations(catalog, schemaName string, tables []*schema.Table) ([]*schema.Relation, error) {
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

	rows, err := d.db.Query(query, catalog, schemaName)
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
			}
		}
		if relation.ParentTable != nil {
			if parentColumn, err := relation.ParentTable.FindColumnByName(refColumnName); err == nil {
				relation.ParentColumns = append(relation.ParentColumns, parentColumn)
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
func (d *Databricks) getViewDefinition(catalog, schemaName, viewName string) (string, error) {
	// Try to get view definition using SHOW CREATE TABLE
	query := fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`.`%s`", catalog, schemaName, viewName)
	row := d.db.QueryRow(query)
	
	var createStatement string
	if err := row.Scan(&createStatement); err != nil {
		return "", err
	}
	
	return createStatement, nil
}

// Info returns driver information.
func (d *Databricks) Info() (*schema.Driver, error) {
	var version string
	row := d.db.QueryRow(`SELECT version()`)
	if err := row.Scan(&version); err != nil {
		// If version query fails, use a default
		version = "Unknown"
	}

	dct := dict.New()
	dct.Merge(map[string]string{
		"Comment": "Description",
	})

	driver := &schema.Driver{
		Name:            "databricks",
		DatabaseVersion: version,
		Meta: &schema.DriverMeta{
			Dict: &dct,
		},
	}

	return driver, nil
}

// normalizeTableType converts Databricks table types to standard types.
func normalizeTableType(tableType string) string {
	switch strings.ToUpper(tableType) {
	case "BASE TABLE":
		return "BASE TABLE"
	case "VIEW":
		return "VIEW"
	case "MATERIALIZED VIEW":
		return "MATERIALIZED VIEW"
	default:
		return tableType
	}
}

// normalizeConstraintType converts constraint types to standard format.
func normalizeConstraintType(constraintType string) string {
	switch strings.ToUpper(constraintType) {
	case "PRIMARY KEY":
		return "PRIMARY KEY"
	case "FOREIGN KEY":
		return "FOREIGN KEY"
	case "CHECK":
		return "CHECK"
	case "UNIQUE":
		return "UNIQUE"
	default:
		return constraintType
	}
}