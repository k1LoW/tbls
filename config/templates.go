package config

// Templates holds the configurations to override the default
// templates used to render the schema and the docs.
type Templates struct {
	MD   MD   `yaml:"md,omitempty"`
	Dot  Dot  `yaml:"dot,omitempty"`
	PUML PUML `yaml:"puml,omitempty"`
}

// MD holds the paths to the markdown template files.
// If populated the files are used to override the default ones.
type MD struct {
	Index string `yaml:"index,omitempty"`
	Table string `yaml:"table,omitempty"`
}

// Dot holds the paths to the dot template files.
// If populated the files are used to override the default ones.
type Dot struct {
	Schema string `yaml:"schema,omitempty"`
	Table  string `yaml:"table,omitempty"`
}

// PUML holds the paths to the PlantUML template files.
// If populated the files are used to override the default ones.
type PUML struct {
	Schema string `yaml:"schema,omitempty"`
	Table  string `yaml:"table,omitempty"`
}
