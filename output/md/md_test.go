package md

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
)

var tests = []struct {
	name     string
	gotFile  string
	wantFile string
	adjust   bool
}{
	{"README.md", "README.md", "md_test_README.md.golden", false},
	{"a.md", "a.md", "md_test_a.md.golden", false},
	{"--adjust option", "README.md", "md_test_README.md.adjust.golden", true},
}

var testsTemplate = []struct {
	name     string
	gotFile  string
	wantFile string
	adjust   bool
}{
	{"README.md", "README.md", "md_template_test_README.md.golden", false},
	{"a.md", "a.md", "md_template_test_a.md.golden", false},
	{"--adjust option", "README.md", "md_template_test_README.md.adjust.golden", true},
}

func TestOutput(t *testing.T) {
	for _, tt := range tests {
		s := newTestSchema()
		c, err := config.New()
		if err != nil {
			t.Error(err)
		}
		tempDir, _ := ioutil.TempDir("", "tbls")
		force := true
		adjust := tt.adjust
		erFormat := "png"
		defer os.RemoveAll(tempDir)
		err = c.Load(filepath.Join(testdataDir(), "out_test_tbls.yml"), config.DocPath(tempDir), config.Adjust(adjust), config.ERFormat(erFormat))
		if err != nil {
			t.Error(err)
		}
		err = c.MergeAdditionalData(s)
		if err != nil {
			t.Error(err)
		}
		err = Output(s, c, force)
		if err != nil {
			t.Error(err)
		}
		want, err := ioutil.ReadFile(filepath.Join(testdataDir(), tt.wantFile))
		if err != nil {
			t.Error(err)
		}
		got, err := ioutil.ReadFile(filepath.Join(tempDir, tt.gotFile))
		if err != nil {
			log.Fatal(err)
		}
		if string(got) != string(want) {
			t.Errorf("got %v\nwant %v", string(got), string(want))
		}
	}
}

func TestOutputTemplate(t *testing.T) {
	for _, tt := range testsTemplate {
		s := newTestSchema()
		c, err := config.New()
		if err != nil {
			t.Error(err)
		}
		tempDir, _ := ioutil.TempDir("", "tbls")
		force := true
		adjust := tt.adjust
		erFormat := "png"
		defer os.RemoveAll(tempDir)
		err = c.Load(filepath.Join(testdataDir(), "out_templates_test_tbls.yml"), config.DocPath(tempDir), config.Adjust(adjust), config.ERFormat(erFormat))
		if err != nil {
			t.Error(err)
		}
		// use the templates in the testdata directory
		c.Templates.MD.Table = filepath.Join(testdataDir(), c.Templates.MD.Table)
		c.Templates.MD.Index = filepath.Join(testdataDir(), c.Templates.MD.Index)

		err = c.MergeAdditionalData(s)
		if err != nil {
			t.Error(err)
		}
		err = Output(s, c, force)
		if err != nil {
			t.Error(err)
		}
		want, err := ioutil.ReadFile(filepath.Join(testdataDir(), tt.wantFile))
		if err != nil {
			t.Error(err)
		}
		got, err := ioutil.ReadFile(filepath.Join(tempDir, tt.gotFile))
		if err != nil {
			log.Fatal(err)
		}
		if string(got) != string(want) {
			t.Errorf("got %v\nwant %v", string(got), string(want))
		}
	}
}

func TestDiffSchemaAndDocs(t *testing.T) {
	for _, tt := range tests {
		func() {
			s := newTestSchema()
			c, err := config.New()
			if err != nil {
				t.Error(err)
			}
			docPath, _ := ioutil.TempDir("", "tbls")
			force := true
			adjust := tt.adjust
			erFormat := "png"
			defer os.RemoveAll(docPath)
			err = c.Load(filepath.Join(testdataDir(), "out_test_tbls.yml"), config.DocPath(docPath), config.Adjust(adjust), config.ERFormat(erFormat))
			if err != nil {
				t.Error(err)
			}
			err = c.MergeAdditionalData(s)
			if err != nil {
				t.Error(err)
			}
			err = Output(s, c, force)
			if err != nil {
				t.Error(err)
			}
			want := ""
			got, err := DiffSchemaAndDocs(docPath, s, c)
			if err != nil {
				t.Error(err)
			}
			if got != want {
				t.Errorf("got %v\nwant %v", got, want)
			}
		}()
	}
}

func TestDiffSchemas(t *testing.T) {
	s := newTestSchema()
	s2 := newTestSchema()
	c, err := config.New()
	if err != nil {
		t.Error(err)
	}
	want := ""
	got, err := DiffSchemas(s, s2, c, c)
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}

	s2.Name = "modified"
	got2, err := DiffSchemas(s, s2, c, c)
	if err != nil {
		t.Error(err)
	}
	if got2 == "" {
		t.Error("diff should not be empty")
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
