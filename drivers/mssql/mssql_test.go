//go:build mssql

package mssql

import (
	"database/sql"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func TestTriggersOrder(t *testing.T) {
	if _, err := db.Exec(`DROP TABLE IF EXISTS tbls_trigger_order`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE tbls_trigger_order (a int)`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(`DROP TABLE IF EXISTS tbls_trigger_order`)
	}()
	// Create the triggers in descending name order so that creation order and name order differ.
	// CREATE TRIGGER must be the first statement in a batch, so each one is executed on its own.
	if _, err := db.Exec(`CREATE TRIGGER trg_tbls_trigger_order_b ON tbls_trigger_order AFTER INSERT AS BEGIN SET NOCOUNT ON; END`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER trg_tbls_trigger_order_a ON tbls_trigger_order AFTER INSERT AS BEGIN SET NOCOUNT ON; END`); err != nil {
		t.Fatal(err)
	}

	driver := New(db)
	sc := &schema.Schema{Name: "testdb"}
	if err := driver.Analyze(sc); err != nil {
		t.Fatal(err)
	}

	tbl, err := sc.FindTableByName("tbls_trigger_order")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, trig := range tbl.Triggers {
		got = append(got, trig.Name)
	}
	want := []string{"trg_tbls_trigger_order_b", "trg_tbls_trigger_order_a"}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Error(diff)
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

func TestFunctionArgumentsOrder(t *testing.T) {
	if _, err := db.Exec(`DROP PROCEDURE IF EXISTS tbls_param_order`); err != nil {
		t.Fatal(err)
	}
	// Parameter names sort as @alpha, @mid, @zeta, so declaration order and name
	// order differ and an unordered aggregate would be visible.
	if _, err := db.Exec(`CREATE PROCEDURE tbls_param_order @zeta int, @alpha nvarchar(10), @mid bit AS BEGIN SET NOCOUNT ON; END`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(`DROP PROCEDURE IF EXISTS tbls_param_order`)
	}()

	driver := New(db)
	sc := &schema.Schema{Name: "testdb"}
	if err := driver.Analyze(sc); err != nil {
		t.Fatal(err)
	}

	var got string
	for _, f := range sc.Functions {
		if f.Name == "dbo.tbls_param_order" {
			got = f.Arguments
			break
		}
	}
	want := "@zeta int, @alpha nvarchar, @mid bit"
	if diff := cmp.Diff(got, want); diff != "" {
		t.Error(diff)
	}
}

func TestIndexIncludedColumnsOrder(t *testing.T) {
	if _, err := db.Exec(`DROP TABLE IF EXISTS tbls_index_include`); err != nil {
		t.Fatal(err)
	}
	// Three INCLUDE columns all carry key_ordinal = 0, so key_ordinal alone does
	// not order them and index_column_id is what makes the aggregate total.
	if _, err := db.Exec(`CREATE TABLE tbls_index_include (k int NOT NULL, z int, y int, x int)`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(`DROP TABLE IF EXISTS tbls_index_include`)
	}()
	if _, err := db.Exec(`CREATE INDEX ix_tbls_index_include ON tbls_index_include (k) INCLUDE (z, y, x)`); err != nil {
		t.Fatal(err)
	}

	driver := New(db)
	sc := &schema.Schema{Name: "testdb"}
	if err := driver.Analyze(sc); err != nil {
		t.Fatal(err)
	}

	tbl, err := sc.FindTableByName("tbls_index_include")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, idx := range tbl.Indexes {
		if idx.Name == "ix_tbls_index_include" {
			got = idx.Columns
			break
		}
	}
	// key_ordinal = 0 sorts the included columns ahead of the key column; within
	// them index_column_id follows the INCLUDE list.
	want := []string{"z", "y", "x", "k"}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Error(diff)
	}
}
