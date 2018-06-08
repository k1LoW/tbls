package dot

import (
	"time"

	"github.com/jessevdk/go-assets"
)

var _Assets5bd148e6149bb9adcdddfcf8cc46d6e3047dbe26 = "digraph {{ .Table.Name }} {\n  // Config\n  graph [rankdir=TB, layout=dot, fontname=\"Arial\"];\n  node [shape=record, fontsize=14, margin=0.6, fontname=\"Arial\"];\n  edge [fontsize=10, labelfloat=false, splines=none, fontname=\"Arial\"];\n\n  // Tables\n  {{ .Table.Name }} [shape=none, label=<<table border=\"3\" cellborder=\"1\" cellspacing=\"0\" cellpadding=\"6\">\n                 <tr><td bgcolor=\"#EFEFEF\"><font face=\"Arial Bold\" point-size=\"18\">{{ .Table.Name | html }}</font> <font color=\"#666666\">[{{ .Table.Type | html }}]</font></td></tr>\n                 {{- range $ii, $c := .Table.Columns }}\n                 <tr><td port=\"{{ $c.Name | html }}\" align=\"left\">{{ $c.Name | html }} <font color=\"#666666\">[{{ $c.Type }}]</font></td></tr>\n                 {{- end }}\n              </table>>];\n  {{- range $i, $t := .Tables }}\n  {{ $t.Name }} [shape=none, label=<<table border=\"0\" cellborder=\"1\" cellspacing=\"0\" cellpadding=\"6\">\n                 <tr><td bgcolor=\"#EFEFEF\"><font face=\"Arial Bold\" point-size=\"18\">{{ $t.Name | html }}</font> <font color=\"#666666\">[{{ $t.Type | html }}]</font></td></tr>\n                 {{- range $ii, $c := $t.Columns }}\n                 <tr><td port=\"{{ $c.Name | html }}\" align=\"left\">{{ $c.Name | html }} <font color=\"#666666\">[{{ $c.Type | html }}]</font></td></tr>\n                 {{- end }}\n              </table>>];\n  {{- end }}\n\n  // Relations\n  {{- range $i, $r := .Relations }}\n  {{ $r.Table.Name }}:{{ $c := index $r.Columns 0 }}{{ $c.Name }} -> {{ $r.ParentTable.Name }}:{{ $pc := index $r.ParentColumns 0 }}{{ $pc.Name }} [dir=back, arrowtail=crow, {{ if $r.IsAdditional }}style =\"dashed\",{{ end }} taillabel=<<table cellpadding=\"5\" border=\"0\" cellborder=\"0\"><tr><td>{{ $r.Def | html }}</td></tr></table>>];\n  {{- end }}\n}\n"
var _Assets21532ae17ad95976ac467eeaeab81f2bb1d537e4 = "digraph \"{{ .Schema.Name }}\" {\n  // Config\n  graph [rankdir=TB, layout=dot, fontname=\"Arial\"];\n  node [shape=record, fontsize=14, margin=0.6, fontname=\"Arial\"];\n  edge [fontsize=10, labelfloat=false, splines=none, fontname=\"Arial\"];\n\n  // Tables\n  {{- range $i, $t := .Schema.Tables }}\n  {{ $t.Name }} [shape=none, label=<<table border=\"0\" cellborder=\"1\" cellspacing=\"0\" cellpadding=\"6\">\n                 <tr><td bgcolor=\"#EFEFEF\"><font face=\"Arial Bold\" point-size=\"18\">{{ $t.Name | html }}</font> <font color=\"#666666\">[{{ $t.Type | html }}]</font></td></tr>\n                 {{- range $ii, $c := $t.Columns }}\n                 <tr><td port=\"{{ $c.Name | html }}\" align=\"left\">{{ $c.Name | html }} <font color=\"#666666\">[{{ $c.Type | html }}]</font></td></tr>\n                 {{- end }}\n              </table>>];\n  {{- end }}\n\n  // Relations\n  {{- range $j, $r := .Schema.Relations }}\n  {{ $r.Table.Name }}:{{ $c := index $r.Columns 0 }}{{ $c.Name }} -> {{ $r.ParentTable.Name }}:{{ $pc := index $r.ParentColumns 0 }}{{ $pc.Name }} [dir=back, arrowtail=crow, {{ if $r.IsAdditional }}style=\"dashed\",{{ end }} taillabel=<<table cellpadding=\"5\" border=\"0\" cellborder=\"0\"><tr><td>{{ $r.Def | html }}</td></tr></table>>];\n  {{- end }}\n}\n"

// Assets returns go-assets FileSystem
var Assets = assets.NewFileSystem(map[string][]string{"/": []string{"table.dot.tmpl", "schema.dot.tmpl"}}, map[string]*assets.File{
	"/table.dot.tmpl": &assets.File{
		Path:     "/table.dot.tmpl",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1527750838, 1527750838424881150),
		Data:     []byte(_Assets5bd148e6149bb9adcdddfcf8cc46d6e3047dbe26),
	}, "/schema.dot.tmpl": &assets.File{
		Path:     "/schema.dot.tmpl",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1528193906, 1528193906825648193),
		Data:     []byte(_Assets21532ae17ad95976ac467eeaeab81f2bb1d537e4),
	}, "/": &assets.File{
		Path:     "/",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1528193906, 1528193906825330636),
		Data:     nil,
	}}, "")
