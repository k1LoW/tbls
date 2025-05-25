//go:build mssql

package mssql

import (
	"database/sql"
	"log"
	"testing"

	"github.com/SouhlInc/tbls/schema"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/xo/dburl"
)

var s *schema.Schema
var db *sql.DB
var err error

func TestMain(m *testing.M) {
	s = &schema.Schema{
		Name: "testdb",
	}
	db, err = dburl.Open("ms://SA:MSSQLServer-Passw0rd@localhost:11433/instance?database=testdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	_ = m.Run()
}

func TestAnalyzeView(t *testing.T) {
	driver := New(db)
	err := driver.Analyze(s)
	if err != nil {
		t.Error(err)
	}
	view, err := s.FindTableByName("post_comments")
	if err != nil {
		t.Fatal(err)
	}
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
	if d.Name != "sqlserver" {
		t.Errorf("got %v\nwant %v", d.Name, "sqlserver")
	}
	if d.DatabaseVersion == "" {
		t.Errorf("got not empty string.")
	}
}
