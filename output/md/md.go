package md

import (
	"bytes"
	"fmt"
	"github.com/k1LoW/tbls/schema"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

// Output generate markdown files.
func Output(s *schema.Schema, path string, force bool) error {
	fullPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if !force && outputExists(s, fullPath) {
		return fmt.Errorf("Error: %s", "output files already exists.")
	}

	// README.md
	file, err := os.Create(filepath.Join(fullPath, "README.md"))
	if err != nil {
		return err
	}
	tmpl := template.Must(template.ParseFiles("output/md/index.md.tmpl"))
	err = tmpl.Execute(file, map[string]interface{}{
		"Schema": s,
	})
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", filepath.Join(path, "README.md"))

	// tables
	for _, t := range s.Tables {
		file, err := os.Create(filepath.Join(fullPath, fmt.Sprintf("%s.md", t.Name)))
		if err != nil {
			return err
		}
		tmpl := template.Must(template.ParseFiles("output/md/table.md.tmpl"))
		err = tmpl.Execute(file, map[string]interface{}{
			"Table": t,
		})
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", filepath.Join(path, fmt.Sprintf("%s.md", t.Name)))
	}
	return nil
}

// Diff database and markdown files.
func Diff(s *schema.Schema, path string) error {
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
	tmpl := template.Must(template.ParseFiles("output/md/index.md.tmpl"))
	err = tmpl.Execute(a, map[string]interface{}{
		"Schema": s,
	})
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
		tmpl := template.Must(template.ParseFiles("output/md/table.md.tmpl"))
		err = tmpl.Execute(a, map[string]interface{}{
			"Table": t,
		})
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
