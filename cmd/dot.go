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
	"github.com/k1LoW/tbls/db"
	"github.com/k1LoW/tbls/output/dot"
	"github.com/spf13/cobra"
	"os"
)

// tableName
var tableName string

// dotCmd represents the doc command
var dotCmd = &cobra.Command{
	Use:   "dot [DSN]",
	Short: "generate dot file",
	Long:  `'tbls dot' analyzes a database and generate dot file.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Error: %s", "requires one arg")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		dsn := args[0]
		s, err := db.Analyze(dsn)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if additionalDataPath != "" {
			err = s.LoadAdditionalData(additionalDataPath)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		if sort {
			err = s.Sort()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		if tableName == "" {
			err = dot.OutputSchema(os.Stdout, s)
		} else {
			t, err := s.FindTableByName(tableName)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = dot.OutputTable(os.Stdout, t)
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(dotCmd)
	dotCmd.Flags().StringVarP(&additionalDataPath, "add", "a", "", "additional schema data path")
	dotCmd.Flags().StringVarP(&tableName, "table", "t", "", "table name")
}
