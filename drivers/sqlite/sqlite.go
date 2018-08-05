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

		// constraints
		constraints := []*schema.Constraint{}

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

			if columnPk != "0" {
				constraintDef := fmt.Sprintf("PRIMARY KEY (%s)", columnName)
				constraint := &schema.Constraint{
					Name: columnName,
					Type: "PRIMARY KEY",
					Def:  constraintDef,
				}
				constraints = append(constraints, constraint)
			}
		}

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
				Name: fmt.Sprintf("- (Foreign key ID: %s)", f.ID),
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

		// indexes and constraints(UNIQUE, PRIMARY KEY)
		indexRows, err := db.Query(fmt.Sprintf("PRAGMA index_list(%s)", tableName))
		defer indexRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

		indexes := []*schema.Index{}
		for indexRows.Next() {
			var (
				indexID        string
				indexName      string
				indexIsUnique  string
				indexCreatedBy string
				indexPartial   string
				indexDef       string
			)
			err = indexRows.Scan(
				&indexID,
				&indexName,
				&indexIsUnique,
				&indexCreatedBy,
				&indexPartial,
			)
			if err != nil {
				return errors.WithStack(err)
			}

			if indexCreatedBy == "c" {
				row, err := db.Query(`SELECT sql FROM sqlite_master WHERE type = 'index' AND tbl_name = ? AND name = ?;
`, tableName, indexName)
				for row.Next() {
					err = row.Scan(
						&indexDef,
					)
					if err != nil {
						return errors.WithStack(err)
					}
				}
			} else {
				var (
					colRank            string
					colRankWithinTable string
					col                string
					cols               []string
				)
				row, err := db.Query(fmt.Sprintf("PRAGMA index_info(%s)", indexName))
				for row.Next() {
					err = row.Scan(
						&colRank,
						&colRankWithinTable,
						&col,
					)
					if err != nil {
						return errors.WithStack(err)
					}
					cols = append(cols, col)
				}
				switch indexCreatedBy {
				case "u":
					indexDef = fmt.Sprintf("UNIQUE (%s)", strings.Join(cols, ", "))
					constraint := &schema.Constraint{
						Name: indexName,
						Type: "UNIQUE",
						Def:  indexDef,
					}
					constraints = append(constraints, constraint)
				case "pk":
					// MEMO: Does not work ?
					indexDef = fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(cols, ", "))
					constraint := &schema.Constraint{
						Name: indexName,
						Type: "PRIMARY KEY",
						Def:  indexDef,
					}
					constraints = append(constraints, constraint)
				}
			}

			index := &schema.Index{
				Name: indexName,
				Def:  indexDef,
			}
			indexes = append(indexes, index)
		}

		// constraints(CHECK)

		// triggers
		triggerRows, err := db.Query(`
SELECT name, sql FROM sqlite_master WHERE type = 'trigger' AND tbl_name = ?;
`, tableName)
		defer triggerRows.Close()
		if err != nil {
			return errors.WithStack(err)
		}

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

		table.Columns = columns
		table.Indexes = indexes
		table.Constraints = constraints
		table.Triggers = triggers

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
