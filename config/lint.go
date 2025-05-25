package config

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/SouhlInc/tbls/schema"
)

// Lint is the struct for lint config.
type Lint struct {
	RequireTableComment      RequireTableComment      `yaml:"requireTableComment"`
	RequireColumnComment     RequireColumnComment     `yaml:"requireColumnComment"`
	RequireIndexComment      RequireIndexComment      `yaml:"requireIndexComment"`
	RequireConstraintComment RequireConstraintComment `yaml:"requireConstraintComment"`
	RequireTriggerComment    RequireTriggerComment    `yaml:"requireTriggerComment"`
	RequireTableLabels       RequireTableLabels       `yaml:"requireTableLabels"`
	UnrelatedTable           UnrelatedTable           `yaml:"unrelatedTable"`
	ColumnCount              ColumnCount              `yaml:"columnCount"`
	RequireColumns           RequireColumns           `yaml:"requireColumns"`
	DuplicateRelations       DuplicateRelations       `yaml:"duplicateRelations"`
	RequireForeignKeyIndex   RequireForeignKeyIndex   `yaml:"requireForeignKeyIndex"`
	LabelStyleBigQuery       LabelStyleBigQuery       `yaml:"labelStyleBigQuery"`
	RequireViewpoints        RequireViewpoints        `yaml:"requireViewpoints"`
	// ◆ 追加
	RequireMaskingTypes      RequireMaskingTypes      `yaml:"requireMaskingTypes"`
}

// RuleWarn is struct of Rule error.
type RuleWarn struct {
	Target  string
	Message string
}

// Rule is interfece of `tbls lint` cop.
type Rule interface {
	IsEnabled() bool
	Check(schema *schema.Schema, exclude []string) []RuleWarn
}

// RequireTableComment checks table comment.
type RequireTableComment struct {
	Enabled      bool     `yaml:"enabled"`
	AllOrNothing bool     `yaml:"allOrNothing"`
	Exclude      []string `yaml:"exclude"`
}

// IsEnabled return Rule is enabled or not.
func (r RequireTableComment) IsEnabled() bool {
	return r.Enabled
}

// Check table comment.
func (r RequireTableComment) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return []RuleWarn{}
	}
	msg := "table comment required."

	nt := s.NormalizeTableNames(r.Exclude)
	commented := false

	for _, t := range s.Tables {
		if match(exclude, t.Name) {
			continue
		}
		if match(nt, t.Name) {
			continue
		}
		if t.Comment == "" {
			warns = append(warns, RuleWarn{
				Target:  t.Name,
				Message: msg,
			})
			continue
		}
		commented = true
	}
	if r.AllOrNothing && !commented {
		return []RuleWarn{}
	}
	return warns
}

// RequireColumnComment checks column comment.
type RequireColumnComment struct {
	Enabled       bool     `yaml:"enabled"`
	AllOrNothing  bool     `yaml:"allOrNothing"`
	Exclude       []string `yaml:"exclude"`
	ExcludeTables []string `yaml:"excludeTables"`
}

// IsEnabled return Rule is enabled or not.
func (r RequireColumnComment) IsEnabled() bool {
	return r.Enabled
}

// Check column comment.
func (r RequireColumnComment) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return []RuleWarn{}
	}
	msg := "column comment required."

	nt := s.NormalizeTableNames(r.ExcludeTables)
	commented := false

	for _, t := range s.Tables {
		if match(exclude, t.Name) {
			continue
		}
		if match(nt, t.Name) {
			continue
		}
		for _, c := range t.Columns {
			target := fmt.Sprintf("%s.%s", t.Name, c.Name)
			if match(r.Exclude, c.Name) || match(r.Exclude, target) {
				continue
			}
			if c.Comment == "" {
				warns = append(warns, RuleWarn{
					Target:  target,
					Message: msg,
				})
				continue
			}
			commented = true
		}
	}
	if r.AllOrNothing && !commented {
		return []RuleWarn{}
	}
	return warns
}

