package cmdutil

import (
	"strings"
	"testing"
)

func TestEnvMap(t *testing.T) {
	t.Setenv("TEST_ENV_EMPTY", "")
	t.Setenv("TEST_ENV_SET", "value")
	result := envMap()
	if value, ok := result["TEST_ENV_EMPTY"]; !ok {
		t.Error("Expected TEST_ENV_EMPTY to be set")
	} else if value != "" {
		t.Errorf("Expected TEST_ENV_EMPTY to be an empty string, got %v", value)
	}
	if value, ok := result["TEST_ENV_SET"]; !ok {
		t.Error("Expected TEST_ENV_SET to be set")
	} else if value != "value" {
		t.Errorf("Expected TEST_ENV_SET to be 'value', got %v", value)
	}
}

func TestIsAllowedToExecute(t *testing.T) {
	tests := []struct {
		name          string
		envset        map[string]string
		when          string
		want          bool
		errorContains any
	}{
		{
			name:          "Empty expression",
			envset:        map[string]string{},
			when:          "",
			want:          true,
			errorContains: nil,
		},
		{
			name: "Equality test, true",
			envset: map[string]string{
				"TEST_ENV1": "a",
			},
			when:          "$TEST_ENV1 == 'a'",
			want:          true,
			errorContains: nil,
		},
		{
			name: "Equality test, false",
			envset: map[string]string{
				"TEST_ENV1": "a",
			},
			when:          "$TEST_ENV1 == 'b'",
			want:          false,
			errorContains: nil,
		},
		{
			name: "Containment in $env",
			envset: map[string]string{
				"env":       "should not replace $env",
				"TEST_ENV1": "a",
			},
			when:          `'TEST_ENV1' not in $env`,
			want:          true,
			errorContains: nil,
		},
		{
			name: "Containment in Env",
			envset: map[string]string{
				"TEST_ENV1": "a",
			},
			when:          "'TEST_ENV1' in Env",
			want:          true,
			errorContains: nil,
		},
		{
			name: "Env var name is used in string literal",
			envset: map[string]string{
				"TEST_ENV1": "foo",
				"TEST_ENV2": "$TEST_ENV1",
			},
			when:          `$TEST_ENV2 == '$TEST_ENV1'`,
			want:          true,
			errorContains: nil,
		},
		{
			name:          "Env var not set",
			envset:        map[string]string{},
			when:          `$TEST_ENV_NONESUCH == ""`,
			want:          true,
			errorContains: nil,
		},
		{
			name:          "Invalid expression",
			envset:        map[string]string{},
			when:          `($TEST_ENV1 == "Missing parentheses"`,
			want:          false,
			errorContains: "unexpected token EOF",
		},
		{
			name:          "Expression produces a non-boolean result",
			envset:        map[string]string{},
			when:          `"String literal expression"`,
			want:          false,
			errorContains: "expected bool, but got string",
		},
		{
			name: "Expression references an unknown variable",
			envset: map[string]string{
				"TEST_ENV1": "a",
			},
			when:          `$TEST_ENV1 == NoneSuchVariable`,
			want:          false,
			errorContains: "unknown name NoneSuchVariable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NewWhenEnv = func() *WhenEnv { return &WhenEnv{Env: tt.envset} }
			got, err := IsAllowedToExecute(tt.when)
			if err != nil {
				if tt.errorContains != nil {
					if !strings.Contains(err.Error(), tt.errorContains.(string)) {
						t.Errorf("Error %v does not contain %s", err, tt.errorContains)
					}
				} else {
					t.Error(err)
				}
			} else if tt.errorContains != nil {
				t.Errorf("Expected an error containing %v", tt.errorContains)
			}
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}
}
