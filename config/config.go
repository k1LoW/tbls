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
	"github.com/k1LoW/tbls/dict"
	"github.com/k1LoW/tbls/schema"
	"github.com/minio/minio/pkg/wildcard"
	"github.com/pkg/errors"
)

const DefaultDocPath = "dbdoc"

var DefaultConfigFilePaths = []string{".tbls.yml", "tbls.yml"}

// DefaultERFormat is the default ER diagram format
const DefaultERFormat = "svg"

// DefaultDistance is the default distance between tables that display relations in the ER
var DefaultDistance = 1

// Config is tbls config
type Config struct {
	Name        string               `yaml:"name"`
	Desc        string               `yaml:"desc,omitempty"`
	Labels      []string             `yaml:"labels,omitempty"`
	DSN         DSN                  `yaml:"dsn"`
	DocPath     string               `yaml:"docPath"`
	Format      Format               `yaml:"format,omitempty"`
	ER          ER                   `yaml:"er,omitempty"`
	Include     []string             `yaml:"include,omitempty"`
	Exclude     []string             `yaml:"exclude,omitempty"`
	Lint        Lint                 `yaml:"lint,omitempty"`
	LintExclude []string             `yaml:"lintExclude,omitempty"`
	Relations   []AdditionalRelation `yaml:"relations,omitempty"`
	Comments    []AdditionalComment  `yaml:"comments,omitempty"`
	Dict        dict.Dict            `yaml:"dict,omitempty"`
	Templates   Templates            `yaml:"templates,omitempty"`
	MergedDict  dict.Dict            `yaml:"-"`
	Path        string               `yaml:"-"`
	root        string               `yaml:"-"`
	BaseUrl     string               `yaml:"baseUrl,omitempty"`
}

type DSN struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

// Format is document format setting
type Format struct {
	Adjust bool `yaml:"adjust,omitempty"`
	Sort   bool `yaml:"sort,omitempty"`
}

// ER is er setting
type ER struct {
	Skip     bool   `yaml:"skip,omitempty"`
	Format   string `yaml:"format,omitempty"`
	Comment  bool   `yaml:"comment,omitempty"`
	Distance *int   `yaml:"distance,omitempty"`
	Font     string `yaml:"font,omitempty"`
}

// AdditionalRelation is the struct for table relation from yaml
type AdditionalRelation struct {
	Table         string   `yaml:"table"`
	Columns       []string `yaml:"columns"`
	ParentTable   string   `yaml:"parentTable"`
	ParentColumns []string `yaml:"parentColumns"`
	Def           string   `yaml:"def,omitempty"`
}

// AdditionalComment is the struct for table relation from yaml
type AdditionalComment struct {
	Table              string            `yaml:"table"`
	TableComment       string            `yaml:"tableComment,omitempty"`
	ColumnComments     map[string]string `yaml:"columnComments,omitempty"`
	IndexComments      map[string]string `yaml:"indexComments,omitempty"`
	ConstraintComments map[string]string `yaml:"constraintComments,omitempty"`
	TriggerComments    map[string]string `yaml:"triggerComments,omitempty"`
	Labels             []string          `yaml:"labels,omitempty"`
}

// Option function change Config
type Option func(*Config) error

