package md

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/k1LoW/tbls/schema"
	"github.com/mattn/go-runewidth"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// Output generate markdown files.
func Output(s *schema.Schema, path string, force bool, adjust bool) error {
	fullPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if !force && outputExists(s, fullPath) {
		return fmt.Errorf("Error: %s", "output files already exists.")
	}

	// README.md
	file, err := os.Create(filepath.Join(fullPath, "README.md"))
	defer file.Close()
	if err != nil {
		return err
	}
	f, _ := Assets.Open(filepath.Join("/", "index.md.tmpl"))
	bs, _ := ioutil.ReadAll(f)
	tmpl := template.Must(template.New("index").Funcs(funcMap()).Parse(string(bs)))
	er := false
	if _, err := os.Lstat(filepath.Join(fullPath, "schema.png")); err == nil {
		er = true
	}

	templateData := makeSchemaTemplateData(s, adjust)
	templateData["er"] = er

	err = tmpl.Execute(file, templateData)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", filepath.Join(path, "README.md"))

	// tables
	for _, t := range s.Tables {
		file, err := os.Create(filepath.Join(fullPath, fmt.Sprintf("%s.md", t.Name)))
		if err != nil {
			file.Close()
			return err
		}
		f, _ := Assets.Open(filepath.Join("/", "table.md.tmpl"))
		bs, _ := ioutil.ReadAll(f)
		tmpl := template.Must(template.New(t.Name).Funcs(funcMap()).Parse(string(bs)))
		er := false
		if _, err := os.Lstat(filepath.Join(fullPath, fmt.Sprintf("%s.png", t.Name))); err == nil {
			er = true
		}

		templateData := makeTableTemplateData(t, adjust)
		templateData["er"] = er

		err = tmpl.Execute(file, templateData)
		if err != nil {
			file.Close()
			return err
		}
		fmt.Printf("%s\n", filepath.Join(path, fmt.Sprintf("%s.md", t.Name)))
		file.Close()
	}
	return nil
}

// Diff database and markdown files.
func Diff(s *schema.Schema, path string, adjust bool) error {
	fullPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if !outputExists(s, fullPath) {
		return fmt.Errorf("Error: %s", "target files does not exists.")
	}

	dmp := diffmatchpatch.New()

	// README.md
	a := new(bytes.Buffer)
	f, _ := Assets.Open(filepath.Join("/", "index.md.tmpl"))
	bs, _ := ioutil.ReadAll(f)
	tmpl := template.Must(template.New("index").Funcs(funcMap()).Parse(string(bs)))
	er := false
	if _, err := os.Lstat(filepath.Join(fullPath, "schema.png")); err == nil {
		er = true
	}

	templateData := makeSchemaTemplateData(s, adjust)
	templateData["er"] = er

	err = tmpl.Execute(a, templateData)
	if err != nil {
		return err
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
		fmt.Printf("diff [database] %s\n", filepath.Join(path, "README.md"))
		fmt.Println(dmp.DiffPrettyText(result))
	}

	// tables
	for _, t := range s.Tables {
		a := new(bytes.Buffer)
		f, _ := Assets.Open(filepath.Join("/", "table.md.tmpl"))
		bs, _ := ioutil.ReadAll(f)
		tmpl := template.Must(template.New(t.Name).Funcs(funcMap()).Parse(string(bs)))
		er := false
		if _, err := os.Lstat(filepath.Join(fullPath, fmt.Sprintf("%s.png", t.Name))); err == nil {
			er = true
		}

		templateData := makeTableTemplateData(t, adjust)
		templateData["er"] = er

		err = tmpl.Execute(a, templateData)

		if err != nil {
			return err
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
			fmt.Printf("diff %s %s\n", t.Name, filepath.Join(path, fmt.Sprintf("%s.md", t.Name)))
			fmt.Println(dmp.DiffPrettyText(result))
		}
	}
	return nil
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

	if adjust {
		return map[string]interface{}{
			"Table":       t,
			"Columns":     adjustTable(columnsData),
			"Constraints": adjustTable(constraintsData),
			"Indexes":     adjustTable(indexesData),
		}
	}

	return map[string]interface{}{
		"Table":       t,
		"Columns":     columnsData,
		"Constraints": constraintsData,
		"Indexes":     indexesData,
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
