package config

import (
	"database/sql"
	"testing"

	"github.com/k1LoW/tbls/schema"
)

func TestRequireTableComment(t *testing.T) {
	tests := []struct {
		enabled bool
		exclude []string
		want    int
	}{
		{true, []string{}, 1},
		{false, []string{}, 0},
		{true, []string{"a"}, 0},
	}

	for i, tt := range tests {
		r := RequireTableComment{
			Enabled: tt.enabled,
			Exclude: tt.exclude,
		}
		s := newTestSchema()
		warns := r.Check(s)
		if len(warns) != tt.want {
			t.Errorf("TestRequireTableComment(%d): actual %v\nwant %v", i, len(warns), tt.want)
		}
	}
}

func TestRequireColumnComment(t *testing.T) {
	tests := []struct {
		enabled       bool
		exclude       []string
		excludeTables []string
		want          int
	}{
		{true, []string{}, []string{}, 1},
		{false, []string{}, []string{}, 0},
		{true, []string{"b1"}, []string{}, 0},
		{true, []string{"b.b1"}, []string{}, 0},
		{true, []string{"a.b1"}, []string{}, 1},
		{true, []string{}, []string{"b"}, 0},
	}

	for i, tt := range tests {
		r := RequireColumnComment{
			Enabled:        tt.enabled,
			Exclude:        tt.exclude,
			ExcludedTables: tt.excludeTables,
		}
		s := newTestSchema()
		warns := r.Check(s)
		if len(warns) != tt.want {
			t.Errorf("TestRequireColumnComment(%d): actual %v\nwant %v", i, len(warns), tt.want)
		}
	}
}

func TestUnrelatedTable(t *testing.T) {
	tests := []struct {
		enabled bool
		exclude []string
		want    int
	}{
		{true, []string{}, 1},
		{false, []string{}, 0},
		{true, []string{"b"}, 1},
		{true, []string{"c"}, 0},
	}

	for i, tt := range tests {
		r := UnrelatedTable{
			Enabled: tt.enabled,
			Exclude: tt.exclude,
		}
		s := newTestSchema()
		warns := r.Check(s)
		if len(warns) != tt.want {
			t.Errorf("TestUnrelatedTable(%d):actual %v\nwant %v", i, len(warns), tt.want)
		}
	}
}

func TestColumnCount(t *testing.T) {
	tests := []struct {
		enabled bool
		exclude []string
		want    int
	}{
		{true, []string{}, 1},
		{false, []string{}, 0},
		{true, []string{"c"}, 0},
	}

	for i, tt := range tests {
		r := ColumnCount{
			Enabled: tt.enabled,
			Exclude: tt.exclude,
			Max:     3,
		}
		s := newTestSchema()
		warns := r.Check(s)
		if len(warns) != tt.want {
			t.Errorf("TestColumnCount(%d): actual %v\nwant %v", i, len(warns), tt.want)
		}
	}
}

func TestRequireColumns(t *testing.T) {
	tests := []struct {
		enabled   bool
		excludeA2 []string
		excludeB2 []string
		want      int
	}{
		{true, []string{}, []string{}, 4},
		{false, []string{}, []string{}, 0},
		{true, []string{"b", "c"}, []string{"a", "c"}, 0},
	}

	for i, tt := range tests {
		r := RequireColumns{
			Enabled: tt.enabled,
			Columns: []RequireColumnsColumn{
				RequireColumnsColumn{
					Name:    "a2",
					Exclude: tt.excludeA2,
				},
				RequireColumnsColumn{
					Name:    "b2",
					Exclude: tt.excludeB2,
				},
			},
		}
		s := newTestSchema()
		warns := r.Check(s)
		if len(warns) != tt.want {
			t.Errorf("TestRequireColumns(%d): actual %v\nwant %v", i, len(warns), tt.want)
		}
	}
}

