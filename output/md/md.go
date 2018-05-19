package md

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/k1LoW/tbls/schema"
	"os"
	"path/filepath"
	"text/template"
)

func Output(s *schema.Schema, dir string) error {
	spew.Dump(s)

	// tables
	file, err := os.OpenFile(filepath.Join(dir, "index.md"), os.O_WRONLY|os.O_CREATE, os.ModePerm)
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
	for _, t := range s.Tables {
		fmt.Printf("%s\n", t.Name)
		file, err := os.OpenFile(filepath.Join(dir, fmt.Sprintf("%s.md", t.Name)), os.O_WRONLY|os.O_CREATE, os.ModePerm)
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
