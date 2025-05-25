package config

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/SouhlInc/tbls/schema"
)

func TestRequireTableComment(t *testing.T) {
	tests := []struct {
		enabled       bool
		allOrNothing  bool
		lintExclude   []string
		exclude       []string
		want          int
		wantNoComment int
	}{
		{true, false, []string{}, []string{}, 1, 3},
		{false, false, []string{}, []string{}, 0, 0},
		{true, false, []string{}, []string{"table_a"}, 0, 2},
		{true, false, []string{"table_a"}, []string{}, 0, 2},
		{true, false, []string{"*_a"}, []string{}, 0, 2},

		{true, true, []string{}, []string{}, 1, 0},
		{false, true, []string{}, []string{}, 0, 0},
		{true, true, []string{}, []string{"table_a"}, 0, 0},
		{true, true, []string{"table_a"}, []string{}, 0, 0},
		{true, true, []string{"*_a"}, []string{}, 0, 0},
	}

	for i, tt := range tests {
		r := RequireTableComment{
			Enabled:      tt.enabled,
			AllOrNothing: tt.allOrNothing,
			Exclude:      tt.exclude,
		}
		s := newTestSchema(t)
		if warns := r.Check(s, tt.lintExclude); len(warns) != tt.want {
			t.Errorf("TestRequireTableComment(%d): got %v\nwant %v", i, len(warns), tt.want)
		}
		ns := newTestNoCommentSchema(t)
		if warns := r.Check(ns, tt.lintExclude); len(warns) != tt.wantNoComment {
			t.Errorf("TestRequireTableComment(%d) (no comment schema): got %v\nwant %v", i, len(warns), tt.wantNoComment)
		}
	}
}

func TestRequireColumnComment(t *testing.T) {
	tests := []struct {
		enabled       bool
		allOrNothing  bool
		lintExclude   []string
		exclude       []string
		excludeTables []string
		want          int
		wantNoComment int
	}{
		{true, false, []string{}, []string{}, []string{}, 1, 8},
		{false, false, []string{}, []string{}, []string{}, 0, 0},
		{true, false, []string{}, []string{"column_b1"}, []string{}, 0, 7},
		{true, false, []string{}, []string{"table_b.column_b1"}, []string{}, 0, 7},
		{true, false, []string{}, []string{"table_a.colmun_b1"}, []string{}, 1, 8},
		{true, false, []string{}, []string{}, []string{"table_b"}, 0, 6},
		{true, false, []string{"table_b"}, []string{}, []string{}, 0, 6},
		{true, false, []string{}, []string{"*_b1"}, []string{}, 0, 7},
		{true, false, []string{}, []string{"table_b.*_b1"}, []string{}, 0, 7},

		{true, true, []string{}, []string{}, []string{}, 1, 0},
		{false, true, []string{}, []string{}, []string{}, 0, 0},
		{true, true, []string{}, []string{"column_b1"}, []string{}, 0, 0},
		{true, true, []string{}, []string{"table_b.column_b1"}, []string{}, 0, 0},
		{true, true, []string{}, []string{"table_a.colmun_b1"}, []string{}, 1, 0},
		{true, true, []string{}, []string{}, []string{"table_b"}, 0, 0},
		{true, true, []string{"table_b"}, []string{}, []string{}, 0, 0},
		{true, true, []string{}, []string{"*_b1"}, []string{}, 0, 0},
		{true, true, []string{}, []string{"table_b.*_b1"}, []string{}, 0, 0},
	}

	for i, tt := range tests {
		r := RequireColumnComment{
			Enabled:       tt.enabled,
			AllOrNothing:  tt.allOrNothing,
			Exclude:       tt.exclude,
			ExcludeTables: tt.excludeTables,
		}
		s := newTestSchema(t)
		if warns := r.Check(s, tt.lintExclude); len(warns) != tt.want {
			t.Errorf("TestRequireColumnComment(%d): got %v\nwant %v", i, len(warns), tt.want)
		}

		ns := newTestNoCommentSchema(t)
		if warns := r.Check(ns, tt.lintExclude); len(warns) != tt.wantNoComment {
			t.Errorf("TestRequireColumnComment(%d) (no comment schema): got %v\nwant %v", i, len(warns), tt.wantNoComment)
		}
	}
}