// RequireIndexComment checks index comment.
type RequireIndexComment struct {
	Enabled       bool     `yaml:"enabled"`
	AllOrNothing  bool     `yaml:"allOrNothing"`
	Exclude       []string `yaml:"exclude"`
	ExcludeTables []string `yaml:"excludeTables"`
}

// IsEnabled return Rule is enabled or not.
func (r RequireIndexComment) IsEnabled() bool {
	return r.Enabled
}

// Check index comment.
func (r RequireIndexComment) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return []RuleWarn{}
	}
	msg := "index comment required."

	nt := s.NormalizeTableNames(r.ExcludeTables)
	commented := false

	for _, t := range s.Tables {
		if match(exclude, t.Name) {
			continue
		}
		if match(nt, t.Name) {
			continue
		}
		for _, i := range t.Indexes {
			target := fmt.Sprintf("%s.%s", t.Name, i.Name)
			if match(r.Exclude, i.Name) || match(r.Exclude, target) {
				continue
			}
			if i.Comment == "" {
				warns = append(warns, RuleWarn{
					Target:  target,
					Message: msg,
				})
				continue
			}
			commented = true
		}
	}
	if r.AllOrNothing && !commented {
		return []RuleWarn{}
	}
	return warns
}

// RequireConstraintComment checks constraint comment.
type RequireConstraintComment struct {
	Enabled       bool     `yaml:"enabled"`
	AllOrNothing  bool     `yaml:"allOrNothing"`
	Exclude       []string `yaml:"exclude"`
	ExcludeTables []string `yaml:"excludeTables"`
}

// IsEnabled return Rule is enabled or not.
func (r RequireConstraintComment) IsEnabled() bool {
	return r.Enabled
}

// Check constraint comment.
func (r RequireConstraintComment) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return []RuleWarn{}
	}
	msg := "constraint comment required."

	nt := s.NormalizeTableNames(r.ExcludeTables)
	commented := false

	for _, t := range s.Tables {
		if match(exclude, t.Name) {
			continue
		}
		if match(nt, t.Name) {
			continue
		}
		for _, c := range t.Constraints {
			target := fmt.Sprintf("%s.%s", t.Name, c.Name)
			if match(r.Exclude, c.Name) || match(r.Exclude, target) {
				continue
			}
			if c.Comment == "" {
				warns = append(warns, RuleWarn{
					Target:  target,
					Message: msg,
				})
				continue
			}
			commented = true
		}
	}
	if r.AllOrNothing && !commented {
		return []RuleWarn{}
	}
	return warns
}

// RequireTriggerComment checks trigger comment.
type RequireTriggerComment struct {
	Enabled       bool     `yaml:"enabled"`
	AllOrNothing  bool     `yaml:"allOrNothing"`
	Exclude       []string `yaml:"exclude"`
	ExcludeTables []string `yaml:"excludeTables"`
}

// IsEnabled return Rule is enabled or not.
func (r RequireTriggerComment) IsEnabled() bool {
	return r.Enabled
}

// Check trigger comment.
func (r RequireTriggerComment) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return []RuleWarn{}
	}
	msg := "trigger comment required."

	nt := s.NormalizeTableNames(r.ExcludeTables)
	commented := false

	for _, t := range s.Tables {
		if match(exclude, t.Name) {
			continue
		}
		if match(nt, t.Name) {
			continue
		}
		for _, trig := range t.Triggers {
			target := fmt.Sprintf("%s.%s", t.Name, trig.Name)
			if match(r.Exclude, trig.Name) || match(r.Exclude, target) {
				continue
			}
			if trig.Comment == "" {
				warns = append(warns, RuleWarn{
					Target:  target,
					Message: msg,
				})
				continue
			}
			commented = true
		}
	}
	if r.AllOrNothing && !commented {
		return []RuleWarn{}
	}
	return warns
}

// RequireTableLabels checks table labels.
type RequireTableLabels struct {
	Enabled      bool     `yaml:"enabled"`
	AllOrNothing bool     `yaml:"allOrNothing"`
	Exclude      []string `yaml:"exclude"`
}

// IsEnabled return Rule is enabled or not.
func (r RequireTableLabels) IsEnabled() bool {
	return r.Enabled
}

