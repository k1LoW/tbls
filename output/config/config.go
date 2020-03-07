package config

import (
	"fmt"
	"io"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const noTableComment = "table comment required."
const noColumnComment = "column comment required."

// Config struct for `tbls out`
type Config struct {
	config *config.Config
}

// NewConfig return Config
func NewConfig(c *config.Config) *Config {
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
			if s.FullTableName(table.Name) != s.FullTableName(c.config.Comments[i].Table) {
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
								c.config.Comments[i].ColumnComments[column.Name] = noColumnComment
							}
						}
					} else {
						if c.config.Comments[i].ColumnComments == nil {
							c.config.Comments[i].ColumnComments = map[string]string{}
						}
						c.config.Comments[i].ColumnComments[column.Name] = noColumnComment
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
						a.TableComment = noTableComment
					}
				}
			} else {
				a.TableComment = noColumnComment
			}

			for _, column := range table.Columns {
				if c.config.Lint.RequireColumnComment.IsEnabled() {
					for _, w := range columnWarns {
						if fmt.Sprintf("%s.%s", table.Name, column.Name) == w.Target {
							a.ColumnComments[column.Name] = noColumnComment
						}
					}
				} else {
					a.ColumnComments[column.Name] = noColumnComment
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
