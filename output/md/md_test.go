package md

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
)

var tests = []struct {
	name                   string
	adjust                 bool
	number                 bool
	showOnlyFirstParagraph bool
	tableBName             string
	gotFile                string
	wantFile               string
}{
	{"README.md", false, false, false, "b", "README.md", "md_test_README.md.golden"},
	{"a.md", false, false, false, "b", "a.md", "md_test_a.md.golden"},
	{"--adjust option", true, false, false, "b", "README.md", "md_test_README.md.adjust.golden"},
	{"number", false, true, false, "b", "README.md", "md_test_README.md.number.golden"},
	{"spaceInTableName", false, false, false, "a b", "README.md", "md_test_README.md.space_in_table_name.golden"},
}

var testsTemplate = []struct {
	name                   string
	adjust                 bool
	number                 bool
	showOnlyFirstParagraph bool
	gotFile                string
	wantFile               string
}{
	{"README.md", false, false, false, "README.md", "md_template_test_README.md.golden"},
	{"a.md", false, false, false, "a.md", "md_template_test_a.md.golden"},
	{"--adjust option", true, false, false, "README.md", "md_template_test_README.md.adjust.golden"},
	{"number", false, true, false, "README.md", "md_template_test_README.md.number.golden"},
	{"showOnlyFirstParagraph", false, true, true, "README.md", "md_template_test_README.md.first_para.golden"},
}

func TestOutput(t *testing.T) {
	for _, tt := range tests {
		s := newTestSchema(tt.tableBName)
		c, err := config.New()
		if err != nil {
			t.Error(err)
		}
		tempDir := t.TempDir()
		force := true
		adjust := tt.adjust
		erFormat := "png"
		err = c.Load(filepath.Join(testdataDir(), "out_test_tbls.yml"), config.DocPath(tempDir), config.Adjust(adjust), config.ERFormat(erFormat))
		if err != nil {
			t.Error(err)
		}
		c.Format.Number = tt.number
		err = c.MergeAdditionalData(s)
		if err != nil {
			t.Error(err)
		}
		c.Format.ShowOnlyFirstParagraph = tt.showOnlyFirstParagraph
		err = Output(s, c, force)
		if err != nil {
			t.Error(err)
		}
		want, err := os.ReadFile(filepath.Join(testdataDir(), tt.wantFile))
		if err != nil {
			t.Error(err)
		}
		got, err := os.ReadFile(filepath.Join(tempDir, tt.gotFile))
		if err != nil {
			log.Fatal(err)
		}
		if diff := cmp.Diff(string(got), string(want), nil); diff != "" {
			t.Errorf("diff with %s:\n %s", tt.wantFile, diff)
		}
	}
}

func TestOutputTemplate(t *testing.T) {
	for _, tt := range testsTemplate {
		s := newTestSchema("b")
		c, err := config.New()
		if err != nil {
			t.Error(err)
		}
		tempDir := t.TempDir()
		force := true
		adjust := tt.adjust
		erFormat := "png"
		err = c.Load(filepath.Join(testdataDir(), "out_templates_test_tbls.yml"), config.DocPath(tempDir), config.Adjust(adjust), config.ERFormat(erFormat))
		if err != nil {
			t.Error(err)
		}
		c.Format.Number = tt.number
		// use the templates in the testdata directory
		c.Templates.MD.Table = filepath.Join(testdataDir(), c.Templates.MD.Table)
		c.Templates.MD.Index = filepath.Join(testdataDir(), c.Templates.MD.Index)
		err = c.MergeAdditionalData(s)
		if err != nil {
			t.Error(err)
		}
		c.Format.ShowOnlyFirstParagraph = tt.showOnlyFirstParagraph
		err = Output(s, c, force)
		if err != nil {
			t.Error(err)
		}
		want, err := os.ReadFile(filepath.Join(testdataDir(), tt.wantFile))
		if err != nil {
			t.Error(err)
		}
		got, err := os.ReadFile(filepath.Join(tempDir, tt.gotFile))
		if err != nil {
			log.Fatal(err)
		}
		if diff := cmp.Diff(string(got), string(want), nil); diff != "" {
			t.Errorf("diff with %s:\n %s", tt.wantFile, diff)
		}
	}
}

func TestDiffSchemaAndDocs(t *testing.T) {
	for _, tt := range tests {
		func() {
			s := newTestSchema("b")
			c, err := config.New()
			if err != nil {
				t.Error(err)
			}
			docPath := t.TempDir()
			force := true
			adjust := tt.adjust
			erFormat := "png"
			err = c.Load(filepath.Join(testdataDir(), "out_test_tbls.yml"), config.DocPath(docPath), config.Adjust(adjust), config.ERFormat(erFormat))
			if err != nil {
				t.Error(err)
			}
			c.Format.Number = tt.number
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
	testData := func() (s, s2 *schema.Schema, c *config.Config) {
		s = newTestSchema("b")
		s2 = newTestSchema("b")
		c, _ = config.New()
		return
	}

	{
		s, s2, c := testData()
		want := ""
		got, err := DiffSchemas(s, s2, c, c)
		if err != nil {
			t.Error(err)
		}
		if got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}
	}

	{
		s, s2, c := testData()
		s2.Name = "modified"
		got2, err := DiffSchemas(s, s2, c, c)
		if err != nil {
			t.Error(err)
		}
		if got2 == "" {
			t.Error("diff should not be empty")
		}
	}

	{
		s, s2, c := testData()
		s.Tables = s.Tables[:len(s.Tables)-1]
		got2, err := DiffSchemas(s, s2, c, c)
		if err != nil {
			t.Error(err)
		}
		if got2 == "" {
			t.Error("diff should not be empty")
		}
	}

	{
		s, s2, c := testData()
		s2.Tables = s2.Tables[:len(s2.Tables)-1]
		got2, err := DiffSchemas(s, s2, c, c)
		if err != nil {
			t.Error(err)
		}
		if got2 == "" {
			t.Error("diff should not be empty")
		}
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(filepath.Dir(wd)), "testdata"))
	return dir
}

func newTestSchema(tableBName string) *schema.Schema {
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
			{
				Name:    "a2",
				Comment: "column a2",
			},
		},
	}
	ta.Indexes = []*schema.Index{
		{
			Name:    "PRIMARY KEY",
			Def:     "PRIMARY KEY(a)",
			Table:   &ta.Name,
			Columns: []string{"a"},
		},
	}
	ta.Constraints = []*schema.Constraint{
		{
			Name:  "PRIMARY",
			Table: &ta.Name,
			Def:   "PRIMARY KEY (a)",
		},
	}
	ta.Triggers = []*schema.Trigger{
		{
			Name: "update_a_a2",
			Def:  "CREATE CONSTRAINT TRIGGER update_a_a2 AFTER INSERT OR UPDATE ON a",
		},
	}
	tb := &schema.Table{
		Name:    tableBName,
		Comment: "table b",
		Columns: []*schema.Column{
			cb,
			{
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
