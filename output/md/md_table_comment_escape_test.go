package md

import (
	"strings"
	"testing"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
)

func TestTablesDataEscapesTableComment(t *testing.T) {
	c, err := config.New()
	if err != nil {
		t.Fatal(err)
	}
	m := New(c)
	tables := []*schema.Table{{Name: "users", Comment: `<img src=x> a|b`}}

	data := m.tablesData(tables, false, false, false, false)

	row := strings.Join(data[len(data)-1], "\x00")
	if !strings.Contains(row, `\<img src=x\> a\|b`) {
		t.Errorf("table comment is not escaped: %q", row)
	}
}
