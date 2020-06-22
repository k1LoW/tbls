package gviz

import (
	"bytes"
	"io"

	"github.com/goccy/go-graphviz"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/output/dot"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

// Gviz struct
type Gviz struct {
	config *config.Config
	dot    *dot.Dot
}

// New return Gviz
func New(c *config.Config) *Gviz {
	return &Gviz{
		config: c,
		dot:    dot.New(c),
	}
}

// OutputSchema output dot format for full relation.
func (g *Gviz) OutputSchema(wr io.Writer, s *schema.Schema) error {
	buf := &bytes.Buffer{}
	err := g.dot.OutputSchema(buf, s)
	if err != nil {
		return errors.WithStack(err)
	}
	return g.render(wr, buf.Bytes())
}

// OutputTable output dot format for table.
func (g *Gviz) OutputTable(wr io.Writer, t *schema.Table) error {
	buf := &bytes.Buffer{}
	err := g.dot.OutputTable(buf, t)
	if err != nil {
		return errors.WithStack(err)
	}
	return g.render(wr, buf.Bytes())
}

func (g *Gviz) render(wr io.Writer, b []byte) (e error) {
	gviz := graphviz.New()
	graph, err := graphviz.ParseBytes(b)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if err := gviz.Close(); err != nil {
			e = errors.WithStack(err)
		}
		if err := graph.Close(); err != nil {
			e = errors.WithStack(err)
		}
	}()
	if err := gviz.Render(graph, graphviz.Format(g.config.ER.Format), wr); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
