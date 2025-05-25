package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/SouhlInc/tbls/config"
	"github.com/SouhlInc/tbls/schema"
)

func TestOutputSchema(t *testing.T) {
	s := newTestSchema(t)
	c, err := config.New()
	if err != nil {
		t.Error(err)
	}
	o := New(c)
	buf := &bytes.Buffer{}
	err = o.OutputSchema(buf, s)
	if err != nil {
		t.Error(err)
	}
	want, err := os.ReadFile(filepath.Join(testdataDir(), "config_test.yml.golden"))
	if err != nil {
		t.Error(err)
	}
	got := buf.String()
	if got != string(want) {
		t.Errorf("got\n%v\nwant\n%v", got, string(want))
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(filepath.Dir(wd)), "testdata"))
	return dir
}

func newTestSchema(t *testing.T) *schema.Schema {
	t.Helper()
	ca := &schema.Column{
		Name:    "a",
		Comment: "column a",
	}
	cb := &schema.Column{
		Name:    "b",
		Comment: "column b",
	}

	ta := &schema.Table{
		Name:    "a",
		Comment: "table a",
		Columns: []*schema.Column{
			ca,
			&schema.Column{
				Name:    "a2",
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
				Comment: "column b2",
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

	s := &schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			ta,
			tb,
		},
		Relations: []*schema.Relation{
			r,
		},
	}
	return s
}
