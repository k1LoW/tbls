package dot

import (
	"io"
	"strings"
	"text/template"

	"github.com/gobuffalo/packr"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

var templateFuncs = map[string]interface{}{
	"nl2br": func(text string) string {
		return strings.Replace(strings.Replace(strings.Replace(text, "\r\n", "<br />", -1), "\n", "<br />", -1), "\r", "<br />", -1)
	},
	"nl2space": func(text string) string {
		return strings.Replace(strings.Replace(strings.Replace(text, "\r\n", " ", -1), "\n", " ", -1), "\r", " ", -1)
	},
}

// Dot struct
type Dot struct {
	config *config.Config
}

// NewDot return Dot
func NewDot(c *config.Config) *Dot {
	return &Dot{
		config: c,
	}
}

// OutputSchema output dot format for full relation.
func (d *Dot) OutputSchema(wr io.Writer, s *schema.Schema) error {
	box := packr.NewBox("./templates")
	ts, _ := box.FindString("schema.dot.tmpl")
	tmpl := template.Must(template.New(s.Name).Funcs(templateFuncs).Parse(ts))
	err := tmpl.Execute(wr, map[string]interface{}{
		"Schema":      s,
		"showComment": d.config.ER.Comment,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// OutputTable output dot format for table.
func (d *Dot) OutputTable(wr io.Writer, t *schema.Table) error {
	encountered := make(map[string]bool)
	tables := []*schema.Table{}
	relations := []*schema.Relation{}
	for _, c := range t.Columns {
		for _, r := range c.ParentRelations {
			if !encountered[r.ParentTable.Name] {
				encountered[r.ParentTable.Name] = true
				tables = append(tables, r.ParentTable)
			}
			if !contains(relations, r) {
				relations = append(relations, r)
			}
		}
		for _, r := range c.ChildRelations {
			if !encountered[r.Table.Name] {
				encountered[r.Table.Name] = true
				tables = append(tables, r.Table)
			}
			if !contains(relations, r) {
				relations = append(relations, r)
			}
		}
	}

	box := packr.NewBox("./templates")

	ts, _ := box.FindString("table.dot.tmpl")
	tmpl := template.Must(template.New(t.Name).Funcs(templateFuncs).Parse(ts))
	err := tmpl.Execute(wr, map[string]interface{}{
		"Table":       t,
		"Tables":      tables,
		"Relations":   relations,
		"showComment": d.config.ER.Comment,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func contains(rs []*schema.Relation, e *schema.Relation) bool {
	for _, r := range rs {
		if e == r {
			return true
		}
	}
	return false
}