// DSNURL return Option set Config.DSN.URL
func DSNURL(dsn string) Option {
	return func(c *Config) error {
		c.DSN.URL = dsn
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

// BaseUrl return Option set Config.BaseUrl
func BaseUrl(baseUrl string) Option {
	return func(c *Config) error {
		if baseUrl != "" {
			c.BaseUrl = baseUrl
		}
		return nil
	}
}

// New return Config
func New() (*Config, error) {
	c := Config{}
	err := c.setDefault()
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Load load config with all method
func (c *Config) Load(configPath string, options ...Option) error {
	if err := c.LoadConfigFile(configPath); err != nil {
		return err
	}

	if err := c.LoadEnviron(); err != nil {
		return err
	}

	if err := c.LoadOption(options...); err != nil {
		return err
	}

	if err := c.setDefault(); err != nil {
		return err
	}

	return nil
}

// LoadOptions load options
func (c *Config) LoadOption(options ...Option) error {
	for _, option := range options {
		if err := option(c); err != nil {
			return err
		}
	}
	return nil
}

// set default setting
func (c *Config) setDefault() error {
	if c.DocPath == "" {
		c.DocPath = DefaultDocPath
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
		c.DSN.URL = dsn
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
		for _, p := range DefaultConfigFilePaths {
			if f, err := os.Stat(filepath.Join(c.root, p)); err == nil && !f.IsDir() {
				if path != "" {
					return fmt.Errorf("duplicate config file [%s, %s]", path, p)
				}
				path = p
			}
		}
	}
	if path == "" {
		return nil
	}

	fullPath, err := filepath.Abs(path)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to load config file")
	}

	buf, err := ioutil.ReadFile(filepath.Clean(fullPath))
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to load config file")
	}
	c.Path = filepath.Clean(fullPath)

	return c.LoadConfig(buf)
}

// LoadConfig load config from []byte
func (c *Config) LoadConfig(in []byte) error {
	err := yaml.Unmarshal(in, c)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to load config file")
	}
	c.DSN.URL, err = parseWithEnviron(c.DSN.URL)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to load config file")
	}
	c.DocPath, err = parseWithEnviron(c.DocPath)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to load config file")
	}

	c.MergedDict.Merge(c.Dict.Dump())
	return nil
}

// ModifySchema modify schema.Schema by config
func (c *Config) ModifySchema(s *schema.Schema) error {
	if c.Name != "" {
		s.Name = c.Name
	}
	if c.Desc != "" {
		s.Desc = c.Desc
	}
	if len(c.Labels) > 0 {
		for _, l := range c.Labels {
			s.Labels = s.Labels.Merge(l)
		}
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
	c.mergeDictFromSchema(s)
	return nil
}

// MergeAdditionalData merge relations: comments: to schema.Schema
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
	ni := s.NormalizeTableNames(c.Include)
	ne := s.NormalizeTableNames(c.Exclude)
	for _, t := range s.Tables {
		if len(c.Include) == 0 || contains(c.Include, t.Name) || contains(ni, t.Name) {
			if contains(c.Exclude, t.Name) || contains(ne, t.Name) {
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

func (c *Config) mergeDictFromSchema(s *schema.Schema) {
	if s.Driver != nil && s.Driver.Meta != nil && s.Driver.Meta.Dict != nil {
		c.MergedDict.Merge(s.Driver.Meta.Dict.Dump())
	}
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
				if r.Table.Name != name && r.ParentTable.Name != name {
					childRelations = append(childRelations, r)
				}
			}
			c.ChildRelations = childRelations

			// ParentRelations
			parentRelations := []*schema.Relation{}
			for _, r := range c.ParentRelations {
				if r.Table.Name != name && r.ParentTable.Name != name {
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
		if r.Table.Name != name && r.ParentTable.Name != name {
			relations = append(relations, r)
		}
	}
	s.Relations = relations

	return nil
}

// MaskedDSN return DSN mask password
func (c *Config) MaskedDSN() (string, error) {
	u, err := url.Parse(c.DSN.URL)
	if err != nil {
		return c.DSN.URL, errors.WithStack(err)
	}
	_, pset := u.User.Password()
	if !pset {
		return c.DSN.URL, nil
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
		if len(c.Labels) > 0 {
			for _, l := range c.Labels {
				table.Labels = table.Labels.Merge(l)
			}
		}
		for c, comment := range c.ColumnComments {
			column, err := table.FindColumnByName(c)
			if err != nil {
				return errors.Wrap(err, "failed to add column comment")
			}
			column.Comment = comment
		}
		for i, comment := range c.IndexComments {
			index, err := table.FindIndexByName(i)
			if err != nil {
				return errors.Wrap(err, "failed to add index comment")
			}
			index.Comment = comment
		}
		for c, comment := range c.ConstraintComments {
			constraint, err := table.FindConstraintByName(c)
			if err != nil {
				return errors.Wrap(err, "failed to add constraint comment")
			}
			constraint.Comment = comment
		}
		for t, comment := range c.TriggerComments {
			trigger, err := table.FindTriggerByName(t)
			if err != nil {
				return errors.Wrap(err, "failed to add trigger comment")
			}
			trigger.Comment = comment
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
