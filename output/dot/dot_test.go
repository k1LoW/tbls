package dot

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/SouhlInc/tbls/config"
	"github.com/SouhlInc/tbls/testutil"
	"github.com/tenntenn/golden"
)

func TestOutputSchema(t *testing.T) {
	tests := []struct {
		hideDef         bool
		showColumnTypes *config.ShowColumnTypes
		wantFile        string
	}{
		{false, nil, "dot_test_schema.dot"},
		{true, nil, "dot_test_schema.dot.hidedef"},
		{false, &config.ShowColumnTypes{Related: true}, "dot_test_schema.dot.hide_not_related_column"},
	}
	for _, tt := range tests {
		t.Run(tt.wantFile, func(t *testing.T) {
			s := testutil.NewSchema(t)
			c, err := config.New()
			if err != nil {
				t.Error(err)
			}
			if err := c.LoadConfigFile(filepath.Join(testdataDir(), "out_test_tbls.yml")); err != nil {
				t.Error(err)
			}
			c.ER.HideDef = tt.hideDef
			c.ER.ShowColumnTypes = tt.showColumnTypes
			if err := c.ModifySchema(s); err != nil {
				t.Error(err)
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

func TestOutputSchemaTemplate(t *testing.T) {
	tests := []struct {
		wantFile string
	}{
		{"dot_template_test_schema.dot"},
	}
	for _, tt := range tests {
		t.Run(tt.wantFile, func(t *testing.T) {
			s := testutil.NewSchema(t)
			c, err := config.New()
			if err != nil {
				t.Error(err)
			}
			if err := c.LoadConfigFile(filepath.Join(testdataDir(), "out_templates_test_tbls.yml")); err != nil {
				t.Error(err)
			}

			// use the templates in the testdata directory
			c.Templates.Dot.Schema = filepath.Join(testdataDir(), c.Templates.Dot.Schema)
			if err := c.ModifySchema(s); err != nil {
				t.Error(err)
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
		{"dot_test_a.dot"},
	}
	for _, tt := range tests {
		t.Run(tt.wantFile, func(t *testing.T) {
			s := testutil.NewSchema(t)
			c, err := config.New()
			if err != nil {
				t.Error(err)
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

func TestOutputTableTemplate(t *testing.T) {
	tests := []struct {
		wantFile string
	}{
		{"dot_template_test_a.dot"},
	}
	for _, tt := range tests {
		t.Run(tt.wantFile, func(t *testing.T) {
			s := testutil.NewSchema(t)
			c, err := config.New()
			if err != nil {
				t.Error(err)
			}
			if err := c.LoadConfigFile(filepath.Join(testdataDir(), "out_templates_test_tbls.yml")); err != nil {
				t.Error(err)
			}
			// use the templates in the testdata directory
			c.Templates.Dot.Table = filepath.Join(testdataDir(), c.Templates.Dot.Table)
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

func TestOutputViewpoint(t *testing.T) {
	s := testutil.NewSchema(t)
	for i, v := range s.Viewpoints {
		fn := fmt.Sprintf("dot_test_viewpoint_%d.dot", i)
		t.Run(v.Name, func(t *testing.T) {
			c, err := config.New()
			if err != nil {
				t.Error(err)
			}
			got := &bytes.Buffer{}
			o := New(c)
			if err := o.OutputViewpoint(got, v); err != nil {
				t.Error(err)
			}
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, testdataDir(), fn, got)
				return
			}
			if diff := golden.Diff(t, testdataDir(), fn, got); diff != "" {
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
