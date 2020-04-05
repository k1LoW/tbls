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
	"io"
	"os"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/datasource"
	"github.com/k1LoW/tbls/output"
	tbls_config "github.com/k1LoW/tbls/output/config"
	"github.com/k1LoW/tbls/output/dot"
	"github.com/k1LoW/tbls/output/gviz"
	"github.com/k1LoW/tbls/output/json"
	"github.com/k1LoW/tbls/output/md"
	"github.com/k1LoW/tbls/output/plantuml"
	"github.com/k1LoW/tbls/output/xlsx"
	"github.com/k1LoW/tbls/output/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	format    string
	outPath   string
	tableName string
	distance  int
)

// outCmd represents the doc command
var outCmd = &cobra.Command{
	Use:   "out [DSN]",
	Short: "analyzes a database and output",
	Long:  `'tbls out' analyzes a database and output.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.New()
		if err != nil {
			printError(err)
			os.Exit(1)
		}

		if configPath == "" && additionalDataPath != "" {
			fmt.Println("Warning: `--add` option is deprecated. Use `--config`")
			configPath = additionalDataPath
		}

		options, err := loadOutArgs(args)
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

		var o output.Output

		switch format {
		case "json":
			o = new(json.JSON)
		case "yaml":
			o = new(yaml.YAML)
		case "dot":
			o = dot.New(c)
		case "md":
			o = md.New(c, false)
		case "xlsx":
			o = xlsx.New(c)
		case "plantuml":
			o = plantuml.New(c)
		case "png", "svg", "jpg":
			o = gviz.New(c, format)
		case "config":
			o = tbls_config.New(c)
		default:
			printError(fmt.Errorf("unsupported format '%s'", format))
			os.Exit(1)
		}

		var wr io.Writer
		if outPath != "" {
			file, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) // #nosec
			if err != nil {
				printError(err)
				os.Exit(1)
			}
			defer func() {
				err := file.Close()
				if err != nil {
					printError(err)
					os.Exit(1)
				}
			}()
			wr = file
		} else {
			wr = os.Stdout
		}

		if tableName == "" {
			err = o.OutputSchema(wr, s)
		} else {
			t, ferr := s.FindTableByName(tableName)
			if ferr != nil {
				printError(ferr)
				os.Exit(1)
			}
			err = o.OutputTable(wr, t)
		}

		if err != nil {
			printError(err)
			os.Exit(1)
		}
	},
}

func loadOutArgs(args []string) ([]config.Option, error) {
	options := []config.Option{}
	if len(args) > 1 {
		return options, errors.WithStack(errors.New("too many arguments"))
	}
	if sort {
		options = append(options, config.Sort(sort))
	}
	options = append(options, config.Distance(distance))

	if len(args) == 1 {
		options = append(options, config.DSNURL(args[0]))
	}
	return options, nil
}

func init() {
	rootCmd.AddCommand(outCmd)
	outCmd.Flags().BoolVarP(&sort, "sort", "", false, "sort")
	outCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	outCmd.Flags().StringVarP(&format, "format", "t", "json", "output format")
	outCmd.Flags().StringVarP(&outPath, "out", "o", "", "output file path")
	outCmd.Flags().StringVar(&tableName, "table", "", "table name")
	outCmd.Flags().IntVarP(&distance, "distance", "", config.DefaultDistance, "distance between tables that display associations in the ER")
	outCmd.Flags().StringVarP(&additionalDataPath, "add", "a", "", "additional schema data path (deprecated, use `config`)")
}