func TestDuplicateRelations(t *testing.T) {
	r := DuplicateRelations{
		Enabled: true,
	}
	s := newTestSchema()
	copy := *s.Relations[0]
	copy.Def = "copy"
	s.Relations = append(s.Relations, &copy)
	copy2 := *s.Relations[0]
	copy2.Def = "copy2"
	copy2Table := *copy2.Table
	copy2.Table = &copy2Table
	copy2.Table.Name = "other table"
	s.Relations = append(s.Relations, &copy2)
	warns := r.Check(s)
	want := 1
	if len(warns) != want {
		t.Errorf("actual %v\nwant %v", len(warns), want)
	}
}

func TestRequireForeignKeyIndex(t *testing.T) {
	tests := []struct {
		enabled bool
		exclude []string
		want    int
	}{
		{true, []string{}, 1},
		{false, []string{}, 0},
		{true, []string{"a.a1"}, 0},
		{true, []string{"a1"}, 0},
	}

	for i, tt := range tests {
		r := RequireForeignKeyIndex{
			Enabled: tt.enabled,
			Exclude: tt.exclude,
		}
		s := newTestSchema()
		warns := r.Check(s)
		if len(warns) != tt.want {
			t.Errorf("TestRequireForeignKeyIndex(%d): actual %v\nwant %v", i, len(warns), tt.want)
		}
	}
}

func newTestSchema() *schema.Schema {
	ca := &schema.Column{
		Name:     "a1",
		Type:     "bigint(20)",
		Comment:  "column a",
		Nullable: false,
	}
	cb := &schema.Column{
		Name:     "b1",
		Type:     "text",
		Comment:  "", // empty comment
		Nullable: true,
	}

	ta := &schema.Table{
		Name:    "a",
		Type:    "BASE TABLE",
		Comment: "", // empty comment
		Columns: []*schema.Column{
			ca,
			&schema.Column{
				Name:     "a2",
				Type:     "datetime",
				Comment:  "column a2",
				Nullable: false,
				Default: sql.NullString{
					String: "CURRENT_TIMESTAMP",
					Valid:  true,
				},
			},
		},
	}
	tb := &schema.Table{
		Name:    "b",
		Type:    "BASE TABLE",
		Comment: "table b",
		Columns: []*schema.Column{
			cb,
			&schema.Column{
				Name:     "b2",
				Comment:  "column b2",
				Type:     "text",
				Nullable: true,
			},
		},
	}
	tc := &schema.Table{
		Name:    "c",
		Type:    "BASE TABLE",
		Comment: "table c",
		Columns: []*schema.Column{
			&schema.Column{
				Name:     "c1",
				Type:     "text",
				Comment:  "column c1",
				Nullable: false,
			},
			&schema.Column{
				Name:     "c2",
				Type:     "text",
				Comment:  "column c2",
				Nullable: false,
			},
			&schema.Column{
				Name:     "c3",
				Type:     "text",
				Comment:  "column c3",
				Nullable: false,
			},
			&schema.Column{
				Name:     "c4",
				Type:     "text",
				Comment:  "column c4",
				Nullable: false,
			},
		},
	}

	r := &schema.Relation{
		Table:         ta,
		Columns:       []*schema.Column{ca},
		ParentTable:   tb,
		ParentColumns: []*schema.Column{cb},
	}
	ca.ParentRelations = []*schema.Relation{r}
	cb.ChildRelations = []*schema.Relation{r}

	ta.Indexes = []*schema.Index{
		&schema.Index{
			Name:  "a2_idx",
			Def:   "a2 index",
			Table: &ta.Name,
			Columns: []string{
				"a2",
			},
		},
	}

	ta.Constraints = []*schema.Constraint{
		&schema.Constraint{
			Name:             "a1_b1_fk",
			Type:             schema.FOREIGN_KEY,
			Table:            &ta.Name,
			ReferenceTable:   &tb.Name,
			Columns:          []string{"a1"},
			ReferenceColumns: []string{"b1"},
		},
	}

	s := &schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			ta,
			tb,
			tc,
		},
		Relations: []*schema.Relation{
			r,
		},
		Driver: &schema.Driver{
			Name:            "testdriver",
			DatabaseVersion: "1.0.0",
		},
	}
	return s
}
