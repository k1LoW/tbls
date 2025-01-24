package schema

import (
	"fmt"
	"strings"
)

type Cardinality string

const (
	ZeroOrOne          Cardinality = "zero_or_one"
	ExactlyOne         Cardinality = "exactly_one"
	ZeroOrMore         Cardinality = "zero_or_more"
	OneOrMore          Cardinality = "one_or_more"
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
	"zero_or_one":  ZeroOrOne,
	"exactly_one":  ExactlyOne,
	"zero_or_more": ZeroOrMore,
	"one_or_more":  OneOrMore,
	"one_or_zero":  ZeroOrOne,
	"zero_or_many": ZeroOrMore,
	"one_or_many":  OneOrMore,
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
