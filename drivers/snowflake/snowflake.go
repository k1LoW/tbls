package snowflake

import (
	"database/sql"

	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/schema"
	_ "github.com/snowflakedb/gosnowflake"
)

type Snowflake struct {
	db *sql.DB
}

func New(db *sql.DB) *Snowflake {
	return &Snowflake{
		db: db,
	}
}

func (sf *Snowflake) Analyze(s *schema.Schema) error {
	d, err := sf.Info()
	if err != nil {
		return errors.WithStack(err)
	}
	s.Driver = d

	tableRows, err := sf.db.Query(`SELECT table_name, table_type, comment FROM information_schema.tables WHERE table_schema = ? order by table_name`, s.Name)
	if err != nil {
		return errors.WithStack(err)
	}
	defer tableRows.Close()

	tables := []*schema.Table{}

	for tableRows.Next() {
		var (
			tableName string
			tableType string
			comment   sql.NullString
		)
		if err := tableRows.Scan(&tableName, &tableType, &comment); err != nil {
			return errors.WithStack(err)
		}
		table := &schema.Table{
			Name:    tableName,
			Type:    tableType,
			Comment: comment.String,
		}

		var getDDLObjectType string
		if tableType == "BASE TABLE" {
			getDDLObjectType = "table"
		} else if tableType == "VIEW" {
			getDDLObjectType = "view"
		}
		if getDDLObjectType != "" {
			tableDefRows, err := sf.db.Query(`SELECT GET_DDL(?, ?)`, getDDLObjectType, tableName)
			if err != nil {
				return errors.WithStack(err)
			}
			defer tableDefRows.Close()
			for tableDefRows.Next() {
				var tableDef string
				err := tableDefRows.Scan(&tableDef)
				if err != nil {
					return errors.WithStack(err)
				}
				table.Def = tableDef
			}
		}

		// columns, comments
		columnRows, err := sf.db.Query(`select column_name, column_default, is_nullable, data_type, comment
from information_schema.columns
where table_schema = ? and table_name = ? order by ordinal_position`, s.Name, tableName)
		if err != nil {
			return errors.WithStack(err)
		}
		defer columnRows.Close()
		columns := []*schema.Column{}
		for columnRows.Next() {
			var (
				columnName    string
				columnDefault sql.NullString
				isNullable    string
				dataType      string
				columnComment sql.NullString
			)
			err = columnRows.Scan(
				&columnName,
				&columnDefault,
				&isNullable,
				&dataType,
				&columnComment,
			)
			if err != nil {
				return errors.WithStack(err)
			}
			column := &schema.Column{
				Name:     columnName,
				Type:     dataType,
				Nullable: convertColumnNullable(isNullable),
				Default:  columnDefault,
				Comment:  columnComment.String,
			}
			columns = append(columns, column)
		}
		table.Columns = columns

		tables = append(tables, table)
	}
	s.Tables = tables

	return nil
}

func (s *Snowflake) Info() (*schema.Driver, error) {
	var v string
	row := s.db.QueryRow(`SELECT CURRENT_VERSION();`)
	if err := row.Scan(&v); err != nil {
		return nil, err
	}
	return &schema.Driver{
		Name:            "snowflake",
		DatabaseVersion: v,
	}, nil
}

func convertColumnNullable(str string) bool {
	return str != "NO"
}
