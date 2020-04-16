package config

import (
	"database/sql"
	"testing"

	"github.com/k1LoW/tbls/schema"
)

func TestRequireTableComment(t *testing.T) {
	tests := []struct {
		enabled     bool
		lintExclude []string
		exclude     []string
		want        int
	}{
		{true, []string{}, []string{}, 1},
		{false, []string{}, []string{}, 0},
		{true, []string{}, []string{"table_a"}, 0},
		{true, []string{"table_a"}, []string{}, 0},
		{true, []string{"*_a"}, []string{}, 0},
	}

	for i, tt := range tests {
		r := RequireTableComment{
			Enabled: tt.enabled,
			Exclude: tt.exclude,
		}
		s := newTestSchema()
		warns := r.Check(s, tt.lintExclude)
		if len(warns) != tt.want {
			t.Errorf("TestRequireTableComment(%d): got %v\nwant %v", i, len(warns), tt.want)
		}
	}
}

func TestRequireColumnComment(t *testing.T) {
	tests := []struct {
		enabled       bool
		lintExclude   []string
		exclude       []string
		excludeTables []string
		want          int
	}{
		{true, []string{}, []string{}, []string{}, 1},
		{false, []string{}, []string{}, []string{}, 0},
		{true, []string{}, []string{"column_b1"}, []string{}, 0},
		{true, []string{}, []string{"table_b.column_b1"}, []string{}, 0},
		{true, []string{}, []string{"table_a.colmun_b1"}, []string{}, 1},
		{true, []string{}, []string{}, []string{"table_b"}, 0},
		{true, []string{"table_b"}, []string{}, []string{}, 0},
		{true, []string{}, []string{"*_b1"}, []string{}, 0},
		{true, []string{}, []string{"table_b.*_b1"}, []string{}, 0},
	}

	for i, tt := range tests {
		r := RequireColumnComment{
			Enabled:        tt.enabled,
			Exclude:        tt.exclude,
			ExcludedTables: tt.excludeTables,
		}
		s := newTestSchema()
		warns := r.Check(s, tt.lintExclude)
		if len(warns) != tt.want {
			t.Errorf("TestRequireColumnComment(%d): got %v\nwant %v", i, len(warns), tt.want)
		}
	}
}

func TestUnrelatedTable(t *testing.T) {
	tests := []struct {
		enabled     bool
		lintExclude []string
		exclude     []string
		want        int
		wantMsg     string
	}{
		{true, []string{}, []string{}, 1, "unrelated (isolated) table exists. [table_c]"},
		{false, []string{}, []string{}, 0, ""},
		{true, []string{}, []string{"table_b"}, 1, "unrelated (isolated) table exists. [table_c]"},
		{true, []string{}, []string{"table_c"}, 0, ""},
		{true, []string{"table_c"}, []string{}, 0, ""},
		{true, []string{}, []string{"*_c"}, 0, ""},
		{true, []string{"*_c"}, []string{}, 0, ""},
	}

	for i, tt := range tests {
		r := UnrelatedTable{
			Enabled: tt.enabled,
			Exclude: tt.exclude,
		}
		s := newTestSchema()
		warns := r.Check(s, tt.lintExclude)
		if len(warns) != tt.want {
			t.Errorf("TestUnrelatedTable(%d): got %v\nwant %v", i, len(warns), tt.want)
		}
		if len(warns) == 0 {
			continue
		}
		if warns[0].Message != tt.wantMsg {
			t.Errorf("TestUnrelatedTable(%d): got %v\nwant %v", i, warns[0].Message, tt.wantMsg)
		}
	}
}

func TestColumnCount(t *testing.T) {
	tests := []struct {
		enabled     bool
		lintExclude []string
		exclude     []string
		want        int
	}{
		{true, []string{}, []string{}, 1},
		{false, []string{}, []string{}, 0},
		{true, []string{}, []string{"table_c"}, 0},
		{true, []string{"table_c"}, []string{}, 0},
		{true, []string{}, []string{"*_c"}, 0},
		{true, []string{"*_c"}, []string{}, 0},
	}

	for i, tt := range tests {
		r := ColumnCount{
			Enabled: tt.enabled,
			Exclude: tt.exclude,
			Max:     3,
		}
		s := newTestSchema()
		warns := r.Check(s, tt.lintExclude)
		if len(warns) != tt.want {
			t.Errorf("TestColumnCount(%d): got %v\nwant %v", i, len(warns), tt.want)
		}
	}
}

func TestRequireColumns(t *testing.T) {
	tests := []struct {
		enabled     bool
		lintExclude []string
		excludeA2   []string
		excludeB2   []string
		want        int
	}{
		{true, []string{}, []string{}, []string{}, 4},
		{false, []string{}, []string{}, []string{}, 0},
		{true, []string{}, []string{"table_c"}, []string{"table_c"}, 2},
		{true, []string{}, []string{"table_b", "table_c"}, []string{"table_a", "table_c"}, 0},
		{true, []string{"table_c"}, []string{}, []string{}, 2},
		{true, []string{}, []string{"table_*"}, []string{"table_*"}, 0},
	}

	for i, tt := range tests {
		r := RequireColumns{
			Enabled: tt.enabled,
			Columns: []RequireColumnsColumn{
				RequireColumnsColumn{
					Name:    "column_a2",
					Exclude: tt.excludeA2,
				},
				RequireColumnsColumn{
					Name:    "column_b2",
					Exclude: tt.excludeB2,
				},
			},
		}
		s := newTestSchema()
		warns := r.Check(s, tt.lintExclude)
		if len(warns) != tt.want {
			t.Errorf("TestRequireColumns(%d): got %v\nwant %v", i, len(warns), tt.want)
		}
	}
}

