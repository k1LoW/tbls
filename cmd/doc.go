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
	"path/filepath"
	"strings"

	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/cmdutil"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/datasource"
	"github.com/k1LoW/tbls/output/gviz"
	"github.com/k1LoW/tbls/output/json"
	"github.com/k1LoW/tbls/output/md"
	"github.com/k1LoW/tbls/schema"
	"github.com/spf13/cobra"
)

var (
	withoutER bool
	rmDist    bool
)

// docCmd represents the doc command
var docCmd = &cobra.Command{
	Use:   "doc [DSN] [DOC_PATH]",
	Short: "document a database",
	Long:  `'tbls doc' analyzes a database and generate document in GitHub Friendly Markdown format.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		options, err := loadDocArgs(args)
		if err != nil {
			return err
		}

		if err := c.Load(configPath, options...); err != nil {
			return err
		}

		s, err := datasource.Analyze(c.DSN)
		if err != nil {
			return err
		}

		if err := c.ModifySchema(s); err != nil {
			return err
		}

		if rmDist && c.DocPath != "" {
			if _, err := os.Lstat(c.DocPath); err == nil {
				docs, err := os.ReadDir(c.DocPath)
				if err != nil {
					return errors.WithStack(err)
				}
				for _, f := range docs {
					if err := os.RemoveAll(filepath.Join(c.DocPath, f.Name())); err != nil {
						return errors.WithStack(err)
					}
				}
			}
		}

		if c.NeedToGenerateERImages() {
			if err := gviz.Output(s, c, force); err != nil {
				return err
			}
		}

		if err := md.Output(s, c, force); err != nil {
			return err
		}

		// output schema.json
		if !c.DisableOutputSchema {
			if err := withSchemaFile(s, c); err != nil {
				return err
			}
		}

		return nil
	},
}

func withSchemaFile(s *schema.Schema, c *config.Config) (e error) {
	sf, err := os.Create(c.SchemaFilePath())
	if err != nil {
		return err
	}
	defer func() {
		_ = sf.Close()
	}()
	fmt.Printf("%s\n", c.SchemaFilePath())
	j := json.New(true)
	if err := j.OutputSchema(sf, s); err != nil {
		return err
	}
	return nil
}

func loadDocArgs(args []string) ([]config.Option, error) {
	options := []config.Option{}
	if len(args) > 2 {
		return options, errors.WithStack(errors.New("too many arguments"))
	}
	if adjust {
		options = append(options, config.Adjust(adjust))
	}
	if sort {
		options = append(options, config.Sort(sort))
	}
	options = append(options, config.ERFormat(erFormat))
	if withoutER {
		options = append(options, config.ERSkip(withoutER))
	}
	options = append(options, config.BaseUrl(baseUrl))
	options = append(options, config.Include(append(tables, includes...)))
	options = append(options, config.Exclude(excludes))
	options = append(options, config.IncludeLabels(labels))
	if len(args) == 2 {
		options = append(options, config.DSNURL(args[0]))
		options = append(options, config.DocPath(args[1]))
	}
	if len(args) == 1 {
		options = append(options, config.DSNURL(args[0]))
	}
	if dsn != "" {
		options = append(options, config.DSNURL(dsn))
	}
	return options, nil
}

func init() {
	rootCmd.AddCommand(docCmd)
	docCmd.Flags().StringVarP(&dsn, "dsn", "", "", "data source name")
	docCmd.Flags().BoolVarP(&force, "force", "f", false, "force")
	docCmd.Flags().BoolVarP(&sort, "sort", "", false, "sort")
	docCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	docCmd.Flags().StringVarP(&erFormat, "er-format", "t", "", fmt.Sprintf("ER diagrams output format (%s). default: %s", strings.Join(config.SupportERFormat, ", "), config.DefaultERFormat))
	docCmd.Flags().BoolVarP(&withoutER, "without-er", "", false, "no generate ER diagrams")
	docCmd.Flags().BoolVarP(&adjust, "adjust-table", "j", false, "adjust column width of table")
	docCmd.Flags().StringVarP(&when, "when", "", "", "command execute condition")
	docCmd.Flags().StringVarP(&baseUrl, "base-url", "b", "", "base url for links")
	docCmd.Flags().BoolVarP(&rmDist, "rm-dist", "", false, "remove files in docPath before generating documents")
	docCmd.Flags().StringSliceVarP(&tables, "table", "", []string{}, "target table (tables to include)")
	docCmd.Flags().StringSliceVarP(&includes, "include", "", []string{}, "tables to include")
	docCmd.Flags().StringSliceVarP(&excludes, "exclude", "", []string{}, "tables to exclude")
	docCmd.Flags().StringSliceVarP(&labels, "label", "", []string{}, "table labels to be included")

	if err := docCmd.MarkZshCompPositionalArgumentFile(2); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
