package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/tbls/schema"
)

func TestLoadDefault(t *testing.T) {
	configFilepath := filepath.Join(testdataDir(), "empty.yml")
	config, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	err = config.Load(configFilepath)
	if err != nil {
		t.Fatal(err)
	}
	want := ""
	if config.DSN != want {
		t.Errorf("got %v\nwant %v", config.DSN, want)
	}
	want2 := "dbdoc"
	if config.DocPath != want2 {
		t.Errorf("got %v\nwant %v", config.DocPath, want2)
	}
	want3 := "png"
	if config.ER.Format != want3 {
		t.Errorf("got %v\nwant %v", config.ER.Format, want3)
	}
	want4 := 1
	if *config.ER.Distance != want4 {
		t.Errorf("got %v\nwant %v", config.ER.Distance, want4)
	}
}

func TestLoadConfigFile(t *testing.T) {
	_ = os.Setenv("TBLS_TEST_PG_PASS", "pgpass")
	_ = os.Setenv("TBLS_TEST_PG_DOC_PATH", "sample/pg")
	configFilepath := filepath.Join(testdataDir(), "config_test_tbls_2.yml")
	config, err := NewConfig()
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
	c, err := NewConfig()
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
	want := 1
	got := len(s.Relations)
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
	posts, _ := s.FindTableByName("posts")
	title, _ := posts.FindColumnByName("title")
	want2 := "post title"
	got2 := title.Comment
	if got2 != want2 {
		t.Errorf("got %v\nwant %v", got2, want2)
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
	c, err := NewConfig()
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
	want := 2
	got := len(s.Tables)
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
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
	c, err := NewConfig()
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
	want := 1
	got := len(s.Relations)
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
	posts, _ := s.FindTableByName("posts")
	title, _ := posts.FindColumnByName("title")
	want2 := "post title"
	got2 := title.Comment
	if got2 != want2 {
		t.Errorf("got %v\nwant %v", got2, want2)
	}
	want3 := 2
	got3 := len(s.Tables)
	if got3 != want3 {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	return dir
}
