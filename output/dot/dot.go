package dot

import (
	"embed"
	"io"
	"os"
	"text/template"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/output"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

//go:embed templates/*
var tmpl embed.FS

// Dot struct
type Dot struct {
	config *config.Config
	tmpl   embed.FS
}

// New return Dot
func New(c *config.Config) *Dot {
	return &Dot{
		config: c,
		tmpl:   tmpl,
	}
}

func (d *Dot) schemaTemplate() (string, error) {
	if len(d.config.Templates.Dot.Schema) > 0 {
		tb, err := os.ReadFile(d.config.Templates.Dot.Schema)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	} else {
		tb, err := d.tmpl.ReadFile("templates/schema.dot.tmpl")
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	}
}

func (d *Dot) tableTemplate() (string, error) {
	if len(d.config.Templates.Dot.Table) > 0 {
		tb, err := os.ReadFile(d.config.Templates.Dot.Table)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	} else {
		tb, err := d.tmpl.ReadFile("templates/table.dot.tmpl")
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
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
		"showDef":     !d.config.ER.HideDef,
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
		"showDef":     !d.config.ER.HideDef,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
