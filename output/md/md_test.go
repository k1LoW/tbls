package md

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/tbls/schema"
)

var tests = []struct {
	name         string
	actualFile   string
	expectedFile string
}{
	{"README.md", "README.md", "md_test_README.md.golden"},
	{"a.md", "a.md", "md_test_a.md.golden"},
}

func TestOutput(t *testing.T) {
	s := newTestSchema()
	tempDir, _ := ioutil.TempDir("", "tbls")
	force := true
	adjust := false
	erFormat := "png"
	defer os.RemoveAll(tempDir)
	err := Output(s, tempDir, force, adjust, erFormat)
	if err != nil {
		t.Error(err)
	}
	for _, tt := range tests {
		expected, _ := ioutil.ReadFile(filepath.Join(testdataDir(), tt.expectedFile))
		actual, err := ioutil.ReadFile(filepath.Join(tempDir, tt.actualFile))
		if err != nil {
			log.Fatal(err)
		}
		if string(actual) != string(expected) {
			t.Errorf("actual %v\nwant %v", string(actual), string(expected))
		}
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
