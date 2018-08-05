package sqlite

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

// Sqlite struct
type Sqlite struct{}

type fk struct {
	ID                 string
	ForeignTableName   string
	ColumnNames        []string
	ForeignColumnNames []string
	OnUpdate           string
	OnDelete           string
	Match              string
}

// Analyze SQLite database schema
func (l *Sqlite) Analyze(db *sql.DB, s *schema.Schema) error {
	// tables
	tableRows, err := db.Query(`
SELECT name, type, sql
FROM sqlite_master
WHERE name != 'sqlite_sequence' AND (type = 'table' OR type = 'view');`)
	defer tableRows.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	relations := []*schema.Relation{}

	tables := []*schema.Table{}
	for tableRows.Next() {
		var (
			tableName string
			tableType string
			tableDef  string
		)
		err := tableRows.Scan(&tableName, &tableType, &tableDef)
		if err != nil {
			return errors.WithStack(err)
		}

		table := &schema.Table{
			Name: tableName,
			Type: tableType,
			Def:  tableDef,
		}

		// indexes
		indexRows, err := db.Query(`
SELECT name, sql FROM sqlite_master WHERE type = 'index' AND tbl_name = ?;
`, tableName)
		defer indexRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

		indexes := []*schema.Index{}
		for indexRows.Next() {
			var (
				indexName string
				indexDef  sql.NullString
			)
			err = indexRows.Scan(&indexName, &indexDef)
			if err != nil {
				fmt.Printf("%s\n", tableName)

				return errors.WithStack(err)
			}
			index := &schema.Index{
				Name: indexName,
				Def:  indexDef.String,
			}
			indexes = append(indexes, index)
		}
		table.Indexes = indexes

		// constraints
		constraints := []*schema.Constraint{}

		/// foreign keys
		fkMap := map[string]*fk{}
		fkSlice := []*fk{}

		foreignKeyRows, err := db.Query(fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName))
		defer foreignKeyRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}
		for foreignKeyRows.Next() {
			var (
				foreignKeyID                string
				foreignKeySeq               string
				foreignKeyForeignTableName  string
				foreignKeyColumnName        string
				foreignKeyForeignColumnName string
				foreignKeyOnUpdate          string
				foreignKeyOnDelete          string
				foreignKeyMatch             string
			)
			err = foreignKeyRows.Scan(
				&foreignKeyID,
				&foreignKeySeq,
				&foreignKeyForeignTableName,
				&foreignKeyColumnName,
				&foreignKeyForeignColumnName,
				&foreignKeyOnUpdate,
				&foreignKeyOnDelete,
				&foreignKeyMatch,
			)
			if err != nil {
				return errors.WithStack(err)
			}

			if f, ok := fkMap[foreignKeyID]; ok {
				fkMap[foreignKeyID].ColumnNames = append(f.ColumnNames, foreignKeyColumnName)
				fkMap[foreignKeyID].ForeignColumnNames = append(f.ForeignColumnNames, foreignKeyForeignColumnName)
			} else {
				f := &fk{
					ID:                 foreignKeyID,
					ForeignTableName:   foreignKeyForeignTableName,
					ColumnNames:        []string{foreignKeyColumnName},
					ForeignColumnNames: []string{foreignKeyForeignColumnName},
					OnUpdate:           foreignKeyOnUpdate,
					OnDelete:           foreignKeyOnDelete,
					Match:              foreignKeyMatch,
				}
				fkMap[foreignKeyID] = f
			}
		}
		/// Sort foreign keys by ID
		for _, f := range fkMap {
			fkSlice = append(fkSlice, f)
		}
		sort.SliceStable(fkSlice, func(i, j int) bool {
			return fkSlice[i].ID < fkSlice[j].ID
		})

		for _, f := range fkSlice {
			foreignKeyDef := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s) ON UPDATE %s ON DELETE %s MATCH %s",
				strings.Join(f.ColumnNames, ", "), f.ForeignTableName, strings.Join(f.ForeignColumnNames, ", "), f.OnUpdate, f.OnDelete, f.Match)
			constraint := &schema.Constraint{
				Name: f.ID,
				Type: "FOREIGN KEY",
				Def:  foreignKeyDef,
			}
			relation := &schema.Relation{
				Table: table,
				Def:   foreignKeyDef,
			}
			relations = append(relations, relation)

			constraints = append(constraints, constraint)
		}

		table.Constraints = constraints

		// triggers

		// columns
		columnRows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
		defer columnRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

		columns := []*schema.Column{}
		for columnRows.Next() {
			var (
				columnID      string
				columnName    string
				dataType      string
				columnNotNull string
				columnDefault sql.NullString
				columnPk      string
			)
			err = columnRows.Scan(&columnID, &columnName, &dataType, &columnNotNull, &columnDefault, &columnPk)
			if err != nil {
				return errors.WithStack(err)
			}
			column := &schema.Column{
				Name:     columnName,
				Type:     dataType,
				Nullable: convertColumnNullable(columnNotNull),
				Default:  columnDefault,
			}
			columns = append(columns, column)
		}

		table.Columns = columns

		tables = append(tables, table)
	}

	s.Tables = tables

	return nil
}

func convertColumnNullable(str string) bool {
	if str == "1" {
		return false
	}
	return true
}
