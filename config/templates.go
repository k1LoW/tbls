package config

// Templates holds the configurations to override the default
// templates used to render the schema and the docs.
type Templates struct {
	MD   MD   `yaml:"md,omitempty"`
	Dot  Dot  `yaml:"dot,omitempty"`
	PUML PUML `yaml:"puml,omitempty"`
}

type MD struct {
	Index string `yaml:"index,omitempty"`
	Table string `yaml:"table,omitempty"`
}

type Dot struct {
	Schema string `yaml:"schema,omitempty"`
	Table  string `yaml:"table,omitempty"`
}
type PUML struct {
	Schema string `yaml:"schema,omitempty"`
	Table  string `yaml:"table,omitempty"`
}
