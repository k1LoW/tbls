//go:build clickhouse

package clickhouse

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/k1LoW/tbls/schema"
	"github.com/xo/dburl"
)

var db *sql.DB

const schemaName = "testdb"

func TestMain(m *testing.M) {
	var err error
	db, err = dburl.Open("clickhouse://default@localhost:9000/testdb")
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	exit := m.Run()
	if exit != 0 {
		os.Exit(exit)
	}
}

func TestInfo(t *testing.T) {
	driver := New(db)
	d, err := driver.Info()
	if err != nil {
		t.Errorf("%v", err)
	}

	if d.Name != "clickhouse" {
		t.Error("Driver name should be \"clickhouse\"")
	}

	if d.DatabaseVersion == "" {
		t.Error("DatabaseVersion should not be empty")
	}
}

func TestAnalyzeRegularTable(t *testing.T) {
	s := &schema.Schema{
		Name: schemaName,
	}
	driver := New(db)

	if err := driver.Analyze(s); err != nil {
		t.Errorf("%v", err)
	}

	if len(s.Tables) == 0 {
		t.Error("Tables shouldn't be empty")
	}

	if len(s.Functions) == 0 {
		t.Error("Functions shouldn't be empty")
	}

	table, err := s.FindTableByName("table_name")
	if err != nil {
		t.Errorf("%v", err)
	}

	if len(table.Columns) != 8 {
		t.Error("There should be 8 columns")
	}

	if len(table.Indexes) != 3 {
		t.Error("There should be 3 indexes")
	}
}

func TestAnalyzeDictionary(t *testing.T) {
	s := &schema.Schema{
		Name: schemaName,
	}
	driver := New(db)

	if err := driver.Analyze(s); err != nil {
		t.Errorf("%v", err)
	}

	if len(s.Tables) == 0 {
		t.Error("Tables shouldn't be empty")
	}

	if len(s.Functions) == 0 {
		t.Error("Functions shouldn't be empty")
	}

	table, err := s.FindTableByName("id_value_dictionary")
	if err != nil {
		t.Errorf("%v", err)
	}

	if len(table.Columns) != 2 {
		t.Error("There should be 2 columns")
	}

	if len(table.Indexes) > 0 {
		t.Error("Indexes should be empty")
	}
}

func TestAnalyzeMaterializedView(t *testing.T) {
	s := &schema.Schema{
		Name: schemaName,
	}
	driver := New(db)

	err := driver.Analyze(s)
	if err != nil {
		t.Errorf("%v", err)
	}

	if len(s.Tables) == 0 {
		t.Error("Tables shouldn't be empty")
	}

	if len(s.Functions) == 0 {
		t.Error("Functions shouldn't be empty")
	}

	table, err := s.FindTableByName("materialized_view")
	if err != nil {
		t.Errorf("%v", err)
	}

	if len(table.Columns) != 2 {
		t.Error("There should be 2 columns")
	}

	if len(table.Indexes) > 0 {
		t.Error("Indexes should be empty")
	}
}

func TestAnalyzeView(t *testing.T) {
	s := &schema.Schema{
		Name: schemaName,
	}
	driver := New(db)

	err := driver.Analyze(s)
	if err != nil {
		t.Errorf("%v", err)
	}

	if len(s.Tables) == 0 {
		t.Error("Tables shouldn't be empty")
	}

	if len(s.Functions) == 0 {
		t.Error("Functions shouldn't be empty")
	}

	table, err := s.FindTableByName("view")
	if err != nil {
		t.Errorf("%v", err)
	}

	if len(table.Columns) != 5 {
		t.Error("There should be 5 columns")
	}

	if len(table.Indexes) > 0 {
		t.Error("Indexes should be empty")
	}
}
