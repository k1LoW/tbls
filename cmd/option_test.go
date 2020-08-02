package cmd

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
	}
	for _, tt := range tests {
		got, gotRemains := pickOption(tt.args, tt.opts)
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
		if diff := cmp.Diff(gotRemains, tt.wantRemains, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}
