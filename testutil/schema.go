package testutil

import (
	"testing"

	"github.com/k1LoW/tbls/schema"
)

func NewSchema(t *testing.T) *schema.Schema {
	const (
		tableAName     = "a"
		tableBName     = "b"
		tableViewName  = "view"
		labelBlueName  = "blue"
		labelRedName   = "red"
		labelGreenName = "green"
		enumName       = "enum"
	)

	labelBlue := &schema.Label{
		Name:    labelBlueName,
		Virtual: false,
	}
	labelRed := &schema.Label{
		Name:    labelRedName,
		Virtual: false,
	}
	labelGreen := &schema.Label{
		Name:    labelGreenName,
		Virtual: true,
	}

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

	cView := &schema.Column{
		Name:    "view_column",
		Type:    "INTEGER",
		Comment: "column of view",
	}

	ta := &schema.Table{
		Name:    tableAName,
		Comment: "table a",
		Columns: []*schema.Column{
			ca,
			&schema.Column{
				Name:    "a2",
				Type:    "TEXT",
				Comment: "column a2",
			},
		},
		Labels: []*schema.Label{labelBlue, labelGreen},
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
		Name:    tableBName,
		Comment: "table b",
		Columns: []*schema.Column{
			cb,
			&schema.Column{
				Name:    "b2",
				Type:    "TEXT",
				Comment: "column b2",
			},
		},
		Labels: []*schema.Label{labelRed, labelGreen},
	}

	tView := &schema.Table{
		Name:    tableViewName,
		Comment: "view",
		Columns: []*schema.Column{
			cView,
		},
		Type: "VIEW",
		Def:  "CREATE VIEW view AS SELECT a, b FROM a JOIN b ON a.a = b.b",
		ReferencedTables: []*schema.Table{
			ta,
			tb,
		},
	}

	enum := &schema.Enum{
		Name:   enumName,
		Values: []string{"one", "two", "three"},
	}

	r := &schema.Relation{
		Table:             tb,
		Columns:           []*schema.Column{cb},
		Cardinality:       schema.OneOrMore,
		ParentTable:       ta,
		ParentColumns:     []*schema.Column{ca},
		ParentCardinality: schema.ExactlyOne,
		Def:               "FOREIGN KEY (b) REFERENCES a(a)",
		Virtual:           false,
	}
	ca.ChildRelations = []*schema.Relation{r}
	cb.ParentRelations = []*schema.Relation{r}

	s := &schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			ta,
			tb,
			tView,
		},
		Enums: []*schema.Enum{
			enum,
		},
		Relations: []*schema.Relation{
			r,
		},
		Viewpoints: schema.Viewpoints{
			&schema.Viewpoint{
				Name: "table a b",
				Desc: "select table a and b",
				Tables: []string{
					tableAName,
					tableBName,
				},
			},
			&schema.Viewpoint{
				Name: "label blue",
				Desc: "select label blue",
				Labels: []string{
					labelBlueName,
				},
			},
			&schema.Viewpoint{
				Name: "label green",
				Desc: "select label green",
				Labels: []string{
					labelGreenName,
				},
				Groups: []*schema.ViewpointGroup{
					&schema.ViewpointGroup{
						Name: "label red",
						Desc: "select label red",
						Labels: []string{
							labelRedName,
						},
					},
				},
			},
			&schema.Viewpoint{
				Name: "table a label red",
				Desc: "select table a and label red\n\n- table a\n- label red",
				Tables: []string{
					tableAName,
				},
				Labels: []string{
					labelRedName,
				},
			},
		},
		Driver: &schema.Driver{
			Name:            "testdriver",
			DatabaseVersion: "1.0.0",
			Meta:            &schema.DriverMeta{},
		},
	}
	if err := s.Repair(); err != nil {
		t.Fatal(err)
	}
	return s
}
