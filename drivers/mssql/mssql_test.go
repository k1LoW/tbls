//go:build mssql

package mssql

import (
	"database/sql"
	"log"
	"testing"

	"github.com/k1LoW/tbls/schema"
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

func TestTableWithNonClusteredPrimaryKey(t *testing.T) {
	driver := New(db)
	err := driver.Analyze(s)
	if err != nil {
		t.Error(err)
	}
	table, err := s.FindTableByName("tableWithOutClusterIndex")
	if err != nil {
		t.Fatal(err)
	}

	if table == nil {
		t.Fatal("tableWithOutClusterIndex not found")
	}

	// Should have 3 indexes (excluding HEAP)
	// 1. PK_tableWithOutClusterIndex (NONCLUSTERED PRIMARY KEY)
	// 2. IX_tableWithOutClusterIndex_testIndex
	// 3. IX_tableWithOutClusterIndex_testIndex_id
	if len(table.Indexes) != 3 {
		t.Errorf("got %v indexes\nwant 3 indexes (HEAP should be excluded)", len(table.Indexes))
		for i, idx := range table.Indexes {
			t.Logf("Index %d: %s - %s", i, idx.Name, idx.Def)
		}
	}

	for _, idx := range table.Indexes {
		if idx.Name == "" {
			t.Errorf("Found index with empty name (HEAP index should be filtered out)")
		}
	}

	var foundPK bool
	for _, idx := range table.Indexes {
		if idx.Name == "PK_tableWithOutClusterIndex" {
			foundPK = true
			if !containsString(idx.Def, "NONCLUSTERED") {
				t.Errorf("Primary key index definition should contain 'NONCLUSTERED', got: %s", idx.Def)
			}
			if !containsString(idx.Def, "PRIMARY KEY") {
				t.Errorf("Primary key index definition should contain 'PRIMARY KEY', got: %s", idx.Def)
			}
		}
	}

	if !foundPK {
		t.Error("Primary key index PK_tableWithOutClusterIndex not found")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsStringHelper(s, substr)))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
