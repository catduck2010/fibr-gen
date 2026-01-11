package core

import (
	"fmt"
	"testing"

	"github.com/xuri/excelize/v2"
)

// TestExcelizeFile_BasicOperations tests the basic wrapper methods of ExcelizeFile.
// Since ExcelizeFile is a thin wrapper around excelize, this test primarily ensures
// that the delegation is correctly wired and the interface contract is met.
func TestExcelizeFile_BasicOperations(t *testing.T) {
	// 1. Setup: Create an in-memory excelize file and wrap it
	f := excelize.NewFile()
	adapter := &ExcelizeFile{file: f}
	defer adapter.Close()

	sheet := "Sheet1"
	// Ensure sheet exists (default in new file)
	idx, err := adapter.GetSheetIndex(sheet)
	if err != nil {
		t.Fatalf("GetSheetIndex failed: %v", err)
	}
	if idx == -1 {
		// create if missing (though NewFile creates Sheet1)
		_, err = adapter.NewSheet(sheet)
		if err != nil {
			t.Fatalf("NewSheet failed: %v", err)
		}
	}

	// 2. Test Set/Get CellValue
	cell := "A1"
	val := "Hello World"
	if err := adapter.SetCellValue(sheet, cell, val); err != nil {
		t.Errorf("SetCellValue failed: %v", err)
	}

	got, err := adapter.GetCellValue(sheet, cell)
	if err != nil {
		t.Errorf("GetCellValue failed: %v", err)
	}
	if got != val {
		t.Errorf("GetCellValue = %q, want %q", got, val)
	}

	// 3. Test NewSheet & DeleteSheet
	newSheet := "TestSheet"
	if _, err := adapter.NewSheet(newSheet); err != nil {
		t.Errorf("NewSheet failed: %v", err)
	}
	idx, err = adapter.GetSheetIndex(newSheet)
	if err != nil || idx == -1 {
		t.Errorf("GetSheetIndex for new sheet failed or not found")
	}

	adapter.DeleteSheet(newSheet)
	idx, err = adapter.GetSheetIndex(newSheet)
	if err == nil && idx != -1 {
		t.Errorf("DeleteSheet failed, sheet still exists index: %d", idx)
	}

	// 4. Test MergeCell & GetMergeCells
	// Merge B2:C3
	if err := adapter.SetCellValue(sheet, "B2", "Merged"); err != nil {
		t.Fatalf("SetCellValue B2 failed: %v", err)
	}
	if err := adapter.MergeCell(sheet, "B2", "C3"); err != nil {
		t.Errorf("MergeCell failed: %v", err)
	}

	merges, err := adapter.GetMergeCells(sheet)
	if err != nil {
		t.Errorf("GetMergeCells failed: %v", err)
	}
	found := false
	for _, m := range merges {
		// excelize v2.8+ uses GetStartAxis() / GetEndAxis() or direct field access depending on version.
		// Our code uses GetStartAxis/GetEndAxis in generator.go, so we should verify basic range string.
		// Note: The mock/interface returns []excelize.MergeCell.
		// Let's check if the range matches "B2:C3" logic.
		// The exact string representation depends on excelize version implementation of GetRange() or manual check.
		// We'll just check if we have at least one merge and it covers B2.
		if m.GetStartAxis() == "B2" && m.GetEndAxis() == "C3" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected merge B2:C3 not found in %v", merges)
	}

	// 5. Test InsertRows
	// Setup A10=Row10, A11=Row11
	adapter.SetCellValue(sheet, "A10", "Row10")
	adapter.SetCellValue(sheet, "A11", "Row11")

	// Insert 2 rows at 11. Old A11 should move to A13.
	if err := adapter.InsertRows(sheet, 11, 2); err != nil {
		t.Errorf("InsertRows failed: %v", err)
	}

	valA10, _ := adapter.GetCellValue(sheet, "A10")
	valA13, _ := adapter.GetCellValue(sheet, "A13")
	valA11, _ := adapter.GetCellValue(sheet, "A11") // Should be empty

	if valA10 != "Row10" {
		t.Errorf("Row 10 moved unexpectedly? Got %s", valA10)
	}
	if valA13 != "Row11" {
		t.Errorf("Row 11 did not move to Row 13 correctly? Got %s", valA13)
	}
	if valA11 != "" {
		t.Errorf("Row 11 should be empty after insert, got %s", valA11)
	}

	// 6. Test InsertCols
	// Setup E1=ColE
	adapter.SetCellValue(sheet, "E1", "ColE")
	// Insert 1 col at E. Old E1 should move to F1.
	if err := adapter.InsertCols(sheet, "E", 1); err != nil {
		t.Errorf("InsertCols failed: %v", err)
	}
	valF1, _ := adapter.GetCellValue(sheet, "F1")
	if valF1 != "ColE" {
		t.Errorf("Column insert failed, expected 'ColE' at F1, got %s", valF1)
	}
}

func TestExcelizeFile_Dimensions(t *testing.T) {
	f := excelize.NewFile()
	adapter := &ExcelizeFile{file: f}
	sheet := "Sheet1"

	// Initial dimension might be just A1 or empty
	dim, err := adapter.GetSheetDimension(sheet)
	if err != nil {
		t.Errorf("GetSheetDimension failed: %v", err)
	}
	// Just ensure it returns something valid-ish (e.g. "A1:A1" or similar)
	if dim == "" {
		t.Errorf("GetSheetDimension returned empty string")
	}

	// Add data at C5
	adapter.SetCellValue(sheet, "C5", "End")
	dim, _ = adapter.GetSheetDimension(sheet)
	// Expect range ending in C5, e.g., "A1:C5"
	// Exact format depends on excelize, but usually "A1:C5"
	fmt.Printf("Dimension: %s\n", dim)
}
