/*
Copyright © 2020 Ken'ichiro Oyama <k1lowxb@gmail.com>

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
	"encoding/json"
	"fmt"
	"os"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/coverage"
	"github.com/k1LoW/tbls/datasource"
	"github.com/labstack/gommon/color"
	"github.com/mattn/go-runewidth"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var cformat string

// coverageCmd represents the coverage command
var coverageCmd = &cobra.Command{
	Use:   "coverage [DSN]",
	Short: "measure document coverage",
	Long:  `'tbls coverage' measure document coverage.`,
	Run: func(cmd *cobra.Command, args []string) {
		if allow, err := isAllowedToExecute(when); !allow || err != nil {
			if err != nil {
				printError(err)
				os.Exit(1)
			}
			return
		}

		c, err := config.New()
		if err != nil {
			printError(err)
			os.Exit(1)
		}

		options, err := loadCoverageArgs(args)
		if err != nil {
			printError(err)
			os.Exit(1)
		}

		err = c.Load(configPath, options...)
		if err != nil {
			printError(err)
			os.Exit(1)
		}

		s, err := datasource.Analyze(c.DSN)
		if err != nil {
			printError(err)
			os.Exit(1)
		}

		err = c.ModifySchema(s)
		if err != nil {
			printError(err)
			os.Exit(1)
		}

		cover := coverage.Measure(s)

		max := runewidth.StringWidth("All tables")
		for _, t := range cover.Tables {
			l := runewidth.StringWidth(t.Name)
			if l+1 > max {
				max = l + 1
			}
		}

		switch cformat {
		case "json":
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			err := encoder.Encode(cover)
			if err != nil {
				printError(err)
				os.Exit(1)
			}
		default:
			fmtName := fmt.Sprintf("%%-%ds", max)
			fmt.Printf("%s  %s\n", color.White(fmt.Sprintf(fmtName, "Table"), color.B), color.White("Coverage", color.B))
			fmt.Printf("%s  %g%%\n", fmt.Sprintf(fmtName, "All tables"), cover.Coverage)
			for _, t := range cover.Tables {
				fmt.Printf(" %s %g%%\n", fmt.Sprintf(fmtName, t.Name), t.Coverage)
			}
		}
	},
}

func loadCoverageArgs(args []string) ([]config.Option, error) {
	options := []config.Option{}
	if len(args) > 1 {
		return options, errors.WithStack(errors.New("too many arguments"))
	}
	if len(args) == 1 {
		options = append(options, config.DSNURL(args[0]))
	}
	return options, nil
}

func init() {
	rootCmd.AddCommand(coverageCmd)
	coverageCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	coverageCmd.Flags().StringVarP(&cformat, "format", "t", "", "output format")
	convergeCmd.Flags().StringVarP(&when, "when", "", "", "command execute condition")
}
