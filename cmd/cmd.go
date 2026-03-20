package cmd

import (
	"os"

	"github.com/k1LoW/tbls/cmdutil"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/datasource"
	"github.com/k1LoW/tbls/schema"
)

func getSchemaFromJSONorDSN(c *config.Config) (*schema.Schema, error) {
	if _, err := os.Stat(c.SchemaFilePath()); err == nil {
		cmdutil.Verbosef("Reading schema from existing JSON file: %s", c.SchemaFilePath())
		s, err := datasource.AnalyzeJSONStringOrFile(c.SchemaFilePath())
		if err != nil {
			return nil, err
		}
		if err := c.FilterTables(s); err != nil {
			return nil, err
		}
		cmdutil.Verbosef("Schema loaded from JSON: %d tables", len(s.Tables))
		return s, nil
	}
	cmdutil.Verbosef("Analyzing database schema from DSN")
	s, err := datasource.Analyze(c.DSN)
	if err != nil {
		return nil, err
	}
	cmdutil.Verbosef("Schema analysis complete: %d tables found", len(s.Tables))
	if err := c.ModifySchema(s); err != nil {
		return nil, err
	}
	return s, nil
}
