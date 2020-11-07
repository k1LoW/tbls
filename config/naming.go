package config

import (
	"fmt"
	"github.com/gertd/go-pluralize"
	"strings"
)

var (
	pluralizeClient = pluralize.NewClient()
	namingStrategy  = &NamingStrategy{
		ParentTable:  defaultParentTableNamer,
		ParentColumn: defaultParentColumnNamer,
	}
)

// Namer is a function type which is given a string and return a string
type Namer func(string) string

// NamingStrategy represents naming strategies
type NamingStrategy struct {
	ParentTable  Namer
	ParentColumn Namer
}

// SelectNamingStrategy sets the naming strategy
func SelectNamingStrategy(name string) {
	if name == "" {
		return
	}
	switch name {
	// TODO: Add case if added naming strategy
	default:
		fmt.Println("Not exists naming strategy")
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

// ToParentTableName convert string to table name
func ToParentTableName(name string) string {
	return namingStrategy.ParentTableName(name)
}

// ToParentColumnName convert string to column name
func ToParentColumnName(name string) string {
	return namingStrategy.ParentColumnName(name)
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
