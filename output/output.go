package output

import (
	"io"
	"strings"
	"text/template"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
)

// Output is interface for output
type Output interface {
	OutputSchema(wr io.Writer, s *schema.Schema) error
	OutputTable(wr io.Writer, s *schema.Table) error
}

func Funcs(c *config.Config) map[string]interface{} {
	return template.FuncMap{
		"nl2br": func(text string) string {
			r := strings.NewReplacer("\r\n", "<br>", "\n", "<br>", "\r", "<br>")
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
		"lookup": func(text string) string {
			return c.Dict.Lookup(text)
		},
	}
}
