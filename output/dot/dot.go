package dot

import (
	"fmt"
	"github.com/k1LoW/tbls/schema"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

// Output generate dot file for full relation.
func Output(s *schema.Schema, path string, force bool) error {
	fullPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if !force && outputExists(s, fullPath) {
		return fmt.Errorf("Error: %s", "output file already exists.")
	}

	file, err := os.Create(filepath.Join(fullPath, "schema.dot"))
	defer file.Close()
	if err != nil {
		return err
	}
	f, _ := Assets.Open(filepath.Join("/", "schema.dot.tmpl"))
	bs, _ := ioutil.ReadAll(f)
	tmpl, err := template.New("index").Parse(string(bs))
	if err != nil {
		return err
	}
	err = tmpl.Execute(file, map[string]interface{}{
		"Schema": s,
	})
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", filepath.Join(path, "schema.dot"))

	return nil
}

func outputExists(s *schema.Schema, path string) bool {
	if _, err := os.Lstat(filepath.Join(path, "schema.dot")); err == nil {
		return true
	}
	return false
}
