package sqlite

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"regexp"

	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/ddl"
	"github.com/k1LoW/tbls/schema"
	"github.com/samber/lo"
)

var reFK = regexp.MustCompile(`FOREIGN KEY \((.+)\) REFERENCES ([^\s\)]+)\s?\(([^\)]+)\)`)
var reFTS = regexp.MustCompile(`(?i)USING\s+fts([34])`)

var shadowTables []string

// Sqlite struct
type Sqlite struct {
	db *sql.DB
}

// New return new Sqlite
func New(db *sql.DB) *Sqlite {
	return &Sqlite{
		db: db,
	}
}

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
func (l *Sqlite) Analyze(s *schema.Schema) (err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	d, err := l.Info()
	if err != nil {
		return errors.WithStack(err)
	}
	s.Driver = d

	// tables
	tableRows, err := l.db.Query(`
SELECT name, type, sql
FROM sqlite_master
WHERE name != 'sqlite_sequence' AND (type = 'table' OR type = 'view');`)
	if err != nil {
		return errors.WithStack(err)
	}
	defer tableRows.Close()

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

		if reFTS.MatchString(tableDef) {
			tableType = "virtual table"
			matches := reFTS.FindStringSubmatch(tableDef)
			if len(matches) < 1 {
				return fmt.Errorf("can not parse table definition: %s", tableDef)
			}
			shadowTables = append(shadowTables, fmt.Sprintf("%s_content", tableName))
			shadowTables = append(shadowTables, fmt.Sprintf("%s_segdir", tableName))
			shadowTables = append(shadowTables, fmt.Sprintf("%s_segments", tableName))
			if matches[1] == "4" {
				shadowTables = append(shadowTables, fmt.Sprintf("%s_stat", tableName))
				shadowTables = append(shadowTables, fmt.Sprintf("%s_docsize", tableName))
			}
		}

		table := &schema.Table{
			Name: tableName,
			Type: tableType,
			Def:  tableDef,
		}

		// constraints
		constraints := []*schema.Constraint{}

		// columns
		columnRows, err := l.db.Query(fmt.Sprintf("PRAGMA table_info(`%s`)", tableName))
		if err != nil {
			return errors.WithStack(err)
		}
		defer columnRows.Close()

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
					Name:    columnName,
					Type:    "PRIMARY KEY",
					Def:     constraintDef,
					Table:   &table.Name,
					Columns: []string{columnName},
				}
				constraints = append(constraints, constraint)
			}
		}

		/// foreign keys
		fkMap := map[string]*fk{}
		fkSlice := []*fk{}

		foreignKeyRows, err := l.db.Query(fmt.Sprintf("PRAGMA foreign_key_list(`%s`)", tableName))
		if err != nil {
			return errors.WithStack(err)
		}
		defer foreignKeyRows.Close()
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
				strings.Join(f.ColumnNames, ", "), f.ForeignTableName, strings.Join(f.ForeignColumnNames, ", "), f.OnUpdate, f.OnDelete, f.Match) // #nosec
			constraint := &schema.Constraint{
				Name:              fmt.Sprintf("- (Foreign key ID: %s)", f.ID),
				Type:              schema.TypeFK,
				Def:               foreignKeyDef,
				Table:             &table.Name,
				Columns:           f.ColumnNames,
				ReferencedTable:   &f.ForeignTableName,
				ReferencedColumns: f.ForeignColumnNames,
			}
			relation := &schema.Relation{
				Table: table,
				Def:   foreignKeyDef,
			}
			relations = append(relations, relation)

			constraints = append(constraints, constraint)
		}

		// indexes and constraints(UNIQUE, PRIMARY KEY)
		indexRows, err := l.db.Query(fmt.Sprintf("PRAGMA index_list(`%s`)", tableName))
		if err != nil {
			return errors.WithStack(err)
		}
		defer indexRows.Close()

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

			var (
				colRank            string
				colRankWithinTable string
				col                sql.NullString
				cols               []string
			)
			row, err := l.db.Query(fmt.Sprintf("PRAGMA index_info(`%s`)", indexName))
			if err != nil {
				return errors.WithStack(err)
			}
			for row.Next() {
				err = row.Scan(
					&colRank,
					&colRankWithinTable,
					&col,
				)
				if err != nil {
					return errors.WithStack(err)
				}
				if col.Valid {
					cols = append(cols, col.String)
				}
			}

			switch indexCreatedBy {
			case "c":
				row, err := l.db.Query(`SELECT sql FROM sqlite_master WHERE type = 'index' AND tbl_name = ? AND name = ?;
`, tableName, indexName)
				if err != nil {
					return errors.WithStack(err)
				}
				for row.Next() {
					err = row.Scan(
						&indexDef,
					)
					if err != nil {
						return errors.WithStack(err)
					}
				}
			case "u":
				indexDef = fmt.Sprintf("UNIQUE (%s)", strings.Join(cols, ", "))
				constraint := &schema.Constraint{
					Name:    indexName,
					Type:    "UNIQUE",
					Def:     indexDef,
					Table:   &table.Name,
					Columns: cols,
				}
				constraints = append(constraints, constraint)
			case "pk":
				indexDef = fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(cols, ", "))
				constraint := &schema.Constraint{
					Name:    indexName,
					Type:    "PRIMARY KEY",
					Def:     indexDef,
					Table:   &table.Name,
					Columns: cols,
				}
				constraints = append(constraints, constraint)
			}

			index := &schema.Index{
				Name:    indexName,
				Def:     indexDef,
				Table:   &table.Name,
				Columns: cols,
			}
			indexes = append(indexes, index)
		}

		// triggers
		triggerRows, err := l.db.Query(`
SELECT name, sql FROM sqlite_master WHERE type = 'trigger' AND tbl_name = ?;
`, tableName)
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

		table.Columns = columns
		table.Indexes = indexes

		// constraints(CHECK)
		checkConstraints := parseCheckConstraints(table, tableDef)
		constraints = append(constraints, checkConstraints...)

		table.Constraints = constraints
		table.Triggers = triggers

		tables = append(tables, table)
	}

	filtered := []*schema.Table{}
	for _, t := range tables {
		if !lo.Contains(shadowTables, t.Name) {
			filtered = append(filtered, t)
		}
	}

	s.Tables = filtered

	// Relations
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
		if t.Type != "view" {
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

// Info return schema.Driver
func (l *Sqlite) Info() (*schema.Driver, error) {
	var v string
	row := l.db.QueryRow(`SELECT sqlite_version();`)
	err := row.Scan(&v)
	if err != nil {
		return nil, err
	}

	d := &schema.Driver{
		Name:            "sqlite",
		DatabaseVersion: v,
	}
	return d, nil
}

func convertColumnNullable(str string) bool {
	return str != "1"
}

func parseCheckConstraints(table *schema.Table, sql string) []*schema.Constraint {
	// tokenize
	re := regexp.MustCompile(`\s+`)
	separator := "__SEP__"
	space := "__SP__"
	r1 := strings.NewReplacer("(", fmt.Sprintf("%s(%s", separator, separator), ")", fmt.Sprintf("%s)%s", separator, separator), ",", fmt.Sprintf("%s,%s", separator, separator))
	r2 := strings.NewReplacer(" ", fmt.Sprintf("%s%s%s", separator, space, separator))
	tokens := strings.Split(r1.Replace(r2.Replace(re.ReplaceAllString(sql, " "))), separator)

	r3 := strings.NewReplacer(space, " ")
	constraints := []*schema.Constraint{}
	def := ""
	counter := 0
	for _, v := range tokens {
		if counter == 0 && (v == "CHECK" || v == "check") {
			def = v
			continue
		}
		if def != "" && v == space {
			def = def + v
			continue
		}
		if def != "" && v == "(" {
			def = def + v
			counter = counter + 1
			continue
		}
		if def != "" && v == ")" {
			def = def + v
			counter = counter - 1
			if counter == 0 {
				replaced := r3.Replace(def)
				constraint := &schema.Constraint{
					Name:  "-",
					Type:  "CHECK",
					Def:   replaced,
					Table: &table.Name,
				}
				for _, c := range table.Columns {
					if strings.Count(replaced, c.Name) > strings.Count(replaced, fmt.Sprintf("%s(", c.Name)) { // to distinguish between 'length' and 'length('
						constraint.Columns = append(constraint.Columns, c.Name)
					}
				}

				constraints = append(constraints, constraint)
				def = ""
			}
			continue
		}
		if def != "" && counter > 0 {
			def = def + v
		}
	}

	return constraints
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
