package dot

import (
	"time"

	"github.com/jessevdk/go-assets"
)

var _Assets21532ae17ad95976ac467eeaeab81f2bb1d537e4 = "digraph {{ .Schema.Name }} {\n  // Config\n  graph [rankdir=TB, layout=dot, margin=0.2, fontname=\"Arial\"];\n  node [shape=record, fontsize=14, fontname=\"Arial\"];\n  edge [fontsize=10, labelfloat=false, splines=none, fontname=\"Arial\"];\n\n  // Tables\n  {{- range $i, $t := .Schema.Tables }}\n  {{ $t.Name }} [shape=none, margin=0.2, label=<<table border=\"0\" cellborder=\"1\" cellspacing=\"0\" cellpadding=\"6\">\n                 <tr><td bgcolor=\"#EFEFEF\"><font face=\"Arial Bold\" point-size=\"18\">{{ $t.Name }}</font> <font color=\"#666666\">[{{ $t.Type }}]</font></td></tr>\n                 {{- range $ii, $c := $t.Columns }}\n                 <tr><td port=\"{{ $c.Name }}\" align=\"left\">{{ $c.Name }} <font color=\"#666666\">[{{ $c.Type }}]</font></td></tr>\n                 {{- end }}\n              </table>>];\n  {{- end }}\n\n  // Relations\n  {{- range $j, $r := .Schema.Relations }}\n  {{ $r.Table.Name }}:{{ $c := index $r.Columns 0 }}{{ $c.Name }} -> {{ $r.ParentTable.Name }}:{{ $pc := index $r.ParentColumns 0 }}{{ $pc.Name }} [dir=back, arrowtail=crow, taillabel=<<table cellpadding=\"5\" border=\"0\" cellborder=\"0\"><tr><td>{{ $r.Def }}</td></tr></table>>];\n  {{- end }}\n}\n"

// Assets returns go-assets FileSystem
var Assets = assets.NewFileSystem(map[string][]string{"/": []string{"schema.dot.tmpl"}}, map[string]*assets.File{
	"/schema.dot.tmpl": &assets.File{
		Path:     "/schema.dot.tmpl",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1527431106, 1527431106000000000),
		Data:     []byte(_Assets21532ae17ad95976ac467eeaeab81f2bb1d537e4),
	}, "/": &assets.File{
		Path:     "/",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1527431106, 1527431106000000000),
		Data:     nil,
	}}, "")