// Check table labels.
func (r RequireTableLabels) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return []RuleWarn{}
	}
	msg := "table labels required."
	labeled := false

	nt := s.NormalizeTableNames(r.Exclude)

	for _, t := range s.Tables {
		if match(exclude, t.Name) {
			continue
		}
		if match(nt, t.Name) {
			continue
		}
		if len(t.Labels) == 0 {
			target := t.Name
			warns = append(warns, RuleWarn{
				Target:  target,
				Message: msg,
			})
			continue
		}
		labeled = true
	}

	if r.AllOrNothing && !labeled {
		return []RuleWarn{}
	}
	return warns
}

// UnrelatedTable checks isolated table.
type UnrelatedTable struct {
	Enabled      bool     `yaml:"enabled"`
	AllOrNothing bool     `yaml:"allOrNothing"`
	Exclude      []string `yaml:"exclude"`
}

// IsEnabled return Rule is enabled or not.
func (r UnrelatedTable) IsEnabled() bool {
	return r.Enabled
}

// Check table relation.
func (r UnrelatedTable) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return []RuleWarn{}
	}
	msgFmt := "unrelated (isolated) table exists. %s"

	nt := s.NormalizeTableNames(r.Exclude)
	related := false
	ut := map[string]*schema.Table{}
	for _, t := range s.Tables {
		if match(exclude, t.Name) {
			continue
		}
		if match(nt, t.Name) {
			continue
		}
		ut[t.Name] = t
	}
	before := len(ut)
	for _, rl := range s.Relations {
		delete(ut, rl.Table.Name)
		delete(ut, rl.ParentTable.Name)
	}
	after := len(ut)
	if before != after {
		related = true
	}
	if len(ut) > 0 {
		us := []string{}
		for _, t := range ut {
			us = append(us, t.Name)
		}
		warns = append(warns, RuleWarn{
			Target:  s.Name,
			Message: fmt.Sprintf(msgFmt, us),
		})
	}
	if r.AllOrNothing && !related {
		return []RuleWarn{}
	}
	return warns
}

// ColumnCount checks table column count.
type ColumnCount struct {
	Enabled bool     `yaml:"enabled"`
	Max     int      `yaml:"max"`
	Exclude []string `yaml:"exclude"`
}

// IsEnabled return Rule is enabled or not.
func (r ColumnCount) IsEnabled() bool {
	return r.Enabled
}

// Check table column count.
func (r ColumnCount) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return warns
	}
	msgFmt := "too many columns. [%d/%d]"

	nt := s.NormalizeTableNames(r.Exclude)
	for _, t := range s.Tables {
		if match(exclude, t.Name) {
			continue
		}
		if match(nt, t.Name) {
			continue
		}
		if len(t.Columns) > r.Max {
			warns = append(warns, RuleWarn{
				Target:  t.Name,
				Message: fmt.Sprintf(msgFmt, len(t.Columns), r.Max),
			})
		}
	}
	return warns
}

// RequireColumns checks if the table has specified columns.
type RequireColumns struct {
	Enabled bool                   `yaml:"enabled"`
	Columns []RequireColumnsColumn `yaml:"columns"`
}

// RequireColumnsColumn is required column.
type RequireColumnsColumn struct {
	Name    string   `yaml:"name"`
	Exclude []string `yaml:"exclude"`
}

// IsEnabled return Rule is enabled or not.
func (r RequireColumns) IsEnabled() bool {
	return r.Enabled
}

// Check the existence of a table columns.
func (r RequireColumns) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return warns
	}
	for _, t := range s.Tables {
		if match(exclude, t.Name) {
			continue
		}
		for _, cc := range r.Columns {
			exclude := false
			if match(cc.Exclude, t.Name) {
				exclude = true
			}
			if exclude {
				continue
			}
			exists := false
			msgFmt := "column '%s' required."
			for _, c := range t.Columns {
				if c.Name == cc.Name {
					exists = true
				}
			}
			if !exists {
				warns = append(warns, RuleWarn{
					Target:  t.Name,
					Message: fmt.Sprintf(msgFmt, cc.Name),
				})
			}
		}
	}
	return warns
}