func TestRequireIndexComment(t *testing.T) {
	tests := []struct {
		enabled       bool
		allOrNothing  bool
		lintExclude   []string
		exclude       []string
		excludeTables []string
		want          int
		wantNoComment int
	}{
		{true, false, []string{}, []string{}, []string{}, 1, 1},
		{false, false, []string{}, []string{}, []string{}, 0, 0},
		{true, false, []string{"table_a"}, []string{}, []string{}, 0, 0},
		{true, false, []string{}, []string{"a2_idx"}, []string{}, 0, 0},
		{true, false, []string{}, []string{"table_a.a2_idx"}, []string{}, 0, 0},
		{true, false, []string{}, []string{}, []string{"table_a"}, 0, 0},

		{true, true, []string{}, []string{}, []string{}, 0, 0},
		{false, true, []string{}, []string{}, []string{}, 0, 0},
		{true, true, []string{"table_a"}, []string{}, []string{}, 0, 0},
		{true, true, []string{}, []string{"a2_idx"}, []string{}, 0, 0},
		{true, true, []string{}, []string{"table_a.a2_idx"}, []string{}, 0, 0},
		{true, true, []string{}, []string{}, []string{"table_a"}, 0, 0},
	}

	for i, tt := range tests {
		r := RequireIndexComment{
			Enabled:       tt.enabled,
			AllOrNothing:  tt.allOrNothing,
			Exclude:       tt.exclude,
			ExcludeTables: tt.excludeTables,
		}
		s := newTestSchema(t)
		if warns := r.Check(s, tt.lintExclude); len(warns) != tt.want {
			t.Errorf("TestRequireIndexComment(%d): got %v\nwant %v", i, len(warns), tt.want)
		}

		ns := newTestNoCommentSchema(t)
		if warns := r.Check(ns, tt.lintExclude); len(warns) != tt.wantNoComment {
			t.Errorf("TestRequireIndexComment(%d) (no comment schema): got %v\nwant %v", i, len(warns), tt.wantNoComment)
		}
	}
}

func TestRequireConstraintComment(t *testing.T) {
	tests := []struct {
		enabled       bool
		allOrNothing  bool
		lintExclude   []string
		exclude       []string
		excludeTables []string
		want          int
		wantNoComment int
	}{
		{true, false, []string{}, []string{}, []string{}, 1, 2},
		{false, false, []string{}, []string{}, []string{}, 0, 0},
		{true, false, []string{"table_a"}, []string{}, []string{}, 0, 0},
		{true, false, []string{}, []string{"a1_b1_fk"}, []string{}, 1, 1},
		{true, false, []string{}, []string{"a1_unique"}, []string{}, 0, 1},
		{true, false, []string{}, []string{"table_a.a1_b1_fk"}, []string{}, 1, 1},

		{true, true, []string{}, []string{}, []string{}, 1, 0},
		{false, true, []string{}, []string{}, []string{}, 0, 0},
		{true, true, []string{"table_a"}, []string{}, []string{}, 0, 0},
		{true, true, []string{}, []string{"a1_b1_fk"}, []string{}, 0, 0},
		{true, true, []string{}, []string{"a1_unique"}, []string{}, 0, 0},
		{true, true, []string{}, []string{"table_a.a1_b1_fk"}, []string{}, 0, 0},
	}

	for i, tt := range tests {
		r := RequireConstraintComment{
			Enabled:       tt.enabled,
			AllOrNothing:  tt.allOrNothing,
			Exclude:       tt.exclude,
			ExcludeTables: tt.excludeTables,
		}
		s := newTestSchema(t)
		if warns := r.Check(s, tt.lintExclude); len(warns) != tt.want {
			t.Errorf("TestRequireConstraintComment(%d): got %v\nwant %v", i, len(warns), tt.want)
		}

		ns := newTestNoCommentSchema(t)
		if warns := r.Check(ns, tt.lintExclude); len(warns) != tt.wantNoComment {
			t.Errorf("TestRequireConstraintComment(%d) (no comment schema): got %v\nwant %v", i, len(warns), tt.wantNoComment)
		}
	}
}

