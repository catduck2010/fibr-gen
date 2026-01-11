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
