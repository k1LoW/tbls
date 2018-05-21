package schema

import (
	"database/sql"
	"fmt"
	"sort"
)

type Index struct {
	Name string
	Def  string
}

type Constrait struct {
	Name string
	Type string
	Def  string
}

type Column struct {
	Name            string
	Type            string
	Nullable        bool
	Default         sql.NullString
	Comment         string
	ParentRelations []*Relation
	ChildRelations  []*Relation
}

type Table struct {
	Name       string
	Type       string
	Comment    string
	Columns    []*Column
	Indexes    []*Index
	Constraits []*Constrait
}

type Relation struct {
	Table         *Table
	Columns       []*Column
	ParentTable   *Table
	ParentColumns []*Column
	Def           string
}

type Schema struct {
	Name      string
	Tables    []*Table
	Relations []*Relation
}

func (s *Schema) FindTableByName(name string) (*Table, error) {
	for _, t := range s.Tables {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, fmt.Errorf("Error: not found table '%s'", name)
}

func (t *Table) FindColumnByName(name string) (*Column, error) {
	for _, c := range t.Columns {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, fmt.Errorf("Error: not found column '%s'", name)
}

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
		sort.SliceStable(t.Constraits, func(i, j int) bool {
			return t.Constraits[i].Name < t.Constraits[j].Name
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
