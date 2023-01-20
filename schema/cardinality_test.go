package schema

import (
	"testing"
)

func TestToCardinality(t *testing.T) {
	tests := []struct {
		in      string
		want    Cardinality
		wantErr bool
	}{
		{"zero or one", ZeroOrOne, false},
		{"Zero or One", ZeroOrOne, false},
		{"0", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ToCardinality(tt.in)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("got error: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Error("want error")
			}
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}
}
