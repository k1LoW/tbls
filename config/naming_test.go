package config

import "testing"

func TestSingularTableParentTableNamer(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"user_id", "user"},
	}

	for _, tt := range tests {
		got := singularTableParentTableNamer(tt.name)

		if got != tt.want {
			t.Errorf("name %v\ngot %v\nwant %v", tt.name, got, tt.want)
		}
	}
}

func TestSingularTableParentColumnNamer(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// normal
		{"user_id", "id"},
	}

	for _, tt := range tests {
		got := singularTableParentColumnNamer(tt.name)

		if got != tt.want {
			t.Errorf("name %v\ngot %v\nwant %v", tt.name, got, tt.want)
		}
	}
}

func TestIdenticalParentColumnNamer(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// normal
		{"user_id", "user_id"},
	}

	for _, tt := range tests {
		got := identicalParentColumnNamer(tt.name)

		if got != tt.want {
			t.Errorf("name %v\ngot %v\nwant %v", tt.name, got, tt.want)
		}
	}
}

func TestInvertedSingularTableParentTableNamer(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"id_user", "user"},
		{"id_category", "category"},
		{"id_users", "user"},
		{"user_id", ""},
		{"id", ""},
		{"userid", ""},
	}

	for _, tt := range tests {
		got := invertedSingularTableParentTableNamer(tt.name)

		if got != tt.want {
			t.Errorf("name %v\ngot %v\nwant %v", tt.name, got, tt.want)
		}
	}
}

func TestSelectNamingStrategy_InvertedSingularTableName(t *testing.T) {
	strategy, err := SelectNamingStrategy("invertedSingularTableName")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		columnName      string
		wantParentTable string
		wantParentCol   string
	}{
		{"id_user", "user", "id"},
		{"id_category", "category", "id"},
	}

	for _, tt := range tests {
		gotTable := strategy.ParentTableName(tt.columnName)
		gotCol := strategy.ParentColumnName(tt.columnName)

		if gotTable != tt.wantParentTable {
			t.Errorf("ParentTableName(%v)\ngot %v\nwant %v", tt.columnName, gotTable, tt.wantParentTable)
		}
		if gotCol != tt.wantParentCol {
			t.Errorf("ParentColumnName(%v)\ngot %v\nwant %v", tt.columnName, gotCol, tt.wantParentCol)
		}
	}
}
