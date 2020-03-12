package config

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/k1LoW/tbls/schema"
	"github.com/minio/minio/pkg/wildcard"
	"github.com/pkg/errors"
)

const defaultConfigFilePath = ".tbls.yml"
const defaultDocPath = "dbdoc"

// DefaultERFormat is the default ER diagram format
const DefaultERFormat = "png"

// DefaultDistance is the default distance between tables that display relations in the ER
var DefaultDistance = 1

// Config is tbls config
type Config struct {
	Name        string               `yaml:"name"`
	DSN         string               `yaml:"dsn"`
	DocPath     string               `yaml:"docPath"`
	Format      Format               `yaml:"format"`
	ER          ER                   `yaml:"er"`
	Include     []string             `yaml:"include"`
	Exclude     []string             `yaml:"exclude"`
	Lint        Lint                 `yaml:"lint"`
	LintExclude []string             `yaml:"lintExclude"`
	Relations   []AdditionalRelation `yaml:"relations"`
	Comments    []AdditionalComment  `yaml:"comments"`
}

// Format is document format setting
type Format struct {
	Adjust bool `yaml:"adjust"`
	Sort   bool `yaml:"sort"`
}

// ER is er setting
type ER struct {
	Skip     bool   `yaml:"skip"`
	Format   string `yaml:"format"`
	Comment  bool   `yaml:"comment"`
	Distance *int   `yaml:"distance"`
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

// Option function change Config
type Option func(*Config) error

// DSN return Option set Config.DSN
func DSN(dsn string) Option {
	return func(c *Config) error {
		c.DSN = dsn
		return nil
	}
}

// DocPath return Option set Config.DocPath
func DocPath(docPath string) Option {
	return func(c *Config) error {
		c.DocPath = docPath
		return nil
	}
}

// Adjust return Option set Config.Format.Adjust
func Adjust(adjust bool) Option {
	return func(c *Config) error {
		if adjust {
			c.Format.Adjust = adjust
		}
		return nil
	}
}

// Sort return Option set Config.Format.Sort
func Sort(sort bool) Option {
	return func(c *Config) error {
		if sort {
			c.Format.Sort = sort
		}
		return nil
	}
}

// ERSkip return Option set Config.ER.Skip
func ERSkip(skip bool) Option {
	return func(c *Config) error {
		c.ER.Skip = skip
		return nil
	}
}

// ERFormat return Option set Config.ER.Format
func ERFormat(erFormat string) Option {
	return func(c *Config) error {
		if erFormat != "" {
			c.ER.Format = erFormat
		}
		return nil
	}
}

// Distance return Option set Config.ER.Distance
func Distance(distance int) Option {
	return func(c *Config) error {
		c.ER.Distance = &distance
		return nil
	}
}

// NewConfig return Config
func NewConfig() (*Config, error) {
	c := Config{}
	err := c.setDefault()
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Load load config with all method
func (c *Config) Load(configPath string, options ...Option) error {
	err := c.LoadConfigFile(configPath)
	if err != nil {
		return err
	}

	err = c.LoadEnviron()
	if err != nil {
		return err
	}

	for _, option := range options {
		err = option(c)
		if err != nil {
			return err
		}
	}

	err = c.setDefault()
	if err != nil {
		return err
	}

	return nil
}

// set default setting
func (c *Config) setDefault() error {
	if c.DocPath == "" {
		c.DocPath = defaultDocPath
	}

	if c.ER.Format == "" {
		c.ER.Format = DefaultERFormat
	}

	if c.ER.Distance == nil {
		c.ER.Distance = &DefaultDistance
	}

	return nil
}

// LoadEnviron load environment variables
func (c *Config) LoadEnviron() error {
	dsn := os.Getenv("TBLS_DSN")
	if dsn != "" {
		c.DSN = dsn
	}
	docPath := os.Getenv("TBLS_DOC_PATH")
	if docPath != "" {
		c.DocPath = docPath
	}
	return nil
}

// LoadConfigFile load config file
func (c *Config) LoadConfigFile(path string) error {
	if path == "" {
		path = defaultConfigFilePath
		if _, err := os.Lstat(path); err != nil {
			return nil
		}
	}

	fullPath, err := filepath.Abs(path)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to load config file")
	}

	buf, err := ioutil.ReadFile(filepath.Clean(fullPath))
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

// ModifySchema modify schema.Schema by config
func (c *Config) ModifySchema(s *schema.Schema) error {
	if c.Name != "" {
		s.Name = c.Name
	}
	err := c.MergeAdditionalData(s)
	if err != nil {
		return err
	}
	err = c.FilterTables(s)
	if err != nil {
		return err
	}
	if c.Format.Sort {
		err = s.Sort()
		if err != nil {
			return err
		}
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

// FilterTables filter tables from schema.Schema
func (c *Config) FilterTables(s *schema.Schema) error {
	for _, t := range s.Tables {
		if len(c.Include) == 0 || contains(c.Include, t.Name) {
			if contains(c.Exclude, t.Name) {
				err := excludeTableFromSchema(t.Name, s)
				if err != nil {
					return errors.Wrap(errors.WithStack(err), fmt.Sprintf("failed to filter table '%s'", t.Name))
				}
			}
		} else {
			err := excludeTableFromSchema(t.Name, s)
			if err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("failed to filter table '%s'", t.Name))
			}
		}
	}
	return nil
}

func excludeTableFromSchema(name string, s *schema.Schema) error {
	// Tables
	tables := []*schema.Table{}
	for _, t := range s.Tables {
		if t.Name != name {
			tables = append(tables, t)
		}
		for _, c := range t.Columns {
			// ChildRelations
			childRelations := []*schema.Relation{}
			for _, r := range c.ChildRelations {
				if r.Table.Name != name {
					childRelations = append(childRelations, r)
				}
			}
			c.ChildRelations = childRelations

			// ParentRelations
			parentRelations := []*schema.Relation{}
			for _, r := range c.ParentRelations {
				if r.Table.Name != name {
					parentRelations = append(parentRelations, r)
				}
			}
			c.ParentRelations = parentRelations
		}
	}
	s.Tables = tables

	// Relations
	relations := []*schema.Relation{}
	for _, r := range s.Relations {
		if r.Table.Name != name {
			relations = append(relations, r)
		}
	}
	s.Relations = relations

	return nil
}

// MaskedDSN return DSN mask password
func (c *Config) MaskedDSN() (string, error) {
	u, err := url.Parse(c.DSN)
	if err != nil {
		return c.DSN, errors.WithStack(err)
	}
	tmp := "-----tbls-----"
	u.User = url.UserPassword(u.User.Username(), tmp)
	return strings.Replace(u.String(), tmp, "*****", 1), nil
}

func mergeAdditionalRelations(s *schema.Schema, relations []AdditionalRelation) error {
	for _, r := range relations {
		relation := &schema.Relation{
			Virtual: true,
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
		if !strings.Contains(kv, "=") {
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

func contains(s []string, e string) bool {
	for _, v := range s {
		if wildcard.MatchSimple(v, e) {
			return true
		}
	}
	return false
}
