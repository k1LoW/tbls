package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/SouhlInc/tbls/schema"
	"github.com/tenntenn/golden"
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
	t.Setenv("TBLS_TEST_PG_PASS", "pgpass")
	t.Setenv("TBLS_TEST_PG_DOC_PATH", "sample/pg")
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
	want := "duplicate config file [.tbls.yml, tbls.yml, .tbls.yaml, tbls.yaml]"
	if fmt.Sprintf("%v", got) != want {
		t.Errorf("got %v\nwant %v", got, want)
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
	if err := c.LoadConfigFile(filepath.Join(testdataDir(), "config_test_tbls.yml")); err != nil {
		t.Error(err)
	}
	if err := c.MergeAdditionalData(&s); err != nil {
		t.Error(err)
	}
	if want := 1; len(s.Relations) != want {
		t.Errorf("got %v\nwant %v", len(s.Relations), want)
	}
	users, err := s.FindTableByName("users")
	if err != nil {
		t.Fatal(err)
	}
	if want := "users comment by tbls"; users.Comment != want {
		t.Errorf("got %v\nwant %v", users.Comment, want)
	}

	posts, err := s.FindTableByName("posts")
	if err != nil {
		t.Fatal(err)
	}
	title, err := posts.FindColumnByName("title")
	if err != nil {
		t.Fatal(err)
	}
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
	c, err := New()
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		include       []string
		exclude       []string
		labels        []string
		distance      int
		wantTables    int
		wantRelations int
	}{
		{[]string{}, []string{}, []string{}, 0, 5, 3},
		{[]string{}, []string{"schema_migrations"}, []string{}, 0, 4, 3},
		{[]string{}, []string{"users"}, []string{}, 0, 4, 1},
		{[]string{"users"}, []string{}, []string{}, 0, 1, 0},
		{[]string{"user*"}, []string{}, []string{}, 0, 2, 1},
		{[]string{"*options"}, []string{}, []string{}, 0, 1, 0},
		{[]string{"*"}, []string{"user_options"}, []string{}, 0, 4, 2},
		{[]string{"not_exist"}, []string{}, []string{}, 0, 0, 0},
		{[]string{"not_exist", "*"}, []string{}, []string{}, 0, 5, 3},
		{[]string{"users"}, []string{"*"}, []string{}, 0, 1, 0},
		{[]string{"use*"}, []string{"use*"}, []string{}, 0, 2, 1},
		{[]string{"use*"}, []string{"user*"}, []string{}, 0, 0, 0},
		{[]string{"user*"}, []string{"user_*"}, []string{}, 0, 1, 0},
		{[]string{"*", "user*"}, []string{"user_*"}, []string{}, 0, 4, 2},

		{[]string{"users"}, []string{}, []string{}, 1, 3, 2},
		{[]string{"user_options"}, []string{}, []string{}, 1, 2, 1},
		{[]string{"user_options"}, []string{}, []string{}, 2, 3, 2},
		{[]string{"user_options"}, []string{}, []string{}, 3, 4, 3},
		{[]string{}, []string{}, []string{}, 9, 5, 3},
		{[]string{"posts"}, []string{}, []string{}, 9, 4, 3},
		{[]string{""}, []string{"*"}, []string{}, 9, 0, 0},

		{[]string{}, []string{}, []string{"private"}, 0, 2, 1},
		{[]string{}, []string{}, []string{"option"}, 0, 2, 0},
		{[]string{}, []string{}, []string{"public", "private"}, 0, 4, 3},
		{[]string{}, []string{"users"}, []string{"private"}, 0, 1, 0},
		{[]string{}, []string{"user*"}, []string{"option"}, 0, 1, 0},
		{[]string{"users"}, []string{}, []string{"private"}, 0, 2, 1},
		{[]string{}, []string{}, []string{"p*"}, 0, 4, 3},
		{[]string{"users"}, []string{}, []string{"pri*"}, 0, 2, 1},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d.%v%v%v", i, tt.include, tt.exclude, tt.labels), func(t *testing.T) {
			s := newTestSchemaViaJSON(t)
			c.Include = tt.include
			c.Exclude = tt.exclude
			c.includeLabels = tt.labels
			c.Distance = tt.distance
			err = c.FilterTables(s)
			if err != nil {
				t.Error(err)
			}
			if got := len(s.Tables); got != tt.wantTables {
				t.Errorf("got %v\nwant %v", got, tt.wantTables)
			}
			if got := len(s.Relations); got != tt.wantRelations {
				t.Errorf("got %v\nwant %v", got, tt.wantRelations)
			}
		})
	}
}

