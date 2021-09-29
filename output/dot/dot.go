package dot

import (
	"io"
	"os"
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

// New return Dot
func New(c *config.Config) *Dot {
	return &Dot{
		config: c,
		box:    packr.New("dot", "./templates"),
	}
}

func (d *Dot) schemaTemplate() (string, error) {
	if len(d.config.Templates.Dot.Schema) > 0 {
		tb, err := os.ReadFile(d.config.Templates.Dot.Schema)
		if err != nil {
			return string(tb), errors.WithStack(err)
		}
		return string(tb), nil
	} else {
		ts, err := d.box.FindString("schema.dot.tmpl")
		if err != nil {
			return ts, errors.WithStack(err)
		}
		return ts, nil
	}
}

func (d *Dot) tableTemplate() (string, error) {
	if len(d.config.Templates.Dot.Table) > 0 {
		tb, err := os.ReadFile(d.config.Templates.Dot.Table)
		if err != nil {
			return string(tb), errors.WithStack(err)
		}
		return string(tb), nil
	} else {
		ts, err := d.box.FindString("table.dot.tmpl")
		if err != nil {
			return ts, errors.WithStack(err)
		}
		return ts, nil
	}
}

// OutputSchema output dot format for full relation.
func (d *Dot) OutputSchema(wr io.Writer, s *schema.Schema) error {
	ts, err := d.schemaTemplate()
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

	ts, err := d.tableTemplate()
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
