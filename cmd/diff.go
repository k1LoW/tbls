// Copyright © 2018 Ken'ichiro Oyama <k1lowxb@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/SouhlInc/tbls/cmdutil"
	"github.com/SouhlInc/tbls/config"
	"github.com/SouhlInc/tbls/datasource"
	"github.com/SouhlInc/tbls/output/md"
	"github.com/SouhlInc/tbls/schema"
	"github.com/spf13/cobra"
)

// diffCmd represents the diff command.
var diffCmd = &cobra.Command{
	Use:   "diff [DSN] [DSN_OR_DOC_PATH]",
	Short: "diff database and ( document or database )",
	Long:  `'tbls diff' shows the difference between database schema and ( generated document or other database schema ).`,
	Args:  cobra.MaximumNArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		if allow, err := cmdutil.IsAllowedToExecute(when); !allow || err != nil {
			if err != nil {
				return err
			}
			return nil
		}

		c, err := config.New()
		if err != nil {
			return err
		}

		c2, err := config.New()
		if err != nil {
			return err
		}

		var (
			s       *schema.Schema
			s2      *schema.Schema
			docPath string
			diff    string
		)

		options := loadDiffOpts()

		switch len(args) {
		case 2:
			if _, err := os.Lstat(args[1]); err == nil {
				// a:dsn and b:path
				if err := c.Load(configPath, append(options, config.DSNURL(args[0]))...); err != nil {
					return err
				}
				c2 = nil
				docPath = args[1]
			} else if !strings.Contains(args[1], "://") {
				// a:dsn and b:path
				if err := c.Load(configPath, append(options, config.DSNURL(args[0]))...); err != nil {
					return err
				}
				c2 = nil
				docPath = args[1]
			} else {
				// a:dsn and b:dsn
				if err := c.Load(configPath, append(options, config.DSNURL(args[0]))...); err != nil {
					return err
				}
				if err := c2.Load(configPath, append(options, config.DSNURL(args[1]))...); err != nil {
					return err
				}
				docPath = ""
			}
		case 1:
			if err := c.Load(configPath); err != nil {
				return err
			}
			if _, err := os.Lstat(args[0]); err == nil {
				// a:path and b:dsn in config
				c2 = nil
				docPath = args[0]
			} else {
				// a:dsn in config and b:dsn
				if err := c2.Load(configPath, append(options, config.DSNURL(args[0]))...); err != nil {
					return err
				}
				docPath = ""
			}
		case 0:
			// a:path in config and b:dsn in config
			if err := c.Load(configPath); err != nil {
				return err
			}
			c2 = nil
			docPath = ""
		}

		s, err = datasource.Analyze(c.DSN)
		if err != nil {
			return err
		}
		if err := c.ModifySchema(s); err != nil {
			return err
		}

		if c2 != nil {
			s2, err = datasource.Analyze(c2.DSN)
			if err != nil {
				return err
			}
			if err := c2.ModifySchema(s2); err != nil {
				return err
			}
		}

		switch {
		case docPath != "":
			diff, err = md.DiffSchemaAndDocs(docPath, s, c)
		case s2 != nil:
			diff, err = md.DiffSchemas(s, s2, c, c2)
		default:
			diff, err = md.DiffSchemaAndDocs(c.DocPath, s, c)
		}
		if err != nil {
			return err
		}
		fmt.Print(diff)
		if diff != "" {
			os.Exit(1)
		}

		return nil
	},
}

func loadDiffOpts() []config.Option {
	options := []config.Option{}
	if adjust {
		options = append(options, config.Adjust(adjust))
	}
	if sort {
		options = append(options, config.Sort(sort))
	}
	options = append(options, config.ERFormat(erFormat))
	return options
}

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().BoolVarP(&sort, "sort", "", false, "sort")
	diffCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	diffCmd.Flags().StringVarP(&erFormat, "er-format", "t", "", fmt.Sprintf("ER diagrams output format (png, svg, jpg, ...). default: %s", config.DefaultERFormat))
	diffCmd.Flags().BoolVarP(&adjust, "adjust-table", "j", false, "adjust column width of table")
	diffCmd.Flags().StringVarP(&when, "when", "", "", "command execute condition")
	if err := diffCmd.MarkZshCompPositionalArgumentFile(2); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
