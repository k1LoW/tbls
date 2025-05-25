package md

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/k1LoW/errors"
	"github.com/SouhlInc/tbls/config"
	"github.com/SouhlInc/tbls/output"
	"github.com/SouhlInc/tbls/output/mermaid"
	"github.com/SouhlInc/tbls/schema"
	"github.com/mattn/go-runewidth"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/samber/lo"
	"gitlab.com/golang-commonmark/mdurl"
)

// mdEscRep is a replacer for markdown escape.
// Add when a case of actual display collapse appears.
var mdEscRep = strings.NewReplacer("`", "\\`", `|`, `\|`, "<", `\<`, ">", `\>`)

var _ output.Output = &Md{}

//go:embed templates/*
var tmpl embed.FS

// Md struct.
type Md struct {
	config *config.Config
	tmpl   embed.FS
}

// New return Md.
func New(c *config.Config) *Md {
	return &Md{
		config: c,
		tmpl:   tmpl,
	}
}

// OutputSchema output .md format for all tables.
func (m *Md) OutputSchema(wr io.Writer, s *schema.Schema) error {
	ts, err := m.indexTemplate()
	if err != nil {
		return errors.WithStack(err)
	}
	tmpl := template.Must(template.New("index").Funcs(output.Funcs(&m.config.MergedDict)).Parse(ts))
	templateData := m.makeSchemaTemplateData(s)
	templateData["er"] = !m.config.ER.Skip
	templateData["showOnlyFirstParagraph"] = m.config.Format.ShowOnlyFirstParagraph
	switch m.config.ER.Format {
	case "mermaid":
		buf := new(bytes.Buffer)
		mmd := mermaid.New(m.config)
		if err := mmd.OutputSchema(buf, s); err != nil {
			return err
		}
		templateData["erDiagram"] = fmt.Sprintf("```mermaid\n%s```", buf.String())
	default:
		templateData["erDiagram"] = fmt.Sprintf("![er](%sschema.%s)", m.config.BaseURL, m.config.ER.Format)
	}
	if err := tmpl.Execute(wr, templateData); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// OutputTable output md format for table.
func (m *Md) OutputTable(wr io.Writer, t *schema.Table) error {
	ts, err := m.tableTemplate()
	if err != nil {
		return errors.WithStack(err)
	}
	tmpl := template.Must(template.New(t.Name).Funcs(output.Funcs(&m.config.MergedDict)).Parse(ts))
	templateData := m.makeTableTemplateData(t)
	templateData["er"] = !m.config.ER.Skip
	switch m.config.ER.Format {
	case "mermaid":
		buf := new(bytes.Buffer)
		mmd := mermaid.New(m.config)
		if err := mmd.OutputTable(buf, t); err != nil {
			return err
		}
		templateData["erDiagram"] = fmt.Sprintf("```mermaid\n%s```", buf.String())
	default:
		templateData["erDiagram"] = fmt.Sprintf("![er](%s%s.%s)", m.config.BaseURL, mdurl.Encode(t.Name), m.config.ER.Format)
	}

	if err := tmpl.Execute(wr, templateData); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// OutputViewpoint output md format for viewpoint.
func (m *Md) OutputViewpoint(wr io.Writer, i int, v *schema.Viewpoint) error {
	ts, err := m.viewpointTemplate()
	if err != nil {
		return errors.WithStack(err)
	}
	tmpl := template.Must(template.New("viewpoint").Funcs(output.Funcs(&m.config.MergedDict)).Parse(ts))
	templateData, err := m.makeViewpointTemplateData(v)
	if err != nil {
		return errors.WithStack(err)
	}
	templateData["er"] = !m.config.ER.Skip
	switch m.config.ER.Format {
	case "mermaid":
		buf := new(bytes.Buffer)
		mmd := mermaid.New(m.config)
		if err := mmd.OutputSchema(buf, v.Schema); err != nil {
			return err
		}
		templateData["erDiagram"] = fmt.Sprintf("```mermaid\n%s```", buf.String())
	default:
		templateData["erDiagram"] = fmt.Sprintf("![er](%sviewpoint-%d.%s)", m.config.BaseURL, i, m.config.ER.Format)
	}
	if err := tmpl.Execute(wr, templateData); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Output generate markdown files.
func Output(s *schema.Schema, c *config.Config, force bool) (e error) {
	docPath := c.DocPath

	fullPath, err := filepath.Abs(docPath)
	if err != nil {
		return errors.WithStack(err)
	}

	if !force && outputExists(s, fullPath) {
		return errors.New("output files already exists")
	}

	err = os.MkdirAll(fullPath, 0755) // #nosec
	if err != nil {
		return errors.WithStack(err)
	}

	// README.md
	f, err := os.Create(filepath.Clean(filepath.Join(fullPath, "README.md")))
	defer func() {
		err := f.Close()
		if err != nil {
			e = err
		}
	}()
	if err != nil {
		return errors.WithStack(err)
	}
	md := New(c)
	if err := md.OutputSchema(f, s); err != nil {
		return errors.WithStack(err)
	}
	fmt.Printf("%s\n", filepath.Join(docPath, "README.md"))

	// tables
	for _, t := range s.Tables {
		f, err := os.Create(filepath.Clean(filepath.Join(fullPath, fmt.Sprintf("%s.md", t.Name))))
		if err != nil {
			_ = f.Close()
			return errors.WithStack(err)
		}
		if err := md.OutputTable(f, t); err != nil {
			_ = f.Close()
			return errors.WithStack(err)
		}
		fmt.Printf("%s\n", filepath.Join(docPath, fmt.Sprintf("%s.md", t.Name)))
		if err := f.Close(); err != nil {
			return errors.WithStack(err)
		}
	}

	// viewpoints
	for i, v := range s.Viewpoints {
		fn := fmt.Sprintf("viewpoint-%d.md", i)
		f, err := os.Create(filepath.Clean(filepath.Join(fullPath, fn)))
		if err != nil {
			_ = f.Close()
			return errors.WithStack(err)
		}
		if err := md.OutputViewpoint(f, i, v); err != nil {
			_ = f.Close()
			return errors.WithStack(err)
		}
		fmt.Printf("%s\n", filepath.Join(docPath, fn))
		if err := f.Close(); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// DiffSchemas show diff databases.
func DiffSchemas(s, s2 *schema.Schema, c, c2 *config.Config) (string, error) {
	var diff string
	md := New(c)
	md2 := New(c2)

	// README.md
	a := new(bytes.Buffer)
	if err := md.OutputSchema(a, s); err != nil {
		return "", errors.WithStack(err)
	}

	b := new(bytes.Buffer)
	if err := md2.OutputSchema(b, s2); err != nil {
		return "", errors.WithStack(err)
	}

	mdsnA, err := c.MaskedDSN()
	if err != nil {
		return "", errors.WithStack(err)
	}
	from := fmt.Sprintf("tbls doc %s", mdsnA)

	mdsnB, err := c2.MaskedDSN()
	if err != nil {
		return "", errors.WithStack(err)
	}
	to := fmt.Sprintf("tbls doc %s", mdsnB)

	d := difflib.UnifiedDiff{
		A:        difflib.SplitLines(a.String()),
		B:        difflib.SplitLines(b.String()),
		FromFile: from,
		ToFile:   to,
		Context:  3,
	}

	text, _ := difflib.GetUnifiedDiffString(d)
	if text != "" {
		diff += fmt.Sprintf("diff '%s' '%s'\n", from, to)
		diff += text
	}

	// tables
	diffed := map[string]struct{}{}
	for _, t := range s.Tables {

		tName := t.Name
		diffed[tName] = struct{}{}

		a := new(bytes.Buffer)
		if err := md.OutputTable(a, t); err != nil {
			return "", errors.WithStack(err)
		}
		from := fmt.Sprintf("%s %s", mdsnA, tName)

		b := new(bytes.Buffer)
		t2, err := s2.FindTableByName(tName)
		if err == nil {
			if err := md2.OutputTable(b, t2); err != nil {
				return "", errors.WithStack(err)
			}
		}
		to := fmt.Sprintf("%s %s", mdsnB, tName)

		d := difflib.UnifiedDiff{
			A:        difflib.SplitLines(a.String()),
			B:        difflib.SplitLines(b.String()),
			FromFile: from,
			ToFile:   to,
			Context:  3,
		}

		text, _ := difflib.GetUnifiedDiffString(d)
		if text != "" {
			diff += fmt.Sprintf("diff '%s' '%s'\n", from, to)
			diff += text
		}
	}
	for _, t := range s2.Tables {
		tName := t.Name
		if _, ok := diffed[tName]; ok {
			continue
		}
		a := ""
		from := fmt.Sprintf("%s %s", mdsnA, tName)

		b := new(bytes.Buffer)
		if err := md2.OutputTable(b, t); err != nil {
			return "", errors.WithStack(err)
		}
		to := fmt.Sprintf("%s %s", mdsnB, tName)

		d := difflib.UnifiedDiff{
			A:        difflib.SplitLines(a),
			B:        difflib.SplitLines(b.String()),
			FromFile: from,
			ToFile:   to,
			Context:  3,
		}

		text, _ := difflib.GetUnifiedDiffString(d)
		if text != "" {
			diff += fmt.Sprintf("diff '%s' '%s'\n", from, to)
			diff += text
		}
	}

	return diff, nil
}

// DiffSchemaAndDocs show diff markdown files and database.
func DiffSchemaAndDocs(docPath string, s *schema.Schema, c *config.Config) (string, error) {
	var diff string
	fullPath, err := filepath.Abs(docPath)
	if err != nil {
		return "", errors.WithStack(err)
	}
	diffed := map[string]struct{}{}

	// README.md
	md := New(c)
	buf := new(bytes.Buffer)
	if err := md.OutputSchema(buf, s); err != nil {
		return "", errors.WithStack(err)
	}

	targetPath := filepath.Join(fullPath, "README.md")
	a, err := os.ReadFile(filepath.Clean(targetPath))
	if err != nil {
		a = []byte{}
	}

	mdsn, err := c.MaskedDSN()
	if err != nil {
		return "", errors.WithStack(err)
	}
	to := fmt.Sprintf("tbls doc %s", mdsn)

	from := filepath.Join(docPath, "README.md")

	d := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(a)),
		B:        difflib.SplitLines(buf.String()),
		FromFile: from,
		ToFile:   to,
		Context:  3,
	}

	text, _ := difflib.GetUnifiedDiffString(d)
	if text != "" {
		diff += fmt.Sprintf("diff '%s' '%s'\n", from, to)
		diff += text
	}
	diffed["README.md"] = struct{}{}

	// tables
	for _, t := range s.Tables {
		buf := new(bytes.Buffer)
		to := fmt.Sprintf("%s %s", mdsn, t.Name)
		if err := md.OutputTable(buf, t); err != nil {
			return "", errors.WithStack(err)
		}
		fn := fmt.Sprintf("%s.md", t.Name)
		targetPath := filepath.Join(fullPath, fn)
		a, err := os.ReadFile(filepath.Clean(targetPath))
		if err != nil {
			a = []byte{}
		}
		from := filepath.Join(docPath, fn)

		d := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(a)),
			B:        difflib.SplitLines(buf.String()),
			FromFile: from,
			ToFile:   to,
			Context:  3,
		}

		text, _ := difflib.GetUnifiedDiffString(d)
		if text != "" {
			diff += fmt.Sprintf("diff '%s' '%s'\n", from, to)
			diff += text
		}
		diffed[fn] = struct{}{}
	}

	// viewpoints
	for i, v := range s.Viewpoints {
		buf := new(bytes.Buffer)
		n := fmt.Sprintf("viewpoint-%d", i)
		fn := fmt.Sprintf("viewpoint-%d.md", i)
		to := fmt.Sprintf("%s %s", mdsn, n)
		if err := md.OutputViewpoint(buf, i, v); err != nil {
			return "", errors.WithStack(err)
		}
		targetPath := filepath.Join(fullPath, fn)
		a, err := os.ReadFile(filepath.Clean(targetPath))
		if err != nil {
			a = []byte{}
		}
		from := filepath.Join(docPath, fn)

		d := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(a)),
			B:        difflib.SplitLines(buf.String()),
			FromFile: from,
			ToFile:   to,
			Context:  3,
		}

		text, _ := difflib.GetUnifiedDiffString(d)
		if text != "" {
			diff += fmt.Sprintf("diff '%s' '%s'\n", from, to)
			diff += text
		}
		diffed[fn] = struct{}{}
	}

	files, err := os.ReadDir(fullPath)
	if err != nil {
		return "", errors.WithStack(err)
	}
	for _, f := range files {
		if _, ok := diffed[f.Name()]; ok {
			continue
		}
		if filepath.Ext(f.Name()) != ".md" {
			continue
		}

		fname := f.Name()
		targetPath := filepath.Join(fullPath, fname)
		a, err := os.ReadFile(filepath.Clean(targetPath))
		if err != nil {
			return "", errors.WithStack(err)
		}
		from := filepath.Join(docPath, f.Name())

		b := ""
		to := fmt.Sprintf("%s %s", mdsn, filepath.Base(fname[:len(fname)-len(filepath.Ext(fname))]))

		d := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(a)),
			B:        difflib.SplitLines(b),
			FromFile: from,
			ToFile:   to,
			Context:  3,
		}

		text, _ := difflib.GetUnifiedDiffString(d)
		if text != "" {
			diff += fmt.Sprintf("diff '%s' '%s'\n", from, to)
			diff += text
		}
	}
	return diff, nil
}

