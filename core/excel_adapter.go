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
	NewSheet(name string) (int, error)
	SaveAs(name string) error
	SetCellStyle(sheet, hcell, vcell string, styleID int) error
	SetCellValue(sheet, cell string, value interface{}) error
}

type excelizeFile struct {
	file *excelize.File
}

func openExcelFile(path string) (ExcelFile, error) {
	file, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	return &excelizeFile{file: file}, nil
}

func (e *excelizeFile) Close() error {
	return e.file.Close()
}

func (e *excelizeFile) CopySheet(from, to int) error {
	return e.file.CopySheet(from, to)
}

func (e *excelizeFile) DeleteSheet(name string) {
	e.file.DeleteSheet(name)
}

func (e *excelizeFile) GetCellStyle(sheet, cell string) (int, error) {
	return e.file.GetCellStyle(sheet, cell)
}

func (e *excelizeFile) GetCellValue(sheet, cell string) (string, error) {
	return e.file.GetCellValue(sheet, cell)
}

func (e *excelizeFile) GetSheetDimension(sheet string) (string, error) {
	return e.file.GetSheetDimension(sheet)
}

func (e *excelizeFile) GetSheetIndex(name string) (int, error) {
	return e.file.GetSheetIndex(name)
}

func (e *excelizeFile) InsertCols(sheet, col string, columns int) error {
	return e.file.InsertCols(sheet, col, columns)
}

func (e *excelizeFile) InsertRows(sheet string, row, rows int) error {
	return e.file.InsertRows(sheet, row, rows)
}

func (e *excelizeFile) NewSheet(name string) (int, error) {
	return e.file.NewSheet(name)
}

func (e *excelizeFile) SaveAs(name string) error {
	return e.file.SaveAs(name)
}

func (e *excelizeFile) SetCellStyle(sheet, hcell, vcell string, styleID int) error {
	return e.file.SetCellStyle(sheet, hcell, vcell, styleID)
}

func (e *excelizeFile) SetCellValue(sheet, cell string, value interface{}) error {
	return e.file.SetCellValue(sheet, cell, value)
}
