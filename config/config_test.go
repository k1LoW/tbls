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
	err = config.Load(configFilepath, []string{})
	if err != nil {
		t.Fatal(err)
	}
	expected := ""
	if config.DSN != expected {
		t.Errorf("actual %v\nwant %v", config.DSN, expected)
	}
	expected2 := "schema"
	if config.DocPath != expected2 {
		t.Errorf("actual %v\nwant %v", config.DocPath, expected2)
	}
}

func TestLoadConfigFile(t *testing.T) {
	_ = os.Setenv("TBLS_TEST_PG_PASS", "pgpass")
	_ = os.Setenv("TBLS_TEST_PG_DOC_PATH", "sample/pg")
	configFilepath := filepath.Join(testdataDir(), "env_testdb_config.yml")
	config, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	err = config.LoadConfigFile(configFilepath)
	if err != nil {
		t.Fatal(err)
	}
	expected := "pg://root:pgpass@localhost:55432/testdb?sslmode=disable"
	if config.DSN != expected {
		t.Errorf("actual %v\nwant %v", config.DSN, expected)
	}
	expected2 := "sample/pg"
	if config.DocPath != expected2 {
		t.Errorf("actual %v\nwant %v", config.DocPath, expected2)
	}
}

var tests = []struct {
	value    string
	expected string
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
		actual, err := parseWithEnviron(tt.value)
		if err != nil {
			t.Fatal(err)
		}
		if actual != tt.expected {
			t.Errorf("actual %v\nwant %v", actual, tt.expected)
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
	err = c.LoadConfigFile(filepath.Join(testdataDir(), "schema_test_additional_data.yml"))
	if err != nil {
		t.Error(err)
	}
	err = c.MergeAdditionalData(&s)
	if err != nil {
		t.Error(err)
	}
	expected := 1
	actual := len(s.Relations)
	if actual != expected {
		t.Errorf("actual %v\nwant %v", actual, expected)
	}
	posts, _ := s.FindTableByName("posts")
	title, _ := posts.FindColumnByName("title")
	expected2 := "post title"
	actual2 := title.Comment
	if actual2 != expected2 {
		t.Errorf("actual %v\nwant %v", actual2, expected2)
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	return dir
}