// DuplicateRelations checks duplicate table relations.
type DuplicateRelations struct {
	Enabled bool `yaml:"enabled"`
}

// IsEnabled return Rule is enabled or not.
func (r DuplicateRelations) IsEnabled() bool {
	return r.Enabled
}

// Check duplicate table relations.
func (r DuplicateRelations) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return warns
	}
	relations := make(map[[4]string]bool)
	msgFmt := "duplicate relations. [%s -> %s]"

	for _, r := range s.Relations {
		if match(exclude, r.Table.Name) {
			continue
		}
		if match(exclude, r.ParentTable.Name) {
			continue
		}
		columns := []string{}
		parentColumns := []string{}
		for _, c := range r.Columns {
			columns = append(columns, c.Name)
		}
		sort.SliceStable(columns, func(i, j int) bool { return columns[i] < columns[j] })
		for _, c := range r.ParentColumns {
			parentColumns = append(parentColumns, c.Name)
		}
		sort.SliceStable(parentColumns, func(i, j int) bool { return parentColumns[i] < parentColumns[j] })

		key := [4]string{r.Table.Name, r.ParentTable.Name, fmt.Sprintf("%v", columns), fmt.Sprintf("%v", parentColumns)}
		if _, dup := relations[key]; dup {
			warns = append(warns, RuleWarn{
				Target:  r.Table.Name,
				Message: fmt.Sprintf(msgFmt, r.Table.Name, r.ParentTable.Name),
			})
		}
		relations[key] = true
	}

	return warns
}

// RequireForeignKeyIndex checks if the foreign key columns have an index.
type RequireForeignKeyIndex struct {
	Enabled bool     `yaml:"enabled"`
	Exclude []string `yaml:"exclude"`
}

// IsEnabled return Rule is enabled or not.
func (r RequireForeignKeyIndex) IsEnabled() bool {
	return r.Enabled
}

// Check if the foreign key columns have an index.
func (r RequireForeignKeyIndex) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return warns
	}
	msgFmt := "foreign key columns do not have an index. [%s]"

	for _, t := range s.Tables {
		if match(exclude, t.Name) {
			continue
		}
		for _, c := range t.Constraints {
			if c.Type != schema.TypeFK {
				continue
			}
			for _, c1 := range c.Columns {
				target := fmt.Sprintf("%s.%s", t.Name, c1)
				if match(r.Exclude, c1) || match(r.Exclude, target) {
					continue
				}
				exist := false
				for _, i := range t.Indexes {
					for _, c2 := range i.Columns {
						if c1 == c2 {
							exist = true
						}
					}
				}
				if !exist {
					warns = append(warns, RuleWarn{
						Target:  target,
						Message: fmt.Sprintf(msgFmt, t.Name),
					})
				}
			}
		}
	}

	return warns
}

// LabelStyleBigQuery checks if labels are in BigQuery style ( https://cloud.google.com/resource-manager/docs/creating-managing-labels#requirements ).
type LabelStyleBigQuery struct {
	Enabled bool     `yaml:"enabled"`
	Exclude []string `yaml:"exclude"`
}

// IsEnabled return Rule is enabled or not.
func (r LabelStyleBigQuery) IsEnabled() bool {
	return r.Enabled
}

// Check if labels are in BigQuery style.
func (r LabelStyleBigQuery) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return warns
	}
	msgFmtSchema := "required to be in BigQuery `key:value` style. [label `%s` in database `%s`]"
	msgFmt := "required to be in BigQuery `key:value` style. [label `%s` in table `%s`]"

	for _, l := range s.Labels {
		if !checkLabelStyleBigQuery(l.Name) {
			target := fmt.Sprintf("%s.Labels.%s", s.Name, l.Name)
			warns = append(warns, RuleWarn{
				Target:  target,
				Message: fmt.Sprintf(msgFmtSchema, l.Name, s.Name),
			})
		}
	}

	nt := s.NormalizeTableNames(r.Exclude)
	for _, t := range s.Tables {
		if match(exclude, t.Name) {
			continue
		}
		if match(nt, t.Name) {
			continue
		}
		for _, l := range t.Labels {
			if !checkLabelStyleBigQuery(l.Name) {
				target := fmt.Sprintf("%s.Labels.%s", t.Name, l.Name)
				warns = append(warns, RuleWarn{
					Target:  target,
					Message: fmt.Sprintf(msgFmt, l.Name, t.Name),
				})
			}
		}
	}

	return warns
}

