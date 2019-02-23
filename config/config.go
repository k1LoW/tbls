package config

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const configDefaultPath = ".tbls.yml"

// Config is tbls config
type Config struct {
	DSN       string               `yaml:"dsn"`
	DocPath   string               `yaml:"dataPath"`
	Lint      Lint                 `yaml:"lint"`
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

	c.DSN, err = parseWithEnviron(c.DSN)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to load config file")
	}
	c.DocPath, err = parseWithEnviron(c.DocPath)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to load config file")
	}
	return nil
}

// MergeAdditionalData merge additional* to schema.Schema
func (c *Config) MergeAdditionalData(s *schema.Schema) error {
	err := mergeAdditionalRelations(s, c.Relations)
	if err != nil {
		return err
	}
	err = mergeAdditionalComments(s, c.Comments)
	if err != nil {
		return err
	}
	return nil
}

func mergeAdditionalRelations(s *schema.Schema, relations []AdditionalRelation) error {
	for _, r := range relations {
		relation := &schema.Relation{
			IsAdditional: true,
		}
		if r.Def != "" {
			relation.Def = r.Def
		} else {
			relation.Def = "Additional Relation"
		}
		var err error
		relation.Table, err = s.FindTableByName(r.Table)
		if err != nil {
			return errors.Wrap(err, "failed to add relation")
		}
		for _, c := range r.Columns {
			column, err := relation.Table.FindColumnByName(c)
			if err != nil {
				return errors.Wrap(err, "failed to add relation")
			}
			relation.Columns = append(relation.Columns, column)
			column.ParentRelations = append(column.ParentRelations, relation)
		}
		relation.ParentTable, err = s.FindTableByName(r.ParentTable)
		if err != nil {
			return errors.Wrap(err, "failed to add relation")
		}
		for _, c := range r.ParentColumns {
			column, err := relation.ParentTable.FindColumnByName(c)
			if err != nil {
				return errors.Wrap(err, "failed to add relation")
			}
			relation.ParentColumns = append(relation.ParentColumns, column)
			column.ChildRelations = append(column.ChildRelations, relation)
		}

		s.Relations = append(s.Relations, relation)
	}
	return nil
}

func mergeAdditionalComments(s *schema.Schema, comments []AdditionalComment) error {
	for _, c := range comments {
		table, err := s.FindTableByName(c.Table)
		if err != nil {
			return errors.Wrap(err, "failed to add table comment")
		}
		if c.TableComment != "" {
			table.Comment = c.TableComment
		}
		for c, comment := range c.ColumnComments {
			column, err := table.FindColumnByName(c)
			if err != nil {
				return errors.Wrap(err, "failed to add column comment")
			}
			column.Comment = comment
		}
	}
	return nil
}

func parseWithEnviron(v string) (string, error) {
	r := regexp.MustCompile(`\${\s*([^{}]+)\s*}`)
	r2 := regexp.MustCompile(`{{([^\.])`)
	r3 := regexp.MustCompile(`__TBLS__(.)`)
	replaced := r.ReplaceAllString(v, "{{.$1}}")
	replaced2 := r2.ReplaceAllString(replaced, "__TBLS__$1")
	tmpl, err := template.New("config").Parse(replaced2)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, envMap())
	if err != nil {
		return "", err
	}
	return r3.ReplaceAllString(buf.String(), "{{$1"), nil
}

func envMap() map[string]string {
	m := map[string]string{}
	for _, kv := range os.Environ() {
		if strings.Index(kv, "=") == -1 {
			continue
		}
		parts := strings.SplitN(kv, "=", 2)
		k := parts[0]
		if len(parts) < 2 {
			m[k] = ""
			continue
		}
		m[k] = parts[1]
	}
	return m
}
