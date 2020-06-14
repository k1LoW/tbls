package mysql

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/k1LoW/tbls/schema"
	"github.com/xo/dburl"
)

var s *schema.Schema
var db *sql.DB

func TestMain(m *testing.M) {
	s = &schema.Schema{
		Name: "testdb",
	}
	db, _ = dburl.Open("my://root:mypass@localhost:33306/testdb")
	defer db.Close()
	exit := m.Run()
	if exit != 0 {
		os.Exit(exit)
	}
}

func TestAnalyzeView(t *testing.T) {
	driver, err := New(db)
	if err != nil {
		t.Fatal(err)
	}

	err = driver.Analyze(s)
	if err != nil {
		t.Fatal(err)
	}
	view, _ := s.FindTableByName("post_comments")
	want := view.Def
	if want == "" {
		t.Errorf("got not empty string.")
	}
}

func TestInfo(t *testing.T) {
	driver, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	d, err := driver.Info()
	if err != nil {
		t.Fatal(err)
	}
	if d.Name != "mysql" {
		t.Errorf("got %v\nwant %v", d.Name, "mysql")
	}
	if d.DatabaseVersion == "" {
		t.Errorf("got not empty string.")
	}
}
