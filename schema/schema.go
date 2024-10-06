package schema

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/dict"
	"github.com/samber/lo"
)

const (
	TypeFK = "FOREIGN KEY"
)

const (
	ColumnExtraDef    = "ExtraDef"
	ColumnOccurrences = "Occurrences"
	ColumnPercents    = "Percents"
	ColumnChildren    = "Children"
	ColumnParents     = "Parents"
	ColumnComment     = "Comment"
	ColumnLabels      = "Labels"
)

var DefaultHideColumns = []string{ColumnExtraDef, ColumnOccurrences, ColumnPercents, ColumnLabels}
var HideableColumns = []string{ColumnExtraDef, ColumnOccurrences, ColumnPercents, ColumnChildren, ColumnParents, ColumnComment, ColumnLabels}

type Label struct {
	Name    string
	Virtual bool
}

type Labels []*Label

func (labels Labels) Merge(name string) Labels {
	if labels.Contains(name) {
		return labels
	}
	return append(labels, &Label{Name: name, Virtual: true})
}

func (labels Labels) Contains(name string) bool {
	return lo.ContainsBy(labels, func(item *Label) bool {
		return item.Name == name
	})
}

// Viewpoint is the struct for viewpoint information
type Viewpoint struct {
	Name     string            `json:"name,omitempty"`
	Desc     string            `json:"desc,omitempty"`
	Labels   []string          `json:"labels,omitempty"`
	Tables   []string          `json:"tables,omitempty"`
	Distance int               `json:"distance,omitempty"`
	Groups   []*ViewpointGroup `json:"groups,omitempty"`

	Schema *Schema `json:"-"`
}

type ViewpointGroup struct {
	Name   string   `json:"name,omitempty"`
	Desc   string   `json:"desc,omitempty"`
	Labels []string `json:"labels,omitempty"`
	Tables []string `json:"tables,omitempty"`
	Color  string   `json:"color,omitempty"`
}

type Viewpoints []*Viewpoint

func (vs Viewpoints) Merge(in *Viewpoint) Viewpoints {
	for i, v := range vs {
		if sameElements(v.Labels, in.Labels) && sameElements(v.Tables, in.Tables) {
			vs[i] = in
			return vs
		}
		if v.Name == in.Name {
			vs[i] = in
			return vs
		}
	}
	return append(vs, in)
}

// Index is the struct for database index
type Index struct {
	Name    string   `json:"name"`
	Def     string   `json:"def"`
	Table   *string  `json:"table"`
	Columns []string `json:"columns"`
	Comment string   `json:"comment"`
}

// Constraint is the struct for database constraint
type Constraint struct {
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	Def               string   `json:"def"`
	Table             *string  `json:"table"`
	ReferencedTable   *string  `json:"referenced_table" yaml:"referencedTable"`
	Columns           []string `json:"columns"`
	ReferencedColumns []string `json:"referenced_columns" yaml:"referencedColumns"`
	Comment           string   `json:"comment"`
}

// Trigger is the struct for database trigger
type Trigger struct {
	Name    string `json:"name"`
	Def     string `json:"def"`
	Comment string `json:"comment"`
}

// Column is the struct for table column
type Column struct {
	Name            string          `json:"name"`
	Type            string          `json:"type"`
	Nullable        bool            `json:"nullable"`
	Default         sql.NullString  `json:"default"`
	Comment         string          `json:"comment"`
	ExtraDef        string          `json:"extra_def,omitempty" yaml:"extraDef,omitempty"`
	Occurrences     sql.NullInt32   `json:"occurrences,omitempty" yaml:"occurrences,omitempty"`
	Percents        sql.NullFloat64 `json:"percents,omitempty" yaml:"percents,omitempty"`
	Labels          Labels          `json:"labels,omitempty"`
	ParentRelations []*Relation     `json:"-"`
	ChildRelations  []*Relation     `json:"-"`
	PK              bool            `json:"-"`
	FK              bool            `json:"-"`
	HideForER       bool            `json:"-"`
}

type TableViewpoint struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
}

// Table is the struct for database table
type Table struct {
	Name             string            `json:"name"`
	Type             string            `json:"type"`
	Comment          string            `json:"comment"`
	Columns          []*Column         `json:"columns"`
	Viewpoints       []*TableViewpoint `json:"viewpoints"`
	Indexes          []*Index          `json:"indexes"`
	Constraints      []*Constraint     `json:"constraints"`
	Triggers         []*Trigger        `json:"triggers"`
	Def              string            `json:"def"`
	Labels           Labels            `json:"labels,omitempty"`
	ReferencedTables []*Table          `json:"referenced_tables,omitempty" yaml:"referencedTables,omitempty"`
	External         bool              `json:"-"` // Table external to the schema
}

