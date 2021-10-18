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