func (m *Md) indexTemplate() (string, error) {
	if m.config.Templates.MD.Index != "" {
		tb, err := os.ReadFile(m.config.Templates.MD.Index)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	}
	tb, err := m.tmpl.ReadFile("templates/index.md.tmpl")
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(tb), nil
}

func (m *Md) tableTemplate() (string, error) {
	if m.config.Templates.MD.Table != "" {
		tb, err := os.ReadFile(m.config.Templates.MD.Table)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	}
	tb, err := m.tmpl.ReadFile("templates/table.md.tmpl")
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(tb), nil
}

func (m *Md) viewpointTemplate() (string, error) {
	if m.config.Templates.MD.Viewpoint != "" {
		tb, err := os.ReadFile(m.config.Templates.MD.Viewpoint)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	}
	tb, err := m.tmpl.ReadFile("templates/viewpoint.md.tmpl")
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(tb), nil
}

func (m *Md) makeSchemaTemplateData(s *schema.Schema) map[string]interface{} {
	number := m.config.Format.Number
	adjust := m.config.Format.Adjust
	showOnlyFirstParagraph := m.config.Format.ShowOnlyFirstParagraph
	hasTableWithLabels := s.HasTableWithLabels()

	// Tables
	tablesData := m.tablesData(s.Tables, number, adjust, showOnlyFirstParagraph, hasTableWithLabels)

	// Functions
	functionsData := m.functionsData(s.Functions, number, adjust)

	// Viewpoints
	viewpointsData := m.viewpointsData(s.Viewpoints, number, adjust, showOnlyFirstParagraph)

	// Enums
	enumData := m.enumData(s.Enums)

	return map[string]interface{}{
		"Schema":     s,
		"Tables":     tablesData,
		"Functions":  functionsData,
		"Viewpoints": viewpointsData,
		"Enums":      enumData,
	}
}

