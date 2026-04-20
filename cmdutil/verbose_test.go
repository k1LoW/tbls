package cmdutil

import (
	"strings"
	"testing"
)

func TestMaskDSN(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"postgres url with password", "postgres://user:secret@host:5432/db", "postgres://user:****@host:5432/db"},
		{"postgres url with query", "postgres://user:secret@host/db?sslmode=disable", "postgres://user:****@host/db?sslmode=disable"},
		{"url without password", "postgres://user@host/db", "postgres://user@host/db"},
		{"no credentials", "postgres://host/db", "postgres://host/db"},
		{"key-value dsn", "host=h user=u password=secret dbname=d", "host=h user=u password=**** dbname=d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskDSN(tt.in)
			if got != tt.want {
				t.Errorf("MaskDSN(%q) = %q, want %q", tt.in, got, tt.want)
			}
			if strings.Contains(got, "secret") {
				t.Errorf("MaskDSN(%q) leaked password: %q", tt.in, got)
			}
		})
	}
}

func TestSetVerboseAndIsVerbose(t *testing.T) {
	t.Cleanup(func() { SetVerbose(false) })
	SetVerbose(false)
	if IsVerbose() {
		t.Error("IsVerbose() = true after SetVerbose(false)")
	}
	SetVerbose(true)
	if !IsVerbose() {
		t.Error("IsVerbose() = false after SetVerbose(true)")
	}
}
