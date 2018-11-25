package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var configDefaultPath = ".tbls.yml"

// Config is tbls config
type Config struct {
	DSN       string               `yaml:"dsn"`
	DocPath   string               `yaml:"dataPath"`
	Relations []AdditionalRelation `yaml:"relations"`
	Comments  []AdditionalComment  `yaml:"comments"`
}

// AdditionalRelation is the struct for table relation from yaml
type AdditionalRelation struct {
	Table         string   `yaml:"table"`
	Columns       []string `yaml:"columns"`
	ParentTable   string   `yaml:"parentTable"`
	ParentColumns []string `yaml:"parentColumns"`
	Def           string   `yaml:"def"`
}

// AdditionalComment is the struct for table relation from yaml
type AdditionalComment struct {
	Table          string            `yaml:"table"`
	TableComment   string            `yaml:"tableComment"`
	ColumnComments map[string]string `yaml:"columnComments"`
}

// NewConfig return Config
func NewConfig() (*Config, error) {
	docPath := os.Getenv("TBLS_DOC_PATH")
	if docPath == "" {
		docPath = "."
	}

	c := Config{
		DSN:     os.Getenv("TBLS_DSN"),
		DocPath: docPath,
	}
	return &c, nil
}

// LoadArgs load args slice
func (c *Config) LoadArgs(args []string) error {
	if len(args) == 2 {
		c.DSN = args[0]
		c.DocPath = args[1]
	}
	if len(args) > 2 {
		return errors.WithStack(errors.New("too many arguments"))
	}
	if len(args) == 1 {
		if c.DSN == "" {
			c.DSN = args[0]
		} else {
			c.DocPath = args[0]
		}
	}
	return nil
}

// LoadConfigFile load config file
func (c *Config) LoadConfigFile(path string) error {
	if path == "" {
		path = configDefaultPath
		if _, err := os.Lstat(path); err != nil {
			return nil
		}
	}

	fullPath, err := filepath.Abs(path)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to load config file")
	}

	buf, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to load config file")
	}

	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to load config file")
	}
	return nil
}
