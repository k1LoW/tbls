package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/k1LoW/tbls/schema"
)

func TestLoadDefault(t *testing.T) {
	configFilepath := filepath.Join(testdataDir(), "empty.yml")
	config, err := New()
	if err != nil {
		t.Fatal(err)
	}
	err = config.Load(configFilepath)
	if err != nil {
		t.Fatal(err)
	}

	if want := ""; config.DSN.URL != want {
		t.Errorf("got %v\nwant %v", config.DSN.URL, want)
	}
	if want := "dbdoc"; config.DocPath != want {
		t.Errorf("got %v\nwant %v", config.DocPath, want)
	}
	if want := "svg"; config.ER.Format != want {
		t.Errorf("got %v\nwant %v", config.ER.Format, want)
	}
	if want := 1; *config.ER.Distance != want {
		t.Errorf("got %v\nwant %v", config.ER.Distance, want)
	}
}

func TestLoadConfigFile(t *testing.T) {
	_ = os.Setenv("TBLS_TEST_PG_PASS", "pgpass")
	_ = os.Setenv("TBLS_TEST_PG_DOC_PATH", "sample/pg")
	configFilepath := filepath.Join(testdataDir(), "config_test_tbls_2.yml")
	config, err := New()
	if err != nil {
		t.Fatal(err)
	}
	err = config.LoadConfigFile(configFilepath)
	if err != nil {
		t.Fatal(err)
	}

	if want := "pg://root:pgpass@localhost:55432/testdb?sslmode=disable"; config.DSN.URL != want {
		t.Errorf("got %v\nwant %v", config.DSN.URL, want)
	}

	if want := "sample/pg"; config.DocPath != want {
		t.Errorf("got %v\nwant %v", config.DocPath, want)
	}

	if want := "INDEX"; config.MergedDict.Lookup("Indexes") != want {
		t.Errorf("got %v\nwant %v", config.MergedDict.Lookup("Indexes"), want)
	}
}

func TestDuplicateConfigFile(t *testing.T) {
	config := &Config{
		root: filepath.Join(testdataDir(), "config"),
	}
	got := config.LoadConfigFile("")
	want := "duplicate config file [.tbls.yml, tbls.yml]"
	if fmt.Sprintf("%v", got) != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

var tests = []struct {
	value string
	want  string
}{
	{"${TBLS_ONE}/${TBLS_TWO}", "one/two"},
	{"${TBLS_ONE}/${TBLS_TWO}/${TBLS_NONE}", "one/two/"},
	{"${{TBLS_ONE}}", "${{TBLS_ONE}}"},
	{"{{.TBLS_ONE}}/{{.TBLS_TWO}}", "one/two"},
}

func TestParseWithEnvirion(t *testing.T) {
	_ = os.Setenv("TBLS_ONE", "one")
	_ = os.Setenv("TBLS_TWO", "two")
	for _, tt := range tests {
		got, err := parseWithEnviron(tt.value)
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func TestMergeAditionalData(t *testing.T) {
	s := schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			&schema.Table{
				Name:    "users",
				Comment: "users comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "username",
						Type: "text",
					},
				},
				Indexes: []*schema.Index{
					&schema.Index{
						Name: "user_index",
					},
				},
				Constraints: []*schema.Constraint{
					&schema.Constraint{
						Name: "PRIMARY",
					},
				},
			},
			&schema.Table{
				Name:    "posts",
				Comment: "posts comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "user_id",
						Type: "int",
					},
					&schema.Column{
						Name: "title",
						Type: "text",
					},
				},
				Triggers: []*schema.Trigger{
					&schema.Trigger{
						Name: "update_posts_title",
					},
				},
			},
		},
	}
	c, err := New()
	if err != nil {
		t.Error(err)
	}
	err = c.LoadConfigFile(filepath.Join(testdataDir(), "config_test_tbls.yml"))
	if err != nil {
		t.Error(err)
	}
	err = c.MergeAdditionalData(&s)
	if err != nil {
		t.Error(err)
	}
	if want := 1; len(s.Relations) != want {
		t.Errorf("got %v\nwant %v", len(s.Relations), want)
	}
	users, _ := s.FindTableByName("users")
	posts, _ := s.FindTableByName("posts")
	title, _ := posts.FindColumnByName("title")
	if want := "post title"; title.Comment != want {
		t.Errorf("got %v\nwant %v", title.Comment, want)
	}

	index, err := users.FindIndexByName("user_index")
	if err != nil {
		t.Fatal(err)
	}
	if want := "user index"; index.Comment != want {
		t.Errorf("got %v want %v", index.Comment, want)
	}

	constraint, err := users.FindConstraintByName("PRIMARY")
	if err != nil {
		t.Fatal(err)
	}
	if want := "PRIMARY(id)"; constraint.Comment != want {
		t.Errorf("got %v want %v", constraint.Comment, want)
	}

	trigger, err := posts.FindTriggerByName("update_posts_title")
	if err != nil {
		t.Fatal(err)
	}
	if want := "update posts title"; trigger.Comment != want {
		t.Errorf("got %v want %v", trigger.Comment, want)
	}
}

