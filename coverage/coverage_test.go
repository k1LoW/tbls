package coverage

import (
	"database/sql"
	"testing"

	"github.com/SouhlInc/tbls/schema"
)

func TestMeasure(t *testing.T) {
	s := newTestSchema(t)
	got := Measure(s)
	if want := 10; got.Covered != want {
		t.Errorf("got %v want %v", got.Covered, want)
	}
	if want := 17; got.Total != want {
		t.Errorf("got %v want %v", got.Total, want)
	}
}

func TestRound(t *testing.T) {
	tests := []struct {
		in   float64
		want float64
	}{
		{0.3, 0.3},
		{0.33, 0.3},
		{0.333333, 0.3},
		{0.34, 0.3},
		{0.35, 0.4},
	}
	for _, tt := range tests {
		got := round(tt.in)
		if got != tt.want {
			t.Errorf("got %v want %v", got, tt.want)
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
		},
		&schema.Constraint{
			Name:            "a1_unique",
			Type:            "UNIQUE",
			Table:           &ta.Name,
			ReferencedTable: nil,
			Columns:         []string{"column_a1"},
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
	}
	return s
}
