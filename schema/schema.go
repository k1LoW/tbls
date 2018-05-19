package schema

import (
	"database/sql"
	"fmt"
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
	NotNull         bool
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
