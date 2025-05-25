package md

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/SouhlInc/tbls/config"
	"github.com/SouhlInc/tbls/schema"
	"github.com/SouhlInc/tbls/testutil"
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
	{"a.md", "png", false, false, true, false, "b", "a.md", "md_test_a.md"},
	{"--adjust option", "png", true, false, true, false, "b", "README.md", "md_test_README.md.adjust"},
	{"number", "png", false, true, true, false, "b", "README.md", "md_test_README.md.number"},
	{"spaceInTableName", "png", false, false, true, false, "a b", "README.md", "md_test_README.md.space_in_table_name"},
	{"mermaid README.md", "mermaid", false, false, false, false, "b", "README.md", "md_test_README.md.mermaid"},
	{"mermaid a.md", "mermaid", false, false, false, false, "b", "a.md", "md_test_a.md.mermaid"},
	{"showOnlyFirstParagraph README.md", "png", false, false, false, true, "b", "README.md", "md_test_README.md.first_para"},
	{"showOnlyFirstParagraph a.md", "png", false, false, false, true, "b", "a.md", "md_test_a.md.first_para"},
	{"view.md", "png", false, false, true, false, "b", "view.md", "md_test_view.md"},
	{"viewpoint-1.md", "png", false, false, false, false, "b", "viewpoint-1.md", "md_test_viewpoint-1.md"},
	{"viewpoint-2.md", "png", false, false, false, false, "b", "viewpoint-2.md", "md_test_viewpoint-2.md"},
	{"viewpoint-1.md", "mermaid", false, false, false, false, "b", "viewpoint-1.md", "md_test_viewpoint-1.md.mermaid"},
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
	{"README.md", false, false, true, false, "README.md", "md_template_test_README.md"},
	{"a.md", false, false, true, false, "a.md", "md_template_test_a.md"},
	{"view.md", false, false, true, false, "view.md", "md_template_test_view.md"},
	{"--adjust option", true, false, true, false, "README.md", "md_template_test_README.md.adjust"},
	{"number", false, true, true, false, "README.md", "md_template_test_README.md.number"},
	{"showOnlyFirstParagraph", false, true, true, true, "README.md", "md_template_test_README.md.first_para"},
}

func TestOutput(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testutil.NewSchema(t)
			tb, err := s.FindTableByName("b")
			if err != nil {
				t.Fatal(err)
			}
			tb.Name = tt.tableBName
			for _, v := range s.Viewpoints {
				if vti := slices.Index(v.Tables, "b"); vti != -1 {
					v.Tables[vti] = tt.tableBName
				}
			}
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
			s := testutil.NewSchema(t)
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
			s := testutil.NewSchema(t)
			c, err := config.New()
			if err != nil {
				t.Error(err)
			}
			docPath := t.TempDir()
			force := true
			opts := []config.Option{
				config.DocPath(docPath),
				config.Adjust(tt.adjust),
				config.ERFormat(tt.format),
				config.ERSkip(tt.skipER),
			}
			if err := c.Load(filepath.Join(testdataDir(), "out_test_tbls.yml"), opts...); err != nil {
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
		s = testutil.NewSchema(t)
		s2 = testutil.NewSchema(t)
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
