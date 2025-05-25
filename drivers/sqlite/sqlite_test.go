//go:build sqlite

package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/SouhlInc/tbls/schema"
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
	driver := New(db)
	err := driver.Analyze(s)
	if err != nil {
		t.Errorf("%v", err)
	}
	view, _ := s.FindTableByName("post_comments")
	want := view.Def
	if want == "" {
		t.Errorf("got not empty string.")
	}
}

func TestInfo(t *testing.T) {
	driver := New(db)
	d, err := driver.Info()
	if err != nil {
		t.Errorf("%v", err)
	}
	if d.Name != "sqlite" {
		t.Errorf("got %v\nwant %v", d.Name, "sqlite")
	}
	if d.DatabaseVersion == "" {
		t.Errorf("got not empty string.")
	}
}

func TestParseCheckConstraints(t *testing.T) {
	table := &schema.Table{
		Name: "check_constraints",
		Columns: []*schema.Column{
			&schema.Column{
				Name: "id",
			},
			&schema.Column{
				Name: "col",
			},
			&schema.Column{
				Name: "brackets",
			},
			&schema.Column{
				Name: "checkcheck",
			},
			&schema.Column{
				Name: "downcase",
			},
			&schema.Column{
				Name: "nl",
			},
		},
	}
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
	tableName := "check_constraints"
	want := []*schema.Constraint{
		&schema.Constraint{
			Name:    "-",
			Type:    "CHECK",
			Def:     "CHECK(length(col) > 4)",
			Table:   &tableName,
			Columns: []string{"col"},
		},
		&schema.Constraint{
			Name:    "-",
			Type:    "CHECK",
			Def:     "CHECK(((length(brackets) > 4)))",
			Table:   &tableName,
			Columns: []string{"brackets"},
		},
		&schema.Constraint{
			Name:    "-",
			Type:    "CHECK",
			Def:     "CHECK(length(checkcheck) > 4)",
			Table:   &tableName,
			Columns: []string{"checkcheck"},
		},
		&schema.Constraint{
			Name:    "-",
			Type:    "CHECK",
			Def:     "check(length(downcase) > 4)",
			Table:   &tableName,
			Columns: []string{"downcase"},
		},
		&schema.Constraint{
			Name:    "-",
			Type:    "CHECK",
			Def:     "check(length(nl) > 4 OR nl != 'ln')",
			Table:   &tableName,
			Columns: []string{"nl"},
		},
	}
	got := parseCheckConstraints(table, sql)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %#v\nwant: %#v", got, want)
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(filepath.Dir(wd)), "testdata"))
	return dir
}
