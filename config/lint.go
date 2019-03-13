package config

import (
	"fmt"

	"github.com/k1LoW/tbls/schema"
)

// Lint is the struct for lint config
type Lint struct {
	RequireTableComment  RequireTableComment  `yaml:"requireTableComment"`
	RequireColumnComment RequireColumnComment `yaml:"requireColumnComment"`
	UnrelatedTable       UnrelatedTable       `yaml:"unrelatedTable"`
	ColumnCount          ColumnCount          `yaml:"columnCount"`
}

// RuleWarn is struct of Rule error
type RuleWarn struct {
	Target  string
	Message string
}

// Rule is interfece of `tbls lint` cop
type Rule interface {
	IsEnabled() bool
	Check(*schema.Schema) []RuleWarn
}

// RequireTableComment check table comment
type RequireTableComment struct {
	Enabled bool     `yaml:"enabled"`
	Exclude []string `yaml:"exclude"`
}

// IsEnabled return Rule is enabled or not
func (r RequireTableComment) IsEnabled() bool {
	return r.Enabled
}

// Check table comment
func (r RequireTableComment) Check(s *schema.Schema) []RuleWarn {
	msg := "table comment required."
	warns := []RuleWarn{}
	for _, t := range s.Tables {
		if !contains(r.Exclude, t.Name) && t.Comment == "" {
			warns = append(warns, RuleWarn{
				Target:  t.Name,
				Message: msg,
			})
		}
	}
	return warns
}

// RequireColumnComment check column comment
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
func (r RequireColumnComment) Check(s *schema.Schema) []RuleWarn {
	msg := "column comment required."
	warns := []RuleWarn{}
	for _, t := range s.Tables {
		if contains(r.ExcludedTables, t.Name) {
			continue
		}
		for _, c := range t.Columns {
			if !contains(r.Exclude, c.Name) && c.Comment == "" {
				warns = append(warns, RuleWarn{
					Target:  fmt.Sprintf("%s.%s", t.Name, c.Name),
					Message: msg,
				})
			}
		}
	}
	return warns
}

// UnrelatedTable check isolated table
type UnrelatedTable struct {
	Enabled bool     `yaml:"enabled"`
	Exclude []string `yaml:"exclude"`
}

// IsEnabled return Rule is enabled or not
func (r UnrelatedTable) IsEnabled() bool {
	return r.Enabled
}

// Check table relation
func (r UnrelatedTable) Check(s *schema.Schema) []RuleWarn {
	msgFmt := "unrelated (isolated) table exists. [%d]"
	warns := []RuleWarn{}
	tableMap := map[string]*schema.Table{}
	for _, t := range s.Tables {
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

// ColumnCount check table column count
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
func (r ColumnCount) Check(s *schema.Schema) []RuleWarn {
	msgFmt := "too many columns. [%d/%d]"
	warns := []RuleWarn{}
	for _, t := range s.Tables {
		if !contains(r.Exclude, t.Name) && len(t.Columns) > r.Max {
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
