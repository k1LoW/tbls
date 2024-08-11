package mermaid

import (
	"embed"
	"io"
	"os"
	"text/template"

	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/output"
	"github.com/k1LoW/tbls/schema"
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

func (m *Mermaid) schemaTemplate() (string, error) {
	if len(m.config.Templates.Mermaid.Schema) > 0 {
		tb, err := os.ReadFile(m.config.Templates.Mermaid.Schema)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	} else {
		tb, err := m.tmpl.ReadFile("templates/schema.mermaid.tmpl")
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	}
}

func (m *Mermaid) tableTemplate() (string, error) {
	if len(m.config.Templates.Mermaid.Table) > 0 {
		tb, err := os.ReadFile(m.config.Templates.Mermaid.Table)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	} else {
		tb, err := m.tmpl.ReadFile("templates/table.mermaid.tmpl")
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	}
}

// OutputSchema output dot format for full relation.
func (m *Mermaid) OutputSchema(wr io.Writer, s *schema.Schema) error {
	ts, err := m.schemaTemplate()
	if err != nil {
		return errors.WithStack(err)
	}
	tmpl := template.Must(template.New(s.Name).Funcs(output.Funcs(&m.config.MergedDict)).Parse(ts))
	err = tmpl.Execute(wr, map[string]interface{}{
		"Schema":          s,
		"showComment":     m.config.ER.Comment,
		"showDef":         !m.config.ER.HideDef,
		"showColumnTypes": m.config.ER.ShowColumnTypes,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// OutputTable output dot format for table.
func (m *Mermaid) OutputTable(wr io.Writer, t *schema.Table) error {
	tables, relations, err := t.CollectTablesAndRelations(*m.config.ER.Distance, true)
	if err != nil {
		return errors.WithStack(err)
	}
	ts, err := m.tableTemplate()
	if err != nil {
		return errors.WithStack(err)
	}
	tmpl := template.Must(template.New(t.Name).Funcs(output.Funcs(&m.config.MergedDict)).Parse(ts))
	err = tmpl.Execute(wr, map[string]interface{}{
		"Table":           tables[0],
		"Tables":          tables[1:],
		"Relations":       relations,
		"showComment":     m.config.ER.Comment,
		"showDef":         !m.config.ER.HideDef,
		"showColumnTypes": m.config.ER.ShowColumnTypes,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
