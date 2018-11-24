package schema

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/k1LoW/tbls/config"
	"github.com/pkg/errors"
)

// Index is the struct for database index
type Index struct {
	Name string `json:"name"`
	Def  string `json:"def"`
}

// Constraint is the struct for database constraint
type Constraint struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Def  string `json:"def"`
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
}

// Relation is the struct for table relation
type Relation struct {
	Table         *Table    `json:"table"`
	Columns       []*Column `json:"columns"`
	ParentTable   *Table    `json:"parent_table"`
	ParentColumns []*Column `json:"parent_columns"`
	Def           string    `json:"def"`
	IsAdditional  bool      `json:"is_additional"`
}

// Driver is the struct for tbls driver information
type Driver struct {
	Name            string `json:"name"`
	DatabaseVersion string `json:"database_version"`
}

// Schema is the struct for database schema
type Schema struct {
	Name      string      `json:"name"`
	Tables    []*Table    `json:"tables"`
	Relations []*Relation `json:"relations"`
	Driver    *Driver     `json:"driver"`
}

// MarshalJSON return custom JSON byte
func (c Column) MarshalJSON() ([]byte, error) {
	if c.Default.Valid {
		return json.Marshal(&struct {
			Name            string      `json:"name"`
			Type            string      `json:"type"`
			Nullable        bool        `json:"nullable"`
			Default         string      `json:"default"`
			Comment         string      `json:"comment"`
			ParentRelations []*Relation `json:"-"`
			ChildRelations  []*Relation `json:"-"`
		}{
			Name:            c.Name,
			Type:            c.Type,
			Nullable:        c.Nullable,
			Default:         c.Default.String,
			Comment:         c.Comment,
			ParentRelations: c.ParentRelations,
			ChildRelations:  c.ChildRelations,
		})
	}
	return json.Marshal(&struct {
		Name            string      `json:"name"`
		Type            string      `json:"type"`
		Nullable        bool        `json:"nullable"`
		Default         *string     `json:"default"`
		Comment         string      `json:"comment"`
		ParentRelations []*Relation `json:"-"`
		ChildRelations  []*Relation `json:"-"`
	}{
		Name:            c.Name,
		Type:            c.Type,
		Nullable:        c.Nullable,
		Default:         nil,
		Comment:         c.Comment,
		ParentRelations: c.ParentRelations,
		ChildRelations:  c.ChildRelations,
	})
}

// UnmarshalJSON ...
func (c *Column) UnmarshalJSON(data []byte) error {
	s := struct {
		Name            string      `json:"name"`
		Type            string      `json:"type"`
		Nullable        bool        `json:"nullable"`
		Default         *string     `json:"default"`
		Comment         string      `json:"comment"`
		ParentRelations []*Relation `json:"-"`
		ChildRelations  []*Relation `json:"-"`
	}{}
	json.Unmarshal(data, &s)
	c.Name = s.Name
	c.Type = s.Type
	c.Nullable = s.Nullable
	if s.Default != nil {
		c.Default.Valid = true
		c.Default.String = *s.Default
	} else {
		c.Default.Valid = false
		c.Default.String = ""
	}
	c.Comment = s.Comment
	return nil
}

// FindTableByName find table by table name
func (s *Schema) FindTableByName(name string) (*Table, error) {
	for _, t := range s.Tables {
		if t.Name == name {
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

// LoadAdditionalData load additional data (relations, comments) from yaml file
func (s *Schema) LoadAdditionalData(c *config.Config) error {
	err := addAdditionalRelations(s, c.Relations)
	if err != nil {
		return err
	}
	err = addAdditionalComments(s, c.Comments)
	if err != nil {
		return err
	}

	return nil
}

// Repair column relations
func (s *Schema) Repair() error {
	for _, r := range s.Relations {
		t, err := s.FindTableByName(r.Table.Name)
		if err != nil {
			return errors.Wrap(err, "failed to repair relation")
		}
		for _, rc := range r.Columns {
			c, err := t.FindColumnByName(rc.Name)
			if err != nil {
				return errors.Wrap(err, "failed to repair relation")
			}
			c.ParentRelations = append(c.ParentRelations, r)
			rc = c
		}
		r.Table = t
		pt, err := s.FindTableByName(r.ParentTable.Name)
		if err != nil {
			return errors.Wrap(err, "failed to repair relation")
		}
		for _, rc := range r.ParentColumns {
			pc, err := pt.FindColumnByName(rc.Name)
			if err != nil {
				return errors.Wrap(err, "failed to repair relation")
			}
			pc.ChildRelations = append(pc.ChildRelations, r)
			rc = pc
		}
		r.ParentTable = pt
	}
	return nil
}

func addAdditionalRelations(s *Schema, relations []config.AdditionalRelation) error {
	for _, r := range relations {
		relation := &Relation{
			IsAdditional: true,
		}
		if r.Def != "" {
			relation.Def = r.Def
		} else {
			relation.Def = "Additional Relation"
		}
		var err error
		relation.Table, err = s.FindTableByName(r.Table)
		if err != nil {
			return errors.Wrap(err, "failed to add relation")
		}
		for _, c := range r.Columns {
			column, err := relation.Table.FindColumnByName(c)
			if err != nil {
				return errors.Wrap(err, "failed to add relation")
			}
			relation.Columns = append(relation.Columns, column)
			column.ParentRelations = append(column.ParentRelations, relation)
		}
		relation.ParentTable, err = s.FindTableByName(r.ParentTable)
		if err != nil {
			return errors.Wrap(err, "failed to add relation")
		}
		for _, c := range r.ParentColumns {
			column, err := relation.ParentTable.FindColumnByName(c)
			if err != nil {
				return errors.Wrap(err, "failed to add relation")
			}
			relation.ParentColumns = append(relation.ParentColumns, column)
			column.ChildRelations = append(column.ChildRelations, relation)
		}

		s.Relations = append(s.Relations, relation)
	}
	return nil
}

func addAdditionalComments(s *Schema, comments []config.AdditionalComment) error {
	for _, c := range comments {
		table, err := s.FindTableByName(c.Table)
		if err != nil {
			return errors.Wrap(err, "failed to add table comment")
		}
		if c.TableComment != "" {
			table.Comment = c.TableComment
		}
		for c, comment := range c.ColumnComments {
			column, err := table.FindColumnByName(c)
			if err != nil {
				return errors.Wrap(err, "failed to add column comment")
			}
			column.Comment = comment
		}
	}
	return nil
}
