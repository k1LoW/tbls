package output

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/k1LoW/tbls/dict"
	"github.com/k1LoW/tbls/schema"
	"gitlab.com/golang-commonmark/mdurl"
)

// Output is interface for output
type Output interface {
	OutputSchema(wr io.Writer, s *schema.Schema) error
	OutputTable(wr io.Writer, s *schema.Table) error
}

func Funcs(d *dict.Dict) map[string]interface{} {
	return template.FuncMap{
		"nl2br": func(text string) string {
			r := strings.NewReplacer("\r\n", "<br>", "\n", "<br>", "\r", "<br>")
			return r.Replace(text)
		},
		"nl2br_slash": func(text string) string {
			r := strings.NewReplacer("\r\n", "<br />", "\n", "<br />", "\r", "<br />")
			return r.Replace(text)
		},
		"nl2mdnl": func(text string) string {
			r := strings.NewReplacer("\r\n", "  \n", "\n", "  \n", "\r", "  \n")
			return r.Replace(text)
		},
		"nl2space": func(text string) string {
			r := strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ")
			return r.Replace(text)
		},
		"escape_nl": func(text string) string {
			r := strings.NewReplacer("\r\n", "\\n", "\n", "\\n", "\r", "\\n")
			return r.Replace(text)
		},
		"show_only_first_paragraph": ShowOnlyFirstParagraph,
		"lookup": func(text string) string {
			return d.Lookup(text)
		},
		"label_join": LabelJoin,
		"escape": func(text string) string {
			return mdurl.Encode(text)
		},
	}
}

func ShowOnlyFirstParagraph(text string) string {
	if strings.Contains(text, "\r\n\r\n") {
		splitted := strings.SplitN(text, "\r\n\r\n", 2)
		return splitted[0]
	}
	if strings.Contains(text, "\r\r") {
		splitted := strings.SplitN(text, "\r\r", 2)
		return splitted[0]
	}
	splitted := strings.SplitN(text, "\n\n", 2)
	return splitted[0]
}

func LabelJoin(labels schema.Labels) string {
	if len(labels) == 0 {
		return ""
	}
	m := []string{}
	for _, l := range labels {
		m = append(m, l.Name)
	}
	return fmt.Sprintf("`%s`", strings.Join(m, "` `"))
}
