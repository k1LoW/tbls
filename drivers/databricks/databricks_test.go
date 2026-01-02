package databricks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/tbls/schema"
)

func TestParseArrayString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty string",
			input: "",
			want:  []string{},
		},
		{
			name:  "empty array",
			input: "[]",
			want:  []string{},
		},
		{
			name:  "single element",
			input: "[col1]",
			want:  []string{"col1"},
		},
		{
			name:  "multiple elements",
			input: "[col1, col2]",
			want:  []string{"col1", "col2"},
		},
		{
			name:  "quoted elements",
			input: `["col1", "col2"]`,
			want:  []string{"col1", "col2"},
		},
		{
			name:  "mixed spacing",
			input: "[ col1 ,  col2  ]",
			want:  []string{"col1", "col2"},
		},
		{
			name:  "single quoted element",
			input: `["single_column"]`,
			want:  []string{"single_column"},
		},
		{
			name:  "three elements with quotes",
			input: `["first_col", "second_col", "third_col"]`,
			want:  []string{"first_col", "second_col", "third_col"},
		},
		{
			name:  "elements with underscores",
			input: "[customer_id, order_date]",
			want:  []string{"customer_id", "order_date"},
		},
		{
			name:  "whitespace around brackets",
			input: " [ col1 , col2 ] ",
			want:  []string{"col1", "col2"},
		},
	}

	dbx := &Databricks{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dbx.parseArrayString(tt.input)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("parseArrayString() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBuildConstraintDefinition(t *testing.T) {
	tests := []struct {
		name              string
		constraintType    string
		columns           []string
		referencedTable   string
		referencedColumns []string
		want              string
	}{
		{
			name:           "empty columns",
			constraintType: "PRIMARY KEY",
			columns:        []string{},
			want:           "",
		},
		{
			name:           "primary key single column",
			constraintType: "PRIMARY KEY",
			columns:        []string{"id"},
			want:           "PRIMARY KEY (id)",
		},
		{
			name:           "primary key multiple columns",
			constraintType: "PRIMARY KEY",
			columns:        []string{"customer_id", "order_id"},
			want:           "PRIMARY KEY (customer_id, order_id)",
		},
		{
			name:           "primary key case insensitive",
			constraintType: "primary key",
			columns:        []string{"id"},
			want:           "PRIMARY KEY (id)",
		},
		{
			name:           "unique constraint single column",
			constraintType: "UNIQUE",
			columns:        []string{"email"},
			want:           "UNIQUE (email)",
		},
		{
			name:           "unique constraint multiple columns",
			constraintType: "UNIQUE",
			columns:        []string{"first_name", "last_name", "birth_date"},
			want:           "UNIQUE (first_name, last_name, birth_date)",
		},
		{
			name:           "unique constraint case insensitive",
			constraintType: "unique",
			columns:        []string{"username"},
			want:           "UNIQUE (username)",
		},
		{
			name:              "foreign key with reference single column",
			constraintType:    "FOREIGN KEY",
			columns:           []string{"customer_id"},
			referencedTable:   "customers",
			referencedColumns: []string{"id"},
			want:              "FOREIGN KEY (customer_id) REFERENCES customers(id)",
		},
		{
			name:              "foreign key with reference multiple columns",
			constraintType:    "FOREIGN KEY",
			columns:           []string{"customer_id", "order_id"},
			referencedTable:   "orders",
			referencedColumns: []string{"cust_id", "id"},
			want:              "FOREIGN KEY (customer_id, order_id) REFERENCES orders(cust_id, id)",
		},
		{
			name:           "foreign key without reference",
			constraintType: "FOREIGN KEY",
			columns:        []string{"customer_id"},
			want:           "FOREIGN KEY (customer_id)",
		},
		{
			name:              "foreign key with empty referenced table",
			constraintType:    "FOREIGN KEY",
			columns:           []string{"customer_id"},
			referencedTable:   "",
			referencedColumns: []string{"id"},
			want:              "FOREIGN KEY (customer_id)",
		},
		{
			name:              "foreign key with empty referenced columns",
			constraintType:    "FOREIGN KEY",
			columns:           []string{"customer_id"},
			referencedTable:   "customers",
			referencedColumns: []string{},
			want:              "FOREIGN KEY (customer_id)",
		},
		{
			name:              "foreign key case insensitive",
			constraintType:    "foreign key",
			columns:           []string{"ref_id"},
			referencedTable:   "ref_table",
			referencedColumns: []string{"id"},
			want:              "FOREIGN KEY (ref_id) REFERENCES ref_table(id)",
		},
		{
			name:           "check constraint",
			constraintType: "CHECK",
			columns:        []string{"age"},
			want:           "CHECK (age)",
		},
		{
			name:           "check constraint multiple columns",
			constraintType: "CHECK",
			columns:        []string{"start_date", "end_date"},
			want:           "CHECK (start_date, end_date)",
		},
		{
			name:           "check constraint case insensitive",
			constraintType: "check",
			columns:        []string{"status"},
			want:           "CHECK (status)",
		},
		{
			name:           "custom constraint type",
			constraintType: "CUSTOM_TYPE",
			columns:        []string{"col1", "col2"},
			want:           "CUSTOM_TYPE (col1, col2)",
		},
		{
			name:           "constraint with special characters in column names",
			constraintType: "PRIMARY KEY",
			columns:        []string{"user_id", "created_at", "updated_at"},
			want:           "PRIMARY KEY (user_id, created_at, updated_at)",
		},
		{
			name:              "foreign key with table and column names containing underscores",
			constraintType:    "FOREIGN KEY",
			columns:           []string{"user_profile_id"},
			referencedTable:   "user_profiles",
			referencedColumns: []string{"profile_id"},
			want:              "FOREIGN KEY (user_profile_id) REFERENCES user_profiles(profile_id)",
		},
		{
			name:           "mixed case constraint type normalization",
			constraintType: "Primary Key",
			columns:        []string{"id"},
			want:           "PRIMARY KEY (id)",
		},
		{
			name:           "unknown constraint type",
			constraintType: "SOME_UNKNOWN_TYPE",
			columns:        []string{"test_col"},
			want:           "SOME_UNKNOWN_TYPE (test_col)",
		},
	}

	dbx := &Databricks{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dbx.buildConstraintDefinition(tt.constraintType, tt.columns, tt.referencedTable, tt.referencedColumns)
			if got != tt.want {
				t.Errorf("buildConstraintDefinition() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHasStructColumns(t *testing.T) {
	tests := []struct {
		name    string
		columns []*schema.Column
		want    bool
	}{
		{
			name:    "empty columns",
			columns: []*schema.Column{},
			want:    false,
		},
		{
			name: "no struct columns",
			columns: []*schema.Column{
				{Name: "id", Type: "INT"},
				{Name: "name", Type: "STRING"},
				{Name: "created_at", Type: "TIMESTAMP"},
			},
			want: false,
		},
		{
			name: "single struct column",
			columns: []*schema.Column{
				{Name: "id", Type: "INT"},
				{Name: "metadata", Type: "STRUCT<name:STRING,value:INT>"},
			},
			want: true,
		},
		{
			name: "struct column uppercase",
			columns: []*schema.Column{
				{Name: "data", Type: "STRUCT<field1:STRING>"},
			},
			want: true,
		},
		{
			name: "struct column lowercase",
			columns: []*schema.Column{
				{Name: "data", Type: "struct<field1:string>"},
			},
			want: true,
		},
		{
			name: "array of struct",
			columns: []*schema.Column{
				{Name: "id", Type: "INT"},
				{Name: "items", Type: "ARRAY(STRUCT<id:INT,name:STRING>)"},
			},
			want: true,
		},
		{
			name: "array of struct lowercase",
			columns: []*schema.Column{
				{Name: "items", Type: "array(struct<id:int,name:string>)"},
			},
			want: true,
		},
		{
			name: "multiple struct columns",
			columns: []*schema.Column{
				{Name: "id", Type: "INT"},
				{Name: "address", Type: "STRUCT<street:STRING,city:STRING>"},
				{Name: "contacts", Type: "ARRAY(STRUCT<type:STRING,value:STRING>)"},
			},
			want: true,
		},
		{
			name: "array of primitive type",
			columns: []*schema.Column{
				{Name: "tags", Type: "ARRAY(STRING)"},
			},
			want: false,
		},
		{
			name: "map type",
			columns: []*schema.Column{
				{Name: "properties", Type: "MAP<STRING,STRING>"},
			},
			want: false,
		},
		{
			name: "mixed types without struct",
			columns: []*schema.Column{
				{Name: "id", Type: "BIGINT"},
				{Name: "data", Type: "JSON"},
				{Name: "tags", Type: "ARRAY(STRING)"},
				{Name: "props", Type: "MAP<STRING,INT>"},
			},
			want: false,
		},
	}

	dbx := &Databricks{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dbx.hasStructColumns(tt.columns)
			if got != tt.want {
				t.Errorf("hasStructColumns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnrichStructColumns(t *testing.T) {
	tests := []struct {
		name         string
		tableName    string
		catalog      string
		schema       string
		inputColumns []*schema.Column
		apiResponse  TableInfo
		wantColumns  []*schema.Column
	}{
		{
			name:      "simple struct",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "id", Type: "INT"},
				{Name: "parent", Type: "STRUCT"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "parent",
						TypeName: "STRUCT",
						TypeJSON: `{"type":"struct","fields":[{"name":"field1","type":"string","nullable":true,"metadata":{"comment":"first field"}},{"name":"field2","type":"integer","nullable":false}]}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "id", Type: "INT"},
				{Name: "parent", Type: "STRUCT"},
				{Name: "parent.field1", Type: "STRING", Nullable: true, Comment: "first field"},
				{Name: "parent.field2", Type: "INTEGER", Nullable: false, Comment: ""},
			},
		},
		{
			name:      "nested struct",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "data", Type: "STRUCT"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "data",
						TypeName: "STRUCT",
						TypeJSON: `{"type":"struct","fields":[{"name":"address","type":{"type":"struct","fields":[{"name":"street","type":"string","nullable":true},{"name":"city","type":"string","nullable":true}]},"nullable":true}]}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "data", Type: "STRUCT"},
				{Name: "data.address", Type: "STRUCT", Nullable: true, Comment: ""},
				{Name: "data.address.street", Type: "STRING", Nullable: true, Comment: ""},
				{Name: "data.address.city", Type: "STRING", Nullable: true, Comment: ""},
			},
		},
		{
			name:      "array of struct",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "items", Type: "ARRAY(STRUCT)"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "items",
						TypeName: "ARRAY",
						TypeJSON: `{"type":{"type":"array","elementType":{"type":"struct","fields":[{"name":"id","type":"integer","nullable":false},{"name":"name","type":"string","nullable":true}]}}}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "items", Type: "ARRAY(STRUCT)"},
				{Name: "items.id", Type: "INTEGER", Nullable: false, Comment: ""},
				{Name: "items.name", Type: "STRING", Nullable: true, Comment: ""},
			},
		},
		{
			name:      "deeply nested struct",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "root", Type: "STRUCT"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "root",
						TypeName: "STRUCT",
						TypeJSON: `{"type":"struct","fields":[{"name":"level1","type":{"type":"struct","fields":[{"name":"level2","type":{"type":"struct","fields":[{"name":"value","type":"string","nullable":true}]},"nullable":true}]},"nullable":true}]}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "root", Type: "STRUCT"},
				{Name: "root.level1", Type: "STRUCT", Nullable: true, Comment: ""},
				{Name: "root.level1.level2", Type: "STRUCT", Nullable: true, Comment: ""},
				{Name: "root.level1.level2.value", Type: "STRING", Nullable: true, Comment: ""},
			},
		},
		{
			name:      "struct with metadata comments",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "user", Type: "STRUCT"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "user",
						TypeName: "STRUCT",
						TypeJSON: `{"type":"struct","fields":[{"name":"user_id","type":"bigint","nullable":false,"metadata":{"comment":"Unique user identifier"}},{"name":"email","type":"string","nullable":true,"metadata":{"comment":"User email address"}}]}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "user", Type: "STRUCT"},
				{Name: "user.user_id", Type: "BIGINT", Nullable: false, Comment: "Unique user identifier"},
				{Name: "user.email", Type: "STRING", Nullable: true, Comment: "User email address"},
			},
		},
		{
			name:      "non-struct column ignored",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "id", Type: "INT"},
				{Name: "name", Type: "STRING"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{Name: "id", TypeName: "INT"},
					{Name: "name", TypeName: "STRING"},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "id", Type: "INT"},
				{Name: "name", Type: "STRING"},
			},
		},
		{
			name:      "table name with schema prefix",
			tableName: "my_schema.test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "data", Type: "STRUCT"},
			},
			apiResponse: TableInfo{
				FullName: "main.my_schema.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "data",
						TypeName: "STRUCT",
						TypeJSON: `{"type":"struct","fields":[{"name":"field1","type":"string","nullable":true}]}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "data", Type: "STRUCT"},
				{Name: "data.field1", Type: "STRING", Nullable: true, Comment: ""},
			},
		},
		{
			name:      "empty struct",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "empty", Type: "STRUCT"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "empty",
						TypeName: "STRUCT",
						TypeJSON: `{"type":"struct","fields":[]}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "empty", Type: "STRUCT"},
			},
		},
		{
			name:      "struct with missing field name",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "data", Type: "STRUCT"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "data",
						TypeName: "STRUCT",
						TypeJSON: `{"type":"struct","fields":[{"type":"string","nullable":true},{"name":"valid","type":"integer","nullable":false}]}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "data", Type: "STRUCT"},
				{Name: "data.valid", Type: "INTEGER", Nullable: false, Comment: ""},
			},
		},
		{
			name:      "array of primitive type not expanded",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "tags", Type: "ARRAY(STRING)"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "tags",
						TypeName: "ARRAY",
						TypeJSON: `{"type":"array","elementType":"string"}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "tags", Type: "ARRAY(STRING)"},
			},
		},
		{
			name:      "flat array of struct with nested expansion",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "items", Type: "ARRAY(STRUCT)"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "items",
						TypeName: "ARRAY",
						TypeJSON: `{"type":"array","elementType":{"type":"struct","fields":[{"name":"id","type":"integer","nullable":false},{"name":"name","type":"string","nullable":true}]}}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "items", Type: "ARRAY(STRUCT)"},
				{Name: "items.id", Type: "INTEGER", Nullable: false, Comment: ""},
				{Name: "items.name", Type: "STRING", Nullable: true, Comment: ""},
			},
		},
		{
			name:      "malformed TypeJSON skipped",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "bad_json", Type: "STRUCT"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "bad_json",
						TypeName: "STRUCT",
						TypeJSON: `{invalid json`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "bad_json", Type: "STRUCT"},
			},
		},
		{
			name:      "TypeJSON with unknown type",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "unknown", Type: "STRUCT"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "unknown",
						TypeName: "STRUCT",
						TypeJSON: `{"unknown":"value"}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "unknown", Type: "UNKNOWN"},
			},
		},
		{
			name:      "array with nested non-struct type",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "items", Type: "ARRAY"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "items",
						TypeName: "ARRAY",
						TypeJSON: `{"type":{"type":"array","elementType":{"type":"integer"}}}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "items", Type: "ARRAY(INTEGER)"},
			},
		},
		{
			name:      "struct with non-map field",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "data", Type: "STRUCT"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "data",
						TypeName: "STRUCT",
						TypeJSON: `{"type":"struct","fields":["invalid_field",{"name":"valid","type":"string","nullable":true}]}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "data", Type: "STRUCT"},
				{Name: "data.valid", Type: "STRING", Nullable: true, Comment: ""},
			},
		},
		{
			name:      "nested type with unknown structType",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "custom", Type: "STRUCT"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "custom",
						TypeName: "STRUCT",
						TypeJSON: `{"type":{"type":"custom_type"}}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "custom", Type: "CUSTOM_TYPE"},
			},
		},
		{
			name:      "array with elementType missing type field",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "items", Type: "ARRAY"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "items",
						TypeName: "ARRAY",
						TypeJSON: `{"type":{"type":"array","elementType":{"name":"something"}}}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "items", Type: "ARRAY"},
			},
		},
		{
			name:      "type with numeric value",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "weird", Type: "STRUCT"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "weird",
						TypeName: "STRUCT",
						TypeJSON: `{"type":123}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "weird", Type: "UNKNOWN"},
			},
		},
		{
			name:      "nested map type without valid type field",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "data", Type: "STRUCT"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "data",
						TypeName: "STRUCT",
						TypeJSON: `{"type":{"name":"something"}}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "data", Type: "UNKNOWN"},
			},
		},
		{
			name:      "array with nested map type non-struct",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "numbers", Type: "ARRAY"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "numbers",
						TypeName: "ARRAY",
						TypeJSON: `{"type":{"type":"array","elementType":{"type":"decimal"}}}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "numbers", Type: "ARRAY(DECIMAL)"},
			},
		},
		{
			name:      "map with string key and string value",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "metadata", Type: "MAP"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "metadata",
						TypeName: "MAP",
						TypeJSON: `{"type":{"type":"map","keyType":"string","valueType":"string"}}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "metadata", Type: "MAP(STRING, STRING)"},
			},
		},
		{
			name:      "map with string key and int value",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "counters", Type: "MAP"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "counters",
						TypeName: "MAP",
						TypeJSON: `{"type":{"type":"map","keyType":"string","valueType":"integer"}}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "counters", Type: "MAP(STRING, INTEGER)"},
			},
		},
		{
			name:      "map with complex value type",
			tableName: "test_table",
			catalog:   "main",
			schema:    "default",
			inputColumns: []*schema.Column{
				{Name: "settings", Type: "MAP"},
			},
			apiResponse: TableInfo{
				FullName: "main.default.test_table",
				Columns: []ColumnInfo{
					{
						Name:     "settings",
						TypeName: "MAP",
						TypeJSON: `{"type":{"type":"map","keyType":"string","valueType":{"type":"struct"}}}`,
					},
				},
			},
			wantColumns: []*schema.Column{
				{Name: "settings", Type: "MAP(STRING, STRUCT)"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(tt.apiResponse)
			}))
			defer server.Close()

			mockClient := &mockTablesAPIClient{
				server:   server,
				response: tt.apiResponse,
			}

			dbx := &Databricks{
				tablesAPIClient: mockClient,
				explicitSchema:  false,
			}

			table := &schema.Table{
				Name:    tt.tableName,
				Columns: tt.inputColumns,
			}

			err := dbx.enrichStructColumns(context.Background(), tt.catalog, tt.schema, table)
			if err != nil {
				t.Fatalf("enrichStructColumns() error = %v", err)
			}

			if diff := cmp.Diff(tt.wantColumns, table.Columns); diff != "" {
				t.Errorf("enrichStructColumns() columns mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

type mockTablesAPIClient struct {
	server   *httptest.Server
	response TableInfo
}

func (m *mockTablesAPIClient) GetTable(_ context.Context, _, _, _ string) (*TableInfo, error) {
	return &m.response, nil
}

func TestFormatType(t *testing.T) {
	tests := []struct {
		name     string
		typeData map[string]any
		want     string
	}{
		{
			name: "simple string type",
			typeData: map[string]any{
				"type": "string",
			},
			want: "STRING",
		},
		{
			name: "simple integer type",
			typeData: map[string]any{
				"type": "integer",
			},
			want: "INTEGER",
		},
		{
			name: "struct type",
			typeData: map[string]any{
				"type": map[string]any{
					"type": "struct",
				},
			},
			want: "STRUCT",
		},
		{
			name: "array of string",
			typeData: map[string]any{
				"type": map[string]any{
					"type":        "array",
					"elementType": "string",
				},
			},
			want: "ARRAY(STRING)",
		},
		{
			name: "array of struct",
			typeData: map[string]any{
				"type": map[string]any{
					"type": "array",
					"elementType": map[string]any{
						"type": "struct",
					},
				},
			},
			want: "ARRAY(STRUCT)",
		},
		{
			name: "map with string key and string value",
			typeData: map[string]any{
				"type": map[string]any{
					"type":      "map",
					"keyType":   "string",
					"valueType": "string",
				},
			},
			want: "MAP(STRING, STRING)",
		},
		{
			name: "map with string key and integer value",
			typeData: map[string]any{
				"type": map[string]any{
					"type":      "map",
					"keyType":   "string",
					"valueType": "integer",
				},
			},
			want: "MAP(STRING, INTEGER)",
		},
		{
			name: "map with integer key and double value",
			typeData: map[string]any{
				"type": map[string]any{
					"type":      "map",
					"keyType":   "integer",
					"valueType": "double",
				},
			},
			want: "MAP(INTEGER, DOUBLE)",
		},
		{
			name: "map with struct value",
			typeData: map[string]any{
				"type": map[string]any{
					"type":    "map",
					"keyType": "string",
					"valueType": map[string]any{
						"type": "struct",
					},
				},
			},
			want: "MAP(STRING, STRUCT)",
		},
		{
			name: "map with array value",
			typeData: map[string]any{
				"type": map[string]any{
					"type":    "map",
					"keyType": "string",
					"valueType": map[string]any{
						"type": "array",
					},
				},
			},
			want: "MAP(STRING, ARRAY)",
		},
		{
			name: "unknown type",
			typeData: map[string]any{
				"type": map[string]any{
					"type": "custom_unknown",
				},
			},
			want: "CUSTOM_UNKNOWN",
		},
		{
			name:     "missing type field",
			typeData: map[string]any{},
			want:     "UNKNOWN",
		},
		{
			name: "empty string type flat",
			typeData: map[string]any{
				"type": "",
			},
			want: "UNKNOWN",
		},
		{
			name: "empty string type nested",
			typeData: map[string]any{
				"type": map[string]any{
					"type": "",
				},
			},
			want: "UNKNOWN",
		},
		{
			name: "flat array of string",
			typeData: map[string]any{
				"type":        "array",
				"elementType": "string",
			},
			want: "ARRAY(STRING)",
		},
		{
			name: "flat array of integer",
			typeData: map[string]any{
				"type":        "array",
				"elementType": "integer",
			},
			want: "ARRAY(INTEGER)",
		},
		{
			name: "flat array of struct",
			typeData: map[string]any{
				"type": "array",
				"elementType": map[string]any{
					"type": "struct",
				},
			},
			want: "ARRAY(STRUCT)",
		},
		{
			name: "flat array without elementType",
			typeData: map[string]any{
				"type": "array",
			},
			want: "ARRAY",
		},
		{
			name: "flat map with string key and value",
			typeData: map[string]any{
				"type":      "map",
				"keyType":   "string",
				"valueType": "string",
			},
			want: "MAP(STRING, STRING)",
		},
		{
			name: "flat map with integer key and double value",
			typeData: map[string]any{
				"type":      "map",
				"keyType":   "integer",
				"valueType": "double",
			},
			want: "MAP(INTEGER, DOUBLE)",
		},
		{
			name: "flat map with complex value type",
			typeData: map[string]any{
				"type":    "map",
				"keyType": "string",
				"valueType": map[string]any{
					"type": "struct",
				},
			},
			want: "MAP(STRING, STRUCT)",
		},
		{
			name: "flat map without keyType or valueType",
			typeData: map[string]any{
				"type": "map",
			},
			want: "MAP",
		},
	}

	dbx := &Databricks{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dbx.formatType(tt.typeData)
			if got != tt.want {
				t.Errorf("formatType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		explicitSchema bool
	}{
		{
			name:           "with explicit schema",
			explicitSchema: true,
		},
		{
			name:           "without explicit schema",
			explicitSchema: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockTablesAPIClient{}

			dbx := New(nil, mockClient, tt.explicitSchema)

			if dbx == nil {
				t.Fatal("New() returned nil")
			}

			if dbx.tablesAPIClient != mockClient {
				t.Error("tablesAPIClient not set correctly")
			}

			if dbx.explicitSchema != tt.explicitSchema {
				t.Errorf("explicitSchema = %v, want %v", dbx.explicitSchema, tt.explicitSchema)
			}
		})
	}
}