func TestRequireTriggerComment(t *testing.T) {
	tests := []struct {
		enabled       bool
		allOrNothing  bool
		lintExclude   []string
		exclude       []string
		excludeTables []string
		want          int
		wantNoComment int
	}{
		{true, false, []string{}, []string{}, []string{}, 1, 2},
		{false, false, []string{}, []string{}, []string{}, 0, 0},
		{true, false, []string{"table_a"}, []string{}, []string{}, 0, 0},
		{true, false, []string{}, []string{"update_table_a_column_a2"}, []string{}, 0, 1},
		{true, false, []string{}, []string{"table_a.update_table_a_column_a2"}, []string{}, 0, 1},
		{true, false, []string{}, []string{}, []string{"table_a"}, 0, 0},

		{true, true, []string{}, []string{}, []string{}, 1, 0},
		{false, true, []string{}, []string{}, []string{}, 0, 0},
		{true, true, []string{"table_a"}, []string{}, []string{}, 0, 0},
		{true, true, []string{}, []string{"update_table_a_column_a2"}, []string{}, 0, 0},
		{true, true, []string{}, []string{"table_a.update_table_a_column_a2"}, []string{}, 0, 0},
		{true, true, []string{}, []string{}, []string{"table_a"}, 0, 0},
	}

	for i, tt := range tests {
		r := RequireTriggerComment{
			Enabled:       tt.enabled,
			AllOrNothing:  tt.allOrNothing,
			Exclude:       tt.exclude,
			ExcludeTables: tt.excludeTables,
		}
		s := newTestSchema(t)
		if warns := r.Check(s, tt.lintExclude); len(warns) != tt.want {
			t.Errorf("TestRequireTriggerComment(%d): got %v\nwant %v", i, len(warns), tt.want)
		}

		ns := newTestNoCommentSchema(t)
		if warns := r.Check(ns, tt.lintExclude); len(warns) != tt.wantNoComment {
			t.Errorf("TestRequireTriggerComment(%d) (no comment schema): got %v\nwant %v", i, len(warns), tt.wantNoComment)
		}
	}
}

