package schema

import (
	"database/sql"
)

type Index struct {
	Name string
	Def  string
}

type Column struct {
	Name    string
	Type    string
	NotNull bool
	Default sql.NullString
	Comment string
}

type Table struct {
	Name    string
	Type    string
	Comment string
	Columns []*Column
	Indexes []*Index
}

type Schema struct {
	Name   string
	Tables []*Table
}
