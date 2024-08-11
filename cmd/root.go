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
	"os/exec"
	"path/filepath"
	sortpkg "sort"
	"strconv"
	"strings"

	"github.com/k1LoW/errors"
	"github.com/k1LoW/tbls/cmdutil"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/output/json"
	"github.com/k1LoW/tbls/version"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// adjust is a flag on whether to adjust the notation width of the table
var adjust bool

// force is a flag on whether to force generate
var force bool

// sort is a flag on whether to sort tables, columns, and more
var sort bool

// configPath is a config file path
var configPath string

// erFormat is a option that ER diagram file format
var erFormat string

// when is a option that command execute condition
var when string

// base url for links
var baseUrl string

// tables to include
var includes []string
var tables []string

// tables to excludes
var excludes []string

// table labels to be included
var labels []string

// dsn
var dsn string

const rootUsageTemplate = `Usage:{{if .Runnable}}{{if ne .UseLine "tbls [flags]" }}
  {{.UseLine}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

var subCmds = []string{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:                "tbls",
	Short:              "tbls is a CI-Friendly tool for document a database, written in Go.",
	Long:               `tbls is a CI-Friendly tool for document a database, written in Go.`,
	SilenceErrors:      true,
	SilenceUsage:       true,
	Args:               cobra.ArbitraryArgs,
	DisableFlagParsing: true,
	ValidArgsFunction:  genValidArgsFunc("tbls"),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, args := cmdutil.PickOption(args, []string{"-c", "--config"})
		when, args := cmdutil.PickOption(args, []string{"--when"})
		dsn, args := cmdutil.PickOption(args, []string{"--dsn"})
		if allow, err := cmdutil.IsAllowedToExecute(when); !allow || err != nil {
			if err != nil {
				return err
			}
			return nil
		}

		if len(args) == 0 {
			cmd.Println(cmd.UsageString())
			return nil
		}

		envs := os.Environ()
		subCmd := args[0]
		path, err := exec.LookPath(version.Name + "-" + subCmd)
		if err != nil {
			if strings.HasPrefix(subCmd, "-") {
				cmd.PrintErrf("Error: unknown flag: '%s'\n", subCmd)
				cmd.HelpFunc()(cmd, args)
				return nil
			}
			cmd.PrintErrf("Error: unknown command \"%s\" for \"%s\"\n", subCmd, version.Name)
			cmd.PrintErrf("Run '%s --help' for usage.\n", version.Name)
			return nil
		}
		args = args[1:]

		cfg, err := config.New()
		if err != nil {
			return err
		}
		opts := []config.Option{}
		if dsn != "" {
			opts = append(opts, config.DSNURL(dsn))
		}
		if err := cfg.Load(configPath, opts...); err != nil {
			return err
		}

		s, err := getSchemaFromJSONorDSN(cfg)
		if err == nil {
			envs = append(envs, fmt.Sprintf("TBLS_DSN=%s", cfg.DSN.URL))
			envs = append(envs, fmt.Sprintf("TBLS_CONFIG_PATH=%s", cfg.Path))
			o := json.New(true)
			tmpfile, err := os.CreateTemp("", "TBLS_SCHEMA")
			if err != nil {
				return err
			}
			defer os.Remove(tmpfile.Name())
			if err := o.OutputSchema(tmpfile, s); err != nil {
				return err
			}
			envs = append(envs, fmt.Sprintf("TBLS_SCHEMA=%s", tmpfile.Name()))
		}

		c := exec.Command(path, args...) // #nosec
		c.Env = envs
		c.Stdout = os.Stdout
		c.Stdin = os.Stdin
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}
		return nil
	},
}

func Execute() {
	var err error
	subCmds, err = getExtSubCmds("tbls")
	if err != nil {
		printError(err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		printError(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetUsageTemplate(rootUsageTemplate)
	rootCmd.Flags().StringVarP(&when, "when", "", "", "command execute condition")
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	rootCmd.Flags().StringVarP(&dsn, "dsn", "", "", "data source name")
}

// genValidArgsFunc
func genValidArgsFunc(prefix string) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		toC := toComplete
		if len(args) > 0 {
			toC = args[0]
		}
		completions := []string{}
		for _, subCmd := range subCmds {
			trimed := strings.TrimPrefix(subCmd, fmt.Sprintf("%s-", prefix))
			switch {
			case len(args) == 0 && toComplete == "":
				completions = append(completions, fmt.Sprintf("%s\t%s", trimed, subCmd))
			case trimed == toC && len(args) > 0:
				// exec external sub-command "__complete"
				subCmdArgs := []string{"__complete"}
				subCmdArgs = append(subCmdArgs, args[1:]...)
				subCmdArgs = append(subCmdArgs, toComplete)
				out, err := exec.Command(subCmd, subCmdArgs...).Output() // #nosec
				if err != nil {
					return []string{}, cobra.ShellCompDirectiveError
				}
				splited := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
				completions = append(completions, splited[:len(splited)-1]...)
			case trimed != strings.TrimPrefix(trimed, toC):
				completions = append(completions, fmt.Sprintf("%s\t%s", trimed, subCmd))
			}
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// getExtSubCmds
func getExtSubCmds(prefix string) ([]string, error) {
	subCmds := []string{}
	paths := lo.Uniq(filepath.SplitList(os.Getenv("PATH")))
	for _, p := range paths {
		if strings.TrimSpace(p) == "" {
			continue
		}
		entries, err := os.ReadDir(p)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if !strings.HasPrefix(e.Name(), fmt.Sprintf("%s-", prefix)) {
				continue
			}
			fi, err := e.Info()
			if err != nil {
				continue
			}
			mode := fi.Mode()
			if mode&0111 == 0 {
				continue
			}
			subCmds = append(subCmds, e.Name())
		}
	}
	sortpkg.Strings(subCmds)
	return lo.Uniq(subCmds), nil
}

func printError(err error) {
	env := os.Getenv("DEBUG")
	debug, _ := strconv.ParseBool(env)
	if env != "" && debug {
		fmt.Println(err, errors.StackTraces(err))
	} else {
		fmt.Println(err)
	}
}
