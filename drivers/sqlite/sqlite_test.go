package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/k1LoW/tbls/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xo/dburl"
	"path/filepath"
)

var s *schema.Schema
var db *sql.DB

func TestMain(m *testing.M) {
	s = &schema.Schema{
		Name: "testdb.sqlite3",
	}
	dir, _ := os.Getwd()
	sqliteFilepath, _ := filepath.Abs(filepath.Join(filepath.Dir(filepath.Dir(dir)), "test", "testdb.sqlite3"))

	db, _ = dburl.Open(fmt.Sprintf("sq://%s", sqliteFilepath))
	defer db.Close()
	exit := m.Run()
	if exit != 0 {
		os.Exit(exit)
	}
}

func TestAnalyzeView(t *testing.T) {
	driver := new(Sqlite)
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
