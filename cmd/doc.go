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
	"github.com/k1LoW/tbls/output/md"
	"github.com/k1LoW/tbls/schema"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// noViz
var noViz bool

// docCmd represents the doc command
var docCmd = &cobra.Command{
	Use:   "doc [DSN] [DOCUMENT_PATH]",
	Short: "document a database",
	Long:  `'tbls doc' analyzes a database and generate document in GitHub Friendly Markdown format.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("Error: %s", "requires two args")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		dsn := args[0]
		outputPath := args[1]
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

		if !noViz {
			_, err = exec.Command("which", "dot").Output()
			if err == nil {
				err := withDot(s, outputPath, force)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}

		err = md.Output(s, outputPath, force)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func withDot(s *schema.Schema, outputPath string, force bool) error {
	fullPath, err := filepath.Abs(outputPath)
	if err != nil {
		return err
	}

	if !force && outputPngExists(s, fullPath) {
		return fmt.Errorf("Error: %s", "output png files already exists.")
	}

	fmt.Printf("%s\n", filepath.Join(outputPath, "schema.png"))
	tmpfile, _ := ioutil.TempFile("", "tblstmp")
	c := exec.Command("dot", "-Tpng", "-o", filepath.Join(fullPath, "schema.png"), tmpfile.Name())

	err = dot.OutputSchema(tmpfile, s)
	if err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return err
	}
	err = tmpfile.Close()
	if err != nil {
		os.Remove(tmpfile.Name())
		return err
	}
	err = c.Run()
	if err != nil {
		os.Remove(tmpfile.Name())
		return err
	}
	os.Remove(tmpfile.Name())

	// tables
	for _, t := range s.Tables {
		fmt.Printf("%s\n", filepath.Join(outputPath, fmt.Sprintf("%s.png", t.Name)))
		tmpfile, _ := ioutil.TempFile("", "tblstmp")
		c := exec.Command("dot", "-Tpng", "-o", filepath.Join(fullPath, fmt.Sprintf("%s.png", t.Name)), tmpfile.Name())
		err = dot.OutputTable(tmpfile, t)
		if err != nil {
			tmpfile.Close()
			os.Remove(tmpfile.Name())
			return err
		}
		err = tmpfile.Close()
		if err != nil {
			os.Remove(tmpfile.Name())
			return err
		}
		err = c.Run()
		if err != nil {
			os.Remove(tmpfile.Name())
			return err
		}
		os.Remove(tmpfile.Name())
	}

	return nil
}

func outputPngExists(s *schema.Schema, path string) bool {
	// schema.png
	if _, err := os.Lstat(filepath.Join(path, "schema.png")); err == nil {
		return true
	}
	// tables
	for _, t := range s.Tables {
		if _, err := os.Lstat(filepath.Join(path, fmt.Sprintf("%s.png", t.Name))); err == nil {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(docCmd)
	docCmd.Flags().BoolVarP(&force, "force", "f", false, "force")
	docCmd.Flags().BoolVarP(&sort, "sort", "", false, "sort")
	docCmd.Flags().BoolVarP(&noViz, "no-viz", "", false, "don't use Graphviz 'dot' command")
	docCmd.Flags().StringVarP(&additionalDataPath, "add", "a", "", "additional schema data path")
}