func (m *Md) makeTableTemplateData(t *schema.Table) map[string]interface{} {
	number := m.config.Format.Number
	adjust := m.config.Format.Adjust
	hideColumns := m.config.Format.HideColumnsWithoutValues
	showOnlyFirstParagraph := m.config.Format.ShowOnlyFirstParagraph

	// Columns
	columnsData := [][]string{}
	columnsHeader := []string{}
	columnsHeaderLine := []string{}
	m.adjustColumnHeader(&columnsHeader, &columnsHeaderLine, true, "Name")
	m.adjustColumnHeader(&columnsHeader, &columnsHeaderLine, true, "Type")
	m.adjustColumnHeader(&columnsHeader, &columnsHeaderLine, true, "Default")
	m.adjustColumnHeader(&columnsHeader, &columnsHeaderLine, true, "Nullable")
	m.adjustColumnHeader(&columnsHeader, &columnsHeaderLine, t.ShowColumn(schema.ColumnExtraDef, hideColumns), "Extra Definition")
	m.adjustColumnHeader(&columnsHeader, &columnsHeaderLine, t.ShowColumn(schema.ColumnOccurrences, hideColumns), "Occurrences")
	m.adjustColumnHeader(&columnsHeader, &columnsHeaderLine, t.ShowColumn(schema.ColumnPercents, hideColumns), "Percents")
	m.adjustColumnHeader(&columnsHeader, &columnsHeaderLine, t.ShowColumn(schema.ColumnChildren, hideColumns), "Children")
	m.adjustColumnHeader(&columnsHeader, &columnsHeaderLine, t.ShowColumn(schema.ColumnParents, hideColumns), "Parents")
	m.adjustColumnHeader(&columnsHeader, &columnsHeaderLine, t.ShowColumn(schema.ColumnComment, hideColumns), "Comment")
	m.adjustColumnHeader(&columnsHeader, &columnsHeaderLine, t.ShowColumn(schema.ColumnLabels, hideColumns), "Labels")

	columnsData = append(columnsData, columnsHeader, columnsHeaderLine)

	for _, c := range t.Columns {
		childRelations := []string{}
		cEncountered := map[string]bool{}
		for _, r := range c.ChildRelations {
			if _, ok := cEncountered[r.Table.Name]; ok {
				continue
			}
			childRelations = append(childRelations, fmt.Sprintf("[%s](%s%s.md)", r.Table.Name, m.config.BaseURL, mdurl.Encode(r.Table.Name)))
			cEncountered[r.Table.Name] = true
		}
		parentRelations := []string{}
		pEncountered := map[string]bool{}
		for _, r := range c.ParentRelations {
			if _, ok := pEncountered[r.ParentTable.Name]; ok {
				continue
			}
			parentRelations = append(parentRelations, fmt.Sprintf("[%s](%s%s.md)", r.ParentTable.Name, m.config.BaseURL, mdurl.Encode(r.ParentTable.Name)))
			pEncountered[r.ParentTable.Name] = true
		}

		data := []string{
			c.Name,
			c.Type,
			c.Default.String,
			fmt.Sprintf("%v", c.Nullable),
		}
		adjustData(&data, t.ShowColumn(schema.ColumnExtraDef, hideColumns), mdEscRep.Replace(c.ExtraDef))
		adjustData(&data, t.ShowColumn(schema.ColumnOccurrences, hideColumns), fmt.Sprint(c.Occurrences.Int32))
		adjustData(&data, t.ShowColumn(schema.ColumnPercents, hideColumns), fmt.Sprintf("%.1f", c.Percents.Float64))
		adjustData(&data, t.ShowColumn(schema.ColumnChildren, hideColumns), strings.Join(childRelations, " "))
		adjustData(&data, t.ShowColumn(schema.ColumnParents, hideColumns), strings.Join(parentRelations, " "))
		adjustData(&data, t.ShowColumn(schema.ColumnComment, hideColumns), mdEscRep.Replace(c.Comment))
		adjustData(&data, t.ShowColumn(schema.ColumnLabels, hideColumns), output.LabelJoin(c.Labels))
		columnsData = append(columnsData, data)
	}

	// Viewpoints
	viewpointsData := [][]string{
		[]string{
			m.config.MergedDict.Lookup("Name"),
			m.config.MergedDict.Lookup("Definition"),
		},
		[]string{"----", "----------"},
	}

	for _, v := range t.Viewpoints {
		desc := v.Desc
		if showOnlyFirstParagraph {
			desc = output.ShowOnlyFirstParagraph(desc)
		}
		data := []string{
			fmt.Sprintf("[%s](viewpoint-%d.md)", v.Name, v.Index),
			desc,
		}

		viewpointsData = append(viewpointsData, data)
	}

	// Constraints
	constraintsData := [][]string{
		[]string{
			m.config.MergedDict.Lookup("Name"),
			m.config.MergedDict.Lookup("Type"),
			m.config.MergedDict.Lookup("Definition"),
		},
		[]string{"----", "----", "----------"},
	}
	cComment := false
	for _, c := range t.Constraints {
		if c.Comment != "" {
			cComment = true
		}
	}
	if cComment {
		constraintsData[0] = append(constraintsData[0], m.config.MergedDict.Lookup("Comment"))
		constraintsData[1] = append(constraintsData[1], "-------")
	}
	for _, c := range t.Constraints {
		data := []string{
			c.Name,
			c.Type,
			c.Def,
		}
		if cComment {
			data = append(data, c.Comment)
		}
		constraintsData = append(constraintsData, data)
	}

	// Indexes
	indexesData := [][]string{
		[]string{
			m.config.MergedDict.Lookup("Name"),
			m.config.MergedDict.Lookup("Definition"),
		},
		[]string{"----", "----------"},
	}
	iComment := false
	for _, i := range t.Indexes {
		if i.Comment != "" {
			iComment = true
		}
	}
	if iComment {
		indexesData[0] = append(indexesData[0], m.config.MergedDict.Lookup("Comment"))
		indexesData[1] = append(indexesData[1], "-------")
	}
	for _, i := range t.Indexes {
		data := []string{
			i.Name,
			i.Def,
		}
		if iComment {
			data = append(data, i.Comment)
		}
		indexesData = append(indexesData, data)
	}

	// Triggers
	triggersData := [][]string{
		{
			m.config.MergedDict.Lookup("Name"),
			m.config.MergedDict.Lookup("Definition"),
		},
		{"----", "----------"},
	}
	tComment := false
	for _, t := range t.Triggers {
		if t.Comment != "" {
			tComment = true
		}
	}
	if tComment {
		triggersData[0] = append(triggersData[0], m.config.MergedDict.Lookup("Comment"))
		triggersData[1] = append(triggersData[1], "-------")
	}
	for _, t := range t.Triggers {
		data := []string{
			t.Name,
			t.Def,
		}
		if tComment {
			data = append(data, t.Comment)
		}
		triggersData = append(triggersData, data)
	}

	// Referenced Tables
	hasReferencedTableWithLabels := false
	for _, rt := range t.ReferencedTables {
		if len(rt.Labels) > 0 {
			hasReferencedTableWithLabels = true
			break
		}
	}

	referencedTables := m.tablesData(t.ReferencedTables, number, adjust, showOnlyFirstParagraph, hasReferencedTableWithLabels)

	if number {
		columnsData = m.addNumberToTable(columnsData)
		constraintsData = m.addNumberToTable(constraintsData)
		indexesData = m.addNumberToTable(indexesData)
		triggersData = m.addNumberToTable(triggersData)
		referencedTables = m.addNumberToTable(referencedTables)
	}

	if adjust {
		return map[string]interface{}{
			"Table":            t,
			"Columns":          adjustTable(columnsData),
			"Viewpoints":       adjustTable(viewpointsData),
			"Constraints":      adjustTable(constraintsData),
			"Indexes":          adjustTable(indexesData),
			"Triggers":         adjustTable(triggersData),
			"ReferencedTables": adjustTable(referencedTables),
		}
	}

	return map[string]interface{}{
		"Table":            t,
		"Columns":          columnsData,
		"Viewpoints":       viewpointsData,
		"Constraints":      constraintsData,
		"Indexes":          indexesData,
		"Triggers":         triggersData,
		"ReferencedTables": referencedTables,
	}
}

