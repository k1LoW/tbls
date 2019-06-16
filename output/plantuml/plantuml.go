package plantuml

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/gobuffalo/packr/v2"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

var templateFuncs = map[string]interface{}{
	"escape_nl": func(text string) string {
		return strings.Replace(strings.Replace(strings.Replace(text, "\r\n", "\\n", -1), "\n", "\\n", -1), "\r", "\\n", -1)
	},
	"nl2space": func(text string) string {
		return strings.Replace(strings.Replace(strings.Replace(text, "\r\n", " ", -1), "\n", " ", -1), "\r", " ", -1)
	},
}

// PlantUML struct
type PlantUML struct {
	config *config.Config
	box    *packr.Box
}

// NewPlantUML return PlantUML
func NewPlantUML(c *config.Config) *PlantUML {
	return &PlantUML{
		config: c,
		box:    packr.New("plantuml", "./templates"),
	}
}

// OutputSchema output dot format for full relation.
func (p *PlantUML) OutputSchema(wr io.Writer, s *schema.Schema) error {
	for _, t := range s.Tables {
		addPrefix(t)
	}

	ts, _ := p.box.FindString("schema.puml.tmpl")
	tmpl := template.Must(template.New(s.Name).Funcs(templateFuncs).Parse(ts))
	err := tmpl.Execute(wr, map[string]interface{}{
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
	addPrefix(t)
	encountered := make(map[string]bool)
	tables := []*schema.Table{}
	relations := []*schema.Relation{}
	for _, c := range t.Columns {
		for _, r := range c.ParentRelations {
			if !encountered[r.ParentTable.Name] {
				encountered[r.ParentTable.Name] = true
				addPrefix(r.ParentTable)
				tables = append(tables, r.ParentTable)
			}
			if !contains(relations, r) {
				relations = append(relations, r)
			}
		}
		for _, r := range c.ChildRelations {
			if !encountered[r.Table.Name] {
				encountered[r.Table.Name] = true
				addPrefix(r.Table)
				tables = append(tables, r.Table)
			}
			if !contains(relations, r) {
				relations = append(relations, r)
			}
		}
	}

	ts, _ := p.box.FindString("table.puml.tmpl")
	tmpl := template.Must(template.New(t.Name).Funcs(templateFuncs).Parse(ts))
	err := tmpl.Execute(wr, map[string]interface{}{
		"Table":       t,
		"Tables":      tables,
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