func TestDuplicateRelations(t *testing.T) {
	tests := []struct {
		enabled     bool
		lintExclude []string
		want        int
	}{
		{true, []string{}, 1},
		{false, []string{}, 0},
		{true, []string{"table_a"}, 0},
		{true, []string{"*_a"}, 0},
	}

	for i, tt := range tests {
		r := DuplicateRelations{
			Enabled: tt.enabled,
		}
		s := newTestSchema()
		copy := *s.Relations[0]
		copy.Def = "copy"
		s.Relations = append(s.Relations, &copy)
		copy2 := *s.Relations[0]
		copy2.Def = "copy2"
		copy2Table := *copy2.Table
		copy2.Table = &copy2Table
		copy2.Table.Name = "other_table"
		s.Relations = append(s.Relations, &copy2)
		warns := r.Check(s, tt.lintExclude)
		if len(warns) != tt.want {
			t.Errorf("TestDuplicateRelations(%d): got %v\nwant %v", i, len(warns), tt.want)
		}
	}
}

func TestRequireForeignKeyIndex(t *testing.T) {
	tests := []struct {
		enabled     bool
		lintExclude []string
		exclude     []string
		want        int
	}{
		{true, []string{}, []string{}, 1},
		{false, []string{}, []string{}, 0},
		{true, []string{}, []string{"table_a.column_a1"}, 0},
		{true, []string{}, []string{"column_a1"}, 0},
		{true, []string{"table_a"}, []string{}, 0},
		{true, []string{}, []string{"*_a1"}, 0},
		{true, []string{"*_a"}, []string{}, 0},
	}

	for i, tt := range tests {
		r := RequireForeignKeyIndex{
			Enabled: tt.enabled,
			Exclude: tt.exclude,
		}
		s := newTestSchema()
		warns := r.Check(s, tt.lintExclude)
		if len(warns) != tt.want {
			t.Errorf("TestRequireForeignKeyIndex(%d): got %v\nwant %v", i, len(warns), tt.want)
		}
	}
}

func TestLabelStyleBigQuery(t *testing.T) {
	tests := []struct {
		enabled     bool
		lintExclude []string
		want        int
	}{
		{true, []string{}, 2},
		{false, []string{}, 0},
		{true, []string{"table_a"}, 1},
	}
	for i, tt := range tests {
		r := LabelStyleBigQuery{
			Enabled: tt.enabled,
		}
		s := newTestSchema()
		warns := r.Check(s, tt.lintExclude)
		if len(warns) != tt.want {
			t.Errorf("TestLabelStyleBigQuery(%d): got %v\nwant %v", i, len(warns), tt.want)
		}
	}
}

func TestCheckLabelStyleBigQuery(t *testing.T) {
	tests := []struct {
		label string
		want  bool
	}{
		{"env:prod", true},
		{"env:", true},
		{"e:p", true},
		{"env", false},
		{":prod", false},
		{"Env:prod", false},
		{"0nv:prod", false},
		{"env:0rod", true},
		{"-nv:prod", false},
		{"env:-rod", true},
		{"(nv:prod", false},
		{"en v:prod", false},
		{"env:pr od", false},
		{"env:テスト", true},
		{"e変数:テスト", true},
	}

	for _, tt := range tests {
		got := checkLabelStyleBigQuery(tt.label)
		if got != tt.want {
			t.Errorf("%v got %v want %v", tt.label, got, tt.want)
		}
	}
}

func newTestSchema() *schema.Schema {
	ca := &schema.Column{
		Name:     "column_a1",
		Type:     "bigint(20)",
		Comment:  "column a",
		Nullable: false,
	}
	cb := &schema.Column{
		Name:     "column_b1",
		Type:     "text",
		Comment:  "", // empty comment
		Nullable: true,
	}

	ta := &schema.Table{
		Name: "table_a",
		Labels: schema.Labels{
			&schema.Label{Name: "bq-invalid", Virtual: false},
		},
		Type:    "BASE TABLE",
		Comment: "", // empty comment
		Columns: []*schema.Column{
			ca,
			&schema.Column{
				Name:     "column_a2",
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
		Name:    "table_b",
		Type:    "BASE TABLE",
		Comment: "table b",
		Columns: []*schema.Column{
			cb,
			&schema.Column{
				Name:     "column_b2",
				Comment:  "column b2",
				Type:     "text",
				Nullable: true,
			},
		},
	}
	tc := &schema.Table{
		Name:    "table_c",
		Type:    "BASE TABLE",
		Comment: "table c",
		Columns: []*schema.Column{
			&schema.Column{
				Name:     "column_c1",
				Type:     "text",
				Comment:  "column c1",
				Nullable: false,
			},
			&schema.Column{
				Name:     "column_c2",
				Type:     "text",
				Comment:  "column c2",
				Nullable: false,
			},
			&schema.Column{
				Name:     "column_c3",
				Type:     "text",
				Comment:  "column c3",
				Nullable: false,
			},
			&schema.Column{
				Name:     "column_c4",
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
				"column_a2",
			},
		},
	}

	ta.Constraints = []*schema.Constraint{
		&schema.Constraint{
			Name:             "a1_b1_fk",
			Type:             schema.TypeFK,
			Table:            &ta.Name,
			ReferenceTable:   &tb.Name,
			Columns:          []string{"column_a1"},
			ReferenceColumns: []string{"column_b1"},
		},
		&schema.Constraint{
			Name:           "a1_unique",
			Type:           "UNIQUE",
			Table:          &ta.Name,
			ReferenceTable: nil,
			Columns:        []string{"column_a1"},
		},
	}

	s := &schema.Schema{
		Name: "testschema",
		Labels: schema.Labels{
			&schema.Label{Name: "bq-invalid", Virtual: false},
		},
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
