package dot

import (
	"github.com/k1LoW/tbls/schema"
	"io"
	"io/ioutil"
	"path/filepath"
	"text/template"
)

// Output generate dot format for full relation.
func Output(wr io.Writer, s *schema.Schema, table string) error {
	f, _ := Assets.Open(filepath.Join("/", "schema.dot.tmpl"))
	bs, _ := ioutil.ReadAll(f)
	tmpl, err := template.New("index").Parse(string(bs))
	if err != nil {
		return err
	}
	err = tmpl.Execute(wr, map[string]interface{}{
		"Schema": s,
	})
	if err != nil {
		return err
	}

	return nil
}
