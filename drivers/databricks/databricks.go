package databricks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	// Import databricks-sql-go driver for database/sql registration.
	_ "github.com/databricks/databricks-sql-go"
	"github.com/k1LoW/errors"
	"github.com/rs/zerolog"

	"github.com/k1LoW/tbls/schema"
)

func init() {
	// required to silence some internal logging on the databricks sdk
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

type Databricks struct {
	db              *sql.DB
	tablesAPIClient TablesAPIClient
	explicitSchema  bool
}

type TablesAPIClient interface {
	GetTable(ctx context.Context, catalog, schema, tableName string) (*TableInfo, error)
}

type TableInfo struct {
	FullName string       `json:"full_name"`
	Columns  []ColumnInfo `json:"columns"`
}

type ColumnInfo struct {
	Name     string `json:"name"`
	TypeName string `json:"type_name"`
	TypeText string `json:"type_text"`
	TypeJSON string `json:"type_json"`
	Position int    `json:"position"`
	Nullable bool   `json:"nullable"`
}

func New(db *sql.DB, apiClient TablesAPIClient, explicitSchema bool) *Databricks {
	return &Databricks{
		db:              db,
		tablesAPIClient: apiClient,
		explicitSchema:  explicitSchema,
	}
}

func (dbx *Databricks) Analyze(s *schema.Schema) (err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	d, err := dbx.Info()
	if err != nil {
		return err
	}
	s.Driver = d

	currentCatalog, currentSchema, err := dbx.getCurrentContext()
	if err != nil {
		return err
	}

	var targetSchema sql.NullString
	if dbx.explicitSchema {
		targetSchema = sql.NullString{String: currentSchema, Valid: true}
		s.Name = fmt.Sprintf("%s.%s", currentCatalog, currentSchema)
		s.Driver.Meta.CurrentSchema = currentSchema
	} else {
		targetSchema = sql.NullString{Valid: false}
		s.Name = currentCatalog
	}

	tables, err := dbx.getTables(currentCatalog, targetSchema)
	if err != nil {
		return err
	}

	columnsByTable, err := dbx.getAllColumns(currentCatalog, targetSchema)
	if err != nil {
		return err
	}

	constraintsByTable, err := dbx.getAllConstraints(currentCatalog, targetSchema)
	if err != nil {
		return err
	}

	for _, table := range tables {
		if columns, exists := columnsByTable[table.Name]; exists {
			table.Columns = columns
		}

		if constraints, exists := constraintsByTable[table.Name]; exists {
			table.Constraints = constraints
		}

		if dbx.hasStructColumns(table.Columns) {
			if err := dbx.enrichStructColumns(context.Background(), currentCatalog, currentSchema, table); err != nil {
				return err
			}
		}
	}

	s.Tables = tables

	relations, err := dbx.getRelations(currentCatalog, targetSchema, tables)
	if err != nil {
		return err
	}
	s.Relations = relations

	return nil
}

func (dbx *Databricks) getCurrentContext() (string, string, error) {
	var catalog, schema string

	catRow := dbx.db.QueryRow(`SELECT current_catalog()`)
	if err := catRow.Scan(&catalog); err != nil {
		return "", "", err
	}

	schemaRow := dbx.db.QueryRow(`SELECT current_schema()`)
	if err := schemaRow.Scan(&schema); err != nil {
		return "", "", err
	}

	return catalog, schema, nil
}

func (dbx *Databricks) getTables(catalog string, schemaName sql.NullString) ([]*schema.Table, error) {
	var query string
	var rows *sql.Rows
	var err error

	if schemaName.Valid {
		query = `
			SELECT 
				table_schema,
				table_name, 
				table_type,
				COALESCE(comment, '') as table_comment
			FROM system.information_schema.tables 
			WHERE table_catalog = ? AND table_schema = ?
			ORDER BY table_name`
		rows, err = dbx.db.Query(query, catalog, schemaName.String)
	} else {
		query = `
			SELECT 
				table_schema,
				table_name, 
				table_type,
				COALESCE(comment, '') as table_comment
			FROM system.information_schema.tables 
			WHERE table_catalog = ?
			ORDER BY table_schema, table_name`
		rows, err = dbx.db.Query(query, catalog)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []*schema.Table
	for rows.Next() {
		var tableSchema, tableName, tableType, tableComment string
		if err := rows.Scan(&tableSchema, &tableName, &tableType, &tableComment); err != nil {
			return nil, err
		}

		fullTableName := tableName
		if !schemaName.Valid {
			fullTableName = fmt.Sprintf("%s.%s", tableSchema, tableName)
		}

		table := &schema.Table{
			Name:    fullTableName,
			Type:    strings.ToUpper(tableType),
			Comment: tableComment,
		}

		if strings.ToUpper(tableType) == "VIEW" {
			viewDef, err := dbx.getViewDefinition(catalog, tableSchema, tableName)
			if err != nil {
				return nil, err
			}
			table.Def = viewDef
		}

		tables = append(tables, table)
	}

	return tables, nil
}

func (dbx *Databricks) getAllColumns(catalog string, schemaName sql.NullString) (map[string][]*schema.Column, error) {
	var query string
	var rows *sql.Rows
	var err error

	if schemaName.Valid {
		query = `
			SELECT 
				c.table_name,
				c.column_name,
				c.data_type,
				c.is_nullable,
				c.column_default,
				COALESCE(c.comment, '') as column_comment
			FROM system.information_schema.columns c
			INNER JOIN system.information_schema.tables t
			    ON c.table_catalog = t.table_catalog
			    AND c.table_schema = t.table_schema
			    AND c.table_name = t.table_name
			WHERE c.table_catalog = ? AND c.table_schema = ?
			ORDER BY c.table_name, c.ordinal_position`
		rows, err = dbx.db.Query(query, catalog, schemaName.String)
	} else {
		query = `
			SELECT 
				c.table_schema,
				c.table_name,
				c.column_name,
				c.data_type,
				c.is_nullable,
				c.column_default,
				COALESCE(c.comment, '') as column_comment
			FROM system.information_schema.columns c
			INNER JOIN system.information_schema.tables t
			    ON c.table_catalog = t.table_catalog
			    AND c.table_schema = t.table_schema
			    AND c.table_name = t.table_name
			WHERE c.table_catalog = ?
			ORDER BY c.table_schema, c.table_name, c.ordinal_position`
		rows, err = dbx.db.Query(query, catalog)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columnsByTable := make(map[string][]*schema.Column)
	for rows.Next() {
		var tableName, columnName, dataType, isNullable string
		var columnDefault, columnComment sql.NullString
		var tableSchema string

		if schemaName.Valid {
			if err := rows.Scan(&tableName, &columnName, &dataType, &isNullable, &columnDefault, &columnComment); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(&tableSchema, &tableName, &columnName, &dataType, &isNullable, &columnDefault, &columnComment); err != nil {
				return nil, err
			}
			tableName = fmt.Sprintf("%s.%s", tableSchema, tableName)
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

func (dbx *Databricks) getAllConstraints(catalog string, schemaName sql.NullString) (map[string][]*schema.Constraint, error) {
	var query string
	var rows *sql.Rows
	var err error

	if schemaName.Valid {
		query = `
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
			GROUP BY tc.table_name, tc.constraint_name, tc.constraint_type
			ORDER BY tc.table_name, tc.constraint_name`
		rows, err = dbx.db.Query(query, catalog, schemaName.String)
	} else {
		query = `
			SELECT 
				tc.table_schema,
				tc.table_name,
				tc.constraint_name,
				tc.constraint_type,
				COALESCE(COLLECT_LIST(kcu.column_name), ARRAY()) as constraint_columns,
				COALESCE(MAX(kcu2.table_schema), '') as referenced_table_schema,
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
			WHERE tc.table_catalog = ?
			GROUP BY tc.table_schema, tc.table_name, tc.constraint_name, tc.constraint_type
			ORDER BY tc.table_schema, tc.table_name, tc.constraint_name`
		rows, err = dbx.db.Query(query, catalog)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	constraintsByTable := make(map[string][]*schema.Constraint)
	for rows.Next() {
		var tableName, constraintName, constraintType, referencedTableName string
		var constraintColumnsStr, referencedColumnsStr string
		var tableSchema, referencedTableSchema string

		if schemaName.Valid {
			if err := rows.Scan(&tableName, &constraintName, &constraintType, &constraintColumnsStr, &referencedTableName, &referencedColumnsStr); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(&tableSchema, &tableName, &constraintName, &constraintType, &constraintColumnsStr, &referencedTableSchema, &referencedTableName, &referencedColumnsStr); err != nil {
				return nil, err
			}
			tableName = fmt.Sprintf("%s.%s", tableSchema, tableName)
			if referencedTableName != "" {
				referencedTableName = fmt.Sprintf("%s.%s", referencedTableSchema, referencedTableName)
			}
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
	arrayStr = strings.TrimSpace(arrayStr)

	if arrayStr == "" || arrayStr == "[]" {
		return []string{}
	}

	arrayStr = strings.TrimPrefix(arrayStr, "[")
	arrayStr = strings.TrimSuffix(arrayStr, "]")

	if arrayStr == "" {
		return []string{}
	}

	parts := strings.Split(arrayStr, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
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

func (dbx *Databricks) getRelations(catalog string, schemaName sql.NullString, tables []*schema.Table) ([]*schema.Relation, error) {
	var query string
	var rows *sql.Rows
	var err error

	if schemaName.Valid {
		query = `
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
		rows, err = dbx.db.Query(query, catalog, schemaName.String)
	} else {
		query = `
			SELECT 
				rc.constraint_name,
				kcu1.table_schema as table_schema,
				kcu1.table_name as table_name,
				kcu1.column_name as column_name,
				rc.unique_constraint_catalog,
				rc.unique_constraint_schema,
				rc.unique_constraint_name,
				kcu2.table_schema as referenced_table_schema,
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
			ORDER BY rc.constraint_name, kcu1.ordinal_position`
		rows, err = dbx.db.Query(query, catalog)
	}

	if err != nil {
		return nil, err
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
		var tableSchema, refTableSchema string

		if schemaName.Valid {
			if err := rows.Scan(&constraintName, &tableName, &columnName,
				&refCatalog, &refSchema, &refConstraintName, &refTableName, &refColumnName, &ordinalPosition); err != nil {
				continue
			}
		} else {
			if err := rows.Scan(&constraintName, &tableSchema, &tableName, &columnName,
				&refCatalog, &refSchema, &refConstraintName, &refTableSchema, &refTableName, &refColumnName, &ordinalPosition); err != nil {
				continue
			}
			tableName = fmt.Sprintf("%s.%s", tableSchema, tableName)
			refTableName = fmt.Sprintf("%s.%s", refTableSchema, refTableName)
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

	var relations []*schema.Relation
	for _, relation := range relationMap {
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

func (dbx *Databricks) hasStructColumns(columns []*schema.Column) bool {
	for _, col := range columns {
		colType := strings.ToUpper(col.Type)
		if strings.HasPrefix(colType, "STRUCT") || strings.HasPrefix(colType, "ARRAY(STRUCT") {
			return true
		}
	}
	return false
}

func (dbx *Databricks) enrichStructColumns(ctx context.Context, catalog, schemaName string, table *schema.Table) error {
	tableName := table.Name
	tableSchema := schemaName

	if strings.Contains(tableName, ".") {
		parts := strings.SplitN(tableName, ".", 2)
		tableSchema = parts[0]
		tableName = parts[1]
	}

	tableInfo, err := dbx.tablesAPIClient.GetTable(ctx, catalog, tableSchema, tableName)
	if err != nil {
		return err
	}

	columnMap := make(map[string]*ColumnInfo)
	for i := range tableInfo.Columns {
		columnMap[tableInfo.Columns[i].Name] = &tableInfo.Columns[i]
	}

	var expandedColumns []*schema.Column
	for _, col := range table.Columns {
		colInfo, exists := columnMap[col.Name]

		expandedColumns = append(expandedColumns, col)

		if !exists || colInfo.TypeJSON == "" {
			continue
		}

		isComplexType := strings.HasPrefix(strings.ToUpper(colInfo.TypeName), "STRUCT") ||
			strings.HasPrefix(strings.ToUpper(colInfo.TypeName), "ARRAY") ||
			strings.HasPrefix(strings.ToUpper(colInfo.TypeName), "MAP")

		if !isComplexType {
			continue
		}

		var typeData map[string]any
		if err := json.Unmarshal([]byte(colInfo.TypeJSON), &typeData); err != nil {
			continue
		}

		col.Type = dbx.formatType(typeData)

		nestedCols := dbx.extractNestedColumns(typeData, col.Name)
		expandedColumns = append(expandedColumns, nestedCols...)
	}

	table.Columns = expandedColumns

	return nil
}

func (dbx *Databricks) extractNestedColumns(typeData map[string]any, prefix string) []*schema.Column {
	var columns []*schema.Column

	typeObj, ok := typeData["type"]
	if !ok {
		return columns
	}

	nullable := true
	if nullableVal, ok := typeData["nullable"].(bool); ok {
		nullable = nullableVal
	}

	var typeStr string
	var fieldsSource map[string]any

	switch t := typeObj.(type) {
	case string:
		typeStr = t
		fieldsSource = typeData
	case map[string]any:
		typeStr, _ = t["type"].(string)
		fieldsSource = t
	default:
		return columns
	}

	switch typeStr {
	case "struct":
		return dbx.processStructFields(fieldsSource, prefix)
	case "array":
		return dbx.processArrayType(fieldsSource, prefix, nullable)
	}

	return columns
}

func (dbx *Databricks) processStructFields(source map[string]any, prefix string) []*schema.Column {
	var columns []*schema.Column

	fields, ok := source["fields"].([]any)
	if !ok {
		return columns
	}

	for _, f := range fields {
		field, ok := f.(map[string]any)
		if !ok {
			continue
		}

		fieldName, _ := field["name"].(string)
		if fieldName == "" {
			continue
		}

		fullName := fmt.Sprintf("%s.%s", prefix, fieldName)

		fieldType := dbx.formatType(field)
		fieldNullable := true
		if nullableVal, ok := field["nullable"].(bool); ok {
			fieldNullable = nullableVal
		}

		fieldComment := ""
		if metadata, ok := field["metadata"].(map[string]any); ok {
			if comment, ok := metadata["comment"].(string); ok {
				fieldComment = comment
			}
		}

		col := &schema.Column{
			Name:     fullName,
			Type:     fieldType,
			Nullable: fieldNullable,
			Comment:  fieldComment,
		}
		columns = append(columns, col)

		nestedCols := dbx.extractNestedColumns(field, fullName)
		columns = append(columns, nestedCols...)
	}

	return columns
}

func (dbx *Databricks) processArrayType(source map[string]any, prefix string, nullable bool) []*schema.Column {
	elementType, ok := source["elementType"].(map[string]any)
	if !ok {
		return nil
	}

	return dbx.extractNestedColumns(map[string]any{
		"type":     elementType,
		"nullable": nullable,
	}, prefix)
}

func (dbx *Databricks) formatType(typeData map[string]any) string {
	typeObj, ok := typeData["type"]
	if !ok {
		return "UNKNOWN"
	}

	switch t := typeObj.(type) {
	case string:
		return strings.ToUpper(t)

	case map[string]any:
		structType, ok := t["type"].(string)
		if !ok {
			return "UNKNOWN"
		}

		switch structType {
		case "struct":
			return "STRUCT"
		case "array":
			if elementType, ok := t["elementType"].(string); ok {
				return fmt.Sprintf("ARRAY(%s)", strings.ToUpper(elementType))
			} else if elementMap, ok := t["elementType"].(map[string]any); ok {
				if nestedType, ok := elementMap["type"].(string); ok {
					if nestedType == "struct" {
						return "ARRAY(STRUCT)"
					}
					return fmt.Sprintf("ARRAY(%s)", strings.ToUpper(nestedType))
				}
			}
			return "ARRAY"
		case "map":
			keyType := strings.ToUpper(t["keyType"].(string))

			var valueType string
			if vt, ok := t["valueType"].(string); ok {
				valueType = strings.ToUpper(vt)
			} else if valueMap, ok := t["valueType"].(map[string]any); ok {
				valueType = strings.ToUpper(valueMap["type"].(string))
			}

			return fmt.Sprintf("MAP(%s, %s)", keyType, valueType)
		default:
			return strings.ToUpper(structType)
		}

	default:
		return "UNKNOWN"
	}
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
		Meta:            &schema.DriverMeta{},
	}, nil
}
