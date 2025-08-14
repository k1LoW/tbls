package databricks

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
