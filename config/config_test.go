package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/tbls/schema"
)

func TestLoadDefault(t *testing.T) {
	configFilepath := filepath.Join(testdataDir(), "empty.yml")
	config, err := New()
	if err != nil {
		t.Fatal(err)
	}
	err = config.Load(configFilepath)
	if err != nil {
		t.Fatal(err)
	}

	if want := ""; config.DSN != want {
		t.Errorf("got %v\nwant %v", config.DSN, want)
	}
	if want := "dbdoc"; config.DocPath != want {
		t.Errorf("got %v\nwant %v", config.DocPath, want)
	}
	if want := "png"; config.ER.Format != want {
		t.Errorf("got %v\nwant %v", config.ER.Format, want)
	}
	if want := 1; *config.ER.Distance != want {
		t.Errorf("got %v\nwant %v", config.ER.Distance, want)
	}
}

func TestLoadConfigFile(t *testing.T) {
	_ = os.Setenv("TBLS_TEST_PG_PASS", "pgpass")
	_ = os.Setenv("TBLS_TEST_PG_DOC_PATH", "sample/pg")
	configFilepath := filepath.Join(testdataDir(), "config_test_tbls_2.yml")
	config, err := New()
	if err != nil {
		t.Fatal(err)
	}
	err = config.LoadConfigFile(configFilepath)
	if err != nil {
		t.Fatal(err)
	}

	if want := "pg://root:pgpass@localhost:55432/testdb?sslmode=disable"; config.DSN != want {
		t.Errorf("got %v\nwant %v", config.DSN, want)
	}

	if want := "sample/pg"; config.DocPath != want {
		t.Errorf("got %v\nwant %v", config.DocPath, want)
	}

	if want := "INDEX"; config.Dict.Lookup("Indexes") != want {
		t.Errorf("got %v\nwant %v", config.Dict.Lookup("Indexes"), want)
	}
}

func TestDuplicateConfigFile(t *testing.T) {
	config := &Config{
		root: filepath.Join(testdataDir(), "config"),
	}
	got := config.LoadConfigFile("")
	want := "duplicate config file [.tbls.yml, tbls.yml]"
	if fmt.Sprintf("%v", got) != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

var tests = []struct {
	value string
	want  string
}{
	{"${TBLS_ONE}/${TBLS_TWO}", "one/two"},
	{"${TBLS_ONE}/${TBLS_TWO}/${TBLS_NONE}", "one/two/"},
	{"${{TBLS_ONE}}", "${{TBLS_ONE}}"},
	{"{{.TBLS_ONE}}/{{.TBLS_TWO}}", "one/two"},
}

func TestParseWithEnvirion(t *testing.T) {
	_ = os.Setenv("TBLS_ONE", "one")
	_ = os.Setenv("TBLS_TWO", "two")
	for _, tt := range tests {
		got, err := parseWithEnviron(tt.value)
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func TestMergeAditionalData(t *testing.T) {
	s := schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			&schema.Table{
				Name:    "users",
				Comment: "users comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "username",
						Type: "text",
					},
				},
			},
			&schema.Table{
				Name:    "posts",
				Comment: "posts comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "user_id",
						Type: "int",
					},
					&schema.Column{
						Name: "title",
						Type: "text",
					},
				},
			},
		},
	}
	c, err := New()
	if err != nil {
		t.Error(err)
	}
	err = c.LoadConfigFile(filepath.Join(testdataDir(), "config_test_tbls.yml"))
	if err != nil {
		t.Error(err)
	}
	err = c.MergeAdditionalData(&s)
	if err != nil {
		t.Error(err)
	}
	if want := 1; len(s.Relations) != want {
		t.Errorf("got %v\nwant %v", len(s.Relations), want)
	}
	posts, _ := s.FindTableByName("posts")
	title, _ := posts.FindColumnByName("title")
	if want := "post title"; title.Comment != want {
		t.Errorf("got %v\nwant %v", title.Comment, want)
	}
}

func TestFilterTables(t *testing.T) {
	s := schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			&schema.Table{
				Name:    "users",
				Comment: "users comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "username",
						Type: "text",
					},
				},
			},
			&schema.Table{
				Name:    "posts",
				Comment: "posts comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "user_id",
						Type: "int",
					},
					&schema.Column{
						Name: "title",
						Type: "text",
					},
				},
			},
			&schema.Table{
				Name:    "user_options",
				Comment: "user_options comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "user_id",
						Type: "int",
					},
					&schema.Column{
						Name: "email",
						Type: "text",
					},
				},
			},
			&schema.Table{
				Name: "schema_migrations",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "name",
						Type: "text",
					},
				},
			},
		},
	}
	c, err := New()
	if err != nil {
		t.Error(err)
	}
	err = c.LoadConfigFile(filepath.Join(testdataDir(), "config_test_tbls.yml"))
	if err != nil {
		t.Error(err)
	}
	err = c.FilterTables(&s)
	if err != nil {
		t.Error(err)
	}
	if want := 2; len(s.Tables) != want {
		t.Errorf("got %v\nwant %v", len(s.Tables), want)
	}
}

func TestModifySchema(t *testing.T) {
	s := schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			&schema.Table{
				Name:    "users",
				Comment: "users comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "username",
						Type: "text",
					},
				},
			},
			&schema.Table{
				Name:    "posts",
				Comment: "posts comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "user_id",
						Type: "int",
					},
					&schema.Column{
						Name: "title",
						Type: "text",
					},
				},
			},
			&schema.Table{
				Name: "migrations",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "name",
						Type: "text",
					},
				},
			},
		},
	}
	c, err := New()
	if err != nil {
		t.Error(err)
	}
	err = c.LoadConfigFile(filepath.Join(testdataDir(), "config_test_tbls.yml"))
	if err != nil {
		t.Error(err)
	}
	err = c.ModifySchema(&s)
	if err != nil {
		t.Error(err)
	}

	if want := 1; len(s.Relations) != want {
		t.Errorf("got %v\nwant %v", len(s.Relations), want)
	}
	posts, _ := s.FindTableByName("posts")
	title, _ := posts.FindColumnByName("title")
	if want := "post title"; title.Comment != want {
		t.Errorf("got %v\nwant %v", title.Comment, want)
	}
	if want := 2; len(s.Tables) != want {
		t.Errorf("got %v\nwant %v", len(s.Tables), want)
	}
	if want := "mydatabase"; s.Name != want {
		t.Errorf("got %v\nwant %v", s.Name, want)
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	return dir
}
