package xlsx

import (
	"archive/zip"
	"bytes"
	"html"
	"io"
	"strings"
	"testing"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
)

func TestOutputSchemaLinksTableNamesToSheets(t *testing.T) {
	longName := strings.Repeat("x", 40)
	s := &schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			{Name: "users", Columns: []*schema.Column{{Name: "id"}}},
			{Name: "user profiles", Columns: []*schema.Column{{Name: "id"}}},
			{Name: longName, Columns: []*schema.Column{{Name: "id"}}},
		},
	}

	c, err := config.New()
	if err != nil {
		t.Fatal(err)
	}
	buf := &bytes.Buffer{}
	if err := New(c).OutputSchema(buf, s); err != nil {
		t.Fatal(err)
	}

	got := html.UnescapeString(worksheetsXML(t, buf.Bytes()))

	// The index sheet should link each table name to its own sheet. Sheet names
	// are single quoted so names with spaces resolve, and names longer than 31
	// characters are truncated to match the table sheet name.
	want := []string{
		`HYPERLINK("#'users'!A1","users")`,
		`HYPERLINK("#'user profiles'!A1","user profiles")`,
		`HYPERLINK("#'` + longName[:31] + `'!A1","` + longName + `")`,
	}
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("index sheet is missing hyperlink formula %q", w)
		}
	}
}

func worksheetsXML(t *testing.T, b []byte) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	for _, f := range zr.File {
		if !strings.HasPrefix(f.Name, "xl/worksheets/") || !strings.HasSuffix(f.Name, ".xml") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		out.Write(data)
	}
	return out.String()
}
