package plantuml

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
)

func TestOutputSchema(t *testing.T) {
	s := newTestSchema()
	c, err := config.New()
	if err != nil {
		t.Error(err)
	}
	err = c.LoadConfigFile(filepath.Join(testdataDir(), "out_test_tbls.yml"))
	if err != nil {
		t.Error(err)
	}
	err = c.MergeAdditionalData(s)
	if err != nil {
		t.Error(err)
	}
	o := New(c)
	buf := &bytes.Buffer{}
	err = o.OutputSchema(buf, s)
	if err != nil {
		t.Error(err)
	}
	want, _ := os.ReadFile(filepath.Join(testdataDir(), "plantuml_test_schema.puml.golden"))
	got := buf.String()
	if got != string(want) {
		t.Errorf("got %v\nwant %v", got, string(want))
	}
}

func TestOutputSchemaTemplate(t *testing.T) {
	s := newTestSchema()
	c, err := config.New()
	if err != nil {
		t.Error(err)
	}
	err = c.LoadConfigFile(filepath.Join(testdataDir(), "out_templates_test_tbls.yml"))
	if err != nil {
		t.Error(err)
	}
	// use the templates in the testdata directory
	c.Templates.PUML.Schema = filepath.Join(testdataDir(), c.Templates.PUML.Schema)
	err = c.MergeAdditionalData(s)
	if err != nil {
		t.Error(err)
	}
	o := New(c)
	buf := &bytes.Buffer{}
	err = o.OutputSchema(buf, s)
	if err != nil {
		t.Error(err)
	}
	want, _ := os.ReadFile(filepath.Join(testdataDir(), "plantuml_template_test_schema.puml.golden"))
	got := buf.String()
	if got != string(want) {
		t.Errorf("got %v\nwant %v", got, string(want))
	}
}

func TestOutputTable(t *testing.T) {
	s := newTestSchema()
	c, err := config.New()
	if err != nil {
		t.Error(err)
	}
	err = c.LoadConfigFile(filepath.Join(testdataDir(), "out_test_tbls.yml"))
	if err != nil {
		t.Error(err)
	}
	err = c.MergeAdditionalData(s)
	if err != nil {
		t.Error(err)
	}
	ta := s.Tables[0]

	o := New(c)
	buf := &bytes.Buffer{}
	_ = o.OutputTable(buf, ta)
	want, _ := os.ReadFile(filepath.Join(testdataDir(), "plantuml_test_a.puml.golden"))
	got := buf.String()
	if got != string(want) {
		t.Errorf("got %v\nwant %v", got, string(want))
	}
}

func TestOutputTableTemplate(t *testing.T) {
	s := newTestSchema()
	c, err := config.New()
	if err != nil {
		t.Error(err)
	}
	err = c.LoadConfigFile(filepath.Join(testdataDir(), "out_templates_test_tbls.yml"))
	if err != nil {
		t.Error(err)
	}
	// use the templates in the testdata directory
	c.Templates.PUML.Table = filepath.Join(testdataDir(), c.Templates.PUML.Table)
	err = c.MergeAdditionalData(s)
	if err != nil {
		t.Error(err)
	}
	ta := s.Tables[0]

	o := New(c)
	buf := &bytes.Buffer{}
	_ = o.OutputTable(buf, ta)
	want, _ := os.ReadFile(filepath.Join(testdataDir(), "plantuml_template_test_a.puml.golden"))
	got := buf.String()
	if got != string(want) {
		t.Errorf("got %v\nwant %v", got, string(want))
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(filepath.Dir(wd)), "testdata"))
	return dir
}

func newTestSchema() *schema.Schema {
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
		Table:         tb,
		Columns:       []*schema.Column{cb},
		ParentTable:   ta,
		ParentColumns: []*schema.Column{ca},
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
