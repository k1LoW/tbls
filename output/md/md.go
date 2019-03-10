package md

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gobuffalo/packr"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
	"github.com/mattn/go-runewidth"
	"github.com/pkg/errors"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// Md struct
type Md struct {
	adjust   bool
	er       bool
	erFormat string
	box      packr.Box
}

// NewMd return Md
func NewMd(adjust bool, er bool, erFormat string) *Md {
	return &Md{
		adjust:   adjust,
		er:       er,
		erFormat: erFormat,
		box:      packr.NewBox("./templates"),
	}
}

// OutputSchema output .md format for all tables.
func (m *Md) OutputSchema(wr io.Writer, s *schema.Schema) error {
	ts, _ := m.box.FindString("index.md.tmpl")
	tmpl := template.Must(template.New("index").Funcs(funcMap()).Parse(ts))
	templateData := makeSchemaTemplateData(s, m.adjust)
	templateData["er"] = m.er
	templateData["erFormat"] = m.erFormat
	err := tmpl.Execute(wr, templateData)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// OutputTable output md format for table.
func (m *Md) OutputTable(wr io.Writer, t *schema.Table) error {
	ts, _ := m.box.FindString("table.md.tmpl")
	tmpl := template.Must(template.New(t.Name).Funcs(funcMap()).Parse(ts))
	templateData := makeTableTemplateData(t, m.adjust)
	templateData["er"] = m.er
	templateData["erFormat"] = m.erFormat

	err := tmpl.Execute(wr, templateData)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// Output generate markdown files.
func Output(s *schema.Schema, c *config.Config, force bool) error {
	docPath := c.DocPath
	adjust := c.Format.Adjust
	erFormat := c.ER.Format

	fullPath, err := filepath.Abs(docPath)
	if err != nil {
		return errors.WithStack(err)
	}

	if !force && outputExists(s, fullPath) {
		return errors.New("output files already exists")
	}

	_ = os.MkdirAll(fullPath, 0755)

	// README.md
	file, err := os.Create(filepath.Join(fullPath, "README.md"))
	defer file.Close()
	if err != nil {
		return errors.WithStack(err)
	}
	er := false
	if _, err := os.Lstat(filepath.Join(fullPath, fmt.Sprintf("schema.%s", erFormat))); err == nil {
		er = true
	}

	md := NewMd(adjust, er, erFormat)

	err = md.OutputSchema(file, s)
	if err != nil {
		return errors.WithStack(err)
	}
	fmt.Printf("%s\n", filepath.Join(docPath, "README.md"))

	// tables
	for _, t := range s.Tables {
		file, err := os.Create(filepath.Join(fullPath, fmt.Sprintf("%s.md", t.Name)))
		if err != nil {
			file.Close()
			return errors.WithStack(err)
		}

		er := false
		if _, err := os.Lstat(filepath.Join(fullPath, fmt.Sprintf("%s.%s", t.Name, erFormat))); err == nil {
			er = true
		}

		md := NewMd(adjust, er, erFormat)

		err = md.OutputTable(file, t)
		if err != nil {
			file.Close()
			return errors.WithStack(err)
		}
		fmt.Printf("%s\n", filepath.Join(docPath, fmt.Sprintf("%s.md", t.Name)))
		file.Close()
	}
	return nil
}

// Diff database and markdown files.
func Diff(s *schema.Schema, c *config.Config) (string, error) {
	docPath := c.DocPath
	adjust := c.Format.Adjust
	erFormat := c.ER.Format

	var diff string
	fullPath, err := filepath.Abs(docPath)
	if err != nil {
		return "", errors.WithStack(err)
	}

	if !outputExists(s, fullPath) {
		return "", errors.New("target files does not exists")
	}

	dmp := diffmatchpatch.New()

	// README.md
	a := new(bytes.Buffer)
	er := false
	if _, err := os.Lstat(filepath.Join(fullPath, fmt.Sprintf("schema.%s", erFormat))); err == nil {
		er = true
	}

	md := NewMd(adjust, er, erFormat)

	err = md.OutputSchema(a, s)
	if err != nil {
		return "", errors.WithStack(err)
	}

	targetPath := filepath.Join(fullPath, "README.md")
	b, err := ioutil.ReadFile(targetPath)
	if err != nil {
		b = []byte{}
	}

	da, db, dc := dmp.DiffLinesToChars(a.String(), string(b))
	diffs := dmp.DiffMain(da, db, false)
	result := dmp.DiffCharsToLines(diffs, dc)

	if len(result) != 1 || result[0].Type != diffmatchpatch.DiffEqual {
		diff += fmt.Sprintf("diff [database] %s\n", filepath.Join(docPath, "README.md"))
		diff += fmt.Sprintln(dmp.DiffPrettyText(result))
	}

	// tables
	for _, t := range s.Tables {
		a := new(bytes.Buffer)
		er := false
		if _, err := os.Lstat(filepath.Join(fullPath, fmt.Sprintf("%s.%s", t.Name, erFormat))); err == nil {
			er = true
		}

		md := NewMd(adjust, er, erFormat)

		err := md.OutputTable(a, t)
		if err != nil {
			return "", errors.WithStack(err)
		}
		targetPath := filepath.Join(fullPath, fmt.Sprintf("%s.md", t.Name))
		b, err := ioutil.ReadFile(targetPath)
		if err != nil {
			b = []byte{}
		}

		da, db, dc := dmp.DiffLinesToChars(a.String(), string(b))
		diffs := dmp.DiffMain(da, db, false)
		result := dmp.DiffCharsToLines(diffs, dc)
		if len(result) != 1 || result[0].Type != diffmatchpatch.DiffEqual {
			diff += fmt.Sprintf("diff %s %s\n", t.Name, filepath.Join(docPath, fmt.Sprintf("%s.md", t.Name)))
			diff += fmt.Sprintln(dmp.DiffPrettyText(result))
		}
	}
	return diff, nil
}

func outputExists(s *schema.Schema, path string) bool {
	// README.md
	if _, err := os.Lstat(filepath.Join(path, "README.md")); err == nil {
		return true
	}
	// tables
	for _, t := range s.Tables {
		if _, err := os.Lstat(filepath.Join(path, fmt.Sprintf("%s.md", t.Name))); err == nil {
			return true
		}
	}
	return false
}

func funcMap() map[string]interface{} {
	return template.FuncMap{
		"nl2br": func(text string) string {
			r := strings.NewReplacer("\r\n", "<br>", "\n", "<br>", "\r", "<br>")
			return r.Replace(text)
		},
		"nl2mdnl": func(text string) string {
			r := strings.NewReplacer("\r\n", "  \n", "\n", "  \n", "\r", "  \n")
			return r.Replace(text)
		},
	}
}

func makeSchemaTemplateData(s *schema.Schema, adjust bool) map[string]interface{} {
	tablesData := [][]string{
		[]string{"Name", "Columns", "Comment", "Type"},
		[]string{"----", "-------", "-------", "----"},
	}
	for _, t := range s.Tables {
		data := []string{
			fmt.Sprintf("[%s](%s.md)", t.Name, t.Name),
			fmt.Sprintf("%d", len(t.Columns)),
			t.Comment,
			t.Type,
		}
		tablesData = append(tablesData, data)
	}

	if adjust {
		return map[string]interface{}{
			"Schema": s,
			"Tables": adjustTable(tablesData),
		}
	}

	return map[string]interface{}{
		"Schema": s,
		"Tables": tablesData,
	}
}

func makeTableTemplateData(t *schema.Table, adjust bool) map[string]interface{} {
	// Columns
	columnsData := [][]string{
		[]string{"Name", "Type", "Default", "Nullable", "Children", "Parents", "Comment"},
		[]string{"----", "----", "-------", "--------", "--------", "-------", "-------"},
	}
	for _, c := range t.Columns {
		childRelations := []string{}
		for _, r := range c.ChildRelations {
			childRelations = append(childRelations, fmt.Sprintf("[%s](%s.md)", r.Table.Name, r.Table.Name))
		}
		parentRelations := []string{}
		for _, r := range c.ParentRelations {
			parentRelations = append(parentRelations, fmt.Sprintf("[%s](%s.md)", r.ParentTable.Name, r.ParentTable.Name))
		}
		data := []string{
			c.Name,
			c.Type,
			c.Default.String,
			fmt.Sprintf("%v", c.Nullable),
			strings.Join(childRelations, " "),
			strings.Join(parentRelations, " "),
			c.Comment,
		}
		columnsData = append(columnsData, data)
	}

	// Constraints
	constraintsData := [][]string{
		[]string{"Name", "Type", "Definition"},
		[]string{"----", "----", "----------"},
	}
	for _, c := range t.Constraints {
		data := []string{
			c.Name,
			c.Type,
			c.Def,
		}
		constraintsData = append(constraintsData, data)
	}

	// Indexes
	indexesData := [][]string{
		[]string{"Name", "Definition"},
		[]string{"----", "----------"},
	}
	for _, i := range t.Indexes {
		data := []string{
			i.Name,
			i.Def,
		}
		indexesData = append(indexesData, data)
	}

	// Triggers
	triggersData := [][]string{
		[]string{"Name", "Definition"},
		[]string{"----", "----------"},
	}
	for _, i := range t.Triggers {
		data := []string{
			i.Name,
			i.Def,
		}
		triggersData = append(triggersData, data)
	}

	if adjust {
		return map[string]interface{}{
			"Table":       t,
			"Columns":     adjustTable(columnsData),
			"Constraints": adjustTable(constraintsData),
			"Indexes":     adjustTable(indexesData),
			"Triggers":    adjustTable(triggersData),
		}
	}

	return map[string]interface{}{
		"Table":       t,
		"Columns":     columnsData,
		"Constraints": constraintsData,
		"Indexes":     indexesData,
		"Triggers":    triggersData,
	}
}

func adjustTable(data [][]string) [][]string {
	r := strings.NewReplacer("\r\n", "<br>", "\n", "<br>", "\r", "<br>")
	w := make([]int, len(data[0]))
	for i := range data {
		for j := range w {
			l := runewidth.StringWidth(r.Replace(data[i][j]))
			if l > w[j] {
				w[j] = l
			}
		}
	}
	for i := range data {
		for j := range w {
			if i == 1 {
				data[i][j] = strings.Repeat("-", w[j])
			} else {
				data[i][j] = fmt.Sprintf(fmt.Sprintf("%%-%ds", w[j]), r.Replace(data[i][j]))
			}
		}
	}

	return data
}
