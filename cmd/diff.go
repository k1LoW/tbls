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

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/datasource"
	"github.com/k1LoW/tbls/output/md"
	"github.com/spf13/cobra"
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff [DSN] [DOC_PATH]",
	Short: "diff database and document",
	Long:  `'tbls diff' shows the difference between database schema and generated document.`,
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

		err = c.Load(configPath, args)
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

		if sort {
			err = s.Sort()
			if err != nil {
				printError(err)
				os.Exit(1)
			}
		}

		diff, err := md.Diff(s, c.DocPath, adjust, erFormat)
		if err != nil {
			printError(err)
			os.Exit(2)
		}
		fmt.Print(diff)
		if diff != "" {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().BoolVarP(&sort, "sort", "", false, "sort")
	diffCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	diffCmd.Flags().StringVarP(&erFormat, "er-format", "t", "png", "ER diagrams output format [png, svg, jpg, ...]")
	diffCmd.Flags().BoolVarP(&adjust, "adjust-table", "j", false, "adjust column width of table")
	diffCmd.Flags().StringVarP(&additionalDataPath, "add", "a", "", "additional schema data path (deprecated, use `config`)")
}
