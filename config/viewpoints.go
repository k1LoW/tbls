package config

type Viewpoint struct {
	Name   string   `yaml:"name,omitempty"`
	Desc   string   `yaml:"desc,omitempty"`
	Labels []string `yaml:"labels,omitempty"`
	Tables []string `yaml:"tables,omitempty"`
}
