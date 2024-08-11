package schema

import (
	"fmt"
	"strings"

	"github.com/k1LoW/errors"
	"github.com/minio/pkg/wildcard"
	"github.com/samber/lo"
)

type FilterOption struct {
	Include       []string
	Exclude       []string
	IncludeLabels []string
	Distance      int
}

func (s *Schema) Filter(opt *FilterOption) (err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	_, excludes, err := s.SepareteTablesThatAreIncludedOrNot(opt)
	if err != nil {
		return err
	}
	for _, t := range excludes {
		err := excludeTableFromSchema(t.Name, s)
		if err != nil {
			return fmt.Errorf("failed to filter table '%s': %w", t.Name, err)
		}
	}

	return nil
}

func (s *Schema) SepareteTablesThatAreIncludedOrNot(opt *FilterOption) (_ []*Table, _ []*Table, err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	i := append(opt.Include, s.NormalizeTableNames(opt.Include)...)
	e := append(opt.Exclude, s.NormalizeTableNames(opt.Exclude)...)

	includes := []*Table{}
	excludes := []*Table{}
	for _, t := range s.Tables {
		li, mi := matchLength(i, t.Name)
		le, me := matchLength(e, t.Name)
		ml := matchTableOrColumnLabels(opt.IncludeLabels, t)
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
		case len(opt.Include) == 0 && len(opt.IncludeLabels) == 0:
			if me {
				excludes = append(excludes, t)
				continue
			}
			includes = append(includes, t)
		default:
			excludes = append(excludes, t)
		}
	}

	includes2 := []*Table{}
	for _, t := range includes {
		includes2 = append(includes2, t)
		ts, _, err := t.CollectTablesAndRelations(opt.Distance, true)
		if err != nil {
			return nil, nil, err
		}
		for _, tt := range ts {
			if !lo.ContainsBy(includes, func(t *Table) bool {
				return tt.Name == t.Name
			}) {
				includes2 = append(includes2, tt)
			}
		}
	}

	excludes2 := []*Table{}
	for _, t := range excludes {
		if lo.ContainsBy(includes2, func(tt *Table) bool {
			return tt.Name == t.Name
		}) {
			continue
		}
		excludes2 = append(excludes2, t)
	}

	// assert
	if len(s.Tables) != len(includes2)+len(excludes2) {
		return nil, nil, fmt.Errorf("failed to separate tables. expected: %d, actual: %d", len(s.Tables), len(includes2)+len(excludes2))
	}

	return includes2, excludes2, nil
}

func excludeTableFromSchema(name string, s *Schema) error {
	// Tables
	tables := []*Table{}
	for _, t := range s.Tables {
		if t.Name != name {
			tables = append(tables, t)
		}
		for _, c := range t.Columns {
			// ChildRelations
			childRelations := []*Relation{}
			for _, r := range c.ChildRelations {
				if r.Table.Name != name && r.ParentTable.Name != name {
					childRelations = append(childRelations, r)
				}
			}
			c.ChildRelations = childRelations

			// ParentRelations
			parentRelations := []*Relation{}
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
	relations := []*Relation{}
	for _, r := range s.Relations {
		if r.Table.Name != name && r.ParentTable.Name != name {
			relations = append(relations, r)
		}
	}
	s.Relations = relations

	return nil
}

func matchTableOrColumnLabels(il []string, t *Table) bool {
	if matchLabels(il, t.Labels) {
		return true
	}
	for _, c := range t.Columns {
		if matchLabels(il, c.Labels) {
			return true
		}
	}
	return false
}

func matchLabels(il []string, l Labels) bool {
	for _, ll := range l {
		for _, ill := range il {
			if wildcard.MatchSimple(ill, ll.Name) {
				return true
			}
		}
	}
	return false
}

func matchLength(s []string, e string) (int, bool) {
	for _, v := range s {
		if wildcard.MatchSimple(v, e) {
			return len(strings.ReplaceAll(v, "*", "")), true
		}
	}
	return 0, false
}
