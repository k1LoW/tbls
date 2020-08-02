package cmd

import (
	"os"
	"testing"
)

func TestIsAllowedToExecute(t *testing.T) {
	tests := []struct {
		envset map[string]string
		when   string
		want   bool
	}{
		{
			envset: map[string]string{
				"TEST_ENV1": "a",
			},
			when: "$TEST_ENV1 == 'a'",
			want: true,
		},
		{
			envset: map[string]string{
				"TEST_ENV1": "a",
			},
			when: "$TEST_ENV1 == 'b'",
			want: false,
		},
		{
			envset: map[string]string{
				"TEST_ENV1": "a",
			},
			when: `$TEST_ENV1 == "a"`,
			want: true,
		},
	}
	for _, tt := range tests {
		for k, v := range tt.envset {
			os.Setenv(k, v)
		}
		got, err := isAllowedToExecute(tt.when)
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}
