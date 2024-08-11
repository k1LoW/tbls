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

	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/cmdutil"
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

		options, err := loadLintArgs(args)
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

		l := reflect.Indirect(reflect.ValueOf(c.Lint))
		t := l.Type()

		ruleWarns := []config.RuleWarn{}
		for i := 0; i < t.NumField(); i++ {
			r := l.Field(i)
			v, ok := r.Interface().(config.Rule)
			if !ok {
				return fmt.Errorf("invalid rule: %v", r.Interface())
			}
			ruleWarns = append(ruleWarns, v.Check(s, s.NormalizeTableNames(c.LintExclude))...)
		}
		if len(ruleWarns) > 0 {
			for _, warn := range ruleWarns {
				fmt.Printf("%s%s\n", color.Cyan(warn.Target), color.White(fmt.Sprintf(": %s", warn.Message), color.B))
			}
			fmt.Println(color.White(fmt.Sprintf("\n%d detected", len(ruleWarns)), color.B))
			os.Exit(1)
		}

		return nil
	},
}

func loadLintArgs(args []string) ([]config.Option, error) {
	options := []config.Option{}
	if len(args) > 2 {
		return options, errors.WithStack(errors.New("too many arguments"))
	}
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
	rootCmd.AddCommand(lintCmd)
	lintCmd.Flags().StringVarP(&dsn, "dsn", "", "", "data source name")
	lintCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	lintCmd.Flags().StringVarP(&when, "when", "", "", "command execute condition")
	err := lintCmd.MarkZshCompPositionalArgumentFile(2)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
