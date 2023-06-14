package config

import "github.com/k1LoW/tbls/schema"

type Viewpoint struct {
	Name   string   `yaml:"name,omitempty"`
	Desc   string   `yaml:"desc,omitempty"`
	Labels []string `yaml:"labels,omitempty"`
	Tables []string `yaml:"tables,omitempty"`
}

func (v *Viewpoint) FilterTables(s *schema.Schema) error {
	c := &Config{Include: v.Tables, includeLabels: v.Labels}
	return c.FilterTables(s)
}
