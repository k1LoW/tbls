package gviz

import (
	"bytes"
	"context"
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

// Gviz struct.
type Gviz struct {
	config *config.Config
	dot    *dot.Dot
}

// New return Gviz.
func New(c *config.Config) *Gviz {
	return &Gviz{
		config: c,
		dot:    dot.New(c),
	}
}

// OutputSchema generate image for full relation.
func (g *Gviz) OutputSchema(wr io.Writer, s *schema.Schema) error {
	buf := &bytes.Buffer{}
	if err := g.dot.OutputSchema(buf, s); err != nil {
		return errors.WithStack(err)
	}
	return g.render(wr, buf.Bytes())
}

// OutputTable generate image for table.
func (g *Gviz) OutputTable(wr io.Writer, t *schema.Table) error {
	buf := &bytes.Buffer{}
	if err := g.dot.OutputTable(buf, t); err != nil {
		return errors.WithStack(err)
	}
	return g.render(wr, buf.Bytes())
}

// OutputViewpoint generate image for viewpoint.
func (g *Gviz) OutputViewpoint(wr io.Writer, v *schema.Viewpoint) error {
	buf := &bytes.Buffer{}
	if err := g.dot.OutputViewpoint(buf, v); err != nil {
		return errors.WithStack(err)
	}
	return g.render(wr, buf.Bytes())
}

func (g *Gviz) render(wr io.Writer, b []byte) (e error) {
	ctx := context.Background()
	gviz, err := graphviz.New(ctx)
	if err != nil {
		return err
	}
	if g.config.ER.Font != "" {
		faceFunc, err := getFaceFunc(g.config.ER.Font)
		if err != nil {
			return errors.WithStack(err)
		}
		// FIXME: more better way
		graphviz.SetFontLoader(func(_ context.Context, _ *graphviz.Job, font *graphviz.TextFont) (font.Face, error) {
			return faceFunc(font.Size())
		})
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
	if err := gviz.Render(ctx, graph, graphviz.Format(g.config.ER.Format), wr); err != nil {
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

	if !force && outputErExists(s, c, fullPath) {
		return errors.New("output ER diagram files already exists")
	}

	err = os.MkdirAll(fullPath, 0755) // #nosec
	if err != nil {
		return errors.WithStack(err)
	}

	g := New(c)

	schemaPath, err := c.ConstructERSchemaPath(erFormat)
	if err != nil {
		return errors.WithStack(err)
	}
	if schemaPath != "" {
		schemaFullPath := filepath.Join(fullPath, schemaPath)
		
		if err := os.MkdirAll(filepath.Dir(schemaFullPath), 0755); err != nil {
			return errors.WithStack(err)
		}
		
		fmt.Printf("%s\n", filepath.Join(outputPath, schemaPath))
		
		f, err := os.OpenFile(filepath.Clean(schemaFullPath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) // #nosec
		if err != nil {
			return errors.WithStack(err)
		}
		if err := g.OutputSchema(f, s); err != nil {
			_ = f.Close()
			return errors.WithStack(err)
		}
		if err := f.Close(); err != nil {
			return errors.WithStack(err)
		}
	}

	// tables
	for _, t := range s.Tables {
		tablePath, err := c.ConstructERTablePath(t.Name, t.ShortName, erFormat)
		if err != nil {
			return errors.WithStack(err)
		}
		if tablePath == "" {
			continue
		}

		tableFullPath := filepath.Join(fullPath, tablePath)
		
		if err := os.MkdirAll(filepath.Dir(tableFullPath), 0755); err != nil {
			return errors.WithStack(err)
		}
		
		fmt.Printf("%s\n", filepath.Join(outputPath, tablePath))
		
		f, err := os.OpenFile(filepath.Clean(tableFullPath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) // #nosec
		if err != nil {
			return errors.WithStack(err)
		}
		if err := g.OutputTable(f, t); err != nil {
			_ = f.Close()
			return errors.WithStack(err)
		}
		if err := f.Close(); err != nil {
			return errors.WithStack(err)
		}
	}

	// viewpoints
	for i, v := range s.Viewpoints {
		viewpointPath, err := c.ConstructERViewpointPath(v.Name, i, erFormat)
		if err != nil {
			return errors.WithStack(err)
		}
		if viewpointPath == "" {
			continue
		}

		viewpointFullPath := filepath.Join(fullPath, viewpointPath)
		
		if err := os.MkdirAll(filepath.Dir(viewpointFullPath), 0755); err != nil {
			return errors.WithStack(err)
		}
		
		fmt.Printf("%s\n", filepath.Join(outputPath, viewpointPath))
		
		f, err := os.OpenFile(filepath.Clean(viewpointFullPath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) // #nosec
		if err != nil {
			return errors.WithStack(err)
		}
		if err := g.OutputViewpoint(f, v); err != nil {
			_ = f.Close()
			return errors.WithStack(err)
		}
		if err := f.Close(); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// getFaceFunc.
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

func outputErExists(s *schema.Schema, c *config.Config, path string) bool {
	// schema ER diagram
	if schemaPath, err := c.ConstructERSchemaPath(c.ER.Format); err == nil && schemaPath != "" {
		if _, err := os.Lstat(filepath.Join(path, schemaPath)); err == nil {
			return true
		}
	}
	// tables
	for _, t := range s.Tables {
		if tablePath, err := c.ConstructERTablePath(t.Name, t.ShortName, c.ER.Format); err == nil && tablePath != "" {
			if _, err := os.Lstat(filepath.Join(path, tablePath)); err == nil {
				return true
			}
		}
	}
	// viewpoints
	for i, v := range s.Viewpoints {
		if viewpointPath, err := c.ConstructERViewpointPath(v.Name, i, c.ER.Format); err == nil && viewpointPath != "" {
			if _, err := os.Lstat(filepath.Join(path, viewpointPath)); err == nil {
				return true
			}
		}
	}
	return false
}
