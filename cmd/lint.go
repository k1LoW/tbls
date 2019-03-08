// Copyright © 2019 Ken'ichiro Oyama <k1lowxb@gmail.com>
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
	"reflect"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/datasource"
	"github.com/labstack/gommon/color"
	"github.com/spf13/cobra"
)

// lintCmd represents the lint command
var lintCmd = &cobra.Command{
	Use:   "lint [DSN] [DOC_PATH]",
	Short: "check database document",
	Long:  `'tbls lint' check database document.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.NewConfig()
		if err != nil {
			printError(err)
			os.Exit(1)
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

		l := reflect.Indirect(reflect.ValueOf(c.Lint))
		t := l.Type()

		ruleWarns := []config.RuleWarn{}
		for i := 0; i < t.NumField(); i++ {
			var v config.Rule
			r := l.Field(i)
			v = r.Interface().(config.Rule)
			if !v.IsEnabled() {
				continue
			}
			ruleWarns = append(ruleWarns, v.Check(s)...)
		}
		if len(ruleWarns) > 0 {
			for _, warn := range ruleWarns {
				fmt.Println(color.Red(fmt.Sprintf("%s", warn.Message)))
			}
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)
	lintCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
}