// Relation is the struct for table relation
type Relation struct {
	Table             *Table      `json:"table"`
	Columns           []*Column   `json:"columns"`
	ParentTable       *Table      `json:"parent_table" yaml:"parentTable"`
	ParentColumns     []*Column   `json:"parent_columns" yaml:"parentColumns"`
	Cardinality       Cardinality `json:"cardinality"`
	ParentCardinality Cardinality `json:"parent_cardinality" yaml:"parentCardinality"`
	Def               string      `json:"def"`
	Virtual           bool        `json:"virtual"`
	HideForER         bool        `json:"-"`
}

type DriverMeta struct {
	CurrentSchema string     `json:"current_schema,omitempty" yaml:"currentSchema,omitempty"`
	SearchPaths   []string   `json:"search_paths,omitempty" yaml:"searchPaths,omitempty"`
	Dict          *dict.Dict `json:"dict,omitempty"`
}

// Function is the struct for tbls stored procedure/function information
type Function struct {
	Name       string `json:"name"`
	ReturnType string `json:"return_type" yaml:"returnType"`
	Arguments  string `json:"arguments"`
	Type       string `json:"type"`
}

type Enum struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// Driver is the struct for tbls driver information
type Driver struct {
	Name            string      `json:"name"`
	DatabaseVersion string      `json:"database_version" yaml:"databaseVersion"`
	Meta            *DriverMeta `json:"meta"`
}

// Schema is the struct for database schema
type Schema struct {
	Name       string      `json:"name"`
	Desc       string      `json:"desc"`
	Tables     []*Table    `json:"tables"`
	Relations  []*Relation `json:"relations"`
	Functions  []*Function `json:"functions"`
	Enums      []*Enum     `json:"enums,omitempty"`
	Driver     *Driver     `json:"driver"`
	Labels     Labels      `json:"labels,omitempty"`
	Viewpoints Viewpoints  `json:"viewpoints,omitempty"`
}

func (s *Schema) NormalizeTableName(name string) string {
	if s.Driver != nil && s.Driver.Meta != nil && s.Driver.Meta.CurrentSchema != "" && (s.Driver.Name == "postgres" || s.Driver.Name == "redshift") && !strings.Contains(name, ".") {
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
func (s *Schema) FindTableByName(name string) (_ *Table, err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	for _, t := range s.Tables {
		if s.NormalizeTableName(t.Name) == s.NormalizeTableName(name) {
			return t, nil
		}
	}
	return nil, fmt.Errorf("not found table '%s'", name)
}

// FindRelation find relation by columns and parent columns
func (s *Schema) FindRelation(cs, pcs []*Column) (_ *Relation, err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
L:
	for _, r := range s.Relations {
		if len(r.Columns) != len(cs) || len(r.ParentColumns) != len(pcs) {
			continue
		}
		for _, rc := range r.Columns {
			exist := false
			for _, cc := range cs {
				if rc == cc {
					exist = true
				}
			}
			if !exist {
				continue L
			}
		}
		for _, rc := range r.ParentColumns {
			exist := false
			for _, cc := range pcs {
				if rc == cc {
					exist = true
				}
			}
			if !exist {
				continue L
			}
		}
		return r, nil
	}
	return nil, fmt.Errorf("not found relation '%v, %v'", cs, pcs)
}

func (s *Schema) HasTableWithLabels() bool {
	for _, t := range s.Tables {
		if len(t.Labels) > 0 {
			return true
		}
	}
	return false
}

// Sort schema tables, columns, relations, constrains, and viewpoints
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
	sort.SliceStable(s.Functions, func(i, j int) bool {
		if s.Functions[i].Name != s.Functions[j].Name {
			return s.Functions[i].Name < s.Functions[j].Name
		}
		return s.Functions[i].Arguments < s.Functions[j].Arguments
	})
	sort.SliceStable(s.Viewpoints, func(i, j int) bool {
		return s.Viewpoints[i].Name < s.Viewpoints[j].Name
	})
	return nil
}

// Repair column relations
func (s *Schema) Repair() (err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	if err := s.repairWithoutViewpoints(); err != nil {
		return err
	}
	// viewpoints should be created using as complete a schema as possible
	for _, v := range s.Viewpoints {
		cs, err := s.CloneWithoutViewpoints()
		if err != nil {
			return fmt.Errorf("failed to repair viewpoint: %w", err)
		}
		if err := cs.Filter(&FilterOption{
			Include:       v.Tables,
			IncludeLabels: v.Labels,
			Distance:      v.Distance,
		}); err != nil {
			return fmt.Errorf("failed to repair viewpoint: %w", err)
		}
		v.Schema = cs
	}
	return nil
}

func (s *Schema) Clone() (c *Schema, err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	c = &Schema{}
	if err := json.Unmarshal(b, c); err != nil {
		return nil, err
	}
	if err := c.Repair(); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Schema) CloneWithoutViewpoints() (c *Schema, err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	c = &Schema{}
	if err := json.Unmarshal(b, c); err != nil {
		return nil, err
	}
	if err := c.repairWithoutViewpoints(); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Schema) repairWithoutViewpoints() (err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
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
		for _, ct := range t.Constraints {
			if len(ct.Columns) == 0 {
				ct.Columns = nil
			}
			if len(ct.ReferencedColumns) == 0 {
				ct.ReferencedColumns = nil
			}
		}
		if len(t.Triggers) == 0 {
			t.Triggers = nil
		}
		for i, rt := range t.ReferencedTables {
			tt, err := s.FindTableByName(rt.Name)
			if err != nil {
				rt.External = true
				tt = rt
			}
			t.ReferencedTables[i] = tt
		}
	}

	for _, r := range s.Relations {
		t, err := s.FindTableByName(r.Table.Name)
		if err != nil {
			return fmt.Errorf("failed to repair relation: %w", err)
		}
		for i, rc := range r.Columns {
			c, err := t.FindColumnByName(rc.Name)
			if err != nil {
				return fmt.Errorf("failed to repair relation: %w", err)
			}
			c.ParentRelations = append(c.ParentRelations, r)
			r.Columns[i] = c
		}
		r.Table = t
		pt, err := s.FindTableByName(r.ParentTable.Name)
		if err != nil {
			return fmt.Errorf("failed to repair relation: %w", err)
		}
		for i, rc := range r.ParentColumns {
			pc, err := pt.FindColumnByName(rc.Name)
			if err != nil {
				return fmt.Errorf("failed to repair relation: %w", err)
			}
			pc.ChildRelations = append(pc.ChildRelations, r)
			r.ParentColumns[i] = pc
		}
		r.ParentTable = pt
	}
	if len(s.Functions) == 0 {
		s.Functions = nil
	}

	return nil
}

