//go:build postgres

package postgres

import (
	"database/sql"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/SouhlInc/tbls/schema"
	_ "github.com/lib/pq"
	"github.com/xo/dburl"
)

var s *schema.Schema
var db *sql.DB

func TestMain(m *testing.M) {
	s = &schema.Schema{
		Name: "testdb",
	}
	db, _ = dburl.Open("pg://postgres:pgpass@localhost:55413/testdb?sslmode=disable")
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
		t.Errorf("%+v", err)
	}
	view, _ := s.FindTableByName("post_comments")
	want := view.Def
	if want == "" {
		t.Errorf("got not empty string.")
	}
}

func TestExtraDef(t *testing.T) {
	driver := New(db)
	if err := driver.Analyze(s); err != nil {
		t.Fatal(err)
	}
	tbl, _ := s.FindTableByName("comments")
	{
		c, _ := tbl.FindColumnByName("post_id_desc")
		got := c.ExtraDef
		if want := "GENERATED ALWAYS AS (post_id * '-1'::integer) STORED"; got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}
	}
}

func TestInfo(t *testing.T) {
	driver := New(db)
	d, err := driver.Info()
	if err != nil {
		t.Errorf("%v", err)
	}
	if d.Name != "postgres" {
		t.Errorf("got %v\nwant %v", d.Name, "postgres")
	}
	if d.DatabaseVersion == "" {
		t.Errorf("got not empty string.")
	}
}

func TestParseFK(t *testing.T) {
	tests := []struct {
		in              string
		wantCols        []string
		wantParentTable string
		wantParentCols  []string
	}{
		{"FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE", []string{"user_id"}, "users", []string{"id"}},
		{"FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL (user_id)", []string{"user_id"}, "users", []string{"id"}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			gotCols, gotParentTable, gotParentCols, err := parseFK(tt.in)
			if err != nil {
				t.Error(err)
				return
			}
			if diff := cmp.Diff(gotCols, tt.wantCols, nil); diff != "" {
				t.Error(diff)
			}
			if gotParentTable != tt.wantParentTable {
				t.Errorf("got %v want %v", gotParentTable, tt.wantParentTable)
			}
			if diff := cmp.Diff(gotParentCols, tt.wantParentCols, nil); diff != "" {
				t.Error(diff)
			}
		})
	}
}
