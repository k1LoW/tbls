package xlsx

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/k1LoW/tbls/schema"
	"github.com/loadoff/excl"
	"github.com/pkg/errors"
)

// Xlsx struct
type Xlsx struct{}

// OutputSchema output JSON format for full relation.
func (x *Xlsx) OutputSchema(wr io.Writer, s *schema.Schema) error {
	w, err := excl.Create()
	if err != nil {
		return err
	}
	err = createSchemaSheet(w, s)
	if err != nil {
		return err
	}
	for _, t := range s.Tables {
		err = createTableSheet(w, t)
		if err != nil {
			return err
		}
	}
	tf, _ := ioutil.TempFile("", "tbls.xlsx")
	path := tf.Name()
	defer tf.Close()
	w.Save(path)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = wr.Write(b)
	if err != nil {
		return err
	}
	return nil
}

// OutputTable output dot format for table.
func (x *Xlsx) OutputTable(wr io.Writer, t *schema.Table) error {
	w, err := excl.Create()
	if err != nil {
		return err
	}
	err = createTableSheet(w, t)
	if err != nil {
		return err
	}
	tf, _ := ioutil.TempFile("", "tbls.xlsx")
	path := tf.Name()
	defer tf.Close()
	w.Save(path)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = wr.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func createSchemaSheet(w *excl.Workbook, s *schema.Schema) error {
	sheetName := fmt.Sprintf("Tables of %s", s.Name)
	sheet, err := w.OpenSheet(sheetName)
	defer sheet.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	setString(sheet, 1, 1, s.Name).SetFont(excl.Font{Bold: true})

	setString(sheet, 3, 1, "Tables").SetFont(excl.Font{Bold: true})
	setHeader(sheet, 4, []string{"Name", "Columns", "Comment", "Type"})
	n := 5
	for i, t := range s.Tables {
		setStringWithBorder(sheet, n+i, 1, t.Name)
		setNumberWithBorder(sheet, n+i, 2, len(t.Columns))
		setStringWithBorder(sheet, n+i, 3, t.Comment)
		setStringWithBorder(sheet, n+i, 4, t.Type)
	}

	return nil
}

func createTableSheet(w *excl.Workbook, t *schema.Table) error {
	sheetName := t.Name
	sheet, err := w.OpenSheet(sheetName)
	defer sheet.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	setString(sheet, 1, 1, t.Name).SetFont(excl.Font{Bold: true})
	setString(sheet, 2, 1, t.Comment)

	setString(sheet, 4, 1, "Columns").SetFont(excl.Font{Bold: true})
	setHeader(sheet, 5, []string{"Name", "Type", "Default", "Nullable", "Children", "Parents", "Comment"})
	r := 6
	for i, c := range t.Columns {
		setStringWithBorder(sheet, r+i, 1, c.Name)
		setStringWithBorder(sheet, r+i, 2, c.Type)
		setStringWithBorder(sheet, r+i, 3, c.Default.String)
		setStringWithBorder(sheet, r+i, 4, fmt.Sprintf("%v", c.Nullable))
		children := []string{}
		for _, child := range c.ChildRelations {
			children = append(children, child.Table.Name)
		}
		setStringWithBorder(sheet, r+i, 5, strings.Join(children, "\n"))
		parents := []string{}
		for _, parent := range c.ParentRelations {
			parents = append(parents, parent.ParentTable.Name)
		}
		setStringWithBorder(sheet, r+i, 6, strings.Join(parents, "\n"))
		setStringWithBorder(sheet, r+i, 7, c.Comment)
	}
	r = r + len(t.Columns)

	if len(t.Constraints) > 0 {
		r++
		setString(sheet, r, 1, "Constraints").SetFont(excl.Font{Bold: true})
		r++
		setHeader(sheet, r, []string{"Name", "Type", "Definition"})
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
		setString(sheet, r, 1, "Indexes").SetFont(excl.Font{Bold: true})
		r++
		setHeader(sheet, r, []string{"Name", "Definition"})
		r++
		for i, idx := range t.Indexes {
			setStringWithBorder(sheet, r+i, 1, idx.Name)
			setStringWithBorder(sheet, r+i, 2, idx.Def)
		}
	}
	r = r + len(t.Indexes)

	if len(t.Triggers) > 0 {
		r++
		setString(sheet, r, 1, "Triggers").SetFont(excl.Font{Bold: true})
		r++
		setHeader(sheet, r, []string{"Name", "Definition"})
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
