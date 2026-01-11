package core

import "github.com/xuri/excelize/v2"

// ExcelFile abstracts workbook operations to decouple generator logic from excelize.
type ExcelFile interface {
	Close() error
	CopySheet(from, to int) error
	DeleteSheet(name string)
	GetCellStyle(sheet, cell string) (int, error)
	GetCellValue(sheet, cell string) (string, error)
	GetSheetDimension(sheet string) (string, error)
	GetSheetIndex(name string) (int, error)
	InsertCols(sheet, col string, columns int) error
	InsertRows(sheet string, row, rows int) error
	MergeCell(sheet, hcell, vcell string) error
	GetMergeCells(sheet string) ([]excelize.MergeCell, error)
	NewSheet(name string) (int, error)
	SaveAs(name string) error
	SetCellStyle(sheet, hcell, vcell string, styleID int) error
	SetCellValue(sheet, cell string, value interface{}) error
	GetSheetList() []string
	SetActiveSheet(index int)
	SetSelection(sheetName, cell string) error
}

type ExcelizeFile struct {
	file *excelize.File
}

func openExcelFile(path string) (ExcelFile, error) {
	file, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	return &ExcelizeFile{file: file}, nil
}

func (e *ExcelizeFile) Close() error {
	return e.file.Close()
}

func (e *ExcelizeFile) CopySheet(from, to int) error {
	return e.file.CopySheet(from, to)
}

func (e *ExcelizeFile) DeleteSheet(name string) {
	e.file.DeleteSheet(name)
}

func (e *ExcelizeFile) GetCellStyle(sheet, cell string) (int, error) {
	return e.file.GetCellStyle(sheet, cell)
}

func (e *ExcelizeFile) GetCellValue(sheet, cell string) (string, error) {
	return e.file.GetCellValue(sheet, cell)
}

func (e *ExcelizeFile) GetSheetDimension(sheet string) (string, error) {
	return e.file.GetSheetDimension(sheet)
}

func (e *ExcelizeFile) GetSheetIndex(name string) (int, error) {
	return e.file.GetSheetIndex(name)
}

func (e *ExcelizeFile) InsertCols(sheet, col string, columns int) error {
	return e.file.InsertCols(sheet, col, columns)
}

func (e *ExcelizeFile) InsertRows(sheet string, row, rows int) error {
	return e.file.InsertRows(sheet, row, rows)
}

func (e *ExcelizeFile) MergeCell(sheet, hcell, vcell string) error {
	return e.file.MergeCell(sheet, hcell, vcell)
}

func (e *ExcelizeFile) GetMergeCells(sheet string) ([]excelize.MergeCell, error) {
	return e.file.GetMergeCells(sheet)
}

func (e *ExcelizeFile) NewSheet(name string) (int, error) {
	return e.file.NewSheet(name)
}

func (e *ExcelizeFile) SaveAs(name string) error {
	return e.file.SaveAs(name)
}

func (e *ExcelizeFile) SetCellStyle(sheet, hcell, vcell string, styleID int) error {
	return e.file.SetCellStyle(sheet, hcell, vcell, styleID)
}

func (e *ExcelizeFile) SetCellValue(sheet, cell string, value interface{}) error {
	return e.file.SetCellValue(sheet, cell, value)
}

func (e *ExcelizeFile) GetSheetList() []string {
	return e.file.GetSheetList()
}

func (e *ExcelizeFile) SetActiveSheet(index int) {
	e.file.SetActiveSheet(index)
}

func (e *ExcelizeFile) SetSelection(sheetName, cell string) error {
	// Set active cell and selection to the specified cell (e.g., "A1") using SetPanes
	// We try to preserve existing panes if possible
	panes, err := e.file.GetPanes(sheetName)
	if err == nil {
		// Update selection in existing panes
		panes.Selection = []excelize.Selection{
			{
				ActiveCell: cell,
				SQRef:      cell,
			},
		}
		// If panes are frozen/split, we need to ensure ActivePane is set correctly if needed,
		// but simply updating Selection list should be enough for basic cases.
		// However, SetPanes expects all fields. GetPanes returns them.
		return e.file.SetPanes(sheetName, &panes)
	}

	// Fallback if GetPanes fails (e.g. no panes set yet? or error)
	// If GetPanes returns error, it might mean no panes are set, or sheet doesn't exist.
	// We assume no panes.
	return e.file.SetPanes(sheetName, &excelize.Panes{
		Freeze: false,
		Split:  false,
		Selection: []excelize.Selection{
			{
				ActiveCell: cell,
				SQRef:      cell,
			},
		},
	})
}
