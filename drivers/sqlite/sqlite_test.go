//go:build sqlite

package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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

// posts has multiple triggers in testdata/ddl/sqlite.sql, created in an order that differs from name order.
func TestTriggersOrder(t *testing.T) {
	driver := New(db)
	sc := &schema.Schema{Name: "testdb.sqlite3"}
	if err := driver.Analyze(sc); err != nil {
		t.Fatal(err)
	}

	tbl, err := sc.FindTableByName("posts")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, trig := range tbl.Triggers {
		got = append(got, trig.Name)
	}
	want := []string{"update_posts_updated", "posts_fts_insert", "posts_fts_delete", "posts_fts_update"}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Error(diff)
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

func TestCompositePrimaryKey(t *testing.T) {
	tdb, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "pk.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	defer tdb.Close()
	if _, err := tdb.Exec(`CREATE TABLE user_roles (user_id INTEGER NOT NULL, role_id INTEGER NOT NULL, PRIMARY KEY (user_id, role_id));`); err != nil {
		t.Fatal(err)
	}
	sc := &schema.Schema{Name: "pk.sqlite3"}
	if err := New(tdb).Analyze(sc); err != nil {
		t.Fatal(err)
	}
	tbl, err := sc.FindTableByName("user_roles")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, c := range tbl.Constraints {
		if c.Type == "PRIMARY KEY" {
			got = append(got, c.Def)
		}
	}
	if diff := cmp.Diff(got, []string{"PRIMARY KEY (user_id, role_id)"}); diff != "" {
		t.Error(diff)
	}
}

func TestForeignKeyWithoutParentColumns(t *testing.T) {
	tests := []struct {
		name              string
		ddl               string
		wantDef           string
		wantParentColumns []string
		wantErr           bool
	}{
		{
			name:              "single column",
			ddl:               `CREATE TABLE parent (id INTEGER PRIMARY KEY); CREATE TABLE child (id INTEGER PRIMARY KEY, parent_id INTEGER REFERENCES parent);`,
			wantDef:           "FOREIGN KEY (parent_id) REFERENCES parent (id) ON UPDATE NO ACTION ON DELETE NO ACTION MATCH NONE",
			wantParentColumns: []string{"id"},
		},
		{
			name:              "composite key in primary key order",
			ddl:               `CREATE TABLE parent (a INTEGER, b INTEGER, PRIMARY KEY (b, a)); CREATE TABLE child (x INTEGER, y INTEGER, FOREIGN KEY (x, y) REFERENCES parent);`,
			wantDef:           "FOREIGN KEY (x, y) REFERENCES parent (b, a) ON UPDATE NO ACTION ON DELETE NO ACTION MATCH NONE",
			wantParentColumns: []string{"b", "a"},
		},
		{
			name:    "parent without primary key",
			ddl:     `CREATE TABLE parent (id INTEGER); CREATE TABLE child (parent_id INTEGER REFERENCES parent);`,
			wantErr: true,
		},
		{
			name:    "fewer columns than the parent primary key",
			ddl:     `CREATE TABLE parent (a INTEGER, b INTEGER, c INTEGER, PRIMARY KEY (a, b, c)); CREATE TABLE child (x INTEGER, y INTEGER, FOREIGN KEY (x, y) REFERENCES parent);`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tdb, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "fk.sqlite3"))
			if err != nil {
				t.Fatal(err)
			}
			defer tdb.Close()
			if _, err := tdb.Exec(tt.ddl); err != nil {
				t.Fatal(err)
			}
			sc := &schema.Schema{Name: "fk.sqlite3"}
			err = New(tdb).Analyze(sc)
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "can not resolve the parent key") {
					t.Errorf("got %v, want an error about the parent key", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(sc.Relations) != 1 {
				t.Fatalf("got %d relations, want 1", len(sc.Relations))
			}
			r := sc.Relations[0]
			if r.Def != tt.wantDef {
				t.Errorf("got %q\nwant %q", r.Def, tt.wantDef)
			}
			var got []string
			for _, c := range r.ParentColumns {
				got = append(got, c.Name)
			}
			if diff := cmp.Diff(got, tt.wantParentColumns); diff != "" {
				t.Error(diff)
			}
		})
	}
}