var labelStyleBigQueryKeyRe = regexp.MustCompile(`^[^A-Z0-9 !"#$%&'()*+,-./:;<=>?@\[\\\]^_\{|}~` + "`" + `][^A-Z !"#$%&'()*+,./:;<=>?@\[\\\]^\{|}~` + "`" + `]*$`)
var labelStyleBigQueryValueRe = regexp.MustCompile(`^[^A-Z !"#$%&'()*+,./:;<=>?@\[\\\]^\{|}~` + "`" + `]*$`)

func checkLabelStyleBigQuery(label string) bool {
	if strings.Count(label, ":") != 1 {
		return false
	}
	kv := strings.Split(label, ":")
	k := kv[0]
	v := kv[1]
	if len(k) == 0 || len(k) > 63 {
		return false
	}
	if len(v) > 63 {
		return false
	}
	if !labelStyleBigQueryKeyRe.MatchString(k) {
		return false
	}
	if !labelStyleBigQueryValueRe.MatchString(v) {
		return false
	}
	return true
}

// RequireViewpoints checks if the table is included in any viewpoints.
type RequireViewpoints struct {
	Enabled bool     `yaml:"enabled"`
	Exclude []string `yaml:"exclude"`
}

// IsEnabled return Rule is enabled or not.
func (r RequireViewpoints) IsEnabled() bool {
	return r.Enabled
}

// Check if the table is included in any viewpoints.
func (r RequireViewpoints) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return warns
	}
	msgFmt := "table `%s` is not included in any viewpoints."

T:
	for _, t := range s.Tables {
		if match(exclude, t.Name) {
			continue
		}

		if match(r.Exclude, t.Name) {
			continue
		}

		for _, v := range s.Viewpoints {
			for _, vt := range v.Tables {
				if vt == t.Name {
					continue T
				}
			}
		}
		warns = append(warns, RuleWarn{
			Target:  t.Name,
			Message: fmt.Sprintf(msgFmt, t.Name),
		})
	}

	return warns
}

// ◆ マスキング種別必須ルール ----------------------------------------------
type RequireMaskingTypes struct {
	Enabled       bool     `yaml:"enabled"`
	AllOrNothing  bool     `yaml:"allOrNothing"`
	Exclude       []string `yaml:"exclude"`
	ExcludeTables []string `yaml:"excludeTables"`
}

func (r RequireMaskingTypes) IsEnabled() bool { return r.Enabled }

var allowedMaskingTypes = map[string]struct{}{
	"salon_id": {},
	"random":   {},
	"none":     {},
}

func (r RequireMaskingTypes) Check(s *schema.Schema, exclude []string) []RuleWarn {
	if !r.IsEnabled() {
		return nil
	}
	warns := []RuleWarn{}
	msgMissing := "maskingTypes required."
	msgInvalid := "invalid maskingType '%s'. (allowed: salon_id, random, none)"

	nt := s.NormalizeTableNames(r.ExcludeTables)
	exists := false

	for _, t := range s.Tables {
		if match(exclude, t.Name) || match(nt, t.Name) {
			continue
		}
		for _, c := range t.Columns {
			target := fmt.Sprintf("%s.%s", t.Name, c.Name)
			if match(r.Exclude, c.Name) || match(r.Exclude, target) {
				continue
			}
			if c.MaskingType == "" {
				warns = append(warns, RuleWarn{Target: target, Message: msgMissing})
				continue
			}
			if _, ok := allowedMaskingTypes[c.MaskingType]; !ok {
				warns = append(warns, RuleWarn{
					Target:  target,
					Message: fmt.Sprintf(msgInvalid, c.MaskingType),
				})
				continue
			}
			exists = true
		}
	}
	if r.AllOrNothing && !exists {
		return nil
	}
	return warns
}