func TestModifySchema(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name             string
		desc             string
		labels           []string
		comments         []AdditionalComment
		relations        []AdditionalRelation
		viewpointATables []string
		viewpointBTables []string
		wantRel          int
	}{
		{"", "", []string{}, nil, nil, []string{
			"users",
			"posts",
		}, []string{
			"users",
			"user_options",
		}, 3},
		{"mod_name_and_desc", "this is test schema", []string{}, nil, nil, []string{
			"users",
			"posts",
		}, []string{
			"users",
			"user_options",
		}, 3},
		{"relations", "", []string{}, nil, []AdditionalRelation{
			{
				Table:         "users",
				ParentTable:   "categories",
				Columns:       []string{"id"},
				ParentColumns: []string{"id"},
			},
		}, []string{
			"users",
			"posts",
		}, []string{
			"users",
			"user_options",
		}, 4},
		{"not_override", "", []string{}, nil, []AdditionalRelation{
			{
				Table:         "users",
				ParentTable:   "posts",
				Columns:       []string{"id"},
				ParentColumns: []string{"user_id"},
				Def:           "Additional Relation",
				Override:      false,
			},
		}, []string{
			"users",
			"posts",
		}, []string{
			"users",
			"user_options",
		}, 4},
		{"override", "", []string{}, nil, []AdditionalRelation{
			{
				Table:             "posts",
				ParentTable:       "users",
				Columns:           []string{"user_id"},
				ParentColumns:     []string{"id"},
				Cardinality:       "Zero or one",
				ParentCardinality: "1+",
				Def:               "Override Relation",
				Override:          true,
			},
		}, []string{
			"users",
			"posts",
		}, []string{
			"users",
			"user_options",
		}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.Name = tt.name
			c.Desc = tt.desc
			c.Labels = tt.labels
			c.Comments = tt.comments
			c.Relations = tt.relations
			c.Viewpoints = append(c.Viewpoints, Viewpoint{
				Name:   "A",
				Desc:   "Viewpoint A",
				Tables: tt.viewpointATables,
			})
			c.Viewpoints = append(c.Viewpoints, Viewpoint{
				Name:   "B",
				Desc:   "Viewpoint B",
				Tables: tt.viewpointBTables,
			})

			s := newTestSchemaViaJSON(t)
			if err := c.ModifySchema(s); err != nil {
				t.Error(err)
			}
			got, err := json.MarshalIndent(s, "", "  ")
			if err != nil {
				t.Error(err)
			}
			f := fmt.Sprintf("modify_schema_%s", tt.name)
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, testdataDir(), f, got)
				return
			}
			if diff := golden.Diff(t, testdataDir(), f, got); diff != "" {
				t.Error(diff)
			}
			if got := len(s.Relations); got != tt.wantRel {
				t.Errorf("got %v wantRel %v", got, tt.wantRel)
			}
		})
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
		{
			"bq://project-id/dataset-id?creds=/path/to/google_application_credentials.json",
			"bq://project-id/dataset-id?creds=/path/to/google_application_credentials.json",
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
	strategy, err := SelectNamingStrategy("default")
	if err != nil {
		t.Fatal(err)
	}
	if relation.ParentTable, err = s1.FindTableByName(strategy.ParentTableName("user_id")); err != nil {
		t.Fatal(err)
	}
	if parentColumn, err = relation.ParentTable.FindColumnByName(strategy.ParentColumnName("users")); err != nil {
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
			mergeDetectedRelations(tt.args.s, strategy)
			if !reflect.DeepEqual(tt.args.s.Relations, tt.want.r) {
				t.Errorf("got: %#v\nwant: %#v", tt.args.s.Relations, tt.want.r)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		erFormat string
		wantErr  bool
	}{
		{"", true},
		{"png", false},
		{"mermaid", false},
		{"invalid", true},
	}
	for _, tt := range tests {
		t.Run(tt.erFormat, func(t *testing.T) {
			c, err := New()
			if err != nil {
				t.Fatal(err)
			}
			c.ER.Format = tt.erFormat
			if err := c.validate(); err != nil {
				if !tt.wantErr {
					t.Errorf("got error: %s", err)
				}
				return
			}
			if tt.wantErr {
				t.Error("want error")
			}
		})
	}
}

func TestCheckVersion(t *testing.T) {
	tests := []struct {
		v    string
		c    string
		want error
	}{
		{"1.42.3", ">= 1.42", nil},
		{"1.42.3", "", nil},
		{"1.42.3", ">= 1.42, < 2", nil},
		{"1.42.3", "> 1.42", nil},
		{"1.42.3", "1.42.3", nil},
		{"1.42.3", "1.42.4", errors.New("the required tbls version for the configuration is '1.42.4'. however, the running tbls version is '1.42.3'")},
	}
	for _, tt := range tests {
		cfg, err := New()
		if err != nil {
			t.Fatal(err)
		}
		cfg.RequiredVersion = tt.c
		if got := cfg.checkVersion(tt.v); fmt.Sprintf("%s", got) != fmt.Sprintf("%s", tt.want) {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func TestNeedToGenerateERImages(t *testing.T) {
	tests := []struct {
		c    *Config
		want bool
	}{
		{&Config{ER: ER{Skip: true}}, false},
		{&Config{ER: ER{Format: "png"}}, true},
		{&Config{ER: ER{Format: "mermaid"}}, false},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := tt.c.NeedToGenerateERImages()
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}
}

func TestDetectShowColumnsForER(t *testing.T) {
	tests := []struct {
		showColumnTypes   *ShowColumnTypes
		wantColumnCount   int
		wantRelationCount int
	}{
		{nil, 13, 3},
		{&ShowColumnTypes{Related: true, Primary: false}, 5, 3},
		{&ShowColumnTypes{Related: false, Primary: true}, 0, 0},
		{&ShowColumnTypes{Related: true, Primary: true}, 5, 3},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.showColumnTypes), func(t *testing.T) {
			c, err := New()
			if err != nil {
				t.Fatal(err)
			}
			c.ER.ShowColumnTypes = tt.showColumnTypes
			s := newTestSchemaViaJSON(t)
			if err := c.ModifySchema(s); err != nil {
				t.Fatal(err)
			}
			var (
				gotColumnCount   int
				gotRelationCount int
			)
			for _, tt := range s.Tables {
				for _, cc := range tt.Columns {
					if !cc.HideForER {
						gotColumnCount++
					}
				}
			}
			for _, r := range s.Relations {
				if !r.HideForER {
					gotRelationCount++
				}
			}
			if gotColumnCount != tt.wantColumnCount {
				t.Errorf("got %v\nwant %v", gotColumnCount, tt.wantColumnCount)
			}
			if gotRelationCount != tt.wantRelationCount {
				t.Errorf("got %v\nwant %v", gotRelationCount, tt.wantRelationCount)
			}
		})
	}
}

func newTestSchemaViaJSON(t *testing.T) *schema.Schema {
	t.Helper()
	s := &schema.Schema{}
	file, err := os.Open(filepath.Join(testdataDir(), "test_schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(file)
	if err := dec.Decode(s); err != nil {
		t.Fatal(err)
	}
	if err := s.Repair(); err != nil {
		t.Fatal(err)
	}
	return s
}
