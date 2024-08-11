package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aquasecurity/go-version/pkg/version"
	"github.com/goccy/go-yaml"
	"github.com/k1LoW/errors"
	"github.com/k1LoW/expand"
	"github.com/k1LoW/tbls/dict"
	"github.com/k1LoW/tbls/schema"
	ver "github.com/k1LoW/tbls/version"
	"github.com/minio/pkg/wildcard"
	"github.com/samber/lo"
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
	Viewpoints             []Viewpoint            `yaml:"viewpoints,omitempty"`
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
	Primary bool `yaml:"primary,omitempty"`
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
	if !lo.Contains(SupportERFormat, c.ER.Format) {
		return fmt.Errorf("unsupported ER format: %s", c.ER.Format)
	}
	for i, v := range c.Viewpoints {
		if v.Name == "" {
			return fmt.Errorf("viewpoints[%d] name is required", i)
		}
		if v.Desc == "" {
			return fmt.Errorf("viewpoints[%d] description is required", i)
		}
		for j, g := range v.Groups {
			if g.Name == "" {
				return fmt.Errorf("viewpoints[%d].groups[%d] name is required", i, j)
			}
			if g.Desc == "" {
				return fmt.Errorf("viewpoints[%d].groups[%d] description is required", i, j)
			}
		}
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
func (c *Config) LoadConfigFile(path string) (err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	if path == "" && os.Getenv("TBLS_DSN") == "" {
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
		return fmt.Errorf("failed to load config file: %w", err)
	}

	buf, err := os.ReadFile(filepath.Clean(fullPath))
	if err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}
	c.Path = filepath.Clean(fullPath)

	return c.LoadConfig(buf)
}

// LoadConfig load config from []byte
func (c *Config) LoadConfig(in []byte) (err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	if err := yaml.Unmarshal(expand.ExpandenvYAMLBytes(in), c); err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
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
	// set Labels
	for _, l := range c.Labels {
		s.Labels = s.Labels.Merge(l)
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

	// set Viewpoints
	// viewpoints should be created using as complete a schema as possible
	for _, v := range c.Viewpoints {
		cs, err := s.CloneWithoutViewpoints()
		if err != nil {
			return err
		}
		if err := cs.Filter(&schema.FilterOption{
			Include:       v.Tables,
			IncludeLabels: v.Labels,
			Distance:      v.Distance,
		}); err != nil {
			return err
		}
		groups := []*schema.ViewpointGroup{}
		tables := lo.Map(cs.Tables, func(t *schema.Table, _ int) string {
			return t.Name
		})
		for _, g := range v.Groups {
			gt, _, err := cs.SepareteTablesThatAreIncludedOrNot(&schema.FilterOption{
				Include:       g.Tables,
				IncludeLabels: g.Labels,
			})
			if err != nil {
				return err
			}
			groups = append(groups, &schema.ViewpointGroup{
				Name:   g.Name,
				Desc:   g.Desc,
				Tables: g.Tables,
				Labels: g.Labels,
				Color:  g.Color,
			})
			left, right := lo.Difference(tables, lo.Map(gt, func(t *schema.Table, _ int) string {
				return t.Name
			}))
			if len(right) > 0 {
				return fmt.Errorf("viewpoint group '%s' has duplicate tables %v", g.Name, right)
			}
			tables = left
		}
		s.Viewpoints = s.Viewpoints.Merge(&schema.Viewpoint{
			Name:     v.Name,
			Desc:     v.Desc,
			Labels:   v.Labels,
			Tables:   v.Tables,
			Distance: v.Distance,
			Groups:   groups,
			Schema:   cs,
		})
	}
	for _, v := range s.Viewpoints {
	L:
		for _, l := range v.Labels {
			for _, t := range s.Tables {
				if t.Labels.Contains(l) {
					continue L
				}
				for _, c := range t.Columns {
					if c.Labels.Contains(l) {
						continue L
					}
				}
			}
			return fmt.Errorf("viewpoint '%s' has unknown label '%s'", v.Name, l)
		}
	}
	for vi, v := range s.Viewpoints {
		// Add viewpoints to table

		for _, t := range v.Tables {
			println(v.Name, t)
			table, err := s.FindTableByName(t)
			if err != nil {
				return err
			}
			table.Viewpoints = append(table.Viewpoints, &schema.TableViewpoint{
				Index: vi,
				Name:  v.Name,
				Desc:  v.Desc,
			})
		}
	}

	return nil
}

// MergeAdditionalData merge relations: comments: to schema.Schema
func (c *Config) MergeAdditionalData(s *schema.Schema) error {
	if err := mergeAdditionalRelations(s, c.Relations); err != nil {
		return err
	}
	if err := mergeAdditionalComments(s, c.Comments); err != nil {
		return err
	}
	return nil
}

// FilterTables filter tables from schema.Schema using include: and exclude: and includeLabels
func (c *Config) FilterTables(s *schema.Schema) error {
	return s.Filter(&schema.FilterOption{
		Include:       c.Include,
		Exclude:       c.Exclude,
		IncludeLabels: c.includeLabels,
		Distance:      c.Distance,
	})
}

func (c *Config) mergeDictFromSchema(s *schema.Schema) {
	if s.Driver != nil && s.Driver.Meta != nil && s.Driver.Meta.Dict != nil {
		c.MergedDict.Merge(s.Driver.Meta.Dict.Dump())
	}
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

	if !c.ER.ShowColumnTypes.Related && !c.ER.ShowColumnTypes.Primary {
		return errors.New("er.showColumnTypes: must be true at least one")
	}

	for _, t := range s.Tables {
		for _, cc := range t.Columns {
			if c.ER.ShowColumnTypes.Related && (cc.ChildRelations != nil || cc.ParentRelations != nil) {
				// related
				cc.HideForER = false
			} else if c.ER.ShowColumnTypes.Primary && cc.PK {
				// primary
				cc.HideForER = false
			} else {
				cc.HideForER = true
				for _, r := range cc.ChildRelations {
					r.HideForER = true
				}
				for _, r := range cc.ParentRelations {
					r.HideForER = true
				}
			}
		}
	}

	return nil
}

func mergeAdditionalRelations(s *schema.Schema, relations []AdditionalRelation) (err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	for _, r := range relations {
		c, err := schema.ToCardinality(r.Cardinality)
		if err != nil {
			return fmt.Errorf("failed to add relation: %w", err)
		}
		pc, err := schema.ToCardinality(r.ParentCardinality)
		if err != nil {
			return fmt.Errorf("failed to add relation: %w", err)
		}
		relation := &schema.Relation{
			Cardinality:       c,
			ParentCardinality: pc,
			Virtual:           true,
		}
		if r.Def != "" {
			relation.Def = r.Def
		} else {
			relation.Def = "Additional Relation"
		}
		relation.Table, err = s.FindTableByName(r.Table)
		if err != nil {
			return fmt.Errorf("failed to add relation: %w", err)
		}
		for _, c := range r.Columns {
			column, err := relation.Table.FindColumnByName(c)
			if err != nil {
				return fmt.Errorf("failed to add relation: %w", err)
			}
			relation.Columns = append(relation.Columns, column)
			column.ParentRelations = append(column.ParentRelations, relation)
		}
		relation.ParentTable, err = s.FindTableByName(r.ParentTable)
		if err != nil {
			return fmt.Errorf("failed to add relation: %w", err)
		}
		for _, c := range r.ParentColumns {
			column, err := relation.ParentTable.FindColumnByName(c)
			if err != nil {
				return fmt.Errorf("failed to add relation: %w", err)
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
					return fmt.Errorf("failed to add relation: %w", err)
				}
				cr.ParentCardinality, err = schema.ToCardinality(r.ParentCardinality)
				if err != nil {
					return fmt.Errorf("failed to add relation: %w", err)
				}
			}
		} else {
			s.Relations = append(s.Relations, relation)
		}
	}
	return nil
}

