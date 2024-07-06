//go:build clickhouse

package clickhouse

import (
	"database/sql"
	"fmt"
	"github.com/k1LoW/tbls/schema"
	"github.com/stretchr/testify/assert"
	"github.com/xo/dburl"
	"os"
	"testing"
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
	assert.NoError(t, err)
	assert.Equal(t, "clickhouse", d.Name)
	assert.NotEmpty(t, d.DatabaseVersion)
}

func TestAnalyzeRegularTable(t *testing.T) {
	s := &schema.Schema{
		Name: schemaName,
	}
	driver := New(db)

	err := driver.Analyze(s)
	assert.NoError(t, err)

	assert.NotEmpty(t, s.Tables)
	assert.NotEmpty(t, s.Functions)

	table, err := s.FindTableByName("table_name")
	assert.NoError(t, err)

	assert.Len(t, table.Columns, 8)
	assert.Len(t, table.Indexes, 3)
}

func TestAnalyzeDictionary(t *testing.T) {
	s := &schema.Schema{
		Name: schemaName,
	}
	driver := New(db)

	err := driver.Analyze(s)
	assert.NoError(t, err)

	assert.NotEmpty(t, s.Tables)
	assert.NotEmpty(t, s.Functions)

	table, err := s.FindTableByName("id_value_dictionary")
	assert.NoError(t, err)

	assert.Len(t, table.Columns, 2)
	assert.Empty(t, table.Indexes)
}

func TestAnalyzeMaterializedView(t *testing.T) {
	s := &schema.Schema{
		Name: schemaName,
	}
	driver := New(db)

	err := driver.Analyze(s)
	assert.NoError(t, err)

	assert.NotEmpty(t, s.Tables)
	assert.NotEmpty(t, s.Functions)

	table, err := s.FindTableByName("materialized_view")
	assert.NoError(t, err)

	assert.Len(t, table.Columns, 2)
	assert.Empty(t, table.Indexes)
}

func TestAnalyzeView(t *testing.T) {
	s := &schema.Schema{
		Name: schemaName,
	}
	driver := New(db)

	err := driver.Analyze(s)
	assert.NoError(t, err)

	assert.NotEmpty(t, s.Tables)
	assert.NotEmpty(t, s.Functions)

	table, err := s.FindTableByName("view")
	assert.NoError(t, err)

	assert.Len(t, table.Columns, 5)
	assert.Empty(t, table.Indexes)
}
