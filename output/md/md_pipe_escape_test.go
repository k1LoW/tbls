package md

import (
	"bytes"
	"database/sql"
	"strings"
	"testing"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
)

// countGFMCells returns the number of cells in a GFM table row by counting
// pipe delimiters that are not backslash-escaped. A literal pipe inside a cell
// must be written as "\|"; an unescaped "|" always starts a new cell.
func countGFMCells(row string) int {
	row = strings.TrimSpace(row)
	row = strings.TrimPrefix(row, "|")
	row = strings.TrimSuffix(row, "|")
	n := 1
	for i := 0; i < len(row); i++ {
		if row[i] == '\\' {
			i++
			continue
		}
		if row[i] == '|' {
			n++
		}
	}
	return n
}

// A column's Comment is escaped with mdEscRep, but the Name, Type and Default
// cells on the same row are written raw. A pipe in one of those cells adds an
// extra column and breaks the GFM table.
func TestOutputTablePipeInColumnCells(t *testing.T) {
	col := &schema.Column{
		Name:    "status",
		Type:    "text",
		Default: sql.NullString{String: "'a|b'", Valid: true},
		Comment: "one | two",
	}
	tbl := &schema.Table{
		Name:    "events",
		Columns: []*schema.Column{col},
	}
	s := &schema.Schema{Name: "s", Tables: []*schema.Table{tbl}}
	if err := s.Repair(); err != nil {
		t.Fatal(err)
	}
	c, err := config.New()
	if err != nil {
		t.Fatal(err)
	}
	c.ER.Skip = true

	buf := new(bytes.Buffer)
	if err := New(c).OutputTable(buf, tbl); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	var header, dataRow string
	for _, ln := range strings.Split(out, "\n") {
		l := strings.TrimSpace(ln)
		if strings.HasPrefix(l, "|") && strings.Contains(l, "Nullable") {
			header = l
		}
		if strings.HasPrefix(l, "| status ") {
			dataRow = l
		}
	}
	if header == "" || dataRow == "" {
		t.Fatalf("could not locate header/data row in output:\n%s", out)
	}

	hc := countGFMCells(header)
	dc := countGFMCells(dataRow)
	if hc != dc {
		t.Errorf("GFM cell-count mismatch: header=%d data=%d (a raw '|' splits the cell)\nheader:  %s\ndataRow: %s", hc, dc, header, dataRow)
	}
}
