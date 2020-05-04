// ---------------------------------------------------------------
// The original code is github.com/spf13/cobra/zsh_completion.go
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	sortp "sort"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func genZshCompletion(cmd *cobra.Command, w io.Writer) error {
	tmpl, err := template.New("Main").Funcs(zshCompFuncMap).Parse(zshCompletionText)
	if err != nil {
		return fmt.Errorf("error creating zsh completion template: %v", err)
	}
	return tmpl.Execute(w, cmd.Root())
}

const (
	zshCompArgumentAnnotation   = "cobra_annotations_zsh_completion_argument_annotation"
	zshCompArgumentFilenameComp = "cobra_annotations_zsh_completion_argument_file_completion"
	zshCompArgumentWordComp     = "cobra_annotations_zsh_completion_argument_word_completion"
	zshCompDirname              = "cobra_annotations_zsh_dirname"
)

var (
	zshCompFuncMap = template.FuncMap{
		"genZshFuncName":              zshCompGenFuncName,
		"extractFlags":                zshCompExtractFlag,
		"genFlagEntryForZshArguments": zshCompGenFlagEntryForArguments,
		"extractArgsCompletions":      zshCompExtractArgumentCompletionHintsForRendering,
	}
	zshCompletionText = `
{{/* should accept Command (that contains subcommands) as parameter */}}
{{define "argumentsC" -}}
{{ $cmdPath := genZshFuncName .}}
function {{$cmdPath}} {
  local -a commands
  local -a subcommands

  _arguments -C \{{- range extractFlags .}}
    {{genFlagEntryForZshArguments .}} \{{- end}}
    "1: :->cmnds" \
    "*::arg:->args"

  case $state in
  cmnds)
    commands=({{range .Commands}}{{if not .Hidden}}
      "{{.Name}}:{{.Short}}"{{end}}{{end}}
    )
    subcommands=( ${(@f)"$(print -l ${^path}/*(-*N) | grep tbls- | xargs basename | sort | uniq | sed -e 's/^tbls-\(.*\)$/\1:tbls-\1/')"} ) 
    commands+=($subcommands)    
    _describe "command" commands
    ;;
  esac

  case "$words[1]" in {{- range .Commands}}{{if not .Hidden}}
  {{.Name}})
    {{$cmdPath}}_{{.Name}}
    ;;{{end}}{{end}}
  *)
    if type _tbls-"$words[1]" 1>/dev/null 2>/dev/null
    then
      _tbls-"$words[1]"
    fi
    ;;
  esac
}
{{range .Commands}}{{if not .Hidden}}
{{template "selectCmdTemplate" .}}
{{- end}}{{end}}
{{- end}}

{{/* should accept Command without subcommands as parameter */}}
{{define "arguments" -}}
function {{genZshFuncName .}} {
{{"  _arguments"}}{{range extractFlags .}} \
    {{genFlagEntryForZshArguments . -}}
{{end}}{{range extractArgsCompletions .}} \
    {{.}}{{end}}
}
{{end}}

{{/* dispatcher for commands with or without subcommands */}}
{{define "selectCmdTemplate" -}}
{{if .Hidden}}{{/* ignore hidden*/}}{{else -}}
{{if .Commands}}{{template "argumentsC" .}}{{else}}{{template "arguments" .}}{{end}}
{{- end}}
{{- end}}

{{/* template entry point */}}
{{define "Main" -}}
#compdef _{{.Name}} {{.Name}}

{{template "selectCmdTemplate" .}}
{{end}}
`
)

// zshCompArgsAnnotation is used to encode/decode zsh completion for
// arguments to/from Command.Annotations.
type zshCompArgsAnnotation map[int]zshCompArgHint

type zshCompArgHint struct {
	// Indicates the type of the completion to use. One of:
	// zshCompArgumentFilenameComp or zshCompArgumentWordComp
	Tipe string `json:"type"`

	// A value for the type above (globs for file completion or words)
	Options []string `json:"options"`
}

