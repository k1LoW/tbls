package schema

import (
	"github.com/goccy/go-yaml"
)

// MarshalYAML return custom JSON byte
func (t Table) MarshalYAML() ([]byte, error) {
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

	referencedTables := []string{}
	for _, rt := range t.ReferencedTables {
		referencedTables = append(referencedTables, rt.Name)
	}

	return yaml.Marshal(&struct {
		Name             string        `yaml:"name"`
		Type             string        `yaml:"type"`
		Comment          string        `yaml:"comment"`
		Columns          []*Column     `yaml:"columns"`
		Indexes          []*Index      `yaml:"indexes"`
		Constraints      []*Constraint `yaml:"constraints"`
		Triggers         []*Trigger    `yaml:"triggers"`
		Def              string        `yaml:"def"`
		Labels           Labels        `yaml:"labels,omitempty"`
		ReferencedTables []string      `yaml:"referencedTables,omitempty"`
	}{
		Name:             t.Name,
		Type:             t.Type,
		Comment:          t.Comment,
		Columns:          t.Columns,
		Indexes:          t.Indexes,
		Constraints:      t.Constraints,
		Triggers:         t.Triggers,
		Def:              t.Def,
		Labels:           t.Labels,
		ReferencedTables: referencedTables,
	})
}

// MarshalYAML return custom YAML byte
func (c Column) MarshalYAML() ([]byte, error) {
	if c.Default.Valid {
		return yaml.Marshal(&struct {
			Name            string      `yaml:"name"`
			Type            string      `yaml:"type"`
			Nullable        bool        `yaml:"nullable"`
			Default         string      `yaml:"default"`
			ExtraDef        string      `yaml:"extraDef,omitempty"`
			Comment         string      `yaml:"comment"`
			ParentRelations []*Relation `yaml:"-"`
			ChildRelations  []*Relation `yaml:"-"`
		}{
			Name:            c.Name,
			Type:            c.Type,
			Nullable:        c.Nullable,
			Default:         c.Default.String,
			Comment:         c.Comment,
			ExtraDef:        c.ExtraDef,
			ParentRelations: c.ParentRelations,
			ChildRelations:  c.ChildRelations,
		})
	}
	return yaml.Marshal(&struct {
		Name            string      `yaml:"name"`
		Type            string      `yaml:"type"`
		Nullable        bool        `yaml:"nullable"`
		Default         *string     `yaml:"default"`
		ExtraDef        string      `yaml:"extraDef,omitempty"`
		Comment         string      `yaml:"comment"`
		ParentRelations []*Relation `yaml:"-"`
		ChildRelations  []*Relation `yaml:"-"`
	}{
		Name:            c.Name,
		Type:            c.Type,
		Nullable:        c.Nullable,
		Default:         nil,
		ExtraDef:        c.ExtraDef,
		Comment:         c.Comment,
		ParentRelations: c.ParentRelations,
		ChildRelations:  c.ChildRelations,
	})
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

// UnMarshalYAML unmarshal JSON to schema.Table
func (t *Table) UnMarshalYAML(data []byte) error {
	s := struct {
		Name             string        `yaml:"name"`
		Type             string        `yaml:"type"`
		Comment          string        `yaml:"comment"`
		Columns          []*Column     `yaml:"columns"`
		Indexes          []*Index      `yaml:"indexes"`
		Constraints      []*Constraint `yaml:"constraints"`
		Triggers         []*Trigger    `yaml:"triggers"`
		Def              string        `yaml:"def"`
		Labels           Labels        `yaml:"labels,omitempty"`
		ReferencedTables []string      `yaml:"referencedTables,omitempty"`
	}{}
	err := yaml.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	t.Name = s.Name
	t.Type = s.Type
	t.Comment = s.Comment
	t.Columns = s.Columns
	t.Indexes = s.Indexes
	t.Constraints = s.Constraints
	t.Triggers = s.Triggers
	t.Def = s.Def
	t.Labels = s.Labels
	for _, rt := range s.ReferencedTables {
		t.ReferencedTables = append(t.ReferencedTables, &Table{
			Name: rt,
		})
	}
	return nil
}

// UnmarshalYAML unmarshal YAML to schema.Column
func (c *Column) UnmarshalYAML(data []byte) error {
	s := struct {
		Name            string      `yaml:"name"`
		Type            string      `yaml:"type"`
		Nullable        bool        `yaml:"nullable"`
		Default         *string     `yaml:"default"`
		Comment         string      `yaml:"comment"`
		ExtraDef        string      `yaml:"extraDef,omitempty"`
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
	c.ExtraDef = s.ExtraDef
	c.Comment = s.Comment
	return nil
}

// UnmarshalYAML unmarshal YAML to schema.Column
func (r *Relation) UnmarshalYAML(data []byte) error {
	s := struct {
		Table         string   `yaml:"table"`
		Columns       []string `yaml:"columns"`
		ParentTable   string   `yaml:"parentTable"`
		ParentColumns []string `yaml:"parentColumns"`
		Def           string   `yaml:"def"`
		Virtual       bool     `yaml:"virtual"`
	}{}
	err := yaml.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	r.Table = &Table{
		Name: s.Table,
	}
	r.Columns = []*Column{}
	for _, c := range s.Columns {
		r.Columns = append(r.Columns, &Column{
			Name: c,
		})
	}
	r.ParentTable = &Table{
		Name: s.ParentTable,
	}
	r.ParentColumns = []*Column{}
	for _, c := range s.ParentColumns {
		r.ParentColumns = append(r.ParentColumns, &Column{
			Name: c,
		})
	}
	r.Def = s.Def
	r.Virtual = s.Virtual
	return nil
}
