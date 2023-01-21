package schema

import (
	"fmt"
	"strings"
)

type Cardinality string

const (
	ZeroOrOne          Cardinality = "Zero or one"
	ExactlyOne         Cardinality = "Exactly one"
	ZeroOrMore         Cardinality = "Zero or more"
	OneOrMore          Cardinality = "One or more"
	UnknownCardinality Cardinality = ""
)

var cardinalityAliases = map[string]Cardinality{
	"zero or one":  ZeroOrOne,
	"exactly one":  ExactlyOne,
	"zero or more": ZeroOrMore,
	"one or more":  OneOrMore,
	"one or zero":  ZeroOrOne,
	"zero or many": ZeroOrMore,
	"one or many":  OneOrMore,
	"many(0)":      ZeroOrMore,
	"many(1)":      OneOrMore,
	"0+":           ZeroOrMore,
	"1+":           OneOrMore,
	"*":            ZeroOrMore,
	"0..*":         ZeroOrMore,
	"0..1":         ZeroOrOne,
	"1..*":         OneOrMore,
	"1":            ExactlyOne,
	"":             UnknownCardinality,
}

func (c Cardinality) String() string {
	return string(c)
}

func ToCardinality(in string) (Cardinality, error) {
	c, ok := cardinalityAliases[strings.ToLower(in)]
	if !ok {
		return "", fmt.Errorf("invalid cardinality: %s", in)
	}
	return c, nil
}