func zshCompExtractArgumentCompletionHintsForRendering(c *cobra.Command) ([]string, error) {
	var result []string
	annotation, err := zshCompGetArgsAnnotations(c)
	if err != nil {
		return nil, err
	}
	for k, v := range annotation {
		s, err := zshCompRenderZshCompArgHint(k, v)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	if len(c.ValidArgs) > 0 {
		if _, positionOneExists := annotation[1]; !positionOneExists {
			s, err := zshCompRenderZshCompArgHint(1, zshCompArgHint{
				Tipe:    zshCompArgumentWordComp,
				Options: c.ValidArgs,
			})
			if err != nil {
				return nil, err
			}
			result = append(result, s)
		}
	}
	sortp.Strings(result)
	return result, nil
}

func zshCompRenderZshCompArgHint(i int, z zshCompArgHint) (string, error) {
	switch t := z.Tipe; t {
	case zshCompArgumentFilenameComp:
		var globs []string
		for _, g := range z.Options {
			globs = append(globs, fmt.Sprintf(`-g "%s"`, g))
		}
		return fmt.Sprintf(`'%d: :_files %s'`, i, strings.Join(globs, " ")), nil
	case zshCompArgumentWordComp:
		var words []string
		for _, w := range z.Options {
			words = append(words, fmt.Sprintf("%q", w))
		}
		return fmt.Sprintf(`'%d: :(%s)'`, i, strings.Join(words, " ")), nil
	default:
		return "", fmt.Errorf("invalid zsh argument completion annotation: %s", t)
	}
}

func zshCompGetArgsAnnotations(c *cobra.Command) (zshCompArgsAnnotation, error) {
	annotation := make(zshCompArgsAnnotation)
	annotationString, ok := c.Annotations[zshCompArgumentAnnotation]
	if !ok {
		return annotation, nil
	}
	err := json.Unmarshal([]byte(annotationString), &annotation)
	if err != nil {
		return annotation, fmt.Errorf("error unmarshaling zsh argument annotation: %v", err)
	}
	return annotation, nil
}

func zshCompGenFuncName(c *cobra.Command) string {
	if c.HasParent() {
		return zshCompGenFuncName(c.Parent()) + "_" + c.Name()
	}
	return "_" + c.Name()
}

func zshCompExtractFlag(c *cobra.Command) []*pflag.Flag {
	var flags []*pflag.Flag
	c.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			flags = append(flags, f)
		}
	})
	c.InheritedFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			flags = append(flags, f)
		}
	})
	return flags
}

// zshCompGenFlagEntryForArguments returns an entry that matches _arguments
// zsh-completion parameters. It's too complicated to generate in a template.
func zshCompGenFlagEntryForArguments(f *pflag.Flag) string {
	if f.Name == "" || f.Shorthand == "" {
		return zshCompGenFlagEntryForSingleOptionFlag(f)
	}
	return zshCompGenFlagEntryForMultiOptionFlag(f)
}

func zshCompGenFlagEntryForSingleOptionFlag(f *pflag.Flag) string {
	var option, multiMark, extras string

	if zshCompFlagCouldBeSpecifiedMoreThenOnce(f) {
		multiMark = "*"
	}

	option = "--" + f.Name
	if option == "--" {
		option = "-" + f.Shorthand
	}
	extras = zshCompGenFlagEntryExtras(f)

	return fmt.Sprintf(`'%s%s[%s]%s'`, multiMark, option, zshCompQuoteFlagDescription(f.Usage), extras)
}

func zshCompGenFlagEntryForMultiOptionFlag(f *pflag.Flag) string {
	var options, parenMultiMark, curlyMultiMark, extras string

	if zshCompFlagCouldBeSpecifiedMoreThenOnce(f) {
		parenMultiMark = "*"
		curlyMultiMark = "\\*"
	}

	options = fmt.Sprintf(`'(%s-%s %s--%s)'{%s-%s,%s--%s}`,
		parenMultiMark, f.Shorthand, parenMultiMark, f.Name, curlyMultiMark, f.Shorthand, curlyMultiMark, f.Name)
	extras = zshCompGenFlagEntryExtras(f)

	return fmt.Sprintf(`%s'[%s]%s'`, options, zshCompQuoteFlagDescription(f.Usage), extras)
}

func zshCompGenFlagEntryExtras(f *pflag.Flag) string {
	if f.NoOptDefVal != "" {
		return ""
	}

	extras := ":" // allow options for flag (even without assistance)
	for key, values := range f.Annotations {
		switch key {
		case zshCompDirname:
			extras = fmt.Sprintf(":filename:_files -g %q", values[0])
		case cobra.BashCompFilenameExt:
			extras = ":filename:_files"
			for _, pattern := range values {
				extras = extras + fmt.Sprintf(` -g "%s"`, pattern)
			}
		}
	}

	return extras
}

func zshCompFlagCouldBeSpecifiedMoreThenOnce(f *pflag.Flag) bool {
	return strings.Contains(f.Value.Type(), "Slice") ||
		strings.Contains(f.Value.Type(), "Array")
}

func zshCompQuoteFlagDescription(s string) string {
	return strings.Replace(s, "'", `'\''`, -1)
}
