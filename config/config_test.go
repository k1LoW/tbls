package config

import (
	"os"
	"path/filepath"
	"testing"
)

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

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	return dir
}
