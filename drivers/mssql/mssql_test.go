package mssql

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/k1LoW/tbls/schema"
	"github.com/xo/dburl"
)

var s *schema.Schema
var db *sql.DB

func TestMain(m *testing.M) {
	s = &schema.Schema{
		Name: "testdb",
	}
	db, _ = dburl.Open("ms://SA:MSSQLServer-Passw0rd@localhost:11433/testdb")
	defer db.Close()
	exit := m.Run()
	if exit != 0 {
		os.Exit(exit)
	}
}

func TestAnalyzeView(t *testing.T) {
	driver := NewMssql(db)
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
	driver := NewMssql(db)
	d, err := driver.Info()
	if err != nil {
		t.Errorf("%v", err)
	}
	if d.Name != "mssql" {
		t.Errorf("actual %v\nwant %v", d.Name, "mssql")
	}
	if d.DatabaseVersion == "" {
		t.Errorf("actual not empty string.")
	}
}
