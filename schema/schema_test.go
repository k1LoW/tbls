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
	expected := "table b"
	actual := table.Comment
	if actual != expected {
		t.Errorf("actual %v\nwant %v", actual, expected)
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
	expected := "column b"
	actual := column.Comment
	if actual != expected {
		t.Errorf("actual %v\nwant %v", actual, expected)
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
	expected := "a"
	actual := schema.Tables[0].Name
	if actual != expected {
		t.Errorf("actual %v\nwant %v", actual, expected)
	}
	expected2 := "a"
	actual2 := schema.Tables[0].Columns[0].Name
	if actual2 != expected2 {
		t.Errorf("actual %v\nwant %v", actual2, expected2)
	}
}

func TestAddAditionalData(t *testing.T) {
	schema := Schema{
		Name: "testschema",
		Tables: []*Table{
			&Table{
				Name:    "users",
				Comment: "users comment",
				Columns: []*Column{
					&Column{
						Name: "id",
						Type: "serial",
					},
					&Column{
						Name: "username",
						Type: "text",
					},
				},
			},
			&Table{
				Name:    "posts",
				Comment: "posts comment",
				Columns: []*Column{
					&Column{
						Name: "id",
						Type: "serial",
					},
					&Column{
						Name: "user_id",
						Type: "int",
					},
					&Column{
						Name: "title",
						Type: "text",
					},
				},
			},
		},
	}
	err := schema.LoadAdditionalData(filepath.Join(testdataDir(), "schema_test_additional_data.yml"))
	if err != nil {
		t.Error(err)
	}
	expected := 1
	actual := len(schema.Relations)
	if actual != expected {
		t.Errorf("actual %v\nwant %v", actual, expected)
	}
	posts, _ := schema.FindTableByName("posts")
	title, _ := posts.FindColumnByName("title")
	expected2 := "post title"
	actual2 := title.Comment
	if actual2 != expected2 {
		t.Errorf("actual %v\nwant %v", actual2, expected2)
	}
}

func TestRepair(t *testing.T) {
	actual := &Schema{}
	file, err := os.Open(filepath.Join(testdataDir(), "json_test_schema.json.golden"))
	if err != nil {
		t.Error(err)
	}
	dec := json.NewDecoder(file)
	dec.Decode(actual)
	expected := newTestSchema()
	err = actual.Repair()
	if err != nil {
		t.Error(err)
	}

	for i, tt := range actual.Tables {
		for j, c := range tt.Columns {
			for _, pr := range actual.Tables[i].Columns[j].ParentRelations {
				if pr.Columns[0] != c {
					t.Errorf("actual %#v\nwant %#v", pr.Columns[0], c)
				}
			}
		}
	}

	if actual != expected {
		t.Errorf("actual %v\nwant %v", actual, expected)
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
	}
	return s
}
