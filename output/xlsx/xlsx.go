package xlsx

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
	"github.com/loadoff/excl"
	"github.com/pkg/errors"
)

// Xlsx struct
type Xlsx struct {
	config *config.Config
}

// New return Xlsx
func New(c *config.Config) *Xlsx {
	return &Xlsx{
		config: c,
	}
}

// OutputSchema output Xlsx format for full relation.
func (x *Xlsx) OutputSchema(wr io.Writer, s *schema.Schema) (e error) {
	w, err := excl.Create()
	if err != nil {
		return err
	}
	err = x.createSchemaSheet(w, s)
	if err != nil {
		return err
	}
	for _, t := range s.Tables {
		err = x.createTableSheet(w, t)
		if err != nil {
			return err
		}
	}
	tf, _ := os.CreateTemp("", "tbls.xlsx")
	path := tf.Name()
	defer func() {
		err := tf.Close()
		if err != nil {
			e = err
		}
	}()
	err = w.Save(path)
	if err != nil {
		return err
	}
	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	_, err = wr.Write(b)
	if err != nil {
		return err
	}
	return nil
}

// OutputTable output Xlsx format for table.
func (x *Xlsx) OutputTable(wr io.Writer, t *schema.Table) (e error) {
	w, err := excl.Create()
	if err != nil {
		return err
	}
	err = x.createTableSheet(w, t)
	if err != nil {
		return err
	}
	tf, _ := os.CreateTemp("", "tbls.xlsx")
	path := tf.Name()
	defer func() {
		err := tf.Close()
		if err != nil {
			e = err
		}
	}()
	err = w.Save(path)
	if err != nil {
		return err
	}
	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	_, err = wr.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func (x *Xlsx) createSchemaSheet(w *excl.Workbook, s *schema.Schema) error {
	sheetName := fmt.Sprintf("%s %s", x.config.MergedDict.Lookup("Tables of"), s.Name)
	if utf8.RuneCountInString(x.config.MergedDict.Lookup(sheetName)) > 31 { // MS Excel assumes a maximum length of 31 characters for sheet name
		sheetName = "Tables"
	}
	sheet, err := w.OpenSheet(x.config.MergedDict.Lookup(sheetName))
	defer sheet.Close()
	if err != nil {
		return errors.WithStack(err)
	}
	setString(sheet, 1, 1, s.Name).SetFont(excl.Font{Bold: true})

	setString(sheet, 3, 1, x.config.MergedDict.Lookup("Tables")).SetFont(excl.Font{Bold: true})
	setHeader(sheet, 4, []string{
		x.config.MergedDict.Lookup("Name"),
		x.config.MergedDict.Lookup("Columns"),
		x.config.MergedDict.Lookup("Comment"),
		x.config.MergedDict.Lookup("Type"),
	})
	n := 5
	for i, t := range s.Tables {
		setStringWithBorder(sheet, n+i, 1, t.Name)
		setNumberWithBorder(sheet, n+i, 2, len(t.Columns))
		setStringWithBorder(sheet, n+i, 3, t.Comment)
		setStringWithBorder(sheet, n+i, 4, t.Type)
	}

	return nil
}

func (x *Xlsx) adjustColumnHeader(columnValues *[]string, hasColumn bool, name string) {
	if hasColumn {
		*columnValues = append(*columnValues, x.config.MergedDict.Lookup(name))
	}
}

func adjustData(hasData bool, sheet *excl.Sheet, row int, column int, value string) int {
	if hasData {
		setStringWithBorder(sheet, row, column, value)
		return column + 1
	}
	return column
}

func (x *Xlsx) createTableSheet(w *excl.Workbook, t *schema.Table) (e error) {
	sheetName := t.Name
	if utf8.RuneCountInString(sheetName) > 31 { // MS Excel assumes a maximum length of 31 characters for sheet name
		r := []rune(sheetName)
		sheetName = string(r[0:31])
	}
	sheet, err := w.OpenSheet(sheetName)
	defer func() {
		err := sheet.Close()
		if err != nil {
			e = err
		}
	}()
	if err != nil {
		return errors.WithStack(err)
	}

	setString(sheet, 1, 1, t.Name).SetFont(excl.Font{Bold: true})
	setString(sheet, 2, 1, t.Comment)

	setString(sheet, 4, 1, x.config.MergedDict.Lookup("Columns")).SetFont(excl.Font{Bold: true})
	columnValues := []string{
		x.config.MergedDict.Lookup("Name"),
		x.config.MergedDict.Lookup("Type"),
		x.config.MergedDict.Lookup("Default"),
		x.config.MergedDict.Lookup("Nullable"),
	}
	x.adjustColumnHeader(&columnValues, t.HasColumnWithExtraDef(), "Extra Definition")
	x.adjustColumnHeader(&columnValues, t.HasColumnWithOccurrences(), "Occurrences")
	x.adjustColumnHeader(&columnValues, t.HasColumnWithPercents(), "Percents")
	x.adjustColumnHeader(&columnValues, t.HasColumnWithChildren(), "Children")
	x.adjustColumnHeader(&columnValues, t.HasColumnWithParents(), "Parents")
	x.adjustColumnHeader(&columnValues, t.HasColumnWithComment(), "Comment")
	setHeader(sheet, 5, columnValues)
	r := 6
	for i, c := range t.Columns {
		setStringWithBorder(sheet, r+i, 1, c.Name)
		setStringWithBorder(sheet, r+i, 2, c.Type)
		setStringWithBorder(sheet, r+i, 3, c.Default.String)
		setStringWithBorder(sheet, r+i, 4, fmt.Sprintf("%v", c.Nullable))
		ci := 5
		ci = adjustData(t.HasColumnWithExtraDef(), sheet, r+i, ci, fmt.Sprintf("%v", c.ExtraDef))
		ci = adjustData(t.HasColumnWithOccurrences(), sheet, r+i, ci, fmt.Sprintf("%d", c.Occurrences.Int32))
		ci = adjustData(t.HasColumnWithPercents(), sheet, r+i, ci, fmt.Sprintf("%.1f", c.Percents.Float64))
		children := []string{}
		for _, child := range c.ChildRelations {
			children = append(children, child.Table.Name)
		}
		ci = adjustData(t.HasColumnWithChildren(), sheet, r+i, ci, strings.Join(children, "\n"))
		parents := []string{}
		for _, parent := range c.ParentRelations {
			parents = append(parents, parent.ParentTable.Name)
		}
		ci = adjustData(t.HasColumnWithParents(), sheet, r+i, ci, strings.Join(parents, "\n"))
		ci = adjustData(t.HasColumnWithComment(), sheet, r+i, ci, c.Comment)
	}
	r = r + len(t.Columns)

	if len(t.Constraints) > 0 {
		r++
		setString(sheet, r, 1, x.config.MergedDict.Lookup("Constraints")).SetFont(excl.Font{Bold: true})
		r++
		setHeader(sheet, r, []string{
			x.config.MergedDict.Lookup("Name"),
			x.config.MergedDict.Lookup("Type"),
			x.config.MergedDict.Lookup("Definition"),
		})
		r++
		for i, c := range t.Constraints {
			setStringWithBorder(sheet, r+i, 1, c.Name)
			setStringWithBorder(sheet, r+i, 2, c.Type)
			setStringWithBorder(sheet, r+i, 3, c.Def)
		}
	}
	r = r + len(t.Constraints)

	if len(t.Indexes) > 0 {
		r++
		setString(sheet, r, 1, x.config.MergedDict.Lookup("Indexes")).SetFont(excl.Font{Bold: true})
		r++
		setHeader(sheet, r, []string{
			x.config.MergedDict.Lookup("Name"),
			x.config.MergedDict.Lookup("Definition"),
		})
		r++
		for i, idx := range t.Indexes {
			setStringWithBorder(sheet, r+i, 1, idx.Name)
			setStringWithBorder(sheet, r+i, 2, idx.Def)
		}
	}
	r = r + len(t.Indexes)

	if len(t.Triggers) > 0 {
		r++
		setString(sheet, r, 1, x.config.MergedDict.Lookup("Triggers")).SetFont(excl.Font{Bold: true})
		r++
		setHeader(sheet, r, []string{
			x.config.MergedDict.Lookup("Name"),
			x.config.MergedDict.Lookup("Definition"),
		})
		r++
		for i, trg := range t.Triggers {
			setStringWithBorder(sheet, r+i, 1, trg.Name)
			setStringWithBorder(sheet, r+i, 2, trg.Def)
		}
	}

	return nil
}

func setHeader(sheet *excl.Sheet, rowNo int, values []string) {
	for i, v := range values {
		sheet.SetColWidth(10, i+1)
		setStringWithBorder(sheet, rowNo, i+1, v).SetFont(excl.Font{Bold: true})
	}
}

func setNumber(sheet *excl.Sheet, rowNo int, colNo int, v int) *excl.Cell {
	row := sheet.GetRow(rowNo)
	return row.SetNumber(v, colNo)
}

func setNumberWithBorder(sheet *excl.Sheet, rowNo int, colNo int, v int) *excl.Cell {
	return setNumber(sheet, rowNo, colNo, v).SetBorder(excl.Border{
		Left:   &excl.BorderSetting{Style: "thin"},
		Right:  &excl.BorderSetting{Style: "thin"},
		Top:    &excl.BorderSetting{Style: "thin"},
		Bottom: &excl.BorderSetting{Style: "thin"},
	})
}

func setString(sheet *excl.Sheet, rowNo int, colNo int, v string) *excl.Cell {
	row := sheet.GetRow(rowNo)
	return row.SetString(v, colNo)
}

func setStringWithBorder(sheet *excl.Sheet, rowNo int, colNo int, v string) *excl.Cell {
	return setString(sheet, rowNo, colNo, v).SetBorder(excl.Border{
		Left:   &excl.BorderSetting{Style: "thin"},
		Right:  &excl.BorderSetting{Style: "thin"},
		Top:    &excl.BorderSetting{Style: "thin"},
		Bottom: &excl.BorderSetting{Style: "thin"},
	})
}

func setFormula(sheet *excl.Sheet, rowNo int, colNo int, v string) *excl.Cell {
	row := sheet.GetRow(rowNo)
	return row.SetFormula(v, colNo)
}

func setFormulaWithBorder(sheet *excl.Sheet, rowNo int, colNo int, v string) *excl.Cell {
	return setFormula(sheet, rowNo, colNo, v).SetBorder(excl.Border{
		Left:   &excl.BorderSetting{Style: "thin"},
		Right:  &excl.BorderSetting{Style: "thin"},
		Top:    &excl.BorderSetting{Style: "thin"},
		Bottom: &excl.BorderSetting{Style: "thin"},
	})
}
