package mermaid

import (
	"embed"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/output"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

//go:embed templates/*
var tmpl embed.FS

// Mermaid struct
type Mermaid struct {
	config *config.Config
	tmpl   embed.FS
}

// New return Mermaid
func New(c *config.Config) *Mermaid {
	return &Mermaid{
		config: c,
		tmpl:   tmpl,
	}
}

func (p *Mermaid) schemaTemplate() (string, error) {
	if len(p.config.Templates.Mermaid.Schema) > 0 {
		tb, err := os.ReadFile(p.config.Templates.Mermaid.Schema)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	} else {
		tb, err := p.tmpl.ReadFile("templates/schema.mermaid.tmpl")
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	}
}

func (p *Mermaid) tableTemplate() (string, error) {
	if len(p.config.Templates.Mermaid.Schema) > 0 {
		tb, err := os.ReadFile(p.config.Templates.Mermaid.Schema)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	} else {
		tb, err := p.tmpl.ReadFile("templates/table.mermaid.tmpl")
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	}
}

// OutputSchema output dot format for full relation.
func (p *Mermaid) OutputSchema(wr io.Writer, s *schema.Schema) error {
	for _, t := range s.Tables {
		err := addSuffix(t)
		if err != nil {
			return err
		}
	}

	ts, err := p.schemaTemplate()
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
func (p *Mermaid) OutputTable(wr io.Writer, t *schema.Table) error {
	tables, relations, err := t.CollectTablesAndRelations(*p.config.ER.Distance, true)

	if err != nil {
		return errors.WithStack(err)
	}
	for _, t := range tables {
		if err := addSuffix(t); err != nil {
			return errors.WithStack(err)
		}
	}
	ts, err := p.tableTemplate()
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

func addSuffix(t *schema.Table) error {
	// PRIMARY KEY
	for _, i := range t.Indexes {
		if !strings.Contains(i.Def, "PRIMARY") {
			continue
		}
		for _, c := range i.Columns {
			column, err := t.FindColumnByName(c)
			if err != nil {
				return err
			}
			column.Name = fmt.Sprintf("%s PK", column.Name)
		}
	}
	// Foreign Key (Relations)
	for _, c := range t.Columns {
		if len(c.ParentRelations) > 0 && !strings.Contains(c.Name, "PK") {
			c.Name = fmt.Sprintf("%s FK", c.Name)
		}
	}
	return nil
}
