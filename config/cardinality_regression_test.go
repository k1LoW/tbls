package config

import (
	"testing"

	"github.com/k1LoW/tbls/schema"
)

func TestModifySchemaParentCardinality(t *testing.T) {
	for _, tt := range []struct {
		name       string
		nullable   []bool
		def        string
		parentName string
		want       schema.Cardinality
	}{
		{"single_required", []bool{false}, "FOREIGN KEY (a) REFERENCES parents(a)", "parents", schema.ExactlyOne},
		{"single_nullable", []bool{true}, "FOREIGN KEY (a) REFERENCES parents(a)", "parents", schema.ZeroOrOne},
		{"simple_required", []bool{false, false}, "FOREIGN KEY (a, b) REFERENCES parents(a, b)", "parents", schema.ExactlyOne},
		{"simple_first_nullable", []bool{true, false}, "FOREIGN KEY (a, b) REFERENCES parents(a, b)", "parents", schema.ZeroOrOne},
		{"simple_second_nullable", []bool{false, true}, "FOREIGN KEY (a, b) REFERENCES parents(a, b)", "parents", schema.ZeroOrOne},
		{"simple_all_nullable", []bool{true, true}, "FOREIGN KEY (a, b) REFERENCES parents(a, b)", "parents", schema.ZeroOrOne},
		{"explicit_simple_mixed", []bool{true, false}, "FOREIGN KEY (a, b) REFERENCES parents(a, b) MATCH SIMPLE", "parents", schema.ZeroOrOne},
		{"full_required", []bool{false, false}, "FOREIGN KEY (a, b) REFERENCES parents(a, b) MATCH FULL", "parents", schema.ExactlyOne},
		{"full_first_nullable", []bool{true, false}, "FOREIGN KEY (a, b) REFERENCES parents(a, b) MATCH FULL", "parents", schema.ExactlyOne},
		{"full_second_nullable", []bool{false, true}, "FOREIGN KEY (a, b) REFERENCES parents(a, b) MATCH FULL", "parents", schema.ExactlyOne},
		{"full_all_nullable", []bool{true, true}, "FOREIGN KEY (a, b) REFERENCES parents(a, b) MATCH FULL", "parents", schema.ZeroOrOne},
		{"quoted_parent_name", []bool{true, false}, `FOREIGN KEY (a, b) REFERENCES "MATCH FULL"(a, b)`, "MATCH FULL", schema.ZeroOrOne},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := parentCardinalitySchema(tt.nullable, tt.parentName, tt.def)
			c, err := New()
			if err != nil {
				t.Fatal(err)
			}
			if err := c.ModifySchema(s); err != nil {
				t.Fatal(err)
			}
			if got := s.Relations[0].ParentCardinality; got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestModifySchemaParentCardinalityOverride(t *testing.T) {
	s := parentCardinalitySchema([]bool{true, false}, "parents", "FOREIGN KEY (a, b) REFERENCES parents(a, b)")
	c, err := New()
	if err != nil {
		t.Fatal(err)
	}
	c.Relations = []AdditionalRelation{{
		Table:             "children",
		Columns:           []string{"a", "b"},
		ParentTable:       "parents",
		ParentColumns:     []string{"a", "b"},
		ParentCardinality: "exactly one",
		Override:          true,
	}}
	if err := c.ModifySchema(s); err != nil {
		t.Fatal(err)
	}
	if got := s.Relations[0].ParentCardinality; got != schema.ExactlyOne {
		t.Errorf("got %s, want explicit override %s", got, schema.ExactlyOne)
	}
}

func parentCardinalitySchema(nullable []bool, parentName, def string) *schema.Schema {
	child := &schema.Table{Name: "children"}
	parent := &schema.Table{Name: parentName}
	for i, allowsNull := range nullable {
		name := string(rune('a' + i))
		child.Columns = append(child.Columns, &schema.Column{Name: name, Type: "integer", Nullable: allowsNull})
		parent.Columns = append(parent.Columns, &schema.Column{Name: name, Type: "integer"})
	}
	relation := &schema.Relation{
		Table:         child,
		Columns:       child.Columns,
		ParentTable:   parent,
		ParentColumns: parent.Columns,
		Def:           def,
	}
	return &schema.Schema{Tables: []*schema.Table{child, parent}, Relations: []*schema.Relation{relation}}
}
