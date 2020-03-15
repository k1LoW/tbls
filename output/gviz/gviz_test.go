package gviz

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
)

func TestOutputSchema(t *testing.T) {
	format := "svg"
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
	o := New(c, format)
	buf := &bytes.Buffer{}
	err = o.OutputSchema(buf, s)
	if err != nil {
		t.Error(err)
	}
	expected, _ := ioutil.ReadFile(filepath.Join(testdataDir(), "svg_test_schema.svg.golden"))
	actual := buf.String()
	if actual != string(expected) {
		t.Errorf("actual %v\nwant %v", actual, string(expected))
	}
}

func TestOutputTable(t *testing.T) {
	format := "svg"
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

	o := New(c, format)
	buf := &bytes.Buffer{}
	_ = o.OutputTable(buf, ta)
	expected, _ := ioutil.ReadFile(filepath.Join(testdataDir(), "svg_test_a.svg.golden"))
	actual := buf.String()
	if actual != string(expected) {
		t.Errorf("actual %v\nwant %v", actual, string(expected))
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
		Driver: &schema.Driver{
			Name:            "testdriver",
			DatabaseVersion: "1.0.0",
		},
	}
	return s
}
