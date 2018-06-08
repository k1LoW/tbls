package dot

import (
	"bytes"
	"github.com/k1LoW/tbls/schema"
	"testing"
)

func TestOutputSchema(t *testing.T) {
	ca := &schema.Column{
		Name:    "a",
		Comment: "column a",
	}
	cb := &schema.Column{
		Name:    "b",
		Comment: "column b",
	}

	ta := &schema.Table{
		Name:    "a",
		Comment: "table a",
		Columns: []*schema.Column{
			ca,
			&schema.Column{
				Name:    "a2",
				Comment: "column a2",
			},
		},
	}
	tb := &schema.Table{
		Name:    "b",
		Comment: "table b",
		Columns: []*schema.Column{
			cb,
			&schema.Column{
				Name:    "b2",
				Comment: "column b2",
			},
		},
	}
	r := &schema.Relation{
		Table:         ta,
		Columns:       []*schema.Column{ca},
		ParentTable:   tb,
		ParentColumns: []*schema.Column{cb},
	}
	ca.ParentRelations = []*schema.Relation{r}
	cb.ChildRelations = []*schema.Relation{r}

	s := &schema.Schema{
		Name: "testschema",
		Tables: []*schema.Table{
			ta,
			tb,
		},
		Relations: []*schema.Relation{
			r,
		},
	}
	buf := &bytes.Buffer{}
	_ = OutputSchema(buf, s)
	expected := `digraph "testschema" {
  // Config
  graph [rankdir=TB, layout=dot, fontname="Arial"];
  node [shape=record, fontsize=14, margin=0.6, fontname="Arial"];
  edge [fontsize=10, labelfloat=false, splines=none, fontname="Arial"];

  // Tables
  a [shape=none, label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="6">
                 <tr><td bgcolor="#EFEFEF"><font face="Arial Bold" point-size="18">a</font> <font color="#666666">[]</font></td></tr>
                 <tr><td port="a" align="left">a <font color="#666666">[]</font></td></tr>
                 <tr><td port="a2" align="left">a2 <font color="#666666">[]</font></td></tr>
              </table>>];
  b [shape=none, label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="6">
                 <tr><td bgcolor="#EFEFEF"><font face="Arial Bold" point-size="18">b</font> <font color="#666666">[]</font></td></tr>
                 <tr><td port="b" align="left">b <font color="#666666">[]</font></td></tr>
                 <tr><td port="b2" align="left">b2 <font color="#666666">[]</font></td></tr>
              </table>>];

  // Relations
  a:a -> b:b [dir=back, arrowtail=crow,  taillabel=<<table cellpadding="5" border="0" cellborder="0"><tr><td></td></tr></table>>];
}
`
	actual := buf.String()
	if actual != expected {
		t.Errorf("actual %v\nwant %v", actual, expected)
	}
}