func (m *Md) makeViewpointTemplateData(v *schema.Viewpoint) (map[string]interface{}, error) {
	number := m.config.Format.Number
	adjust := m.config.Format.Adjust
	showOnlyFirstParagraph := m.config.Format.ShowOnlyFirstParagraph
	hasTableWithLabels := v.Schema.HasTableWithLabels()

	data := m.makeSchemaTemplateData(v.Schema)
	data["Name"] = v.Name
	data["Desc"] = v.Desc

	groups := []map[string]interface{}{}
	nogroup := v.Schema.Tables
	for _, g := range v.Groups {
		tables, _, err := v.Schema.SeparateTablesThatAreIncludedOrNot(&schema.FilterOption{
			Include:       g.Tables,
			IncludeLabels: g.Labels,
		})
		if err != nil {
			return nil, err
		}
		d := map[string]interface{}{
			"Name":   g.Name,
			"Desc":   g.Desc,
			"Tables": m.tablesData(tables, number, adjust, showOnlyFirstParagraph, hasTableWithLabels),
		}
		groups = append(groups, d)
		nogroup = lo.Without(nogroup, tables...)
	}
	if len(v.Groups) > 0 && len(nogroup) > 0 {
		d := map[string]interface{}{
			"Name":   "-",
			"Desc":   "",
			"Tables": m.tablesData(nogroup, number, adjust, showOnlyFirstParagraph, hasTableWithLabels),
		}
		groups = append(groups, d)
	}
	data["Groups"] = groups

	return data, nil
}

