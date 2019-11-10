package config

import (
	"fmt"
	"sort"

	"github.com/k1LoW/tbls/schema"
)

// Lint is the struct for lint config
type Lint struct {
	RequireTableComment    RequireTableComment    `yaml:"requireTableComment"`
	RequireColumnComment   RequireColumnComment   `yaml:"requireColumnComment"`
	UnrelatedTable         UnrelatedTable         `yaml:"unrelatedTable"`
	ColumnCount            ColumnCount            `yaml:"columnCount"`
	RequireColumns         RequireColumns         `yaml:"requireColumns"`
	DuplicateRelations     DuplicateRelations     `yaml:"duplicateRelations"`
	RequireForeignKeyIndex RequireForeignKeyIndex `yaml:"requireForeignKeyIndex"`
}

// RuleWarn is struct of Rule error
type RuleWarn struct {
	Target  string
	Message string
}

// Rule is interfece of `tbls lint` cop
type Rule interface {
	IsEnabled() bool
	Check(schema *schema.Schema, exclude []string) []RuleWarn
}

// RequireTableComment checks table comment
type RequireTableComment struct {
	Enabled bool     `yaml:"enabled"`
	Exclude []string `yaml:"exclude"`
}

// IsEnabled return Rule is enabled or not
func (r RequireTableComment) IsEnabled() bool {
	return r.Enabled
}

// Check table comment
func (r RequireTableComment) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return warns
	}
	msg := "table comment required."

	for _, t := range s.Tables {
		if contains(exclude, t.Name) {
			continue
		}
		if contains(r.Exclude, t.Name) {
			continue
		}
		if t.Comment == "" {
			warns = append(warns, RuleWarn{
				Target:  t.Name,
				Message: msg,
			})
		}
	}
	return warns
}

// RequireColumnComment checks column comment
type RequireColumnComment struct {
	Enabled        bool     `yaml:"enabled"`
	Exclude        []string `yaml:"exclude"`
	ExcludedTables []string `yaml:"excludedTables"`
}

// IsEnabled return Rule is enabled or not
func (r RequireColumnComment) IsEnabled() bool {
	return r.Enabled
}

// Check column comment
func (r RequireColumnComment) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return warns
	}
	msg := "column comment required."

	for _, t := range s.Tables {
		if contains(exclude, t.Name) {
			continue
		}
		if contains(r.ExcludedTables, t.Name) {
			continue
		}
		for _, c := range t.Columns {
			target := fmt.Sprintf("%s.%s", t.Name, c.Name)
			if contains(r.Exclude, c.Name) || contains(r.Exclude, target) {
				continue
			}
			if c.Comment == "" {
				warns = append(warns, RuleWarn{
					Target:  target,
					Message: msg,
				})
			}
		}
	}
	return warns
}

// UnrelatedTable checks isolated table
type UnrelatedTable struct {
	Enabled bool     `yaml:"enabled"`
	Exclude []string `yaml:"exclude"`
}

// IsEnabled return Rule is enabled or not
func (r UnrelatedTable) IsEnabled() bool {
	return r.Enabled
}

// Check table relation
func (r UnrelatedTable) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return warns
	}
	msgFmt := "unrelated (isolated) table exists. [%d]"

	tableMap := map[string]*schema.Table{}
	for _, t := range s.Tables {
		if contains(exclude, t.Name) {
			continue
		}
		if contains(r.Exclude, t.Name) {
			continue
		}
		tableMap[t.Name] = t
	}
	for _, rl := range s.Relations {
		delete(tableMap, rl.Table.Name)
		delete(tableMap, rl.ParentTable.Name)
	}
	if len(tableMap) > 0 {
		warns = append(warns, RuleWarn{
			Target:  s.Name,
			Message: fmt.Sprintf(msgFmt, len(tableMap)),
		})
	}
	return warns
}

// ColumnCount checks table column count
type ColumnCount struct {
	Enabled bool     `yaml:"enabled"`
	Max     int      `yaml:"max"`
	Exclude []string `yaml:"exclude"`
}

// IsEnabled return Rule is enabled or not
func (r ColumnCount) IsEnabled() bool {
	return r.Enabled
}

// Check table column count
func (r ColumnCount) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return warns
	}
	msgFmt := "too many columns. [%d/%d]"

	for _, t := range s.Tables {
		if contains(exclude, t.Name) {
			continue
		}
		if contains(r.Exclude, t.Name) {
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

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}

// RequireColumns checks if the table has specified columns
type RequireColumns struct {
	Enabled bool                   `yaml:"enabled"`
	Columns []RequireColumnsColumn `yaml:"columns"`
}

// RequireColumnsColumn is required column
type RequireColumnsColumn struct {
	Name    string   `yaml:"name"`
	Exclude []string `yaml:"exclude"`
}

// IsEnabled return Rule is enabled or not
func (r RequireColumns) IsEnabled() bool {
	return r.Enabled
}

// Check the existence of a table columns
func (r RequireColumns) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return warns
	}
	for _, t := range s.Tables {
		if contains(exclude, t.Name) {
			continue
		}
		for _, cc := range r.Columns {
			excluded := false
			for _, tt := range cc.Exclude {
				if t.Name == tt {
					excluded = true
				}
			}
			if excluded {
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

// DuplicateRelations checks duplicate table relations
type DuplicateRelations struct {
	Enabled bool `yaml:"enabled"`
}

// IsEnabled return Rule is enabled or not
func (r DuplicateRelations) IsEnabled() bool {
	return r.Enabled
}

// Check duplicate table relations
func (r DuplicateRelations) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return warns
	}
	relations := make(map[[4]string]bool)
	msgFmt := "duplicate relations. [%s -> %s]"

	for _, r := range s.Relations {
		if contains(exclude, r.Table.Name) {
			continue
		}
		if contains(exclude, r.ParentTable.Name) {
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

// RequireForeignKeyIndex checks if the foreign key columns have an index
type RequireForeignKeyIndex struct {
	Enabled bool     `yaml:"enabled"`
	Exclude []string `yaml:"exclude"`
}

// IsEnabled return Rule is enabled or not
func (r RequireForeignKeyIndex) IsEnabled() bool {
	return r.Enabled
}

// Check if the foreign key columns have an index
func (r RequireForeignKeyIndex) Check(s *schema.Schema, exclude []string) []RuleWarn {
	warns := []RuleWarn{}
	if !r.IsEnabled() {
		return warns
	}
	msgFmt := "foreign key columns do not have an index. [%s]"

	for _, t := range s.Tables {
		if contains(exclude, t.Name) {
			continue
		}
		for _, c := range t.Constraints {
			for _, c1 := range c.Columns {
				target := fmt.Sprintf("%s.%s", t.Name, c1)
				if contains(r.Exclude, c1) || contains(r.Exclude, target) {
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
