package schema

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
)

const (
	TypeFK = "FOREIGN KEY"
)

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
}

// Relation is the struct for table relation
type Relation struct {
	Table         *Table    `json:"table"`
	Columns       []*Column `json:"columns"`
	ParentTable   *Table    `json:"parent_table" yaml:"parentTable"`
	ParentColumns []*Column `json:"parent_columns" yaml:"parentColumns"`
	Def           string    `json:"def"`
	Virtual       bool      `json:"virtual" yaml:"virtual"`
}

// Driver is the struct for tbls driver information
type Driver struct {
	Name            string `json:"name"`
	DatabaseVersion string `json:"database_version" yaml:"databaseVersion"`
}

// Schema is the struct for database schema
type Schema struct {
	Name      string      `json:"name"`
	Tables    []*Table    `json:"tables"`
	Relations []*Relation `json:"relations"`
	Driver    *Driver     `json:"driver"`
}

// MarshalJSON return custom JSON byte
func (s Schema) MarshalJSON() ([]byte, error) {
	if len(s.Tables) == 0 {
		s.Tables = []*Table{}
	}
	if len(s.Relations) == 0 {
		s.Relations = []*Relation{}
	}
	return json.Marshal(&struct {
		Name      string      `json:"name"`
		Tables    []*Table    `json:"tables"`
		Relations []*Relation `json:"relations"`
		Driver    *Driver     `json:"driver"`
	}{
		Name:      s.Name,
		Tables:    s.Tables,
		Relations: s.Relations,
		Driver:    s.Driver,
	})
}

// MarshalJSON return custom JSON byte
func (t Table) MarshalJSON() ([]byte, error) {
	if len(t.Columns) == 0 {
		t.Columns = []*Column{}
	}
	if len(t.Indexes) == 0 {
		t.Indexes = []*Index{}
	}
	if len(t.Constraints) == 0 {
		t.Constraints = []*Constraint{}
	}
	if len(t.Triggers) == 0 {
		t.Triggers = []*Trigger{}
	}

	return json.Marshal(&struct {
		Name        string        `json:"name"`
		Type        string        `json:"type"`
		Comment     string        `json:"comment"`
		Columns     []*Column     `json:"columns"`
		Indexes     []*Index      `json:"indexes"`
		Constraints []*Constraint `json:"constraints"`
		Triggers    []*Trigger    `json:"triggers"`
		Def         string        `json:"def"`
	}{
		Name:        t.Name,
		Type:        t.Type,
		Comment:     t.Comment,
		Columns:     t.Columns,
		Indexes:     t.Indexes,
		Constraints: t.Constraints,
		Triggers:    t.Triggers,
		Def:         t.Def,
	})
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

// MarshalJSON return custom JSON byte
func (r Relation) MarshalJSON() ([]byte, error) {
	columns := []string{}
	parentColumns := []string{}
	for _, c := range r.Columns {
		columns = append(columns, c.Name)
	}
	for _, c := range r.ParentColumns {
		parentColumns = append(parentColumns, c.Name)
	}

	return json.Marshal(&struct {
		Table         string   `json:"table"`
		Columns       []string `json:"columns"`
		ParentTable   string   `json:"parent_table"`
		ParentColumns []string `json:"parent_columns"`
		Def           string   `json:"def"`
		Virtual       bool     `json:"virtual"`
	}{
		Table:         r.Table.Name,
		Columns:       columns,
		ParentTable:   r.ParentTable.Name,
		ParentColumns: parentColumns,
		Def:           r.Def,
		Virtual:       r.Virtual,
	})
}

// UnmarshalJSON unmarshal JSON to schema.Column
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
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
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

// MarshalYAML return custom YAML byte
func (c Column) MarshalYAML() ([]byte, error) {
	if c.Default.Valid {
		return yaml.Marshal(&struct {
			Name            string      `yaml:"name"`
			Type            string      `yaml:"type"`
			Nullable        bool        `yaml:"nullable"`
			Default         string      `yaml:"default"`
			Comment         string      `yaml:"comment"`
			ParentRelations []*Relation `yaml:"-"`
			ChildRelations  []*Relation `yaml:"-"`
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
	return yaml.Marshal(&struct {
		Name            string      `yaml:"name"`
		Type            string      `yaml:"type"`
		Nullable        bool        `yaml:"nullable"`
		Default         *string     `yaml:"default"`
		Comment         string      `yaml:"comment"`
		ParentRelations []*Relation `yaml:"-"`
		ChildRelations  []*Relation `yaml:"-"`
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

// UnmarshalYAML unmarshal YAML to schema.Column
func (c *Column) UnmarshalYAML(data []byte) error {
	s := struct {
		Name            string      `yaml:"name"`
		Type            string      `yaml:"type"`
		Nullable        bool        `yaml:"nullable"`
		Default         *string     `yaml:"default"`
		Comment         string      `yaml:"comment"`
		ParentRelations []*Relation `yaml:"-"`
		ChildRelations  []*Relation `yaml:"-"`
	}{}
	err := yaml.Unmarshal(data, &s)
	if err != nil {
		return err
	}
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

// MarshalYAML return custom YAML byte
func (r Relation) MarshalYAML() ([]byte, error) {
	columns := []string{}
	parentColumns := []string{}
	for _, c := range r.Columns {
		columns = append(columns, c.Name)
	}
	for _, c := range r.ParentColumns {
		parentColumns = append(parentColumns, c.Name)
	}

	return yaml.Marshal(&struct {
		Table         string   `yaml:"table"`
		Columns       []string `yaml:"columns"`
		ParentTable   string   `yaml:"parentTable"`
		ParentColumns []string `yaml:"parentColumns"`
		Def           string   `yaml:"def"`
		Virtual       bool     `yaml:"virtual"`
	}{
		Table:         r.Table.Name,
		Columns:       columns,
		ParentTable:   r.ParentTable.Name,
		ParentColumns: parentColumns,
		Def:           r.Def,
		Virtual:       r.Virtual,
	})
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