func mergeAdditionalComments(s *schema.Schema, comments []AdditionalComment) (err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	for _, c := range comments {
		table, err := s.FindTableByName(c.Table)
		if err != nil {
			return fmt.Errorf("failed to add table comment: %w", err)
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
				return fmt.Errorf("failed to add column comment: %w", err)
			}
			column.Comment = comment
		}
		for c, labels := range c.ColumnLabels {
			column, err := table.FindColumnByName(c)
			if err != nil {
				return fmt.Errorf("failed to add column comment: %w", err)
			}
			for _, l := range labels {
				column.Labels = column.Labels.Merge(l)
			}
		}
		for i, comment := range c.IndexComments {
			index, err := table.FindIndexByName(i)
			if err != nil {
				return fmt.Errorf("failed to add index comment: %w", err)
			}
			index.Comment = comment
		}
		for c, comment := range c.ConstraintComments {
			constraint, err := table.FindConstraintByName(c)
			if err != nil {
				return fmt.Errorf("failed to add constraint comment: %w", err)
			}
			constraint.Comment = comment
		}
		for t, comment := range c.TriggerComments {
			trigger, err := table.FindTriggerByName(t)
			if err != nil {
				return fmt.Errorf("failed to add trigger comment: %w", err)
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
		parentTable  *schema.Table
	)

	for _, t := range s.Tables {
		for _, c := range t.Columns {
			relation := &schema.Relation{
				Virtual: true,
				Def:     "Detected Relation",
				Table:   t,
			}

			if parentTable, err = s.FindTableByName(strategy.ParentTableName(c.Name)); err != nil {
				continue
			}

			if parentTable == t {
				continue
			}

			relation.ParentTable = parentTable

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

// detectCardinality detects the cardinality of the relations
// This function should be applied to the completed schema
func detectCardinality(s *schema.Schema) error {
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
					if !lo.Contains(columns, cc) {
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
			// whether the child columns are nullable or not.
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

func match(s []string, e string) bool {
	_, m := matchLength(s, e)
	return m
}
