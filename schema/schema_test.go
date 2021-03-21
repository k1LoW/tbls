package schema

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSchema_FindTableByName(t *testing.T) {
	schema := Schema{
		Name: "testschema",
		Tables: []*Table{
			&Table{
				Name:    "a",
				Comment: "table a",
			},
			&Table{
				Name:    "b",
				Comment: "table b",
			},
		},
	}
	table, _ := schema.FindTableByName("b")
	want := "table b"
	got := table.Comment
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestTable_FindColumnByName(t *testing.T) {
	table := Table{
		Name: "testtable",
		Columns: []*Column{
			&Column{
				Name:    "a",
				Comment: "column a",
			},
			&Column{
				Name:    "b",
				Comment: "column b",
			},
		},
	}
	column, _ := table.FindColumnByName("b")
	want := "column b"
	got := column.Comment
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestTable_FindConstrainsByColumnName(t *testing.T) {
	table := Table{
		Name: "testtable",
		Columns: []*Column{
			&Column{
				Name:    "a",
				Comment: "column a",
			},
			&Column{
				Name:    "b",
				Comment: "column b",
			},
		},
	}
	table.Constraints = []*Constraint{
		&Constraint{
			Name:              "PRIMARY",
			Type:              "PRIMARY KEY",
			Def:               "PRIMARY KEY(a)",
			ReferencedTable:   nil,
			Table:             &table.Name,
			Columns:           []string{"a"},
			ReferencedColumns: []string{},
		},
		&Constraint{
			Name:              "UNIQUE",
			Type:              "UNIQUE",
			Def:               "UNIQUE KEY a (b)",
			ReferencedTable:   nil,
			Table:             &table.Name,
			Columns:           []string{"b"},
			ReferencedColumns: []string{},
		},
	}

	got := table.FindConstrainsByColumnName("a")
	if want := 1; len(got) != want {
		t.Errorf("got %v\nwant %v", len(got), want)
	}
	if want := "PRIMARY"; got[0].Name != want {
		t.Errorf("got %v\nwant %v", got[0].Name, want)
	}
}

func TestSchema_Sort(t *testing.T) {
	schema := Schema{
		Name: "testschema",
		Tables: []*Table{
			&Table{
				Name:    "b",
				Comment: "table b",
			},
			&Table{
				Name:    "a",
				Comment: "table a",
				Columns: []*Column{
					&Column{
						Name:    "b",
						Comment: "column b",
					},
					&Column{
						Name:    "a",
						Comment: "column a",
					},
				},
			},
		},
	}
	_ = schema.Sort()
	want := "a"
	got := schema.Tables[0].Name
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
	want2 := "a"
	got2 := schema.Tables[0].Columns[0].Name
	if got2 != want2 {
		t.Errorf("got %v\nwant %v", got2, want2)
	}
}

func TestRepair(t *testing.T) {
	got := &Schema{}
	file, err := os.Open(filepath.Join(testdataDir(), "json_test_schema.json.golden"))
	if err != nil {
		t.Error(err)
	}
	dec := json.NewDecoder(file)
	err = dec.Decode(got)
	if err != nil {
		t.Error(err)
	}
	want := newTestSchema()
	err = got.Repair()
	if err != nil {
		t.Error(err)
	}

	for i, tt := range got.Tables {
		compareStrings(t, got.Tables[i].Name, want.Tables[i].Name)
		for j := range tt.Columns {
			compareStrings(t, got.Tables[i].Columns[j].Name, want.Tables[i].Columns[j].Name)
			for k := range got.Tables[i].Columns[j].ParentRelations {
				compareStrings(t, got.Tables[i].Columns[j].ParentRelations[k].Table.Name, want.Tables[i].Columns[j].ParentRelations[k].Table.Name)
				compareStrings(t, got.Tables[i].Columns[j].ParentRelations[k].ParentTable.Name, want.Tables[i].Columns[j].ParentRelations[k].ParentTable.Name)
			}
			for k := range got.Tables[i].Columns[j].ChildRelations {
				compareStrings(t, got.Tables[i].Columns[j].ChildRelations[k].Table.Name, want.Tables[i].Columns[j].ChildRelations[k].Table.Name)
				compareStrings(t, got.Tables[i].Columns[j].ChildRelations[k].ParentTable.Name, want.Tables[i].Columns[j].ChildRelations[k].ParentTable.Name)
			}
		}
	}

	if len(got.Relations) != len(want.Relations) {
		t.Errorf("got %#v\nwant %#v", got.Relations, want.Relations)
	}
}

func compareStrings(tb testing.TB, got, want string) {
	tb.Helper()
	if got != want {
		tb.Errorf("got %#v\nwant %#v", got, want)
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	return dir
}

func newTestSchema() *Schema {
	ca := &Column{
		Name:     "a",
		Type:     "bigint(20)",
		Comment:  "column a",
		Nullable: false,
	}
	cb := &Column{
		Name:     "b",
		Type:     "text",
		Comment:  "column b",
		Nullable: true,
	}

	ta := &Table{
		Name:    "a",
		Type:    "BASE TABLE",
		Comment: "table a",
		Columns: []*Column{
			ca,
			&Column{
				Name:     "a2",
				Type:     "datetime",
				Comment:  "column a2",
				Nullable: false,
				Default: sql.NullString{
					String: "CURRENT_TIMESTAMP",
					Valid:  true,
				},
			},
		},
	}

	tb := &Table{
		Name:    "b",
		Type:    "BASE TABLE",
		Comment: "table b",
		Columns: []*Column{
			cb,
			&Column{
				Name:     "b2",
				Comment:  "column b2",
				Type:     "text",
				Nullable: true,
			},
		},
	}
	r := &Relation{
		Table:         ta,
		Columns:       []*Column{ca},
		ParentTable:   tb,
		ParentColumns: []*Column{cb},
	}
	ca.ParentRelations = []*Relation{r}
	cb.ChildRelations = []*Relation{r}

	s := &Schema{
		Name: "testschema",
		Tables: []*Table{
			ta,
			tb,
		},
		Relations: []*Relation{
			r,
		},
		Driver: &Driver{
			Name:            "testdriver",
			DatabaseVersion: "1.0.0",
		},
	}
	return s
}
