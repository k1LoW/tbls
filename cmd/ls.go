/*
Copyright © 2023 Ken'ichiro Oyama <k1lowxb@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/SouhlInc/tbls/config"
	"github.com/SouhlInc/tbls/schema"
	"github.com/minio/pkg/wildcard"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var long bool

// lsCmd represents the ls command.
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list schema resources",
	Long:  `list schema resources.`,
	RunE: func(_ *cobra.Command, args []string) error {
		c, err := config.New()
		if err != nil {
			return err
		}

		options, patterns, err := loadLsArgs(args)
		if err != nil {
			return err
		}

		if err := c.Load(configPath, options...); err != nil {
			return err
		}

		s, err := getSchemaFromJSONorDSN(c)
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		if long {
			table.SetHeader([]string{"NAME", "TYPE", "COMMENT"})
		}
		table.SetAutoFormatHeaders(false)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetNoWhiteSpace(true)
		table.SetTablePadding("\t")
		nlrep := strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ")

		if len(patterns) == 0 {
			for _, t := range s.Tables {
				if long {
					table.Append([]string{t.Name, t.Type, nlrep.Replace(t.Comment)})
				} else {
					table.Append([]string{t.Name})
				}
			}
			table.Render()
			return nil
		}

		matches := []*schema.Table{}
		for _, t := range s.Tables {
			for _, p := range patterns {
				if wildcard.MatchSimple(p, t.Name) {
					matches = append(matches, t)
				}
			}
		}
		if len(matches) == 0 {
			return fmt.Errorf("not found: %v", patterns)
		}
		for _, t := range matches {
			for _, c := range t.Columns {
				if long {
					table.Append([]string{fmt.Sprintf("%s.%s", t.Name, c.Name), c.Type, nlrep.Replace(c.Comment)})
				} else {
					table.Append([]string{fmt.Sprintf("%s.%s", t.Name, c.Name)})
				}
			}
		}
		table.Render()
		return nil
	},
}

func loadLsArgs(args []string) ([]config.Option, []string, error) {
	pattern := args
	options := []config.Option{}
	if dsn != "" {
		options = append(options, config.DSNURL(dsn))
	}
	options = append(options, config.Include(append(tables, includes...)))
	options = append(options, config.Exclude(excludes))
	options = append(options, config.IncludeLabels(labels))
	options = append(options, config.Distance(distance))
	return options, pattern, nil
}

func init() {
	rootCmd.AddCommand(lsCmd)
	lsCmd.Flags().StringVarP(&dsn, "dsn", "", "", "data source name")
	lsCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	lsCmd.Flags().StringSliceVarP(&tables, "table", "", []string{}, "target table (tables to include)")
	lsCmd.Flags().StringSliceVarP(&includes, "include", "", []string{}, "tables to include")
	lsCmd.Flags().StringSliceVarP(&excludes, "exclude", "", []string{}, "tables to exclude")
	lsCmd.Flags().StringSliceVarP(&labels, "label", "", []string{}, "table labels to be included")
	lsCmd.Flags().IntVarP(&distance, "distance", "", 0, "distance between related tables to be displayed")
	lsCmd.Flags().BoolVarP(&long, "long", "l", false, "list in the long format")
}
