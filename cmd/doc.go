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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/datasource"
	"github.com/k1LoW/tbls/output/dot"
	"github.com/k1LoW/tbls/output/md"
	"github.com/k1LoW/tbls/schema"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// withoutER
var withoutER bool

// docCmd represents the doc command
var docCmd = &cobra.Command{
	Use:   "doc [DSN] [DOC_PATH]",
	Short: "document a database",
	Long:  `'tbls doc' analyzes a database and generate document in GitHub Friendly Markdown format.`,
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

		options, err := loadDocArgs(args)
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

		if !c.ER.Skip {
			_, err = exec.Command("which", "dot").Output()
			if err == nil {
				err := withDot(s, c, force)
				if err != nil {
					printError(err)
					os.Exit(1)
				}
			}
		}

		err = md.Output(s, c, force)

		if err != nil {
			printError(err)
			os.Exit(1)
		}
	},
}

func withDot(s *schema.Schema, c *config.Config, force bool) error {
	erFormat := c.ER.Format
	outputPath := c.DocPath
	fullPath, err := filepath.Abs(outputPath)
	if err != nil {
		return errors.WithStack(err)
	}

	if !force && outputErExists(s, fullPath) {
		return errors.New("output ER diagram files already exists")
	}

	_ = os.MkdirAll(fullPath, 0755)

	dotFormatOption := fmt.Sprintf("-T%s", erFormat)
	erFileName := fmt.Sprintf("schema.%s", erFormat)

	fmt.Printf("%s\n", filepath.Join(outputPath, erFileName))
	tmpfile, _ := ioutil.TempFile("", "tblstmp")
	cmd := exec.Command("dot", dotFormatOption, "-o", filepath.Join(fullPath, erFileName), tmpfile.Name())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	dot := new(dot.Dot)

	err = dot.OutputSchema(tmpfile, s)
	if err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return err
	}
	err = tmpfile.Close()
	if err != nil {
		os.Remove(tmpfile.Name())
		return errors.WithStack(err)
	}
	err = cmd.Run()
	if err != nil {
		os.Remove(tmpfile.Name())
		return errors.WithStack(errors.Wrap(err, stderr.String()))
	}
	os.Remove(tmpfile.Name())

	// tables
	for _, t := range s.Tables {
		erFileName := fmt.Sprintf("%s.%s", t.Name, erFormat)

		fmt.Printf("%s\n", filepath.Join(outputPath, erFileName))
		tmpfile, _ := ioutil.TempFile("", "tblstmp")
		c := exec.Command("dot", dotFormatOption, "-o", filepath.Join(fullPath, erFileName), tmpfile.Name())
		var stderr bytes.Buffer
		c.Stderr = &stderr

		err = dot.OutputTable(tmpfile, t)
		if err != nil {
			tmpfile.Close()
			os.Remove(tmpfile.Name())
			return err
		}
		err = tmpfile.Close()
		if err != nil {
			os.Remove(tmpfile.Name())
			return errors.WithStack(err)
		}
		err = c.Run()
		if err != nil {
			os.Remove(tmpfile.Name())
			return errors.WithStack(errors.Wrap(err, stderr.String()))
		}
		os.Remove(tmpfile.Name())
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
	if len(args) == 2 {
		options = append(options, config.DSN(args[0]))
		options = append(options, config.DocPath(args[1]))
	}
	if len(args) == 1 {
		options = append(options, config.DSN(args[0]))
	}
	return options, nil
}

func outputErExists(s *schema.Schema, path string) bool {
	// schema.png
	erFileName := fmt.Sprintf("schema.%s", erFormat)
	if _, err := os.Lstat(filepath.Join(path, erFileName)); err == nil {
		return true
	}
	// tables
	for _, t := range s.Tables {
		erFileName := fmt.Sprintf("%s.%s", t.Name, erFormat)
		if _, err := os.Lstat(filepath.Join(path, erFileName)); err == nil {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(docCmd)
	docCmd.Flags().BoolVarP(&force, "force", "f", false, "force")
	docCmd.Flags().BoolVarP(&sort, "sort", "", false, "sort")
	docCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	docCmd.Flags().StringVarP(&erFormat, "er-format", "t", "", fmt.Sprintf("ER diagrams output format [png, svg, jpg, ...]. default: %s", config.DefaultERFormat))
	docCmd.Flags().BoolVarP(&withoutER, "without-er", "", false, "no generate ER diagrams")
	docCmd.Flags().BoolVarP(&adjust, "adjust-table", "j", false, "adjust column width of table")
	docCmd.Flags().StringVarP(&additionalDataPath, "add", "a", "", "additional schema data path (deprecated, use `config`)")
}
