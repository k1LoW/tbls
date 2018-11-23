package postgres

import (
	"database/sql"
	"os"
	"testing"

	"github.com/k1LoW/tbls/schema"
	_ "github.com/lib/pq"
	"github.com/xo/dburl"
)

var s *schema.Schema
var db *sql.DB

func TestMain(m *testing.M) {
	s = &schema.Schema{
		Name: "testdb",
	}
	db, _ = dburl.Open("pg://postgres:pgpass@localhost:55432/testdb?sslmode=disable")
	defer db.Close()
	exit := m.Run()
	if exit != 0 {
		os.Exit(exit)
	}
}

func TestAnalyzeView(t *testing.T) {
	driver := new(Postgres)
	err := driver.Analyze(db, s)
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
	driver := new(Postgres)
	d, err := driver.Info(db)
	if err != nil {
		t.Errorf("%v", err)
	}
	if d.Name != "postgres" {
		t.Errorf("actual %v\nwant %v", d.Name, "postgres")
	}
	if d.DatabaseVersion == "" {
		t.Errorf("actual not empty string.")
	}
}
