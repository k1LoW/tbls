package dot

import (
	"io"
	"text/template"

	"github.com/gobuffalo/packr/v2"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/output"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

// Dot struct
type Dot struct {
	config *config.Config
	box    *packr.Box
}

// NewDot return Dot
func NewDot(c *config.Config) *Dot {
	return &Dot{
		config: c,
		box:    packr.New("dot", "./templates"),
	}
}

// OutputSchema output dot format for full relation.
func (d *Dot) OutputSchema(wr io.Writer, s *schema.Schema) error {
	ts, err := d.box.FindString("schema.dot.tmpl")
	if err != nil {
		return errors.WithStack(err)
	}
	tmpl := template.Must(template.New(s.Name).Funcs(output.Funcs(&d.config.MergedDict)).Parse(ts))
	err = tmpl.Execute(wr, map[string]interface{}{
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
	tables, relations, err := t.CollectTablesAndRelations(*d.config.ER.Distance, true)
	if err != nil {
		return errors.WithStack(err)
	}

	ts, err := d.box.FindString("table.dot.tmpl")
	if err != nil {
		return errors.WithStack(err)
	}
	tmpl := template.Must(template.New(t.Name).Funcs(output.Funcs(&d.config.MergedDict)).Parse(ts))
	err = tmpl.Execute(wr, map[string]interface{}{
		"Table":       tables[0],
		"Tables":      tables[1:],
		"Relations":   relations,
		"showComment": d.config.ER.Comment,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