// FindColumnByName find column by column name
func (t *Table) FindColumnByName(name string) (_ *Column, err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	for _, c := range t.Columns {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("not found column '%s' on table '%s'", name, t.Name))
}

// FindIndexByName find index by index name
func (t *Table) FindIndexByName(name string) (_ *Index, err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	for _, i := range t.Indexes {
		if i.Name == name {
			return i, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("not found index '%s' on table '%s'", name, t.Name))
}

// FindConstraintByName find constraint by constraint name
func (t *Table) FindConstraintByName(name string) (_ *Constraint, err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	for _, c := range t.Constraints {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("not found constraint '%s' on table '%s'", name, t.Name))
}

// FindTriggerByName find trigger by trigger name
func (t *Table) FindTriggerByName(name string) (_ *Trigger, err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	for _, trig := range t.Triggers {
		if trig.Name == name {
			return trig, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("not found trigger '%s' on table '%s'", name, t.Name))
}

// FindConstrainsByColumnName find constraint by column name
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

func (t *Table) hasColumnWithValues(name string) bool {
	for _, c := range t.Columns {
		switch name {
		case ColumnExtraDef:
			if c.ExtraDef != "" {
				return true
			}
		case ColumnOccurrences:
			if c.Occurrences.Valid {
				return true
			}
		case ColumnPercents:
			if c.Percents.Valid {
				return true
			}
		case ColumnChildren:
			if len(c.ChildRelations) > 0 {
				return true
			}
		case ColumnParents:
			if len(c.ParentRelations) > 0 {
				return true
			}
		case ColumnComment:
			if c.Comment != "" {
				return true
			}
		case ColumnLabels:
			if len(c.Labels) > 0 {
				return true
			}
		}
	}
	return false
}

func (t *Table) ShowColumn(name string, hideColumns []string) bool {
	hideColumns = lo.Uniq(append(DefaultHideColumns, hideColumns...))
	if lo.Contains(hideColumns, name) {
		return t.hasColumnWithValues(name)
	}
	return true
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

func sameElements(a, b []string) bool {
	if len(a) == len(b) && lo.Every(a, b) {
		return true
	}
	return false
}
