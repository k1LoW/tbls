package mermaid

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
	"github.com/tenntenn/golden"
)

func TestOutputSchema(t *testing.T) {
	s := newTestSchema(t)
	c, err := config.New()
	if err != nil {
		t.Error(err)
	}
	if err := c.LoadConfigFile(filepath.Join(testdataDir(), "out_test_tbls.yml")); err != nil {
		t.Error(err)
	}
	if err := c.MergeAdditionalData(s); err != nil {
		t.Error(err)
	}
	o := New(c)
	got := &bytes.Buffer{}
	err = o.OutputSchema(got, s)
	if err != nil {
		t.Error(err)
	}
	f := fmt.Sprintf("mermaid_test_schema")
	if os.Getenv("UPDATE_GOLDEN") != "" {
		golden.Update(t, testdataDir(), f, got)
		return
	}
	if diff := golden.Diff(t, testdataDir(), f, got); diff != "" {
		t.Error(diff)
	}
}

func TestOutputSchemaTemplate(t *testing.T) {
	s := newTestSchema(t)
	c, err := config.New()
	if err != nil {
		t.Error(err)
	}
	if err := c.LoadConfigFile(filepath.Join(testdataDir(), "out_templates_test_tbls.yml")); err != nil {
		t.Error(err)
	}
	// use the templates in the testdata directory
	c.Templates.Mermaid.Schema = filepath.Join(testdataDir(), c.Templates.Mermaid.Schema)
	if err := c.MergeAdditionalData(s); err != nil {
		t.Error(err)
	}
	o := New(c)
	got := &bytes.Buffer{}
	if err := o.OutputSchema(got, s); err != nil {
		t.Error(err)
	}
	f := fmt.Sprintf("mermaid_template_test_schema")
	if os.Getenv("UPDATE_GOLDEN") != "" {
		golden.Update(t, testdataDir(), f, got)
		return
	}
	if diff := golden.Diff(t, testdataDir(), f, got); diff != "" {
		t.Error(diff)
	}
}

func TestOutputTable(t *testing.T) {
	s := newTestSchema(t)
	c, err := config.New()
	if err != nil {
		t.Error(err)
	}
	if err := c.LoadConfigFile(filepath.Join(testdataDir(), "out_test_tbls.yml")); err != nil {
		t.Error(err)
	}
	if err := c.MergeAdditionalData(s); err != nil {
		t.Error(err)
	}
	ta := s.Tables[0]

	o := New(c)
	got := &bytes.Buffer{}
	if err := o.OutputTable(got, ta); err != nil {
		t.Error(err)
	}
	f := fmt.Sprintf("mermaid_test_a")
	if os.Getenv("UPDATE_GOLDEN") != "" {
		golden.Update(t, testdataDir(), f, got)
		return
	}
	if diff := golden.Diff(t, testdataDir(), f, got); diff != "" {
		t.Error(diff)
	}
}

func TestOutputTableTemplate(t *testing.T) {
	s := newTestSchema(t)
	c, err := config.New()
	if err != nil {
		t.Error(err)
	}
	if err := c.LoadConfigFile(filepath.Join(testdataDir(), "out_templates_test_tbls.yml")); err != nil {
		t.Error(err)
	}
	// use the templates in the testdata directory
	c.Templates.Mermaid.Table = filepath.Join(testdataDir(), c.Templates.Mermaid.Table)
	if err := c.MergeAdditionalData(s); err != nil {
		t.Error(err)
	}
	ta := s.Tables[0]

	o := New(c)
	got := &bytes.Buffer{}
	if err := o.OutputTable(got, ta); err != nil {
		t.Error(err)
	}
	f := fmt.Sprintf("mermaid_template_test_a")
	if os.Getenv("UPDATE_GOLDEN") != "" {
		golden.Update(t, testdataDir(), f, got)
		return
	}
	if diff := golden.Diff(t, testdataDir(), f, got); diff != "" {
		t.Error(diff)
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(filepath.Dir(wd)), "testdata"))
	return dir
}

func newTestSchema(t *testing.T) *schema.Schema {
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
