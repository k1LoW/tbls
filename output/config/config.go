package config

import (
	"fmt"
	"io"

	"github.com/goccy/go-yaml"
	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
)

const NoTableComment = "table comment required."
const NoColumnComment = "column comment required."

// Config struct for `tbls out`
type Config struct {
	config *config.Config
}

// New return Config
func New(c *config.Config) *Config {
	return &Config{
		config: c,
	}
}

func (c *Config) OutputSchema(wr io.Writer, s *schema.Schema) error {
	tableWarns := c.config.Lint.RequireTableComment.Check(s, []string{})
	columnWarns := c.config.Lint.RequireColumnComment.Check(s, []string{})

	for _, table := range s.Tables {
		tableExist := false
		for i := range c.config.Comments {
			if s.NormalizeTableName(table.Name) != s.NormalizeTableName(c.config.Comments[i].Table) {
				continue
			}
			tableExist = true
			for _, column := range table.Columns {
				if _, ok := c.config.Comments[i].ColumnComments[column.Name]; !ok {
					if c.config.Lint.RequireColumnComment.IsEnabled() {
						for _, w := range columnWarns {
							if fmt.Sprintf("%s.%s", table.Name, column.Name) == w.Target {
								if c.config.Comments[i].ColumnComments == nil {
									c.config.Comments[i].ColumnComments = map[string]string{}
								}
								c.config.Comments[i].ColumnComments[column.Name] = NoColumnComment
							}
						}
					} else {
						if c.config.Comments[i].ColumnComments == nil {
							c.config.Comments[i].ColumnComments = map[string]string{}
						}
						c.config.Comments[i].ColumnComments[column.Name] = NoColumnComment
					}
				}
			}
		}
		if !tableExist {
			a := config.AdditionalComment{
				Table:          table.Name,
				ColumnComments: map[string]string{},
			}

			if c.config.Lint.RequireTableComment.IsEnabled() {
				for _, w := range tableWarns {
					if table.Name == w.Target {
						a.TableComment = NoTableComment
					}
				}
			} else {
				a.TableComment = NoColumnComment
			}

			for _, column := range table.Columns {
				if c.config.Lint.RequireColumnComment.IsEnabled() {
					for _, w := range columnWarns {
						if fmt.Sprintf("%s.%s", table.Name, column.Name) == w.Target {
							a.ColumnComments[column.Name] = NoColumnComment
						}
					}
				} else {
					a.ColumnComments[column.Name] = NoColumnComment
				}
			}
			c.config.Comments = append(c.config.Comments, a)
		}
	}

	d := yaml.NewEncoder(wr)
	defer d.Close()
	if err := d.Encode(c.config); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (c *Config) OutputTable(wr io.Writer, t *schema.Table) error {
	return errors.New("not supported")
}
