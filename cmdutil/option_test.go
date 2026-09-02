package cmdutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPickOption(t *testing.T) {
	tests := []struct {
		args        []string
		opts        []string
		want        string
		wantRemains []string
	}{
		{[]string{}, []string{}, "", []string{}},
		{[]string{"-a", "-b", "B", "-c"}, []string{"-b"}, "B", []string{"-a", "-c"}},
		{[]string{"-a", "-b=B", "-c"}, []string{"-b"}, "B", []string{"-a", "-c"}},
		{[]string{"-a", "-b=B", "-c"}, []string{"-b", "--bbb"}, "B", []string{"-a", "-c"}},
		{[]string{"-a", "-b=B", "-c"}, []string{"-d"}, "", []string{"-a", "-b=B", "-c"}},
		{[]string{"-b=B"}, []string{"-b"}, "B", []string{}},
		{[]string{"-b", "B"}, []string{"-b"}, "B", []string{}},
		// Option without a value ( trailing option )
		{[]string{"-b"}, []string{"-b"}, "", []string{"-b"}},
		{[]string{"-a", "-b"}, []string{"-b"}, "", []string{"-a", "-b"}},
		// Value containing "="
		{[]string{"-b=B=BB"}, []string{"-b"}, "B=BB", []string{}},
		{[]string{"-a", "--dsn=postgres://u:p@h/db?a=1&b=2", "-c"}, []string{"--dsn"}, "postgres://u:p@h/db?a=1&b=2", []string{"-a", "-c"}},
	}
	for _, tt := range tests {
		got, gotRemains := PickOption(tt.args, tt.opts)
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
		if diff := cmp.Diff(gotRemains, tt.wantRemains, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}
