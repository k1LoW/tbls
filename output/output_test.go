package output

import (
	"testing"

	"github.com/SouhlInc/tbls/config"
	"github.com/SouhlInc/tbls/datasource"
	"github.com/SouhlInc/tbls/schema"
)

func TestDistance(t *testing.T) {
	dsn := config.DSN{URL: "json://../testdata/testdb.json"}
	s, err := datasource.Analyze(dsn)
	if err != nil {
		t.Errorf("%s", err)
	}
	var ut *schema.Table
	for _, t := range s.Tables {
		if t.Name == "public.users" {
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
	want = 7
	if len(tables) != want {
		t.Errorf("got %v\nwant %v", len(tables), want)
	}
	want = 6
	if len(relations) != want {
		t.Errorf("got %v\nwant %v", len(relations), want)
	}

	tables, relations, _ = ut.CollectTablesAndRelations(2, true)
	want = 7
	if len(tables) != want {
		t.Errorf("got %v\nwant %v", len(tables), want)
	}
	want = 11
	if len(relations) != want {
		t.Errorf("got %v\nwant %v", len(relations), want)
	}
}
