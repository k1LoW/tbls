package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aquasecurity/go-version/pkg/version"
	"github.com/goccy/go-yaml"
	"github.com/k1LoW/expand"
	"github.com/k1LoW/tbls/dict"
	"github.com/k1LoW/tbls/schema"
	ver "github.com/k1LoW/tbls/version"
	"github.com/minio/pkg/wildcard"
	"github.com/pkg/errors"
)

const DefaultDocPath = "dbdoc"

var DefaultConfigFilePaths = []string{".tbls.yml", "tbls.yml"}

// DefaultERFormat is the default ER diagram format
const DefaultERFormat = "svg"

var SupportERFormat = []string{"png", "jpg", "svg", "mermaid"}

const SchemaFileName = "schema.json"

// DefaultERDistance is the default distance between tables that display relations in the ER
var DefaultERDistance = 1

// Config is tbls config
type Config struct {
	Name   string   `yaml:"name"`
	Desc   string   `yaml:"desc,omitempty"`
	Labels []string `yaml:"labels,omitempty"`
	DSN    DSN      `yaml:"dsn"`
	// Directory of schema document
	DocPath                string                 `yaml:"docPath"`
	Format                 Format                 `yaml:"format,omitempty"`
	ER                     ER                     `yaml:"er,omitempty"`
	Include                []string               `yaml:"include,omitempty"`
	Exclude                []string               `yaml:"exclude,omitempty"`
	Distance               int                    `yaml:"distance,omitempty"`
	Lint                   Lint                   `yaml:"lint,omitempty"`
	LintExclude            []string               `yaml:"lintExclude,omitempty"`
	Relations              []AdditionalRelation   `yaml:"relations,omitempty"`
	Comments               []AdditionalComment    `yaml:"comments,omitempty"`
	Dict                   dict.Dict              `yaml:"dict,omitempty"`
	Templates              Templates              `yaml:"templates,omitempty"`
	DetectVirtualRelations DetectVirtualRelations `yaml:"detectVirtualRelations,omitempty"`
	BaseUrl                string                 `yaml:"baseUrl,omitempty"`
	RequiredVersion        string                 `yaml:"requiredVersion,omitempty"`
	DisableOutputSchema    bool                   `yaml:"disableOutputSchema,omitempty"`
	MergedDict             dict.Dict              `yaml:"-"`

	// Table labels to be included
	includeLabels []string

	// Path of config file
	Path string `yaml:"-"`
	root string `yaml:"-"`
}

type DSN struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

// Format is document format setting
type Format struct {
	Adjust                   bool     `yaml:"adjust,omitempty"`
	Sort                     bool     `yaml:"sort,omitempty"`
	Number                   bool     `yaml:"number,omitempty"`
	ShowOnlyFirstParagraph   bool     `yaml:"showOnlyFirstParagraph,omitempty"`
	HideColumnsWithoutValues []string `yaml:"hideColumnsWithoutValues,omitempty"`
}

// ER is er setting
type ER struct {
	Skip            bool             `yaml:"skip,omitempty"`
	Format          string           `yaml:"format,omitempty"`
	Comment         bool             `yaml:"comment,omitempty"`
	HideDef         bool             `yaml:"hideDef,omitempty"`
	ShowColumnTypes *ShowColumnTypes `yaml:"showColumnTypes,omitempty"`
	Distance        *int             `yaml:"distance,omitempty"`
	Font            string           `yaml:"font,omitempty"`
}

// ShowColumnTypes is show column setting for ER diagram
type ShowColumnTypes struct {
	Related bool `yaml:"related,omitempty"`
}

// AdditionalRelation is the struct for table relation from yaml
type AdditionalRelation struct {
	Table             string   `yaml:"table"`
	Columns           []string `yaml:"columns"`
	Cardinality       string   `yaml:"cardinality,omitempty"`
	ParentTable       string   `yaml:"parentTable"`
	ParentColumns     []string `yaml:"parentColumns"`
	ParentCardinality string   `yaml:"parentCardinality,omitempty"`
	Def               string   `yaml:"def,omitempty"`
	Override          bool     `yaml:"override,omitempty"`
}

// AdditionalComment is the struct for table relation from yaml
type AdditionalComment struct {
	Table              string              `yaml:"table"`
	TableComment       string              `yaml:"tableComment,omitempty"`
	ColumnComments     map[string]string   `yaml:"columnComments,omitempty"`
	ColumnLabels       map[string][]string `yaml:"columnLabels,omitempty"`
	IndexComments      map[string]string   `yaml:"indexComments,omitempty"`
	ConstraintComments map[string]string   `yaml:"constraintComments,omitempty"`
	TriggerComments    map[string]string   `yaml:"triggerComments,omitempty"`
	Labels             []string            `yaml:"labels,omitempty"`
}