func (m *Md) adjustColumnHeader(columnsHeader *[]string, columnsHeaderLine *[]string, hasColumn bool, name string) {
	if hasColumn {
		*columnsHeader = append(*columnsHeader, m.config.MergedDict.Lookup(name))
		*columnsHeaderLine = append(*columnsHeaderLine, strings.Repeat("-", runewidth.StringWidth(m.config.MergedDict.Lookup(name))))
	}
}

func (m *Md) tablesData(tables []*schema.Table, number, adjust, showOnlyFirstParagraph, hasTableWithLabels bool) [][]string {
	data := [][]string{}
	header := []string{
		m.config.MergedDict.Lookup("Name"),
		m.config.MergedDict.Lookup("Columns"),
		m.config.MergedDict.Lookup("Comment"),
		m.config.MergedDict.Lookup("Type"),
	}
	headerLine := []string{"----", "-------", "-------", "----"}

	if hasTableWithLabels {
		header = append(header, m.config.MergedDict.Lookup("Labels"))
		headerLine = append(headerLine, "------")
	}

	data = append(data,
		header,
		headerLine,
	)

	for _, t := range tables {
		comment := t.Comment
		if showOnlyFirstParagraph {
			comment = output.ShowOnlyFirstParagraph(comment)
		}
		d := []string{
			fmt.Sprintf("[%s](%s%s.md)", t.Name, m.config.BaseURL, mdurl.Encode(t.Name)),
			fmt.Sprintf("%d", len(t.Columns)),
			comment,
			t.Type,
		}
		if hasTableWithLabels {
			d = append(d, output.LabelJoin(t.Labels))
		}
		data = append(data, d)
	}

	if number {
		data = m.addNumberToTable(data)
	}

	if adjust {
		data = adjustTable(data)
	}

	return data
}

