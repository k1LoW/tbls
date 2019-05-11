package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/k1LoW/tbls/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xo/dburl"
)

var s *schema.Schema
var db *sql.DB

func TestMain(m *testing.M) {
	s = &schema.Schema{
		Name: "testdb.sqlite3",
	}
	sqliteFilepath := filepath.Join(testdataDir(), "testdb.sqlite3")

	db, _ = dburl.Open(fmt.Sprintf("sq://%s", sqliteFilepath))
	defer db.Close()
	exit := m.Run()
	if exit != 0 {
		os.Exit(exit)
	}
}

func TestAnalyzeView(t *testing.T) {
	driver := NewSqlite(db)
	err := driver.Analyze(s)
	if err != nil {
		t.Errorf("%v", err)
	}
	view, _ := s.FindTableByName("post_comments")
	expected := view.Def
	if expected == "" {
		t.Errorf("actual not empty string.")
	}
}

func TestInfo(t *testing.T) {
	driver := NewSqlite(db)
	d, err := driver.Info()
	if err != nil {
		t.Errorf("%v", err)
	}
	if d.Name != "sqlite" {
		t.Errorf("actual %v\nwant %v", d.Name, "sqlite")
	}
	if d.DatabaseVersion == "" {
		t.Errorf("actual not empty string.")
	}
}

func TestParseCheckConstraints(t *testing.T) {
	sql := `CREATE TABLE check_constraints (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  col TEXT CHECK(length(col) > 4),
  brackets TEXT UNIQUE NOT NULL CHECK(((length(brackets) > 4))),
  checkcheck TEXT UNIQUE NOT NULL CHECK(length(checkcheck) > 4),
  downcase TEXT UNIQUE NOT NULL check(length(downcase) > 4),
  nl TEXT UNIQUE NOT
    NULL check(length(nl) > 4 OR
      nl != 'ln')
);`
	expected := []*schema.Constraint{
		&schema.Constraint{
			Name: "-",
			Type: "CHECK",
			Def:  "CHECK(length(col) > 4)",
		},
		&schema.Constraint{
			Name: "-",
			Type: "CHECK",
			Def:  "CHECK(((length(brackets) > 4)))",
		},
		&schema.Constraint{
			Name: "-",
			Type: "CHECK",
			Def:  "CHECK(length(checkcheck) > 4)",
		},
		&schema.Constraint{
			Name: "-",
			Type: "CHECK",
			Def:  "check(length(downcase) > 4)",
		},
		&schema.Constraint{
			Name: "-",
			Type: "CHECK",
			Def:  "check(length(nl) > 4 OR nl != 'ln')",
		},
	}
	actual := parseCheckConstraints(sql)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got: %#v\nwant: %#v", actual, expected)
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(filepath.Dir(wd)), "testdata"))
	return dir
}
