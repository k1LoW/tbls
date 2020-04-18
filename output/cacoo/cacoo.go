package cacoo

import (
	"io"
	"strconv"

	csv "github.com/gocarina/gocsv"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
)

// Cacoo struct
type Cacoo struct {
	config *config.Config
}

// New return Cacoo
func New(c *config.Config) *Cacoo {
	return &Cacoo{
		config: c,
	}
}

type CSVRow struct {
	DBMS            string `csv:"dbms"`
	TableSchema     string `csv:"table_schema"`
	TableName       string `csv:"table_name"`
	ColumnName      string `csv:"column_name"`
	OrdinalPosition string `csv:"ordinal_position"`
	DataType        string `csv:"data_type"`
	CharMaxLength   string `csv:"character_maximum_length"`
	ConstraintType  string `csv:"constraint_type"`
	RefTableSchema  string `csv:"referenced_table_schema"`
	RefTableName    string `csv:"referenced_table_name"`
	RefColumnName   string `csv:"referenced_column_name"`
}

// OutputSchema output dot format for full relation.
func (cc *Cacoo) OutputSchema(wr io.Writer, s *schema.Schema) error {
	var dbms string
	switch s.Driver.Name {
	case "postgres", "redshift":
		dbms = "postgres"
	default:
		dbms = "mysql"
	}
	rows := []*CSVRow{}
	for _, t := range s.Tables {
		for i, c := range t.Columns {
			ty := c.Type
			max := ""

			constraints := t.FindConstrainsByColumnName(c.Name)

			if len(c.ParentRelations) > 0 {
				for _, ct := range constraints {
					if ct.Type == "FOREIGN KEY" {
						continue
					}
					row := &CSVRow{
						DBMS:            dbms,
						TableSchema:     s.Name,
						TableName:       t.Name,
						ColumnName:      c.Name,
						OrdinalPosition: strconv.Itoa(i + 1),
						DataType:        ty,
						CharMaxLength:   max,
						ConstraintType:  ct.Type,
						RefTableSchema:  "",
						RefTableName:    "",
						RefColumnName:   "",
					}
					rows = append(rows, row)
				}

				for _, r := range c.ParentRelations {
					for _, pc := range r.ParentColumns {
						cType := "FOREIGN KEY"
						// if r.Virtual {
						//   ct = "Virtual Relation"
						// }
						row := &CSVRow{
							DBMS:            dbms,
							TableSchema:     s.Name,
							TableName:       t.Name,
							ColumnName:      c.Name,
							OrdinalPosition: strconv.Itoa(i + 1),
							DataType:        ty,
							CharMaxLength:   max,
							ConstraintType:  cType,
							RefTableSchema:  s.Name,
							RefTableName:    r.ParentTable.Name,
							RefColumnName:   pc.Name,
						}
						rows = append(rows, row)
					}
				}
			} else if len(constraints) > 0 {
				for _, ct := range constraints {
					if ct.Type == "FOREIGN KEY" {
						continue
					}
					row := &CSVRow{
						DBMS:            dbms,
						TableSchema:     s.Name,
						TableName:       t.Name,
						ColumnName:      c.Name,
						OrdinalPosition: strconv.Itoa(i + 1),
						DataType:        ty,
						CharMaxLength:   max,
						ConstraintType:  ct.Type,
						RefTableSchema:  "",
						RefTableName:    "",
						RefColumnName:   "",
					}
					rows = append(rows, row)
				}
			} else {
				row := &CSVRow{
					DBMS:            dbms,
					TableSchema:     s.Name,
					TableName:       t.Name,
					ColumnName:      c.Name,
					OrdinalPosition: strconv.Itoa(i + 1),
					DataType:        ty,
					CharMaxLength:   max,
					ConstraintType:  "",
					RefTableSchema:  "",
					RefTableName:    "",
					RefColumnName:   "",
				}
				rows = append(rows, row)
			}
		}
	}
	err := csv.Marshal(&rows, wr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// OutputTable output dot format for table.
func (cc *Cacoo) OutputTable(wr io.Writer, t *schema.Table) error {
	return errors.New("not implemented")
}
