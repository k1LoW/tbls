package output

import (
	"testing"

	"github.com/k1LoW/tbls/datasource"
	"github.com/k1LoW/tbls/schema"
)

func TestDistance(t *testing.T) {
	dsn := "json://../testdata/testdb.json"
	s, err := datasource.Analyze(dsn)
	if err != nil {
		t.Errorf("%s", err)
	}
	var ut *schema.Table
	for _, t := range s.Tables {
		if t.Name == "users" {
			ut = t
			break
		}
	}
	tables, relations, _ := ut.CollectTablesAndRelations(0, true)
	want := 1
	if len(tables) != want {
		t.Errorf("got %v\nwant %v", len(tables), want)
	}
	want = 0
	if len(relations) != want {
		t.Errorf("got %v\nwant %v", len(relations), want)
	}

	tables, relations, _ = ut.CollectTablesAndRelations(1, true)
	want = 5
	if len(tables) != want {
		t.Errorf("got %v\nwant %v", len(tables), want)
	}
	want = 4
	if len(relations) != want {
		t.Errorf("got %v\nwant %v", len(relations), want)
	}

	tables, relations, _ = ut.CollectTablesAndRelations(2, true)
	want = 5
	if len(tables) != want {
		t.Errorf("got %v\nwant %v", len(tables), want)
	}
	want = 9
	if len(relations) != want {
		t.Errorf("got %v\nwant %v", len(relations), want)
	}
}
