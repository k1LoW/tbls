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
	"github.com/k1LoW/tbls/output/dot"
	"github.com/k1LoW/tbls/output/json"
	"github.com/k1LoW/tbls/output/md"
	"github.com/k1LoW/tbls/output/xlsx"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	format    string
	outPath   string
	tableName string
)

// outCmd represents the doc command
var outCmd = &cobra.Command{
	Use:   "out [DSN]",
	Short: "analyzes a database and output",
	Long:  `'tbls out' analyzes a database and output.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.NewConfig()
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

		err = c.MergeAdditionalData(s)
		if err != nil {
			printError(err)
			os.Exit(1)
		}

		if c.Format.Sort {
			err = s.Sort()
			if err != nil {
				printError(err)
				os.Exit(1)
			}
		}

		var o output.Output

		switch format {
		case "json":
			o = new(json.JSON)
		case "dot":
			o = new(dot.Dot)
		case "md":
			o = md.NewMd(false, false, "")
		case "xlsx":
			o = new(xlsx.Xlsx)
		default:
			printError(fmt.Errorf("unsupported format '%s'", format))
			os.Exit(1)
		}

		var wr io.Writer
		if outPath != "" {
			file, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				printError(err)
				os.Exit(1)
			}
			defer file.Close()
			wr = file
		} else {
			wr = os.Stdout
		}

		if tableName == "" {
			err = o.OutputSchema(wr, s)
		} else {
			t, err := s.FindTableByName(tableName)
			if err != nil {
				printError(err)
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
	options = append(options, config.Sort(sort))
	if len(args) == 1 {
		options = append(options, config.DSN(args[0]))
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
	outCmd.Flags().StringVarP(&additionalDataPath, "add", "a", "", "additional schema data path (deprecated, use `config`)")
}
