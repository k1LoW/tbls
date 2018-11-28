package config

import (
	"os"
	"testing"
)

var tests = []struct {
	value    string
	expected string
}{
	{"${TBLS_ONE}/${TBLS_TWO}", "one/two"},
	{"${TBLS_ONE}/${TBLS_TWO}/${TBLS_NONE}", "one/two/"},
	{"${{TBLS_ONE}}", "${{TBLS_ONE}}"},
	{"{{.TBLS_ONE}}/{{.TBLS_TWO}}", "one/two"},
}

func TestParseWithEnvirion(t *testing.T) {
	_ = os.Setenv("TBLS_ONE", "one")
	_ = os.Setenv("TBLS_TWO", "two")
	for _, tt := range tests {
		actual, err := parseWithEnviron(tt.value)
		if err != nil {
			t.Fatal(err)
		}
		if actual != tt.expected {
			t.Errorf("actual %v\nwant %v", actual, tt.expected)
		}
	}
}
