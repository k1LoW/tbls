package md

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
	"github.com/tenntenn/golden"
)

var tests = []struct {
	name                   string
	format                 string
	adjust                 bool
	number                 bool
	skipER                 bool
	showOnlyFirstParagraph bool
	tableBName             string
	gotFile                string
	wantFile               string
}{
	{"README.md", "png", false, false, false, false, "b", "README.md", "md_test_README.md"},
	{"a.md", "png", false, false, false, false, "b", "a.md", "md_test_a.md"},
	{"--adjust option", "png", true, false, false, false, "b", "README.md", "md_test_README.md.adjust"},
	{"number", "png", false, true, false, false, "b", "README.md", "md_test_README.md.number"},
	{"spaceInTableName", "png", false, false, false, false, "a b", "README.md", "md_test_README.md.space_in_table_name"},
	{"mermaid README.md", "mermaid", false, false, false, true, "b", "README.md", "md_test_README.md.mermaid"},
	{"mermaid a.md", "mermaid", false, false, false, true, "b", "a.md", "md_test_a.md.mermaid"},
}

var testsTemplate = []struct {
	name                   string
	adjust                 bool
	number                 bool
	skipER                 bool
	showOnlyFirstParagraph bool
	gotFile                string
	wantFile               string
}{
	{"README.md", false, false, false, false, "README.md", "md_template_test_README.md"},
	{"a.md", false, false, false, false, "a.md", "md_template_test_a.md"},
	{"--adjust option", true, false, false, false, "README.md", "md_template_test_README.md.adjust"},
	{"number", false, true, false, false, "README.md", "md_template_test_README.md.number"},
	{"showOnlyFirstParagraph", false, true, true, false, "README.md", "md_template_test_README.md.first_para"},
}

func TestOutput(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestSchema(tt.tableBName)
			c, err := config.New()
			if err != nil {
				t.Error(err)
			}
			tempDir := t.TempDir()
			force := true
			erFormat := tt.format
			opts := []config.Option{
				config.DocPath(tempDir),
				config.Adjust(tt.adjust),
				config.ERFormat(erFormat),
				config.ERSkip(tt.skipER),
			}
			if err := c.Load(filepath.Join(testdataDir(), "out_test_tbls.yml"), opts...); err != nil {
				t.Error(err)
			}
			c.Format.Number = tt.number
			if err := c.ModifySchema(s); err != nil {
				t.Error(err)
			}
			c.Format.ShowOnlyFirstParagraph = tt.showOnlyFirstParagraph
			if err := Output(s, c, force); err != nil {
				t.Error(err)
			}
			got, err := os.ReadFile(filepath.Join(tempDir, tt.gotFile))
			if err != nil {
				t.Fatal(err)
			}
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, testdataDir(), tt.wantFile, got)
				return
			}
			if diff := golden.Diff(t, testdataDir(), tt.wantFile, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestOutputTemplate(t *testing.T) {
	for _, tt := range testsTemplate {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestSchema("b")
			c, err := config.New()
			if err != nil {
				t.Error(err)
			}
			tempDir := t.TempDir()
			force := true
			erFormat := "png"
			opts := []config.Option{
				config.DocPath(tempDir),
				config.Adjust(tt.adjust),
				config.ERFormat(erFormat),
				config.ERSkip(tt.skipER),
			}
			if err := c.Load(filepath.Join(testdataDir(), "out_templates_test_tbls.yml"), opts...); err != nil {
				t.Error(err)
			}
			c.Format.Number = tt.number
			// use the templates in the testdata directory
			c.Templates.MD.Table = filepath.Join(testdataDir(), c.Templates.MD.Table)
			c.Templates.MD.Index = filepath.Join(testdataDir(), c.Templates.MD.Index)
			if err := c.ModifySchema(s); err != nil {
				t.Error(err)
			}
			c.Format.ShowOnlyFirstParagraph = tt.showOnlyFirstParagraph
			if err := Output(s, c, force); err != nil {
				t.Error(err)
			}
			got, err := os.ReadFile(filepath.Join(tempDir, tt.gotFile))
			if err != nil {
				t.Fatal(err)
			}
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, testdataDir(), tt.wantFile, got)
				return
			}
			if diff := golden.Diff(t, testdataDir(), tt.wantFile, got); diff != "" {
				t.Error(diff)
			}
		})
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
			erFormat := tt.format
			if err := c.Load(filepath.Join(testdataDir(), "out_test_tbls.yml"), config.DocPath(docPath), config.Adjust(adjust), config.ERFormat(erFormat)); err != nil {
				t.Error(err)
			}
			c.Format.Number = tt.number
			if err := c.ModifySchema(s); err != nil {
				t.Error(err)
			}
			if err := Output(s, c, force); err != nil {
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
		Type:    "INTEGER",
		Comment: "column a",
	}
	cb := &schema.Column{
		Name:    "b",
		Type:    "TEXT",
		Comment: "column b",
	}

	ta := &schema.Table{
		Name:    "a",
		Comment: "table a",
		Columns: []*schema.Column{
			ca,
			{
				Name:    "a2",
				Type:    "TEXT",
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
				Type:    "DATETIME",
				Comment: "column b2",
			},
		},
	}
	r := &schema.Relation{
		Table:             ta,
		Columns:           []*schema.Column{ca},
		Cardinality:       schema.ZeroOrMore,
		ParentTable:       tb,
		ParentColumns:     []*schema.Column{cb},
		ParentCardinality: schema.ZeroOrOne,
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
