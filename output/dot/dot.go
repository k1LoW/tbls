package dot

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/k1LoW/tbls/schema"
)

// OutputSchema generate dot format for full relation.
func OutputSchema(wr io.Writer, s *schema.Schema) error {
	f, _ := Assets.Open(filepath.Join("/", "schema.dot.tmpl"))
	bs, _ := ioutil.ReadAll(f)
	tmpl, err := template.New(s.Name).Parse(string(bs))
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

// OutputTable generate dot format for table.
func OutputTable(wr io.Writer, t *schema.Table) error {
	encountered := make(map[string]bool)
	tables := []*schema.Table{}
	relations := []*schema.Relation{}
	for _, c := range t.Columns {
		for _, r := range c.ParentRelations {
			if !encountered[r.ParentTable.Name] {
				encountered[r.ParentTable.Name] = true
				tables = append(tables, r.ParentTable)
			}
			relations = append(relations, r)
		}
		for _, r := range c.ChildRelations {
			if !encountered[r.Table.Name] {
				encountered[r.Table.Name] = true
				tables = append(tables, r.Table)
			}
			relations = append(relations, r)
		}
	}

	f, _ := Assets.Open(filepath.Join("/", "table.dot.tmpl"))
	bs, _ := ioutil.ReadAll(f)
	tmpl, err := template.New(t.Name).Parse(string(bs))
	if err != nil {
		return err
	}
	err = tmpl.Execute(wr, map[string]interface{}{
		"Table":     t,
		"Tables":    tables,
		"Relations": relations,
	})
	if err != nil {
		return err
	}

	return nil
}