type DetectVirtualRelations struct {
	Enabled  bool   `yaml:"enabled,omitempty"`
	Strategy string `yaml:"strategy,omitempty"`
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

// Distance return Option set Config.Distance
func Distance(distance int) Option {
	return func(c *Config) error {
		c.Distance = distance
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

// Include return Option set Config.Include
func Include(i []string) Option {
	return func(c *Config) error {
		if len(i) > 0 {
			c.Include = i
		}
		return nil
	}
}

// Exclude return Option set Config.Exclude
func Exclude(e []string) Option {
	return func(c *Config) error {
		if len(e) > 0 {
			c.Exclude = e
		}
		return nil
	}
}

// IncludeLabels return Option set Config.includeLabels
func IncludeLabels(l []string) Option {
	return func(c *Config) error {
		if len(l) > 0 {
			c.includeLabels = l
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

// Load config with all method
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

	if err := c.validate(); err != nil {
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
		c.ER.Distance = &DefaultERDistance
	}

	return nil
}

func (c *Config) checkVersion(sv string) error {
	if sv == "dev" {
		return nil
	}
	if c.RequiredVersion == "" {
		return nil
	}
	cons, err := version.NewConstraints(c.RequiredVersion)
	if err != nil {
		return err
	}
	v, err := version.Parse(sv)
	if err != nil {
		return err
	}
	if !cons.Check(v) {
		return fmt.Errorf("the required tbls version for the configuration is '%s'. however, the running tbls version is '%s'", c.RequiredVersion, sv)
	}

	return nil
}

func (c *Config) validate() error {
	if err := c.checkVersion(ver.Version); err != nil {
		return err
	}
	if !contains(SupportERFormat, c.ER.Format) {
		return fmt.Errorf("unsupported ER format: %s", c.ER.Format)
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
	if path == "" && os.Getenv("TBLS_DSN") == "" {
		for _, p := range DefaultConfigFilePaths {
			if f, err := os.Stat(filepath.Join(c.root, p)); err == nil && !f.IsDir() {
				if path != "" {
					return errors.Errorf("duplicate config file [%s, %s]", path, p)
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

	buf, err := os.ReadFile(filepath.Clean(fullPath))
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to load config file")
	}
	c.Path = filepath.Clean(fullPath)

	return c.LoadConfig(buf)
}

// LoadConfig load config from []byte
func (c *Config) LoadConfig(in []byte) error {
	err := yaml.Unmarshal(expand.ExpandenvYAMLBytes(in), c)
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
	if err := detectCardinality(s); err != nil {
		return err
	}
	if err := detectPKFK(s); err != nil {
		return err
	}
	if err := c.MergeAdditionalData(s); err != nil {
		return err
	}
	if err := c.FilterTables(s); err != nil {
		return err
	}
	if c.Format.Sort {
		if err := s.Sort(); err != nil {
			return err
		}
	}
	if c.DetectVirtualRelations.Enabled {
		strategy, err := SelectNamingStrategy(c.DetectVirtualRelations.Strategy)
		if err != nil {
			return err
		}
		mergeDetectedRelations(s, strategy)
	}
	c.mergeDictFromSchema(s)
	if err := detectCardinality(s); err != nil {
		return err
	}
	if err := c.detectShowColumnsForER(s); err != nil {
		return err
	}
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
	i := append(c.Include, s.NormalizeTableNames(c.Include)...)
	e := append(c.Exclude, s.NormalizeTableNames(c.Exclude)...)

	includes := []*schema.Table{}
	excludes := []*schema.Table{}
	for _, t := range s.Tables {
		li, mi := matchLength(i, t.Name)
		le, me := matchLength(e, t.Name)
		ml := matchLabels(c.includeLabels, t.Labels)
		switch {
		case mi:
			if me && li < le {
				excludes = append(excludes, t)
				continue
			}
			includes = append(includes, t)
		case ml:
			if me {
				excludes = append(excludes, t)
				continue
			}
			includes = append(includes, t)
		case len(c.Include) == 0 && len(c.includeLabels) == 0:
			if me {
				excludes = append(excludes, t)
				continue
			}
			includes = append(includes, t)
		default:
			excludes = append(excludes, t)
		}
	}

	collects := []*schema.Table{}
	for _, t := range includes {
		ts, _, err := t.CollectTablesAndRelations(c.Distance, true)
		if err != nil {
			return err
		}
		for _, tt := range ts {
			if !tt.Contains(includes) {
				collects = append(collects, tt)
			}
		}
	}

	for _, t := range excludes {
		if t.Contains(collects) {
			continue
		}
		err := excludeTableFromSchema(t.Name, s)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("failed to filter table '%s'", t.Name))
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

func (c *Config) SchemaFilePath() string {
	return filepath.Join(c.DocPath, SchemaFileName)
}

func (c *Config) NeedToGenerateERImages() bool {
	if c.ER.Skip {
		return false
	}
	if c.ER.Format == "mermaid" {
		return false
	}
	return true
}

func (c *Config) detectShowColumnsForER(s *schema.Schema) error {
	if c.ER.ShowColumnTypes == nil {
		return nil
	}

	if !c.ER.ShowColumnTypes.Related {
		return errors.New("er.showColumnTypes: must be true at least one")
	}

	for _, t := range s.Tables {
		for _, cc := range t.Columns {
			// related
			if c.ER.ShowColumnTypes.Related && cc.ChildRelations == nil && cc.ParentRelations == nil {
				cc.HideForER = true
			}
		}
	}

	return nil
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

		if r.Override {
			cr, err := s.FindRelation(relation.Columns, relation.ParentColumns)
			if err != nil {
				s.Relations = append(s.Relations, relation)
			} else {
				cr.Virtual = true
				cr.Def = r.Def
				cr.Cardinality, err = schema.ToCardinality(r.Cardinality)
				if err != nil {
					return errors.Wrap(err, "failed to add relation")
				}
				cr.ParentCardinality, err = schema.ToCardinality(r.ParentCardinality)
				if err != nil {
					return errors.Wrap(err, "failed to add relation")
				}
			}
		} else {
			s.Relations = append(s.Relations, relation)
		}
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
		for c, labels := range c.ColumnLabels {
			column, err := table.FindColumnByName(c)
			if err != nil {
				return errors.Wrap(err, "failed to add column comment")
			}
			for _, l := range labels {
				column.Labels = column.Labels.Merge(l)
			}
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

func mergeDetectedRelations(s *schema.Schema, strategy *NamingStrategy) {
	var (
		err          error
		parentColumn *schema.Column
	)

	for _, t := range s.Tables {
		for _, c := range t.Columns {
			relation := &schema.Relation{
				Virtual: true,
				Def:     "Detected Relation",
				Table:   t,
			}

			if relation.ParentTable, err = s.FindTableByName(strategy.ParentTableName(c.Name)); err != nil {
				continue
			}
			if parentColumn, err = relation.ParentTable.FindColumnByName(strategy.ParentColumnName(c.Name)); err != nil {
				continue
			}

			relation.Columns = append(relation.Columns, c)
			relation.ParentColumns = append(relation.ParentColumns, parentColumn)

			if _, err := s.FindRelation(relation.Columns, relation.ParentColumns); err == nil {
				// If the relation already exists, do not create a new virtual relation.
				continue
			}

			c.ParentRelations = append(c.ParentRelations, relation)
			parentColumn.ChildRelations = append(parentColumn.ChildRelations, relation)
			s.Relations = append(s.Relations, relation)
		}
	}
}

func matchLength(s []string, e string) (int, bool) {
	for _, v := range s {
		if wildcard.MatchSimple(v, e) {
			return len(strings.ReplaceAll(v, "*", "")), true
		}
	}
	return 0, false
}

func detectCardinality(s *schema.Schema) error {
	// This function should be applied to the completed schema
	for _, r := range s.Relations {
		// child
		if r.Cardinality == schema.UnknownCardinality {
			unique := false
			columns := []string{}
			for _, c := range r.Columns {
				columns = append(columns, c.Name)
			}
		LL:
			for _, c := range r.Table.Constraints {
				if len(columns) != len(c.Columns) {
					continue
				}
				for _, cc := range c.Columns {
					if !contains(columns, cc) {
						continue LL
					}
				}
				if strings.Contains(strings.ToUpper(c.Def), "UNIQUE") || strings.Contains(strings.ToUpper(c.Def), "PRIMARY KEY") {
					unique = true
				}
			}
			if unique {
				r.Cardinality = schema.ZeroOrOne
			} else {
				r.Cardinality = schema.ZeroOrMore
			}
		}

		// parent
		if r.ParentCardinality == schema.UnknownCardinality {
			// whether the child colums are nullable or not.
			nullable := true
			for _, c := range r.Columns {
				if !c.Nullable {
					nullable = false
				}
			}

			if nullable {
				r.ParentCardinality = schema.ZeroOrOne
			} else {
				r.ParentCardinality = schema.ExactlyOne
			}
		}
	}
	return nil
}

func detectPKFK(s *schema.Schema) error {
	for _, t := range s.Tables {
		// PRIMARY KEY
		for _, i := range t.Indexes {
			if !strings.Contains(i.Def, "PRIMARY") {
				continue
			}
			for _, c := range i.Columns {
				column, err := t.FindColumnByName(c)
				if err != nil {
					return err
				}
				column.PK = true
			}
		}
		// Foreign Key (Relations)
		for _, c := range t.Columns {
			if len(c.ParentRelations) > 0 && !c.PK {
				c.FK = true
			}
		}
	}
	return nil
}

func matchLabels(il []string, l schema.Labels) bool {
	for _, ll := range l {
		for _, ill := range il {
			if ll.Name == ill {
				return true
			}
		}
	}
	return false
}

func match(s []string, e string) bool {
	_, m := matchLength(s, e)
	return m
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}
