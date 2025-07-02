package schema

import (
	"encoding/json"
)

// SchemaJSON is a JSON representation of schema.Schema.
type SchemaJSON struct { // nolint: revive
	Name       string          `json:"name,omitempty"`
	Desc       string          `json:"desc,omitempty"`
	Tables     []*TableJSON    `json:"tables"`
	Relations  []*RelationJSON `json:"relations,omitempty"`
	Functions  []*Function     `json:"functions,omitempty"`
	Enums      []*Enum         `json:"enums,omitempty"`
	Driver     *DriverJSON     `json:"driver,omitempty"`
	Labels     Labels          `json:"labels,omitempty"`
	Viewpoints Viewpoints      `json:"viewpoints,omitempty"`
}

// TableJSON is a JSON representation of schema.Table.
type TableJSON struct {
	Name             string        `json:"name"`
	Type             string        `json:"type"`
	Comment          string        `json:"comment,omitempty"`
	Columns          []*ColumnJSON `json:"columns"`
	Indexes          []*Index      `json:"indexes,omitempty"`
	Constraints      []*Constraint `json:"constraints,omitempty"`
	Triggers         []*Trigger    `json:"triggers,omitempty"`
	Def              string        `json:"def,omitempty"`
	Labels           Labels        `json:"labels,omitempty"`
	ReferencedTables []string      `json:"referenced_tables,omitempty"`
}

// ColumnJSON is a JSON representation of schema.Column.
type ColumnJSON struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Nullable    bool    `json:"nullable"`
	Default     *string `json:"default,omitempty" jsonschema:"anyof_type=string;null"`
	ExtraDef    string  `json:"extra_def,omitempty"`
	Labels      Labels  `json:"labels,omitempty"`
	Comment     string  `json:"comment,omitempty"`
	LogicalName string  `json:"logical_name,omitempty"`
}

// RelationJSON is a JSON representation of schema.Relation.
type RelationJSON struct {
	Table             string   `json:"table"`
	Columns           []string `json:"columns"`
	Cardinality       string   `json:"cardinality,omitempty" jsonschema:"enum=zero_or_one,enum=exactly_one,enum=zero_or_more,enum=one_or_more,enum="`
	ParentTable       string   `json:"parent_table"`
	ParentColumns     []string `json:"parent_columns"`
	ParentCardinality string   `json:"parent_cardinality,omitempty" jsonschema:"enum=zero_or_one,enum=exactly_one,enum=zero_or_more,enum=one_or_more,enum="`
	Def               string   `json:"def"`
	Virtual           bool     `json:"virtual,omitempty"`
}

type DriverJSON struct {
	Name            string          `json:"name"`
	DatabaseVersion string          `json:"database_version,omitempty" yaml:"databaseVersion,omitempty"`
	Meta            *DriverMetaJSON `json:"meta,omitempty"`
}

type DriverMetaJSON struct {
	CurrentSchema string            `json:"current_schema,omitempty" yaml:"currentSchema,omitempty"`
	SearchPaths   []string          `json:"search_paths,omitempty" yaml:"searchPaths,omitempty"`
	Dict          map[string]string `json:"dict,omitempty"`
}

// ToJSONObjct convert schema.Schema to JSON object.
func (s Schema) ToJSONObject() SchemaJSON {
	var tables []*TableJSON
	for _, t := range s.Tables {
		tt := t.ToJSONObject()
		tables = append(tables, &tt)
	}
	var relations []*RelationJSON
	for _, r := range s.Relations {
		rr := r.ToJSONObject()
		relations = append(relations, &rr)
	}
	return SchemaJSON{
		Name:       s.Name,
		Desc:       s.Desc,
		Tables:     tables,
		Relations:  relations,
		Functions:  s.Functions,
		Enums:      s.Enums,
		Driver:     s.Driver.ToJSONObject(),
		Labels:     s.Labels,
		Viewpoints: s.Viewpoints,
	}
}

func (t Table) ToJSONObject() TableJSON {
	var referencedTables []string
	for _, rt := range t.ReferencedTables {
		referencedTables = append(referencedTables, rt.Name)
	}
	var columns []*ColumnJSON
	for _, c := range t.Columns {
		cc := c.ToJSONObject()
		columns = append(columns, &cc)
	}
	return TableJSON{
		Name:             t.Name,
		Type:             t.Type,
		Comment:          t.Comment,
		Columns:          columns,
		Indexes:          t.Indexes,
		Constraints:      t.Constraints,
		Triggers:         t.Triggers,
		Def:              t.Def,
		Labels:           t.Labels,
		ReferencedTables: referencedTables,
	}
}