func TestFilterTables(t *testing.T) {
	s := schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			&schema.Table{
				Name:    "users",
				Comment: "users comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "username",
						Type: "text",
					},
				},
			},
			&schema.Table{
				Name:    "categories",
				Comment: "categories comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "name",
						Type: "text",
					},
				},
			},
			&schema.Table{
				Name:    "posts",
				Comment: "posts comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "user_id",
						Type: "int",
					},
					&schema.Column{
						Name: "category_id",
						Type: "int",
					},
					&schema.Column{
						Name: "title",
						Type: "text",
					},
				},
			},
			&schema.Table{
				Name:    "user_options",
				Comment: "user_options comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "user_id",
						Type: "int",
					},
					&schema.Column{
						Name: "email",
						Type: "text",
					},
				},
			},
			&schema.Table{
				Name: "schema_migrations",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "name",
						Type: "text",
					},
				},
			},
		},
	}
	usersTable, err := s.FindTableByName("users")
	if err != nil {
		t.Error(err)
	}
	categoriesTable, err := s.FindTableByName("categories")
	if err != nil {
		t.Error(err)
	}
	postsTable, err := s.FindTableByName("posts")
	if err != nil {
		t.Error(err)
	}
	userOptionsTable, err := s.FindTableByName("user_options")
	if err != nil {
		t.Error(err)
	}
	s.Relations = []*schema.Relation{
		&schema.Relation{
			Table:       userOptionsTable,
			ParentTable: usersTable,
		},
		&schema.Relation{
			Table:       postsTable,
			ParentTable: categoriesTable,
		},
	}
	c, err := New()
	if err != nil {
		t.Error(err)
	}
	err = c.LoadConfigFile(filepath.Join(testdataDir(), "config_test_tbls.yml"))
	if err != nil {
		t.Error(err)
	}
	err = c.FilterTables(&s)
	if err != nil {
		t.Error(err)
	}
	if want := 2; len(s.Tables) != want {
		t.Errorf("got %v\nwant %v", len(s.Tables), want)
	}
	if want := 0; len(s.Relations) != want {
		t.Errorf("got %v\nwant %v", len(s.Relations), want)
	}
}

func TestModifySchema(t *testing.T) {
	s := schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			&schema.Table{
				Name:    "users",
				Comment: "users comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "username",
						Type: "text",
					},
				},
				Indexes: []*schema.Index{
					&schema.Index{
						Name: "user_index",
					},
				},
				Constraints: []*schema.Constraint{
					&schema.Constraint{
						Name: "PRIMARY",
					},
				},
			},
			&schema.Table{
				Name:    "posts",
				Comment: "posts comment",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "user_id",
						Type: "int",
					},
					&schema.Column{
						Name: "title",
						Type: "text",
					},
				},
				Triggers: []*schema.Trigger{
					&schema.Trigger{
						Name: "update_posts_title",
					},
				},
			},
			&schema.Table{
				Name: "migrations",
				Columns: []*schema.Column{
					&schema.Column{
						Name: "id",
						Type: "serial",
					},
					&schema.Column{
						Name: "name",
						Type: "text",
					},
				},
			},
		},
	}
	c, err := New()
	if err != nil {
		t.Error(err)
	}
	err = c.LoadConfigFile(filepath.Join(testdataDir(), "config_test_tbls.yml"))
	if err != nil {
		t.Error(err)
	}
	err = c.ModifySchema(&s)
	if err != nil {
		t.Error(err)
	}

	if want := 1; len(s.Relations) != want {
		t.Errorf("got %v\nwant %v", len(s.Relations), want)
	}
	posts, _ := s.FindTableByName("posts")
	title, _ := posts.FindColumnByName("title")
	if want := "post title"; title.Comment != want {
		t.Errorf("got %v\nwant %v", title.Comment, want)
	}
	if want := 2; len(s.Tables) != want {
		t.Errorf("got %v\nwant %v", len(s.Tables), want)
	}
	if want := "mydatabase"; s.Name != want {
		t.Errorf("got %v\nwant %v", s.Name, want)
	}
}

