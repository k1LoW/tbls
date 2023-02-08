package md

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/output"
	"github.com/k1LoW/tbls/output/mermaid"
	"github.com/k1LoW/tbls/schema"
	"github.com/mattn/go-runewidth"
	"github.com/pkg/errors"
	"github.com/pmezard/go-difflib/difflib"
	"gitlab.com/golang-commonmark/mdurl"
)

var mdEscRep = strings.NewReplacer("`", "\\`")

//go:embed templates/*
var tmpl embed.FS

// Md struct
type Md struct {
	config *config.Config
	er     bool
	tmpl   embed.FS
}

// New return Md
func New(c *config.Config, er bool) *Md {
	return &Md{
		config: c,
		er:     er,
		tmpl:   tmpl,
	}
}

func (m *Md) indexTemplate() (string, error) {
	if len(m.config.Templates.MD.Index) > 0 {
		tb, err := os.ReadFile(m.config.Templates.MD.Index)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	} else {
		tb, err := m.tmpl.ReadFile("templates/index.md.tmpl")
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	}
}

func (m *Md) tableTemplate() (string, error) {
	if len(m.config.Templates.MD.Table) > 0 {
		tb, err := os.ReadFile(m.config.Templates.MD.Table)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
	} else {
		tb, err := m.tmpl.ReadFile("templates/table.md.tmpl")
		if err != nil {
			return "", errors.WithStack(err)
		}
		return string(tb), nil
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
	groupErDiagrams := map[string]string{}
	switch m.config.ER.Format {
	case "mermaid":
		buf := new(bytes.Buffer)
		mmd := mermaid.New(m.config)
		if err := mmd.OutputSchema(buf, s); err != nil {
			return err
		}
		templateData["er"] = !m.config.ER.Skip
		templateData["erDiagram"] = fmt.Sprintf("```mermaid\n%s```", buf.String())
		for _, tableGroup := range m.config.Format.TableGroups {
			buf := new(bytes.Buffer)
			groupSchema, err := s.NewSchemaForTableGroup(tableGroup.Name, tableGroup.Tables)
			if err != nil {
				return errors.WithStack(err)
			}
			if err := mmd.OutputSchema(buf, groupSchema); err != nil {
				return err
			}
			groupErDiagrams[tableGroup.Name] = fmt.Sprintf("```mermaid\n%s```", buf.String())
		}
	default:
		if m.er {
			templateData["er"] = !m.config.ER.Skip
		} else {
			templateData["er"] = false
		}
		templateData["erDiagram"] = fmt.Sprintf("![er](%sschema.%s)", m.config.BaseUrl, m.config.ER.Format)
		for _, tableGroup := range m.config.Format.TableGroups {
			groupErDiagrams[tableGroup.Name] = fmt.Sprintf("![er](%s%s_group_schema.%s)", m.config.BaseUrl, tableGroup.Name, m.config.ER.Format)
		}
	}

	templateData["groupErDiagrams"] = groupErDiagrams
	templateData["showOnlyFirstParagraph"] = m.config.Format.ShowOnlyFirstParagraph
	err = tmpl.Execute(wr, templateData)
	if err != nil {
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
	switch m.config.ER.Format {
	case "mermaid":
		buf := new(bytes.Buffer)
		mmd := mermaid.New(m.config)
		if err := mmd.OutputTable(buf, t); err != nil {
			return err
		}
		templateData["er"] = !m.config.ER.Skip
		templateData["erDiagram"] = fmt.Sprintf("```mermaid\n%s```", buf.String())
	default:
		if m.er {
			templateData["er"] = !m.config.ER.Skip
		} else {
			templateData["er"] = false
		}
		templateData["erDiagram"] = fmt.Sprintf("![er](%s%s.%s)", m.config.BaseUrl, mdurl.Encode(t.Name), m.config.ER.Format)
	}

	err = tmpl.Execute(wr, templateData)
	if err != nil {
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
	file, err := os.Create(filepath.Clean(filepath.Join(fullPath, "README.md")))
	defer func() {
		err := file.Close()
		if err != nil {
			e = err
		}
	}()
	if err != nil {
		return errors.WithStack(err)
	}
	er := false
	if _, err := os.Lstat(filepath.Join(fullPath, fmt.Sprintf("schema.%s", c.ER.Format))); err == nil {
		er = true
	}

	md := New(c, er)

	err = md.OutputSchema(file, s)
	if err != nil {
		return errors.WithStack(err)
	}
	fmt.Printf("%s\n", filepath.Join(docPath, "README.md"))

	// tables
	for _, t := range s.Tables {
		file, err := os.Create(filepath.Clean(filepath.Join(fullPath, fmt.Sprintf("%s.md", t.Name))))
		if err != nil {
			_ = file.Close()
			return errors.WithStack(err)
		}

		er := false
		if _, err := os.Lstat(filepath.Join(fullPath, fmt.Sprintf("%s.%s", t.Name, c.ER.Format))); err == nil {
			er = true
		}

		md := New(c, er)

		err = md.OutputTable(file, t)
		if err != nil {
			_ = file.Close()
			return errors.WithStack(err)
		}
		fmt.Printf("%s\n", filepath.Join(docPath, fmt.Sprintf("%s.md", t.Name)))
		err = file.Close()
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// DiffSchemas show diff databases.
func DiffSchemas(s, s2 *schema.Schema, c, c2 *config.Config) (string, error) {
	var diff string
	md := New(c, false)

	// README.md
	a := new(bytes.Buffer)
	if err := md.OutputSchema(a, s); err != nil {
		return "", errors.WithStack(err)
	}

	b := new(bytes.Buffer)
	if err := md.OutputSchema(b, s2); err != nil {
		return "", errors.WithStack(err)
	}

	mdsnA, err := c.MaskedDSN()
	if err != nil {
		return "", errors.WithStack(err)
	}
	from := fmt.Sprintf("tbls doc %s", mdsnA)

	mdsnB, err := c.MaskedDSN()
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
			if err := md.OutputTable(b, t2); err != nil {
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
		if err := md.OutputTable(b, t); err != nil {
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

	// README.md
	er := false
	if _, err := os.Lstat(filepath.Join(fullPath, fmt.Sprintf("schema.%s", c.ER.Format))); err == nil {
		er = true
	}

	md := New(c, er)

	b := new(bytes.Buffer)
	err = md.OutputSchema(b, s)
	if err != nil {
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
	diffed := map[string]struct{}{
		"README.md": struct{}{},
	}
	for _, t := range s.Tables {
		b := new(bytes.Buffer)
		er := false
		if _, err := os.Lstat(filepath.Join(fullPath, fmt.Sprintf("%s.%s", t.Name, c.ER.Format))); err == nil {
			er = true
		}
		to := fmt.Sprintf("%s %s", mdsn, t.Name)

		md := New(c, er)

		err := md.OutputTable(b, t)
		if err != nil {
			return "", errors.WithStack(err)
		}
		targetPath := filepath.Join(fullPath, fmt.Sprintf("%s.md", t.Name))
		diffed[fmt.Sprintf("%s.md", t.Name)] = struct{}{}
		a, err := os.ReadFile(filepath.Clean(targetPath))
		if err != nil {
			a = []byte{}
		}
		from := filepath.Join(docPath, fmt.Sprintf("%s.md", t.Name))

		d := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(a)),
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
	files, _ := os.ReadDir(fullPath)
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

func containsTableWithLabels(tables []*schema.Table) bool {
	for _, t := range tables {
		if len(t.Labels) > 0 {
			return true
		}
	}
	return false
}

func (m *Md) makeTablesData(tables []*schema.Table) [][]string {
	var tablesData [][]string

	tablesHeader := []string{
		m.config.MergedDict.Lookup("Name"),
		m.config.MergedDict.Lookup("Columns"),
		m.config.MergedDict.Lookup("Comment"),
		m.config.MergedDict.Lookup("Type"),
	}
	tablesHeaderLine := []string{"----", "-------", "-------", "----"}

	if containsTableWithLabels(tables) {
		tablesHeader = append(tablesHeader, m.config.MergedDict.Lookup("Labels"))
		tablesHeaderLine = append(tablesHeaderLine, "------")
	}

	tablesData = append(tablesData,
		tablesHeader,
		tablesHeaderLine,
	)

	for _, t := range tables {
		comment := t.Comment
		if m.config.Format.ShowOnlyFirstParagraph {
			comment = output.ShowOnlyFirstParagraph(comment)
		}
		data := []string{
			fmt.Sprintf("[%s](%s%s.md)", t.Name, m.config.BaseUrl, mdurl.Encode(t.Name)),
			fmt.Sprintf("%d", len(t.Columns)),
			comment,
			t.Type,
		}
		if containsTableWithLabels(tables) {
			data = append(data, output.LabelJoin(t.Labels))
		}

		tablesData = append(tablesData, data)
	}

	return tablesData
}

func (m *Md) makeSchemaTemplateData(s *schema.Schema) map[string]interface{} {
	number := m.config.Format.Number
	adjust := m.config.Format.Adjust

	tableGroupsMap := map[string][]string{}
	for _, tableGroup := range m.config.Format.TableGroups {
		for _, table := range tableGroup.Tables {
			tableGroupsMap[table] = append(tableGroupsMap[table], tableGroup.Name)
		}
	}

	groupTables := map[string][]*schema.Table{}
	var outsideGroupTables []*schema.Table
	for _, t := range s.Tables {
		if tableGroups, ok := tableGroupsMap[t.Name]; ok {
			for _, tableGroup := range tableGroups {
				groupTables[tableGroup] = append(groupTables[tableGroup], t)
			}
		} else {
			outsideGroupTables = append(outsideGroupTables, t)
		}
	}

	allTablesData := m.makeTablesData(s.Tables)
	groupTablesData := map[string][][]string{}
	for groupName, tables := range groupTables {
		groupTablesData[groupName] = m.makeTablesData(tables)
	}
	outsideGroupTablesData := m.makeTablesData(outsideGroupTables)

	if number {
		allTablesData = m.addNumberToTable(allTablesData)
		for _, tablesData := range groupTablesData {
			tablesData = m.addNumberToTable(tablesData)
		}
		outsideGroupTablesData = m.addNumberToTable(outsideGroupTablesData)
	}

	tablesSubroutineData := [][]string{}
	tablesSubroutineHeader := []string{
		m.config.MergedDict.Lookup("Name"),
		m.config.MergedDict.Lookup("ReturnType"),
		m.config.MergedDict.Lookup("Arguments"),
		m.config.MergedDict.Lookup("Type"),
	}
	tablesSubroutineHeaderLine := []string{"----", "-------", "-------", "----"}
	tablesSubroutineData = append(tablesSubroutineData,
		tablesSubroutineHeader,
		tablesSubroutineHeaderLine,
	)

	for _, t := range s.Functions {
		data := []string{
			t.Name,
			t.ReturnType,
			t.Arguments,
			t.Type,
		}
		tablesSubroutineData = append(tablesSubroutineData, data)
	}

	if adjust {
		allTablesData = adjustTable(allTablesData)
		for _, tablesData := range groupTablesData {
			tablesData = adjustTable(tablesData)
		}
		outsideGroupTablesData = adjustTable(outsideGroupTablesData)
		tablesSubroutineData = adjustTable(tablesSubroutineData)
	}

	var tableGroupNames []string
	for _, tableGroup := range m.config.Format.TableGroups {
		if _, ok := groupTablesData[tableGroup.Name]; ok {
			tableGroupNames = append(tableGroupNames, tableGroup.Name)
		}
	}

	return map[string]interface{}{
		"Schema":             s,
		"AllTables":          allTablesData,
		"TableGroupNames":    tableGroupNames,
		"GroupTables":        groupTablesData,
		"OutsideGroupTables": outsideGroupTablesData,
		"Functions":          tablesSubroutineData,
	}
}

func (m *Md) makeTableTemplateData(t *schema.Table) map[string]interface{} {
	number := m.config.Format.Number
	adjust := m.config.Format.Adjust
	hideColumns := m.config.Format.HideColumnsWithoutValues

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
			childRelations = append(childRelations, fmt.Sprintf("[%s](%s%s.md)", r.Table.Name, m.config.BaseUrl, mdurl.Encode(r.Table.Name)))
			cEncountered[r.Table.Name] = true
		}
		parentRelations := []string{}
		pEncountered := map[string]bool{}
		for _, r := range c.ParentRelations {
			if _, ok := pEncountered[r.ParentTable.Name]; ok {
				continue
			}
			parentRelations = append(parentRelations, fmt.Sprintf("[%s](%s%s.md)", r.ParentTable.Name, m.config.BaseUrl, mdurl.Encode(r.ParentTable.Name)))
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
		adjustData(&data, t.ShowColumn(schema.ColumnComment, hideColumns), c.Comment)
		adjustData(&data, t.ShowColumn(schema.ColumnLabels, hideColumns), output.LabelJoin(c.Labels))
		columnsData = append(columnsData, data)
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
	referencedTables := []string{}
	for _, rt := range t.ReferencedTables {
		if rt.External {
			referencedTables = append(referencedTables, rt.Name)
			continue
		}
		referencedTables = append(referencedTables, fmt.Sprintf("[%s](%s%s.md)", rt.Name, m.config.BaseUrl, mdurl.Encode(rt.Name)))
	}

	if number {
		columnsData = m.addNumberToTable(columnsData)
		constraintsData = m.addNumberToTable(constraintsData)
		indexesData = m.addNumberToTable(indexesData)
		triggersData = m.addNumberToTable(triggersData)
	}

	if adjust {
		return map[string]interface{}{
			"Table":            t,
			"Columns":          adjustTable(columnsData),
			"Constraints":      adjustTable(constraintsData),
			"Indexes":          adjustTable(indexesData),
			"Triggers":         adjustTable(triggersData),
			"ReferencedTables": referencedTables,
		}
	}

	return map[string]interface{}{
		"Table":            t,
		"Columns":          columnsData,
		"Constraints":      constraintsData,
		"Indexes":          indexesData,
		"Triggers":         triggersData,
		"ReferencedTables": referencedTables,
	}
}

func (m *Md) adjustColumnHeader(columnsHeader *[]string, columnsHeaderLine *[]string, hasColumn bool, name string) {
	if hasColumn {
		*columnsHeader = append(*columnsHeader, m.config.MergedDict.Lookup(name))
		*columnsHeaderLine = append(*columnsHeaderLine, strings.Repeat("-", runewidth.StringWidth(m.config.MergedDict.Lookup(name))))
	}
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
		switch {
		case i == 0:
			r = append([]string{m.config.MergedDict.Lookup("#")}, r...)
		case i == 1:
			r = append([]string{strings.Repeat("-", w)}, r...)
		default:
			r = append([]string{strconv.Itoa(i - 1)}, r...)
		}
		data[i] = r
	}

	return data
}