func (c Column) ToJSONObject() ColumnJSON {
	var defaultVal *string
	if c.Default.Valid {
		defaultVal = &c.Default.String
	}
	return ColumnJSON{
		Name:        c.Name,
		Type:        c.Type,
		Nullable:    c.Nullable,
		Default:     defaultVal,
		Comment:     c.Comment,
		ExtraDef:    c.ExtraDef,
		Labels:      c.Labels,
		LogicalName: c.LogicalName,
	}
}

func (r Relation) ToJSONObject() RelationJSON {
	var columns []string
	var parentColumns []string
	for _, c := range r.Columns {
		columns = append(columns, c.Name)
	}
	for _, c := range r.ParentColumns {
		parentColumns = append(parentColumns, c.Name)
	}
	return RelationJSON{
		Table:             r.Table.Name,
		Columns:           columns,
		Cardinality:       r.Cardinality.String(),
		ParentTable:       r.ParentTable.Name,
		ParentColumns:     parentColumns,
		ParentCardinality: r.ParentCardinality.String(),
		Def:               r.Def,
		Virtual:           r.Virtual,
	}
}

func (d *Driver) ToJSONObject() *DriverJSON {
	if d == nil {
		return nil
	}
	return &DriverJSON{
		Name:            d.Name,
		DatabaseVersion: d.DatabaseVersion,
		Meta:            d.Meta.ToJSONObject(),
	}
}

func (d *DriverMeta) ToJSONObject() *DriverMetaJSON {
	if d == nil {
		return nil
	}
	m := &DriverMetaJSON{
		CurrentSchema: d.CurrentSchema,
		SearchPaths:   d.SearchPaths,
	}
	if d.Dict != nil {
		m.Dict = d.Dict.Dump()
	}
	return m
}

// MarshalJSON return custom JSON byte.
func (s Schema) MarshalJSON() ([]byte, error) {
	ss := s.ToJSONObject()
	return json.Marshal(&ss)
}

// MarshalJSON return custom JSON byte.
func (t Table) MarshalJSON() ([]byte, error) {
	tt := t.ToJSONObject()
	return json.Marshal(&tt)
}

// MarshalJSON return custom JSON byte.
func (c Column) MarshalJSON() ([]byte, error) {
	cc := c.ToJSONObject()
	return json.Marshal(&cc)
}

// MarshalJSON return custom JSON byte.
func (r Relation) MarshalJSON() ([]byte, error) {
	rr := r.ToJSONObject()
	return json.Marshal(&rr)
}

// UnmarshalJSON unmarshal JSON to schema.Table.
func (t *Table) UnmarshalJSON(data []byte) error {
	s := struct {
		Name             string        `json:"name"`
		Type             string        `json:"type"`
		Comment          string        `json:"comment,omitempty"`
		Columns          []*Column     `json:"columns"`
		Indexes          []*Index      `json:"indexes,omitempty"`
		Constraints      []*Constraint `json:"constraints,omitempty"`
		Triggers         []*Trigger    `json:"triggers,omitempty"`
		Def              string        `json:"def,omitempty"`
		Labels           Labels        `json:"labels,omitempty"`
		ReferencedTables []string      `json:"referenced_tables,omitempty"`
	}{}
	err := json.Unmarshal(data, &s)
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

// UnmarshalJSON unmarshal JSON to schema.Column.
func (c *Column) UnmarshalJSON(data []byte) error {
	s := struct {
		Name        string  `json:"name"`
		Type        string  `json:"type"`
		Nullable    bool    `json:"nullable"`
		Default     *string `json:"default,omitempty"`
		Comment     string  `json:"comment,omitempty"`
		ExtraDef    string  `json:"extra_def,omitempty"`
		Labels      Labels  `json:"labels,omitempty"`
		LogicalName string  `json:"logical_name,omitempty"`
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
	c.ExtraDef = s.ExtraDef
	c.Labels = s.Labels
	c.Comment = s.Comment
	c.LogicalName = s.LogicalName
	return nil
}

// UnmarshalJSON unmarshal JSON to schema.Relation.
func (r *Relation) UnmarshalJSON(data []byte) error {
	s := struct {
		Table             string   `json:"table"`
		Columns           []string `json:"columns"`
		Cardinality       string   `json:"cardinality,omitempty"`
		ParentTable       string   `json:"parent_table"`
		ParentColumns     []string `json:"parent_columns"`
		ParentCardinality string   `json:"parent_cardinality,omitempty"`
		Def               string   `json:"def"`
		Virtual           bool     `json:"virtual,omitempty"`
	}{}
	err := json.Unmarshal(data, &s)
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
	r.Cardinality, err = ToCardinality(s.Cardinality)
	if err != nil {
		return err
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
	r.ParentCardinality, err = ToCardinality(s.ParentCardinality)
	if err != nil {
		return err
	}
	r.Def = s.Def
	r.Virtual = s.Virtual
	return nil
}