func TestUnrelatedTable(t *testing.T) {
	tests := []struct {
		enabled        bool
		allOrNothing   bool
		lintExclude    []string
		exclude        []string
		want           int
		wantMsg        string
		wantNoRelation int
	}{
		{true, false, []string{}, []string{}, 1, "unrelated (isolated) table exists. [table_c]", 1},
		{false, false, []string{}, []string{}, 0, "", 0},
		{true, false, []string{}, []string{"table_b"}, 1, "unrelated (isolated) table exists. [table_c]", 1},
		{true, false, []string{}, []string{"table_c"}, 0, "", 1},
		{true, false, []string{"table_c"}, []string{}, 0, "", 1},
		{true, false, []string{}, []string{"*_c"}, 0, "", 1},
		{true, false, []string{"*_c"}, []string{}, 0, "", 1},

		{true, true, []string{}, []string{}, 1, "unrelated (isolated) table exists. [table_c]", 0},
		{false, true, []string{}, []string{}, 0, "", 0},
		{true, true, []string{}, []string{"table_b"}, 1, "unrelated (isolated) table exists. [table_c]", 0},
		{true, true, []string{}, []string{"table_c"}, 0, "", 0},
		{true, true, []string{"table_c"}, []string{}, 0, "", 0},
		{true, true, []string{}, []string{"*_c"}, 0, "", 0},
		{true, true, []string{"*_c"}, []string{}, 0, "", 0},
	}

	for i, tt := range tests {
		r := UnrelatedTable{
			Enabled:      tt.enabled,
			AllOrNothing: tt.allOrNothing,
			Exclude:      tt.exclude,
		}
		s := newTestSchema(t)
		warns := r.Check(s, tt.lintExclude)
		if len(warns) != tt.want {
			t.Errorf("TestUnrelatedTable(%d): got %v\nwant %v", i, len(warns), tt.want)
		}
		if len(warns) > 0 {
			if warns[0].Message != tt.wantMsg {
				t.Errorf("TestUnrelatedTable(%d): got %v\nwant %v", i, warns[0].Message, tt.wantMsg)
			}
		}
		ns := newTestNoRelationSchema(t)
		if warns := r.Check(ns, tt.lintExclude); len(warns) != tt.wantNoRelation {
			fmt.Printf("%v\n", warns)
			t.Errorf("TestUnrelatedTable(%d) (no relation): got %v\nwant %v", i, len(warns), tt.wantNoRelation)
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
		s := newTestSchema(t)
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
		s := newTestSchema(t)
		warns := r.Check(s, tt.lintExclude)
		if len(warns) != tt.want {
			t.Errorf("TestRequireColumns(%d): got %v\nwant %v", i, len(warns), tt.want)
		}
	}
}

func TestRequireTableLabels(t *testing.T) {
	tests := []struct {
		enabled      bool
		allOrNothing bool
		exclude      []string
		lintExclude  []string
		want         int
	}{
		{true, false, []string{}, []string{}, 2},
		{false, false, []string{}, []string{}, 0},
		{true, true, []string{}, []string{}, 2},
		{true, true, []string{"table_b"}, []string{}, 1},
		{true, true, []string{"table_*"}, []string{}, 0},
		{true, true, []string{}, []string{"table_c"}, 1},
		{true, true, []string{}, []string{"table_*"}, 0},
	}

	for i, tt := range tests {
		r := RequireTableLabels{
			Enabled:      tt.enabled,
			AllOrNothing: tt.allOrNothing,
			Exclude:      tt.exclude,
		}
		s := newTestSchema(t)
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
		s := newTestSchema(t)
		relationCopy := *s.Relations[0]
		relationCopy.Def = "copy"
		s.Relations = append(s.Relations, &relationCopy)
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
		s := newTestSchema(t)
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
		s := newTestSchema(t)
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

func newTestSchema(_ *testing.T) *schema.Schema {
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
			Name:              "a1_b1_fk",
			Type:              schema.TypeFK,
			Table:             &ta.Name,
			ReferencedTable:   &tb.Name,
			Columns:           []string{"column_a1"},
			ReferencedColumns: []string{"column_b1"},
			Comment:           "a1_b1_fk comment",
		},
		&schema.Constraint{
			Name:            "a1_unique",
			Type:            "UNIQUE",
			Table:           &ta.Name,
			ReferencedTable: nil,
			Columns:         []string{"column_a1"},
			Comment:         "", // empty comment
		},
	}

	ta.Triggers = []*schema.Trigger{
		&schema.Trigger{
			Name:    "update_table_a_column_a1",
			Def:     "CREATE CONSTRAINT TRIGGER update_table_a_column_a1 AFTER INSERT OR UPDATE ON table_a",
			Comment: "Update column_a1 when update table",
		},
		&schema.Trigger{
			Name: "update_table_a_column_a2",
			Def:  "CREATE CONSTRAINT TRIGGER update_table_a_column_a2 AFTER INSERT OR UPDATE ON table_a",
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
		Viewpoints: []*schema.Viewpoint{
			&schema.Viewpoint{
				Name: "testviewpoint",
				Desc: "testviewpoint desc",
				Tables: []string{
					"table_c", // table_a and table_b is not included
				},
			},
		},
	}
	return s
}

func newTestNoCommentSchema(t *testing.T) *schema.Schema {
	t.Helper()
	s := newTestSchema(t)
	for _, t := range s.Tables {
		t.Comment = ""
		for _, c := range t.Columns {
			c.Comment = ""
		}
		for _, i := range t.Indexes {
			i.Comment = ""
		}
		for _, c := range t.Constraints {
			c.Comment = ""
		}
		for _, tri := range t.Triggers {
			tri.Comment = ""
		}
	}
	return s
}

func newTestNoRelationSchema(t *testing.T) *schema.Schema {
	t.Helper()
	s := newTestSchema(t)
	for _, t := range s.Tables {
		for _, c := range t.Columns {
			c.ChildRelations = nil
			c.ParentRelations = nil
		}
	}
	s.Relations = nil
	return s
}

func TestRequireViewpoints(t *testing.T) {
	tests := []struct {
		enabled     bool
		lintExclude []string
		exclude     []string
		want        int
	}{
		{true, []string{}, []string{}, 2},
		{false, []string{}, []string{}, 0},
		{true, []string{"table_a"}, []string{}, 1},
		{true, []string{"*_a"}, []string{}, 1},
		{true, []string{}, []string{"table_b"}, 1},
	}

	for i, tt := range tests {
		r := RequireViewpoints{
			Enabled: tt.enabled,
			Exclude: tt.exclude,
		}
		s := newTestSchema(t)

		s.Tables[0].Type = "VIEW"
		warns := r.Check(s, tt.lintExclude)
		if len(warns) != tt.want {
			t.Errorf("TestRequireViewpoints(%d): got %v\nwant %v", i, len(warns), tt.want)
		}
	}
}
