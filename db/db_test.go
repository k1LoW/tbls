package db

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"testing"
)

func TestAnalyzeSchema(t *testing.T) {
	schema, err := Analyze("my://root:mypass@localhost:33306/testdb")
	if err != nil {
		t.Errorf("%s", err)
	}
	expected := "testdb"
	actual := schema.Name
	if actual != expected {
		t.Errorf("actual %v\nwant %v", actual, expected)
	}
}

func TestAnalyzeTables(t *testing.T) {
	schema, err := Analyze("my://root:mypass@localhost:33306/testdb")
	if err != nil {
		t.Errorf("%s", err)
	}
	expected := 7
	actual := len(schema.Tables)
	if actual != expected {
		t.Errorf("actual %v\nwant %v", actual, expected)
	}
}

func TestAnalyzeRelations(t *testing.T) {
	schema, err := Analyze("my://root:mypass@localhost:33306/testdb")
	if err != nil {
		t.Errorf("%s", err)
	}
	expected := 5
	actual := len(schema.Relations)
	if actual != expected {
		t.Errorf("actual %v\nwant %v", actual, expected)
	}
}
