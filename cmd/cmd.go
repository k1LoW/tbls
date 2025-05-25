package cmd

import (
	"os"

	"github.com/SouhlInc/tbls/config"
	"github.com/SouhlInc/tbls/datasource"
	"github.com/SouhlInc/tbls/schema"
)

func getSchemaFromJSONorDSN(c *config.Config) (*schema.Schema, error) {
	if _, err := os.Stat(c.SchemaFilePath()); err == nil {
		s, err := datasource.AnalyzeJSONStringOrFile(c.SchemaFilePath())
		if err != nil {
			return nil, err
		}
		if err := c.FilterTables(s); err != nil {
			return nil, err
		}
		return s, nil
	}
	s, err := datasource.Analyze(c.DSN)
	if err != nil {
		return nil, err
	}
	if err := c.ModifySchema(s); err != nil {
		return nil, err
	}
	return s, nil
}
