package json

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/SouhlInc/tbls/dict"
	"github.com/SouhlInc/tbls/schema"
	"github.com/SouhlInc/tbls/testutil"
	"github.com/tenntenn/golden"
)

func TestOutputSchema(t *testing.T) {
	s := testutil.NewSchema(t)
	o := new(JSON)
	buf := &bytes.Buffer{}
	if err := o.OutputSchema(buf, s); err != nil {
		t.Error(err)
	}
	got := buf.String()
	f := "json_output_schema"
	if os.Getenv("UPDATE_GOLDEN") != "" {
		golden.Update(t, testdataDir(), f, got)
		return
	}
	if diff := golden.Diff(t, testdataDir(), f, got); diff != "" {
		t.Error(diff)
	}
}

func TestEncodeAndDecode(t *testing.T) {
	s1 := testutil.NewSchema(t)
	o := new(JSON)
	buf := &bytes.Buffer{}
	if err := o.OutputSchema(buf, s1); err != nil {
		t.Error(err)
	}
	s2 := &schema.Schema{}
	dec := json.NewDecoder(buf)
	if err := dec.Decode(s2); err != nil {
		t.Error(err)
	}
	if err := s2.Repair(); err != nil {
		t.Error(err)
	}

	_ = removeColumnRelations(s1)
	_ = removeColumnRelations(s2)

	opt := cmpopts.IgnoreUnexported(dict.New())

	if diff := cmp.Diff(s1, s2, opt); diff != "" {
		t.Errorf("schemas not equal\n%v", diff)
	}
}

func removeColumnRelations(s *schema.Schema) error {
	for _, t := range s.Tables {
		for _, c := range t.Columns {
			c.ParentRelations = nil
			c.ChildRelations = nil
		}
	}
	return nil
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(filepath.Dir(wd)), "testdata"))
	return dir
}
