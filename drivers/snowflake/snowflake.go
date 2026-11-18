package snowflake

import (
	"database/sql"
	"fmt"

	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/schema"

	// Import the Snowflake driver for side effects (database/sql driver registration).
	_ "github.com/snowflakedb/gosnowflake/v2"
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
		switch tableType {
		case "BASE TABLE":
			getDDLObjectType = "table"
		case "VIEW":
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
		columnRows, err := sf.db.Query(`select column_name, column_default, is_nullable, data_type, comment,
character_maximum_length, numeric_precision, numeric_scale,
is_identity, identity_start, identity_increment
from information_schema.columns
where table_schema = ? and table_name = ? order by ordinal_position`, s.Name, tableName)
		if err != nil {
			return errors.WithStack(err)
		}
		defer columnRows.Close()
		columns := []*schema.Column{}
		for columnRows.Next() {
			var (
				columnName             string
				columnDefault          sql.NullString
				isNullable             string
				dataType               string
				columnComment          sql.NullString
				characterMaximumLength sql.NullInt64
				numericPrecision       sql.NullInt64
				numericScale           sql.NullInt64
				isIdentity             sql.NullString
				identityStart          sql.NullString
				identityIncrement      sql.NullString
			)
			err = columnRows.Scan(
				&columnName,
				&columnDefault,
				&isNullable,
				&dataType,
				&columnComment,
				&characterMaximumLength,
				&numericPrecision,
				&numericScale,
				&isIdentity,
				&identityStart,
				&identityIncrement,
			)
			if err != nil {
				return errors.WithStack(err)
			}
			switch dataType {
			case "NUMBER":
				if numericPrecision.Valid && numericScale.Valid {
					dataType = fmt.Sprintf("%s(%d,%d)", dataType, numericPrecision.Int64, numericScale.Int64)
				}
			case "TEXT":
				if characterMaximumLength.Valid {
					dataType = fmt.Sprintf("%s(%d)", dataType, characterMaximumLength.Int64)
				}
			}
			column := &schema.Column{
				Name:     columnName,
				Type:     dataType,
				Nullable: convertColumnNullable(isNullable),
				Default:  columnDefault,
				Comment:  columnComment.String,
			}
			if isIdentity.Valid && isIdentity.String == "YES" {
				column.ExtraDef = "IDENTITY"
				if identityStart.Valid && identityIncrement.Valid {
					column.ExtraDef = fmt.Sprintf("IDENTITY START %s INCREMENT %s", identityStart.String, identityIncrement.String)
				}
			}
			columns = append(columns, column)
		}
		table.Columns = columns

		tables = append(tables, table)
	}
	s.Tables = tables

	return nil
}

func (sf *Snowflake) Info() (*schema.Driver, error) {
	var v string
	row := sf.db.QueryRow(`SELECT CURRENT_VERSION();`)
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
