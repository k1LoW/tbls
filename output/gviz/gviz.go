package gviz

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/beta/freetype/truetype"
	"github.com/goccy/go-graphviz"
	"github.com/k1LoW/ffff"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/output/dot"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
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
	if err := g.dot.OutputSchema(buf, s); err != nil {
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
	if g.config.ER.Font != "" {
		faceFunc, err := getFaceFunc(g.config.ER.Font)
		if err != nil {
			return errors.WithStack(err)
		}
		gviz.SetFontFace(faceFunc)
	}
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

// getFaceFunc
func getFaceFunc(keyword string) (func(size float64) (font.Face, error), error) {
	var (
		faceFunc func(size float64) (font.Face, error)
		path     string
	)

	fi, err := os.Stat(keyword)
	if err == nil && !fi.IsDir() {
		path = keyword
	} else {
		path, err = ffff.FuzzyFindPath(keyword)
		if err != nil {
			return faceFunc, errors.WithStack(err)
		}
	}

	fb, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return faceFunc, errors.WithStack(err)
	}

	if strings.HasSuffix(path, ".otf") {
		// OpenType
		ft, err := sfnt.Parse(fb)
		if err != nil {
			return faceFunc, errors.WithStack(err)
		}
		faceFunc = func(size float64) (font.Face, error) {
			opt := &opentype.FaceOptions{
				Size:    size,
				DPI:     0,
				Hinting: 0,
			}
			return opentype.NewFace(ft, opt)
		}
	} else {
		// TrueType
		ft, err := truetype.Parse(fb)
		if err != nil {
			return faceFunc, errors.WithStack(err)
		}
		faceFunc = func(size float64) (font.Face, error) {
			opt := &truetype.Options{
				Size:              size,
				DPI:               0,
				Hinting:           0,
				GlyphCacheEntries: 0,
				SubPixelsX:        0,
				SubPixelsY:        0,
			}
			return truetype.NewFace(ft, opt), nil
		}
	}
	return faceFunc, nil
}