func (m *Md) functionsData(functions []*schema.Function, number, adjust bool) [][]string {
	data := [][]string{}
	header := []string{
		m.config.MergedDict.Lookup("Name"),
		m.config.MergedDict.Lookup("ReturnType"),
		m.config.MergedDict.Lookup("Arguments"),
		m.config.MergedDict.Lookup("Type"),
	}
	headerLine := []string{"----", "-------", "-------", "----"}
	data = append(data,
		header,
		headerLine,
	)

	for _, f := range functions {
		d := []string{
			f.Name,
			f.ReturnType,
			f.Arguments,
			f.Type,
		}
		data = append(data, d)
	}

	if number {
		data = m.addNumberToTable(data)
	}

	if adjust {
		data = adjustTable(data)
	}

	return data
}

func (m *Md) enumData(enums []*schema.Enum) [][]string {
	data := [][]string{}

	if len(enums) == 0 {
		return data
	}

	header := []string{
		m.config.MergedDict.Lookup("Name"),
		m.config.MergedDict.Lookup("Values"),
	}
	headerLine := []string{"----", "-------"}
	data = append(data,
		header,
		headerLine,
	)

	for _, e := range enums {
		sort.Strings(e.Values)
		d := []string{
			e.Name,
			strings.Join(e.Values, ", "),
		}
		data = append(data, d)
	}

	return data
}

