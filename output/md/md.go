package md

import (
	"fmt"
	"github.com/k1LoW/tbls/schema"
	"os"
	"path/filepath"
	"text/template"
)

func Output(s *schema.Schema, dir string) error {
	// README.md
	file, err := os.Create(filepath.Join(dir, "README.md"))
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
	// tables
	for _, t := range s.Tables {
		file, err := os.Create(filepath.Join(dir, fmt.Sprintf("%s.md", t.Name)))
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
	}
	return nil
}
