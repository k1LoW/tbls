package plantuml

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/gobuffalo/packr/v2"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/output"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

// PlantUML struct
type PlantUML struct {
	config *config.Config
	box    *packr.Box
}

// New return PlantUML
func New(c *config.Config) *PlantUML {
	return &PlantUML{
		config: c,
		box:    packr.New("plantuml", "./templates"),
	}
}

// OutputSchema output dot format for full relation.
func (p *PlantUML) OutputSchema(wr io.Writer, s *schema.Schema) error {
	for _, t := range s.Tables {
		err := addPrefix(t)
		if err != nil {
			return err
		}
	}

	ts, err := p.box.FindString("schema.puml.tmpl")
	if err != nil {
		return errors.WithStack(err)
	}
	tmpl := template.Must(template.New(s.Name).Funcs(output.Funcs(&p.config.MergedDict)).Parse(ts))
	err = tmpl.Execute(wr, map[string]interface{}{
		"Schema":      s,
		"showComment": p.config.ER.Comment,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// OutputTable output dot format for table.
func (p *PlantUML) OutputTable(wr io.Writer, t *schema.Table) error {
	tables, relations, err := t.CollectTablesAndRelations(*p.config.ER.Distance, true)

	if err != nil {
		return errors.WithStack(err)
	}
	for _, t := range tables {
		if err := addPrefix(t); err != nil {
			return errors.WithStack(err)
		}
	}
	ts, err := p.box.FindString("table.puml.tmpl")
	if err != nil {
		return errors.WithStack(err)
	}
	tmpl := template.Must(template.New(t.Name).Funcs(output.Funcs(&p.config.MergedDict)).Parse(ts))
	err = tmpl.Execute(wr, map[string]interface{}{
		"Table":       tables[0],
		"Tables":      tables[1:],
		"Relations":   relations,
		"showComment": p.config.ER.Comment,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func addPrefix(t *schema.Table) error {
	// PRIMARY KEY
	for _, i := range t.Indexes {
		if strings.Index(i.Def, "PRIMARY") < 0 {
			continue
		}
		for _, c := range i.Columns {
			column, err := t.FindColumnByName(c)
			if err != nil {
				return err
			}
			column.Name = fmt.Sprintf("+ %s", column.Name)
		}
	}
	// Foreign Key (Relations)
	for _, c := range t.Columns {
		if len(c.ParentRelations) > 0 && strings.Index(c.Name, "+") < 0 {
			c.Name = fmt.Sprintf("# %s", c.Name)
		}
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
