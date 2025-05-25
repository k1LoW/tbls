package gviz

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/SouhlInc/tbls/config"
	"github.com/SouhlInc/tbls/testutil"
	"github.com/tenntenn/golden"
)

func TestOutputSchema(t *testing.T) {
	tests := []struct {
		wantFile string
	}{
		{"svg_test_schema.svg"},
	}
	for _, tt := range tests {
		t.Run(tt.wantFile, func(t *testing.T) {
			format := "svg"
			s := testutil.NewSchema(t)
			c, err := config.New()
			if err != nil {
				t.Fatal(err)
			}
			option := config.ERFormat(format)
			if err := c.LoadOption(option); err != nil {
				t.Fatal(err)
			}
			if err := c.LoadConfigFile(filepath.Join(testdataDir(), "out_test_tbls.yml")); err != nil {
				t.Fatal(err)
			}
			if err := c.MergeAdditionalData(s); err != nil {
				t.Fatal(err)
			}
			o := New(c)
			got := &bytes.Buffer{}
			if err := o.OutputSchema(got, s); err != nil {
				t.Error(err)
			}
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, testdataDir(), tt.wantFile, got)
				return
			}
			if diff := golden.Diff(t, testdataDir(), tt.wantFile, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestOutputTable(t *testing.T) {
	tests := []struct {
		wantFile string
	}{
		{"svg_test_a.svg"},
	}
	for _, tt := range tests {
		t.Run(tt.wantFile, func(t *testing.T) {
			format := "svg"
			s := testutil.NewSchema(t)
			c, err := config.New()
			if err != nil {
				t.Error(err)
			}
			option := config.ERFormat(format)
			if err := c.LoadOption(option); err != nil {
				t.Fatal(err)
			}
			if err := c.LoadConfigFile(filepath.Join(testdataDir(), "out_test_tbls.yml")); err != nil {
				t.Error(err)
			}
			if err := c.MergeAdditionalData(s); err != nil {
				t.Error(err)
			}
			ta := s.Tables[0]

			o := New(c)
			got := &bytes.Buffer{}
			if err := o.OutputTable(got, ta); err != nil {
				t.Error(err)
			}
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, testdataDir(), tt.wantFile, got)
				return
			}
			if diff := golden.Diff(t, testdataDir(), tt.wantFile, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(filepath.Dir(wd)), "testdata"))
	return dir
}
