package gviz

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/beta/freetype/truetype"
	"github.com/goccy/go-graphviz"
	"github.com/k1LoW/errors"
	"github.com/k1LoW/ffff"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/output/dot"
	"github.com/k1LoW/tbls/schema"
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

// OutputSchema generage image for full relation.
func (g *Gviz) OutputSchema(wr io.Writer, s *schema.Schema) error {
	buf := &bytes.Buffer{}
	if err := g.dot.OutputSchema(buf, s); err != nil {
		return errors.WithStack(err)
	}
	return g.render(wr, buf.Bytes())
}

// OutputTable generage image for table.
func (g *Gviz) OutputTable(wr io.Writer, t *schema.Table) error {
	buf := &bytes.Buffer{}
	if err := g.dot.OutputTable(buf, t); err != nil {
		return errors.WithStack(err)
	}
	return g.render(wr, buf.Bytes())
}

// OutputViewpoint generage image for viewpoint.
func (g *Gviz) OutputViewpoint(wr io.Writer, v *schema.Viewpoint) error {
	buf := &bytes.Buffer{}
	if err := g.dot.OutputViewpoint(buf, v); err != nil {
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

// Output generate images.
func Output(s *schema.Schema, c *config.Config, force bool) (e error) {
	erFormat := c.ER.Format
	outputPath := c.DocPath
	fullPath, err := filepath.Abs(outputPath)
	if err != nil {
		return errors.WithStack(err)
	}

	if !force && outputErExists(s, c.ER.Format, fullPath) {
		return errors.New("output ER diagram files already exists")
	}

	err = os.MkdirAll(fullPath, 0755) // #nosec
	if err != nil {
		return errors.WithStack(err)
	}

	fn := fmt.Sprintf("schema.%s", erFormat)
	fmt.Printf("%s\n", filepath.Join(outputPath, fn))

	f, err := os.OpenFile(filepath.Join(fullPath, fn), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) // #nosec
	if err != nil {
		return errors.WithStack(err)
	}
	g := New(c)
	if err := g.OutputSchema(f, s); err != nil {
		return errors.WithStack(err)
	}

	// tables
	for _, t := range s.Tables {
		fn := fmt.Sprintf("%s.%s", t.Name, erFormat)
		fmt.Printf("%s\n", filepath.Join(outputPath, fn))

		f, err := os.OpenFile(filepath.Join(fullPath, fn), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) // #nosec
		if err != nil {
			return errors.WithStack(err)
		}
		if err := g.OutputTable(f, t); err != nil {
			return errors.WithStack(err)
		}
	}

	// viewpoints
	for i, v := range s.Viewpoints {
		fn := fmt.Sprintf("viewpoint-%d.%s", i, erFormat)
		fmt.Printf("%s\n", filepath.Join(outputPath, fn))
		f, err := os.OpenFile(filepath.Join(fullPath, fn), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) // #nosec
		if err != nil {
			return errors.WithStack(err)
		}
		if err := g.OutputViewpoint(f, v); err != nil {
			return errors.WithStack(err)
		}
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

	fb, err := os.ReadFile(filepath.Clean(path))
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

func outputErExists(s *schema.Schema, erFormat, path string) bool {
	// schema.png
	fn := fmt.Sprintf("schema.%s", erFormat)
	if _, err := os.Lstat(filepath.Join(path, fn)); err == nil {
		return true
	}
	// tables
	for _, t := range s.Tables {
		fn := fmt.Sprintf("%s.%s", t.Name, erFormat)
		if _, err := os.Lstat(filepath.Join(path, fn)); err == nil {
			return true
		}
	}
	// viewpoints
	for i := range s.Viewpoints {
		fn := fmt.Sprintf("viewpoint-%d.%s", i, erFormat)
		if _, err := os.Lstat(filepath.Join(path, fn)); err == nil {
			return true
		}
	}
	return false
}
