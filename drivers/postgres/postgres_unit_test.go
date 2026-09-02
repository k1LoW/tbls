package postgres

import (
	"strings"
	"testing"
)

func TestGlobToLike(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		wantLike string
		wantOK   bool
	}{
		{"plain name", "users", "users", true},
		{"trailing star", "events_*", `events\_%`, true},
		{"leading star", "*_archive", `%\_archive`, true},
		{"escapes percent", "10%off", `10\%off`, true},
		{"escapes backslash", `a\b`, `a\\b`, true},
		{"schema qualified", "public.events_*", `public.events\_%`, true},
		{"rejects question mark", "foo?", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := globToLike(tt.in)
			if ok != tt.wantOK {
				t.Fatalf("globToLike(%q) ok = %v, want %v", tt.in, ok, tt.wantOK)
			}
			if got != tt.wantLike {
				t.Errorf("globToLike(%q) = %q, want %q", tt.in, got, tt.wantLike)
			}
		})
	}
}

func TestQueryForTablesNoExcludes(t *testing.T) {
	p := &Postgres{}
	q := p.queryForTables()
	if strings.Contains(q, "NOT LIKE") {
		t.Errorf("query unexpectedly contains NOT LIKE clause when no excludes set:\n%s", q)
	}
	if !strings.Contains(q, "ORDER BY oid") {
		t.Errorf("query missing ORDER BY clause:\n%s", q)
	}
}

func TestQueryForTablesWithExcludes(t *testing.T) {
	p := &Postgres{}
	p.SetTableExcludes([]string{"events_*", "public.logs_*", "ignored?glob"})
	q := p.queryForTables()

	wants := []string{
		`cls.relname NOT LIKE 'events\_%' ESCAPE '\'`,
		`(ns.nspname || '.' || cls.relname) NOT LIKE 'events\_%' ESCAPE '\'`,
		`cls.relname NOT LIKE 'public.logs\_%' ESCAPE '\'`,
	}
	for _, w := range wants {
		if !strings.Contains(q, w) {
			t.Errorf("query missing expected clause %q\nfull query:\n%s", w, q)
		}
	}
	// The ?-glob pattern should be dropped entirely.
	if strings.Contains(q, "ignored") {
		t.Errorf("query unexpectedly contains pushed-down ?-glob pattern:\n%s", q)
	}
}
