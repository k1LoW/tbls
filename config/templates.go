package config

// Templates holds the configurations to override the default
// templates used to render the schema and the docs.
type Templates struct {
	MD      MD      `yaml:"md,omitempty"`
	Dot     Dot     `yaml:"dot,omitempty"`
	PUML    PUML    `yaml:"puml,omitempty"`
	Mermaid Mermaid `yaml:"mermaid,omitempty"`
}

// OutputPaths holds the configurations for customizing output file paths.
type OutputPaths struct {
	MD OutputPathsMD `yaml:"md,omitempty"`
	ER OutputPathsER `yaml:"er,omitempty"`
}

// OutputPathsMD holds the output file path patterns for markdown files.
// Each field supports template variables to customize file naming and organization.
// nil = use default, empty string = disable generation, non-empty = custom path
type OutputPathsMD struct {
	Index     *string `yaml:"index,omitempty"`     // README.md path
	Table     *string `yaml:"table,omitempty"`     // Table file path pattern (supports {{.Name}})
	Viewpoint *string `yaml:"viewpoint,omitempty"` // Viewpoint file path pattern (supports {{.Name}}, {{.Index}})
	Enum      *string `yaml:"enum,omitempty"`      // Enum file path pattern (supports {{.Name}})
}

// OutputPathsER holds the output file path patterns for ER diagram image files.
// Each field supports template variables to customize file naming and organization.
// nil = use default, empty string = disable generation, non-empty = custom path
type OutputPathsER struct {
	Schema    *string `yaml:"schema,omitempty"`    // Schema ER diagram path pattern (supports {{.Format}})
	Table     *string `yaml:"table,omitempty"`     // Table ER diagram path pattern (supports {{.Name}}, {{.Format}})
	Viewpoint *string `yaml:"viewpoint,omitempty"` // Viewpoint ER diagram path pattern (supports {{.Name}}, {{.Index}}, {{.Format}})
}

// MD holds the paths to the markdown template files.
// If populated the files are used to override the default ones.
type MD struct {
	Index     string `yaml:"index,omitempty"`
	Table     string `yaml:"table,omitempty"`
	Viewpoint string `yaml:"viewpoint,omitempty"`
	Enum      string `yaml:"enum,omitempty"`
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

// Mermaid holds the paths to the Mermaid template files.
// If populated the files are used to override the default ones.
type Mermaid struct {
	Schema string `yaml:"schema,omitempty"`
	Table  string `yaml:"table,omitempty"`
}
