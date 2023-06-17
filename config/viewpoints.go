package config

type Viewpoint struct {
	Name     string           `yaml:"name,omitempty"`
	Desc     string           `yaml:"desc,omitempty"`
	Labels   []string         `yaml:"labels,omitempty"`
	Tables   []string         `yaml:"tables,omitempty"`
	Groups   []ViewpointGroup `yaml:"groups,omitempty"`
	Distance int              `yaml:"distance,omitempty"`
}

type ViewpointGroup struct {
	Name   string   `yaml:"name,omitempty"`
	Desc   string   `yaml:"desc,omitempty"`
	Labels []string `yaml:"labels,omitempty"`
	Tables []string `yaml:"tables,omitempty"`
	Color  string   `yaml:"color,omitempty"`
}
