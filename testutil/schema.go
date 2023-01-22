package testutil

import (
	"testing"

	"github.com/k1LoW/tbls/schema"
)

func NewSchema(t *testing.T) *schema.Schema {
	ca := &schema.Column{
		Name:    "a",
		Type:    "INTEGER",
		Comment: "column a",
	}
	cb := &schema.Column{
		Name:    "b",
		Type:    "INTEGER",
		Comment: "column b",
	}

	ta := &schema.Table{
		Name:    "a",
		Comment: "table a",
		Columns: []*schema.Column{
			ca,
			&schema.Column{
				Name:    "a2",
				Type:    "TEXT",
				Comment: "column a2",
			},
		},
	}
	ta.Indexes = []*schema.Index{
		&schema.Index{
			Name:    "PRIMARY KEY",
			Def:     "PRIMARY KEY(a)",
			Table:   &ta.Name,
			Columns: []string{"a"},
		},
	}
	ta.Constraints = []*schema.Constraint{
		&schema.Constraint{
			Name:  "PRIMARY",
			Table: &ta.Name,
			Def:   "PRIMARY KEY (a)",
		},
	}
	ta.Triggers = []*schema.Trigger{
		&schema.Trigger{
			Name: "update_a_a2",
			Def:  "CREATE CONSTRAINT TRIGGER update_a_a2 AFTER INSERT OR UPDATE ON a",
		},
	}
	tb := &schema.Table{
		Name:    "b",
		Comment: "table b",
		Columns: []*schema.Column{
			cb,
			&schema.Column{
				Name:    "b2",
				Type:    "TEXT",
				Comment: "column b2",
			},
		},
	}
	r := &schema.Relation{
		Table:             tb,
		Columns:           []*schema.Column{cb},
		Cardinality:       schema.OneOrMore,
		ParentTable:       ta,
		ParentColumns:     []*schema.Column{ca},
		ParentCardinality: schema.ExactlyOne,
	}
	ca.ChildRelations = []*schema.Relation{r}
	cb.ParentRelations = []*schema.Relation{r}

	s := &schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			ta,
			tb,
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
