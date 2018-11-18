package dot

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/tbls/schema"
)

func TestOutputSchema(t *testing.T) {
	s := newTestSchema()
	err := s.LoadAdditionalData(filepath.Join(testdataDir(), "md_test_additional_data.yml"))
	if err != nil {
		t.Error(err)
	}
	buf := &bytes.Buffer{}
	err = OutputSchema(buf, s)
	if err != nil {
		t.Error(err)
	}
	expected, _ := ioutil.ReadFile(filepath.Join(testdataDir(), "dot_test_schema.dot.golden"))
	actual := buf.String()
	if actual != string(expected) {
		t.Errorf("actual %v\nwant %v", actual, string(expected))
	}
}

func TestOutputTable(t *testing.T) {
	s := newTestSchema()
	err := s.LoadAdditionalData(filepath.Join(testdataDir(), "md_test_additional_data.yml"))
	if err != nil {
		t.Error(err)
	}
	ta := s.Tables[0]

	buf := &bytes.Buffer{}
	_ = OutputTable(buf, ta)
	expected, _ := ioutil.ReadFile(filepath.Join(testdataDir(), "dot_test_a.dot.golden"))
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
	}
	return s
}
