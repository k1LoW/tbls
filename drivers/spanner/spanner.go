package spanner

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/spanner"
	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/schema"
	"google.golang.org/api/iterator"
)

type Spanner struct {
	ctx    context.Context
	client *spanner.Client
}

// New return new Spanner.
func New(ctx context.Context, client *spanner.Client) (*Spanner, error) {
	return &Spanner{
		ctx:    ctx,
		client: client,
	}, nil
}

type interleave struct {
	tableName       string
	parentTableName string
	onDeleteAction  string
}

func (sp *Spanner) Analyze(s *schema.Schema) error {
	d, err := sp.Info()
	if err != nil {
		return errors.WithStack(err)
	}
	s.Driver = d

	// tables / constraints
	tableStmt := spanner.Statement{SQL: `
SELECT
  TABLE_NAME, PARENT_TABLE_NAME, ON_DELETE_ACTION
FROM
  INFORMATION_SCHEMA.TABLES
WHERE
  TABLE_CATALOG = '' AND TABLE_SCHEMA = '';
`}
	tableIter := sp.client.Single().Query(sp.ctx, tableStmt)
	defer tableIter.Stop()

	tables := []*schema.Table{}
	tableMap := map[string]*schema.Table{}
	interleaves := []interleave{}
	tableType := "BASIC TABLE"
	for {
		tableRaw, err := tableIter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return errors.WithStack(err)
		}
		var (
			tableName       spanner.NullString
			parentTableName spanner.NullString
			onDeleteAction  spanner.NullString
		)
		if err := tableRaw.Columns(&tableName, &parentTableName, &onDeleteAction); err != nil {
			return errors.WithStack(err)
		}
		table := &schema.Table{
			ShortName: tableName.StringVal,
			Name: tableName.StringVal,
			Type: tableType,
		}

		if parentTableName.StringVal != "" {
			interleaves = append(interleaves, interleave{
				tableName:       tableName.StringVal,
				parentTableName: parentTableName.StringVal,
				onDeleteAction:  onDeleteAction.StringVal,
			})
		}

		// columns
		columnStmt := spanner.Statement{
			SQL: `
SELECT
  COLUMN_NAME, IS_NULLABLE, SPANNER_TYPE
FROM
  INFORMATION_SCHEMA.COLUMNS
WHERE
  TABLE_NAME = @tableName AND TABLE_CATALOG = '' AND TABLE_SCHEMA = ''
ORDER BY ORDINAL_POSITION ASC;
`,
			Params: map[string]interface{}{"tableName": tableName},
		}
		columnIter := sp.client.Single().Query(sp.ctx, columnStmt)
		columns := []*schema.Column{}
		for {
			columnRow, err := columnIter.Next()
			if errors.Is(err, iterator.Done) {
				columnIter.Stop()
				break
			}
			if err != nil {
				columnIter.Stop()
				return errors.WithStack(err)
			}
			var (
				columnName string
				isNullable string
				columnType string
			)

			if err := columnRow.Columns(&columnName, &isNullable, &columnType); err != nil {
				columnIter.Stop()
				return errors.WithStack(err)
			}
			column := &schema.Column{
				Name:     columnName,
				Type:     columnType,
				Nullable: convertColumnNullable(isNullable),
			}

			// column options
			optionStmt := spanner.Statement{
				SQL: `
SELECT
  OPTION_NAME, OPTION_VALUE
FROM
  INFORMATION_SCHEMA.COLUMN_OPTIONS
WHERE
  TABLE_NAME = @tableName AND COLUMN_NAME = @columnName AND TABLE_CATALOG = '' AND TABLE_SCHEMA = '';
`,
				Params: map[string]interface{}{"tableName": tableName, "columnName": columnName},
			}
			optionIter := sp.client.Single().Query(sp.ctx, optionStmt)
			for {
				optionRow, err := optionIter.Next()
				if errors.Is(err, iterator.Done) {
					optionIter.Stop()
					break
				}
				if err != nil {
					optionIter.Stop()
					return errors.WithStack(err)
				}
				var (
					optionName  string
					optionValue string
				)
				if err := optionRow.Columns(&optionName, &optionValue); err != nil {
					optionIter.Stop()
					return errors.WithStack(err)
				}
				column.Type = fmt.Sprintf("%s (%s=%s)", column.Type, optionName, optionValue)
			}
			optionIter.Stop()

			columns = append(columns, column)
		}
		columnIter.Stop()
		table.Columns = columns

		// indexes / constraints
		indexStmt := spanner.Statement{
			SQL: `
SELECT
  c.INDEX_NAME, c.INDEX_TYPE, ARRAY_TO_STRING(ARRAY(
   SELECT COLUMN_NAME
   FROM INFORMATION_SCHEMA.INDEX_COLUMNS
   WHERE TABLE_NAME = c.TABLE_NAME AND INDEX_NAME = c.INDEX_NAME AND INDEX_TYPE = c.INDEX_TYPE AND ORDINAL_POSITION IS NOT NULL
   ORDER BY ORDINAL_POSITION ASC
 ), ", ") AS columns,
 ARRAY_TO_STRING(ARRAY(
   SELECT COLUMN_NAME
   FROM INFORMATION_SCHEMA.INDEX_COLUMNS
   WHERE TABLE_NAME = c.TABLE_NAME AND INDEX_NAME = c.INDEX_NAME AND INDEX_TYPE = c.INDEX_TYPE AND ORDINAL_POSITION IS NULL
   ORDER BY INDEX_NAME ASC
 ), ", ") AS storing_columns,
  i.PARENT_TABLE_NAME, i.IS_UNIQUE, i.IS_NULL_FILTERED, i.INDEX_STATE
FROM
  INFORMATION_SCHEMA.INDEX_COLUMNS AS c
INNER JOIN INFORMATION_SCHEMA.INDEXES AS i ON i.TABLE_NAME = c.TABLE_NAME AND i.INDEX_NAME = c.INDEX_NAME
WHERE
  c.TABLE_CATALOG = '' AND c.TABLE_SCHEMA = '' AND c.TABLE_NAME = @tableName
GROUP BY c.TABLE_CATALOG, c.TABLE_SCHEMA, c.TABLE_NAME, c.INDEX_NAME, c.INDEX_TYPE, i.PARENT_TABLE_NAME, i.IS_UNIQUE, i.IS_NULL_FILTERED, i.INDEX_STATE;
`,
			Params: map[string]interface{}{"tableName": tableName},
		}
		indexIter := sp.client.Single().Query(sp.ctx, indexStmt)
		indexes := []*schema.Index{}
		constraints := []*schema.Constraint{}

		for {
			indexRow, err := indexIter.Next()
			if errors.Is(err, iterator.Done) {
				indexIter.Stop()
				break
			}
			if err != nil {
				indexIter.Stop()
				return errors.WithStack(err)
			}
			var (
				indexName       string
				indexType       string
				columns         string
				storingColumns  string
				parentTableName spanner.NullString
				isUnique        bool
				isNullFiltered  bool
				indexState      spanner.NullString
			)
			if err := indexRow.Columns(&indexName, &indexType, &columns, &storingColumns, &parentTableName, &isUnique, &isNullFiltered, &indexState); err != nil {
				indexIter.Stop()
				return errors.WithStack(err)
			}

			switch indexType {
			case "INDEX":
				var (
					strUnique         string
					strNullFiltered   string
					strInterleave     string
					strStoringColumns string
				)
				if isUnique {
					strUnique = "UNIQUE "
				}
				if isNullFiltered {
					strNullFiltered = "NULL_FILTERED "
				}
				if storingColumns != "" {
					strStoringColumns = fmt.Sprintf(" STORING (%s)", storingColumns)
				}
				if parentTableName.StringVal != "" {
					strInterleave = fmt.Sprintf(", INTERLEAVE IN %s", parentTableName.StringVal)
				}

				indexDef := fmt.Sprintf("CREATE %s%sINDEX %s ON %s (%s)%s%s", strUnique, strNullFiltered, indexName, table.Name, columns, strStoringColumns, strInterleave)

				index := &schema.Index{
					Name:    indexName,
					Def:     indexDef,
					Table:   &table.Name,
					Columns: strings.Split(columns, ", "),
				}
				indexes = append(indexes, index)
			case "PRIMARY_KEY":
				constraint := &schema.Constraint{
					Name:              "PRIMARY_KEY",
					Type:              "PRIMARY_KEY",
					Def:               fmt.Sprintf("PRIMARY KEY(%s)", columns),
					Table:             &table.Name,
					Columns:           strings.Split(columns, ", "),
					ReferencedTable:   nil,
					ReferencedColumns: []string{},
				}
				constraints = append(constraints, constraint)
			default:
			}
		}
		indexIter.Stop()
		table.Indexes = indexes
		table.Constraints = constraints

		tables = append(tables, table)
		tableMap[table.Name] = table
	}

	s.Tables = tables

	// bulk get foreign keys
	fkStmt := spanner.Statement{
		SQL: `
SELECT
  kcu.TABLE_NAME,
  rc.CONSTRAINT_NAME,
  STRING_AGG(kcu.COLUMN_NAME, ', ' ORDER BY kcu.ORDINAL_POSITION) AS columns,
  ccu.TABLE_NAME as REFERENCED_TABLE_NAME,
  STRING_AGG(ccu.COLUMN_NAME, ', ' ORDER BY kcu.ORDINAL_POSITION) AS referenced_columns,
  rc.DELETE_RULE
FROM
  INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS AS rc
INNER JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE AS kcu
  ON rc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME
INNER JOIN INFORMATION_SCHEMA.CONSTRAINT_COLUMN_USAGE AS ccu
  ON rc.UNIQUE_CONSTRAINT_NAME = ccu.CONSTRAINT_NAME
WHERE
  kcu.TABLE_CATALOG = '' AND kcu.TABLE_SCHEMA = ''
  AND ccu.TABLE_CATALOG = '' AND ccu.TABLE_SCHEMA = ''
GROUP BY kcu.TABLE_NAME, rc.CONSTRAINT_NAME, ccu.TABLE_NAME, rc.DELETE_RULE;
`,
	}
	fkIter := sp.client.Single().Query(sp.ctx, fkStmt)
	defer fkIter.Stop()

	relations := []*schema.Relation{}

	for {
		fkRow, err := fkIter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return errors.WithStack(err)
		}
		var (
			tableName           string
			constraintName      string
			columns             string
			referencedTableName string
			referencedColumns   string
			deleteRule          string
		)
		if err := fkRow.Columns(&tableName, &constraintName, &columns, &referencedTableName, &referencedColumns, &deleteRule); err != nil {
			return errors.WithStack(err)
		}

		t, ok := tableMap[tableName]
		if !ok {
			continue
		}
		rt, ok := tableMap[referencedTableName]
		if !ok {
			continue
		}

		columnList := strings.Split(columns, ", ")
		referencedColumnList := strings.Split(referencedColumns, ", ")

		constraintDef := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)", columns, referencedTableName, referencedColumns)
		if deleteRule != "" && deleteRule != "NO ACTION" {
			constraintDef = fmt.Sprintf("%s ON DELETE %s", constraintDef, deleteRule)
		}

		constraint := &schema.Constraint{
			Name:              constraintName,
			Type:              schema.TypeFK,
			Def:               constraintDef,
			Table:             &t.Name,
			Columns:           columnList,
			ReferencedTable:   &rt.Name,
			ReferencedColumns: referencedColumnList,
		}
		t.Constraints = append(t.Constraints, constraint)

		relation := &schema.Relation{
			Table:         t,
			Columns:       []*schema.Column{},
			ParentTable:   rt,
			ParentColumns: []*schema.Column{},
			Def:           constraintDef,
			Virtual:       false,
		}

		for _, colName := range columnList {
			column, err := t.FindColumnByName(colName)
			if err != nil {
				return err
			}
			column.ParentRelations = append(column.ParentRelations, relation)
			relation.Columns = append(relation.Columns, column)
		}

		for _, colName := range referencedColumnList {
			column, err := rt.FindColumnByName(colName)
			if err != nil {
				return err
			}
			column.ChildRelations = append(column.ChildRelations, relation)
			relation.ParentColumns = append(relation.ParentColumns, column)
		}

		relations = append(relations, relation)
	}

	// interleaves
	for _, i := range interleaves {
		t, err := s.FindTableByName(i.tableName)
		if err != nil {
			return err
		}
		pt, err := s.FindTableByName(i.parentTableName)
		if err != nil {
			return err
		}
		def := fmt.Sprintf("INTERLEAVE IN PARENT %s ON DELETE %s", i.parentTableName, i.onDeleteAction) // #nosec

		// constraints
		constraint := &schema.Constraint{
			Name:              "INTERLEAVE",
			Type:              "INTERLEAVE",
			Def:               def,
			Table:             &t.Name,
			Columns:           []string{},
			ReferencedTable:   &pt.Name,
			ReferencedColumns: []string{},
		}

		// relations
		relation := &schema.Relation{
			Table:         t,
			Columns:       []*schema.Column{},
			ParentTable:   pt,
			ParentColumns: []*schema.Column{},
			Def:           def,
			Virtual:       false,
		}

		for _, c := range t.Constraints {
			if c.Type == "PRIMARY_KEY" {
				constraint.Columns = c.Columns
				for _, cName := range c.Columns {
					column, err := t.FindColumnByName(cName)
					if err != nil {
						return err
					}
					column.ParentRelations = append(column.ParentRelations, relation)
					relation.Columns = append(relation.Columns, column)
				}
			}
		}
		for _, c := range pt.Constraints {
			if c.Type == "PRIMARY_KEY" {
				constraint.ReferencedColumns = c.Columns
				for _, cName := range c.Columns {
					column, err := pt.FindColumnByName(cName)
					if err != nil {
						return err
					}
					column.ChildRelations = append(column.ChildRelations, relation)
					relation.ParentColumns = append(relation.ParentColumns, column)
				}
			}
		}
		t.Constraints = append(t.Constraints, constraint)
		relations = append(relations, relation)
	}

	s.Relations = relations

	return nil
}

func (sp *Spanner) Info() (*schema.Driver, error) {
	d := &schema.Driver{
		Name:            "spanner",
		DatabaseVersion: "",
	}
	return d, nil
}

func convertColumnNullable(str string) bool {
	return str != "NO"
}
