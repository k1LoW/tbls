package schema

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/k1LoW/tbls/dict"
	"github.com/pkg/errors"
)

const (
	TypeFK = "FOREIGN KEY"
)

type Label struct {
	Name    string
	Virtual bool
}

type Labels []*Label

func (labels Labels) Merge(name string) Labels {
	for _, l := range labels {
		if l.Name == name {
			return labels
		}
	}
	return append(labels, &Label{Name: name, Virtual: true})
}

// Index is the struct for database index
type Index struct {
	Name    string   `json:"name"`
	Def     string   `json:"def"`
	Table   *string  `json:"table"`
	Columns []string `json:"columns"`
}

// Constraint is the struct for database constraint
type Constraint struct {
	Name             string   `json:"name"`
	Type             string   `json:"type"`
	Def              string   `json:"def"`
	Table            *string  `json:"table"`
	ReferenceTable   *string  `json:"reference_table" yaml:"referenceTable"`
	Columns          []string `json:"columns"`
	ReferenceColumns []string `json:"reference_columns" yaml:"referenceColumns"`
}

// Trigger is the struct for database trigger
type Trigger struct {
	Name string `json:"name"`
	Def  string `json:"def"`
}

// Column is the struct for table column
type Column struct {
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	Nullable        bool           `json:"nullable"`
	Default         sql.NullString `json:"default"`
	Comment         string         `json:"comment"`
	ParentRelations []*Relation    `json:"-"`
	ChildRelations  []*Relation    `json:"-"`
}

// Table is the struct for database table
type Table struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Comment     string        `json:"comment"`
	Columns     []*Column     `json:"columns"`
	Indexes     []*Index      `json:"indexes"`
	Constraints []*Constraint `json:"constraints"`
	Triggers    []*Trigger    `json:"triggers"`
	Def         string        `json:"def"`
	Labels      Labels        `json:"labels,omitempty"`
}

// Relation is the struct for table relation
type Relation struct {
	Table         *Table    `json:"table"`
	Columns       []*Column `json:"columns"`
	ParentTable   *Table    `json:"parent_table" yaml:"parentTable"`
	ParentColumns []*Column `json:"parent_columns" yaml:"parentColumns"`
	Def           string    `json:"def"`
	Virtual       bool      `json:"virtual"`
}

type DriverMeta struct {
	CurrentSchema string     `json:"current_schema,omitempty" yaml:"currentSchema,omitempty"`
	SearchPaths   []string   `json:"search_paths,omitempty" yaml:"searchPaths,omitempty"`
	Dict          *dict.Dict `json:"dict,omitempty"`
}

// Driver is the struct for tbls driver information
type Driver struct {
	Name            string      `json:"name"`
	DatabaseVersion string      `json:"database_version" yaml:"databaseVersion"`
	Meta            *DriverMeta `json:"meta"`
}

// Schema is the struct for database schema
type Schema struct {
	Name      string      `json:"name"`
	Desc      string      `json:"desc"`
	Tables    []*Table    `json:"tables"`
	Relations []*Relation `json:"relations"`
	Driver    *Driver     `json:"driver"`
	Labels    Labels      `json:"labels,omitempty"`
}

func (s *Schema) NormalizeTableName(name string) string {
	if s.Driver != nil && (s.Driver.Name == "postgres" || s.Driver.Name == "redshift") && !strings.Contains(name, ".") {
		return fmt.Sprintf("%s.%s", s.Driver.Meta.CurrentSchema, name)
	}
	return name
}

func (s *Schema) NormalizeTableNames(names []string) []string {
	for i, n := range names {
		names[i] = s.NormalizeTableName(n)
	}
	return names
}

// FindTableByName find table by table name
func (s *Schema) FindTableByName(name string) (*Table, error) {
	for _, t := range s.Tables {
		if s.NormalizeTableName(t.Name) == s.NormalizeTableName(name) {
			return t, nil
		}
	}
	return nil, errors.WithStack(fmt.Errorf("not found table '%s'", name))
}

