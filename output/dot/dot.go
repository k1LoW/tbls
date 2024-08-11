package dot

import (
	"embed"
	"io"
	"os"
	"text/template"

	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/output"
	"github.com/k1LoW/tbls/schema"
	"github.com/samber/lo"
)

//go:embed templates/*
var tmpl embed.FS

var defaultColors = []string{
	"#1F91BE",
	"#B2CF3E",
	"#F0BA32",
	"#8858AA",
}

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

// OutputSchema output dot format for full relation.
func (d *Dot) OutputSchema(wr io.Writer, s *schema.Schema) error {
	ts, err := d.schemaTemplate()
	if err != nil {
		return errors.WithStack(err)
	}
	tmpl := template.Must(template.New(s.Name).Funcs(output.Funcs(&d.config.MergedDict)).Parse(ts))
	if err := tmpl.Execute(wr, map[string]interface{}{
		"Name":        s.Name,
		"Tables":      s.Tables,
		"Relations":   s.Relations,
		"showComment": d.config.ER.Comment,
		"showDef":     !d.config.ER.HideDef,
	}); err != nil {
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
	if err := tmpl.Execute(wr, map[string]interface{}{
		"Table":       tables[0],
		"Tables":      tables[1:],
		"Relations":   relations,
		"showComment": d.config.ER.Comment,
		"showDef":     !d.config.ER.HideDef,
	}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// OutputViewpoint output dot format for viewpoint.
func (d *Dot) OutputViewpoint(wr io.Writer, v *schema.Viewpoint) error {
	ts, err := d.schemaTemplate()
	if err != nil {
		return errors.WithStack(err)
	}

	tables := v.Schema.Tables
	groups := []map[string]interface{}{}
	nogroup := v.Schema.Tables
	for i, g := range v.Groups {
		tables, _, err := v.Schema.SepareteTablesThatAreIncludedOrNot(&schema.FilterOption{
			Include:       g.Tables,
			IncludeLabels: g.Labels,
		})
		if err != nil {
			return errors.WithStack(err)
		}
		color := g.Color
		if color == "" {
			color = defaultColors[i%len(defaultColors)]
		}
		d := map[string]interface{}{
			"Name":   g.Name,
			"Desc":   g.Desc,
			"Tables": tables,
			"Color":  color,
		}
		groups = append(groups, d)
		nogroup = lo.Without(nogroup, tables...)
	}
	if len(v.Groups) > 0 && len(nogroup) > 0 {
		tables = nogroup
	}

	tmpl := template.Must(template.New(v.Name).Funcs(output.Funcs(&d.config.MergedDict)).Parse(ts))
	if err := tmpl.Execute(wr, map[string]interface{}{
		"Name":        v.Name,
		"Tables":      tables,
		"Relations":   v.Schema.Relations,
		"Groups":      groups,
		"showComment": d.config.ER.Comment,
		"showDef":     !d.config.ER.HideDef,
	}); err != nil {
		return errors.WithStack(err)
	}

	return nil
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
