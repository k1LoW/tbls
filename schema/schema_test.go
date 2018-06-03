package schema

import (
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
	additionalBytes := []byte(`
---
relations:
  -
    table: posts
    columns:
      - user_id
    parentTable: users
    parentColumns:
      - id
    def: posts->users
comments:
  -
    table: posts
    columnComments:
      title: post title
`)
	_ = schema.AddAdditionalData(additionalBytes)
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