func TestMaskedDSN(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{
			"pg://root:pgpass@localhost:5432/testdb?sslmode=disable",
			"pg://root:*****@localhost:5432/testdb?sslmode=disable",
		},
		{
			"pg://root@localhost:5432/testdb?sslmode=disable",
			"pg://root@localhost:5432/testdb?sslmode=disable",
		},
		{
			"pg://localhost:5432/testdb?sslmode=disable",
			"pg://localhost:5432/testdb?sslmode=disable",
		},
	}

	for _, tt := range tests {
		config, err := New()
		if err != nil {
			t.Fatal(err)
		}
		config.DSN.URL = tt.url
		got, err := config.MaskedDSN()
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	return dir
}

func Test_mergeDetectedRelations(t *testing.T) {
	var (
		err          error
		table        *schema.Table
		column       *schema.Column
		parentColumn *schema.Column
		relations    []*schema.Relation
	)
	s1 := &schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			{
				Name:    "users",
				Comment: "users comment",
				Columns: []*schema.Column{
					{
						Name: "id",
						Type: "serial",
					},
					{
						Name: "username",
						Type: "text",
					},
				},
			},
			{
				Name:    "posts",
				Comment: "posts comment",
				Columns: []*schema.Column{
					{
						Name: "id",
						Type: "serial",
					},
					{
						Name: "user_id",
						Type: "int",
					},
					{
						Name: "title",
						Type: "text",
					},
				},
			},
		},
	}
	s2 := &schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			{
				Name:    "users",
				Comment: "users comment",
				Columns: []*schema.Column{
					{
						Name: "id",
						Type: "serial",
					},
				},
			},
			{
				Name:    "posts",
				Comment: "posts comment",
				Columns: []*schema.Column{
					{
						Name: "id",
						Type: "serial",
					},
					{
						Name: "uid",
						Type: "int",
					},
					{
						Name: "title",
						Type: "text",
					},
				},
			},
		},
	}
	table, err = s1.FindTableByName("posts")
	if err != nil {
		t.Fatal(err)
	}
	column, err = table.FindColumnByName("user_id")
	if err != nil {
		t.Fatal(err)
	}

	relation := &schema.Relation{
		Virtual: true,
		Def:     "Detected Relation",
		Table:   table,
	}
	if relation.ParentTable, err = s1.FindTableByName(ToParentTableName("user_id")); err != nil {
		t.Fatal(err)
	}
	if parentColumn, err = relation.ParentTable.FindColumnByName(ToParentColumnName("users")); err != nil {
		t.Fatal(err)
	}
	relation.Columns = append(relation.Columns, column)
	relation.ParentColumns = append(relation.ParentColumns, parentColumn)

	column.ParentRelations = append(column.ParentRelations, relation)
	parentColumn.ChildRelations = append(parentColumn.ChildRelations, relation)

	relations = append(relations, relation)

	type args struct {
		s *schema.Schema
	}
	type want struct {
		r []*schema.Relation
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Detect relation succeed",
			args: args{
				s: s1,
			},
			want: want{
				r: relations,
			},
		},
		{
			name: "Detect relation failed",
			args: args{
				s: s2,
			},
			want: want{
				r: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mergeDetectedRelations(tt.args.s)
			if !reflect.DeepEqual(tt.args.s.Relations, tt.want.r) {
				t.Errorf("got: %#v\nwant: %#v", tt.args.s.Relations, tt.want.r)
			}
		})
	}
}
