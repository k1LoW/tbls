package config

import (
	"fmt"
	"strings"

	"github.com/gertd/go-pluralize"
)

var (
	pluralizeClient = pluralize.NewClient()
)

// Namer is a function type which is given a string and return a string
type Namer func(string) string

// NamingStrategy represents naming strategies
type NamingStrategy struct {
	ParentTable  Namer
	ParentColumn Namer
}

// SelectNamingStrategy sets the naming strategy
func SelectNamingStrategy(name string) (*NamingStrategy, error) {
	switch name {
	case "", "default":
		// default
		return &NamingStrategy{
			ParentTable:  defaultParentTableNamer,
			ParentColumn: defaultParentColumnNamer,
		}, nil

	case "singularTableName":
		return &NamingStrategy{
			ParentTable:  singularTableParentTableNamer,
			ParentColumn: singularTableParentColumnNamer,
		}, nil

	default:
		return nil, fmt.Errorf("Naming strategy does not exist. strategy: %s\n", name)
	}
}

// ParentTableName alters the given name by Table
func (ns *NamingStrategy) ParentTableName(name string) string {
	return ns.ParentTable(name)
}

// ParentColumnName alters the given name by Column
func (ns *NamingStrategy) ParentColumnName(name string) string {
	return ns.ParentColumn(name)
}

func defaultParentTableNamer(name string) string {
	index := strings.LastIndex(name, "_")

	if index == -1 || name[index+1:] != "id" {
		return ""
	}
	return pluralizeClient.Plural(name[:index])
}

func defaultParentColumnNamer(name string) string {
	return "id"
}

func singularTableParentTableNamer(name string) string {
	index := strings.LastIndex(name, "_")

	if index == -1 || name[index+1:] != "id" {
		return ""
	}
	return pluralizeClient.Singular(name[:index])
}

func singularTableParentColumnNamer(name string) string {
	return "id"
}
