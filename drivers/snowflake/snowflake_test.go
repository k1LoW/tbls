package snowflake

import "testing"

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"USERS", `"USERS"`},
		{"schema_migrations", `"schema_migrations"`},
		{"MixedCase", `"MixedCase"`},
		{`with"quote`, `"with""quote"`},
	}
	for _, tt := range tests {
		if got := quoteIdentifier(tt.in); got != tt.want {
			t.Errorf("quoteIdentifier(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