// FindColumnByName find column by column name
func (t *Table) FindColumnByName(name string) (*Column, error) {
	for _, c := range t.Columns {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, errors.WithStack(fmt.Errorf("not found column '%s.%s'", t.Name, name))
}

func (t *Table) FindConstrainsByColumnName(name string) []*Constraint {
	cts := []*Constraint{}
	for _, ct := range t.Constraints {
		for _, ctc := range ct.Columns {
			if ctc == name {
				cts = append(cts, ct)
			}
		}
	}
	return cts
}

// Sort schema tables, columns, relations, and constrains
func (s *Schema) Sort() error {
	for _, t := range s.Tables {
		for _, c := range t.Columns {
			sort.SliceStable(c.ParentRelations, func(i, j int) bool {
				return c.ParentRelations[i].Table.Name < c.ParentRelations[j].Table.Name
			})
			sort.SliceStable(c.ChildRelations, func(i, j int) bool {
				return c.ChildRelations[i].Table.Name < c.ChildRelations[j].Table.Name
			})
		}
		sort.SliceStable(t.Columns, func(i, j int) bool {
			return t.Columns[i].Name < t.Columns[j].Name
		})
		sort.SliceStable(t.Indexes, func(i, j int) bool {
			return t.Indexes[i].Name < t.Indexes[j].Name
		})
		sort.SliceStable(t.Constraints, func(i, j int) bool {
			return t.Constraints[i].Name < t.Constraints[j].Name
		})
		sort.SliceStable(t.Triggers, func(i, j int) bool {
			return t.Triggers[i].Name < t.Triggers[j].Name
		})
	}
	sort.SliceStable(s.Tables, func(i, j int) bool {
		return s.Tables[i].Name < s.Tables[j].Name
	})
	sort.SliceStable(s.Relations, func(i, j int) bool {
		return s.Relations[i].Table.Name < s.Relations[j].Table.Name
	})
	return nil
}

// Repair column relations
func (s *Schema) Repair() error {
	for _, t := range s.Tables {
		if len(t.Columns) == 0 {
			t.Columns = nil
		}
		if len(t.Indexes) == 0 {
			t.Indexes = nil
		}
		if len(t.Constraints) == 0 {
			t.Constraints = nil
		}
		if len(t.Triggers) == 0 {
			t.Triggers = nil
		}
	}

	for _, r := range s.Relations {
		t, err := s.FindTableByName(r.Table.Name)
		if err != nil {
			return errors.Wrap(err, "failed to repair relation")
		}
		for i, rc := range r.Columns {
			c, err := t.FindColumnByName(rc.Name)
			if err != nil {
				return errors.Wrap(err, "failed to repair relation")
			}
			c.ParentRelations = append(c.ParentRelations, r)
			r.Columns[i] = c
		}
		r.Table = t
		pt, err := s.FindTableByName(r.ParentTable.Name)
		if err != nil {
			return errors.Wrap(err, "failed to repair relation")
		}
		for i, rc := range r.ParentColumns {
			pc, err := pt.FindColumnByName(rc.Name)
			if err != nil {
				return errors.Wrap(err, "failed to repair relation")
			}
			pc.ChildRelations = append(pc.ChildRelations, r)
			r.ParentColumns[i] = pc
		}
		r.ParentTable = pt
	}
	return nil
}

func (t *Table) CollectTablesAndRelations(distance int, root bool) ([]*Table, []*Relation, error) {
	tables := []*Table{}
	relations := []*Relation{}
	tables = append(tables, t)
	if distance == 0 {
		return tables, relations, nil
	}
	distance = distance - 1
	for _, c := range t.Columns {
		for _, r := range c.ParentRelations {
			relations = append(relations, r)
			ts, rs, err := r.ParentTable.CollectTablesAndRelations(distance, false)
			if err != nil {
				return nil, nil, err
			}
			tables = append(tables, ts...)
			relations = append(relations, rs...)
		}
		for _, r := range c.ChildRelations {
			relations = append(relations, r)
			ts, rs, err := r.Table.CollectTablesAndRelations(distance, false)
			if err != nil {
				return nil, nil, err
			}
			tables = append(tables, ts...)
			relations = append(relations, rs...)
		}
	}

	if !root {
		return tables, relations, nil
	}

	uTables := []*Table{}
	encounteredT := make(map[string]bool)
	for _, t := range tables {
		if !encounteredT[t.Name] {
			encounteredT[t.Name] = true
			uTables = append(uTables, t)
		}
	}

	uRelations := []*Relation{}
	encounteredR := make(map[*Relation]bool)
	for _, r := range relations {
		if !encounteredR[r] {
			encounteredR[r] = true
			if !encounteredT[r.ParentTable.Name] || !encounteredT[r.Table.Name] {
				continue
			}
			uRelations = append(uRelations, r)
		}
	}

	return uTables, uRelations, nil
}