func (m *Md) viewpointsData(viewpoints []*schema.Viewpoint, number, adjust, showOnlyFirstParagraph bool) [][]string {
	data := [][]string{}
	header := []string{
		m.config.MergedDict.Lookup("Name"),
		m.config.MergedDict.Lookup("Description"),
	}
	headerLine := []string{"----", "-----------"}
	data = append(data,
		header,
		headerLine,
	)

	for i, v := range viewpoints {
		desc := v.Desc
		if showOnlyFirstParagraph {
			desc = output.ShowOnlyFirstParagraph(desc)
		}
		d := []string{
			fmt.Sprintf("[%s](%sviewpoint-%d.md)", v.Name, m.config.BaseURL, i),
			desc,
		}
		data = append(data, d)
	}

	if number {
		data = m.addNumberToTable(data)
	}

	if adjust {
		data = adjustTable(data)
	}

	return data
}

func adjustData(data *[]string, hasData bool, value string) {
	if hasData {
		*data = append(*data, value)
	}
}

func adjustTable(data [][]string) [][]string {
	r := strings.NewReplacer("\r\n", "<br>", "\n", "<br>", "\r", "<br>")
	w := make([]int, len(data[0]))
	for i := range data {
		for j := range w {
			l := runewidth.StringWidth(r.Replace(data[i][j]))
			if l > w[j] {
				w[j] = l
			}
		}
	}
	for i := range data {
		for j := range w {
			if i == 1 {
				data[i][j] = strings.Repeat("-", w[j])
			} else {
				data[i][j] = fmt.Sprintf(fmt.Sprintf("%%-%ds", w[j]), r.Replace(data[i][j]))
			}
		}
	}

	return data
}

func (m *Md) addNumberToTable(data [][]string) [][]string {
	w := len(data[0])/10 + 1

	for i, r := range data {
		switch i {
		case 0:
			r = append([]string{m.config.MergedDict.Lookup("#")}, r...)
		case 1:
			r = append([]string{strings.Repeat("-", w)}, r...)
		default:
			r = append([]string{strconv.Itoa(i - 1)}, r...)
		}
		data[i] = r
	}

	return data
}

func outputExists(s *schema.Schema, path string) bool {
	// README.md
	if _, err := os.Lstat(filepath.Join(path, "README.md")); err == nil {
		return true
	}
	// tables
	for _, t := range s.Tables {
		if _, err := os.Lstat(filepath.Join(path, fmt.Sprintf("%s.md", t.Name))); err == nil {
			return true
		}
	}
	return false
}
