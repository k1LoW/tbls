package mssql

import "testing"

func TestNormalizeViewDefinition(t *testing.T) {
	tests := []struct {
		name string
		def  string
		want string
	}{
		{
			name: "definition without a trailing semicolon gains one",
			def:  "CREATE VIEW v AS (\n  SELECT 1 AS a\n)",
			want: "CREATE VIEW v AS (\n  SELECT 1 AS a\n);",
		},
		{
			name: "definition already ending in a semicolon is unchanged",
			def:  "CREATE VIEW v AS (\n  SELECT 1 AS a\n);",
			want: "CREATE VIEW v AS (\n  SELECT 1 AS a\n);",
		},
		{
			name: "trailing whitespace is trimmed before the semicolon is added",
			def:  "CREATE VIEW v AS (SELECT 1 AS a)  \n\t\r\n",
			want: "CREATE VIEW v AS (SELECT 1 AS a);",
		},
		{
			name: "trailing whitespace after an existing semicolon is trimmed",
			def:  "CREATE VIEW v AS (SELECT 1 AS a); \n",
			want: "CREATE VIEW v AS (SELECT 1 AS a);",
		},
		{
			name: "empty definition stays empty",
			def:  "",
			want: "",
		},
		{
			name: "whitespace-only definition collapses to empty",
			def:  " \n\t",
			want: "",
		},
		{
			name: "inner semicolons do not suppress the trailing one",
			def:  "CREATE VIEW v AS (SELECT 1 AS a); -- note\nSELECT 2",
			want: "CREATE VIEW v AS (SELECT 1 AS a); -- note\nSELECT 2;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeViewDefinition(tt.def); got != tt.want {
				t.Errorf("normalizeViewDefinition(%q) = %q, want %q", tt.def, got, tt.want)
			}
		})
	}
}
