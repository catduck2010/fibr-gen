package core

import (
	"fibr-gen/config"
	"fmt"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

// MockFetcher for testing
type MockFetcher struct {
	Data map[string][]map[string]interface{}
}

func (m *MockFetcher) Fetch(viewName string, params map[string]string) ([]map[string]interface{}, error) {
	if data, ok := m.Data[viewName]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("view not found: %s", viewName)
}

func TestExpandableBlock_MergedAxis(t *testing.T) {
	// 1. Setup Excel Template
	f := excelize.NewFile()
	sheet := "Sheet1"
	// Ensure Sheet1 exists and is clean
	idx, _ := f.GetSheetIndex("Sheet1")
	if idx == -1 {
		f.NewSheet(sheet)
	}

	// VAxis Template (Merged A2:B2)
	// A2:B2
	f.SetCellValue(sheet, "A2", "ItemTemplate")
	if err := f.MergeCell(sheet, "A2", "B2"); err != nil {
		t.Fatalf("Failed to merge cells: %v", err)
	}

	// HAxis Template (C1)
	f.SetCellValue(sheet, "C1", "Header1")

	// Data Template (C2)
	f.SetCellValue(sheet, "C2", "ValueTemplate")

	// 2. Setup Config
	vAxisConf := config.BlockConfig{
		Name:        "VAxis",
		Type:        config.BlockTypeAxis,
		Direction:   config.DirectionVertical,
		Range:       config.CellRange{Ref: "A2:B2"}, // Merged Range
		VViewName:   "v_items",
		InsertAfter: true, // Expand Rows
	}

	hAxisConf := config.BlockConfig{
		Name:      "HAxis",
		Type:      config.BlockTypeAxis,
		Direction: config.DirectionHorizontal,
		Range:     config.CellRange{Ref: "C1:C1"},
		VViewName: "v_headers",
	}

	templateBlock := config.BlockConfig{
		Name:  "Template1",
		Type:  config.BlockTypeTag,
		Range: config.CellRange{Ref: "C2:C2"},
	}

	// SubBlocks
	expandBlock := &config.BlockConfig{
		Name:      "ExpandBlock",
		Type:      config.BlockTypeExpand,
		Direction: config.DirectionVertical,
		Range:     config.CellRange{Ref: "A1:C2"}, // Bounding Box
		SubBlocks: []config.BlockConfig{vAxisConf, hAxisConf, templateBlock},
	}

	// 3. Setup Context
	mockData := map[string][]map[string]interface{}{
		"v_items": {
			{"id": "1", "name": "Item1"},
			{"id": "2", "name": "Item2"},
			{"id": "3", "name": "Item3"},
		},
		"v_headers": {
			{"col": "Header1"},
		},
	}
	fetcher := &MockFetcher{Data: mockData}

	vViews := map[string]*config.VirtualViewConfig{
		"v_items":   {Name: "v_items", Tags: []config.TagConfig{{Name: "name", Column: "name"}}},
		"v_headers": {Name: "v_headers", Tags: []config.TagConfig{{Name: "col", Column: "col"}}},
	}
	provider := config.NewMemoryConfigRegistry(vViews)

	ctx := NewGenerationContext(&config.WorkbookConfig{}, provider, fetcher, nil)
	gen := NewGenerator(ctx)

	// 4. Run
	adapter := &ExcelizeFile{file: f}
	err := gen.processBlock(adapter, sheet, expandBlock)
	if err != nil {
		t.Fatalf("processBlock failed: %v", err)
	}

	// 5. Verify Merged Cells
	// We expect 3 items.
	// Item1 at A2:B2 (Original, rewritten)
	// Item2 at A3:B3 (Expanded)
	// Item3 at A4:B4 (Expanded)

	mergedCells, err := f.GetMergeCells(sheet)
	if err != nil {
		t.Fatalf("GetMergeCells failed: %v", err)
	}

	foundA3B3 := false
	foundA4B4 := false

	t.Logf("Found %d merged cells", len(mergedCells))
	for _, mc := range mergedCells {
		ref := mc.GetStartAxis() + ":" + mc.GetEndAxis()
		t.Logf("Merged Cell: %s", ref)
		if ref == "A3:B3" {
			foundA3B3 = true
		}
		if ref == "A4:B4" {
			foundA4B4 = true
		}
	}

	if !foundA3B3 {
		t.Errorf("Expected merged cells at A3:B3, but not found")
	}
	if !foundA4B4 {
		t.Errorf("Expected merged cells at A4:B4, but not found")
	}
}

// --- End-to-End Tests ported from test/ directory ---

// Helper to create Demo Template
func setupTemplate_Demo(t *testing.T) *excelize.File {
	f := excelize.NewFile()
	sheet := "Sheet1"
	idx, _ := f.GetSheetIndex("Sheet1")
	if idx == -1 {
		f.NewSheet(sheet)
	}
	// TitleBlock A1:B1
	// Typically headers or title info
	f.SetCellValue(sheet, "A1", "User:")
	f.SetCellValue(sheet, "B1", "{user_name}")
	return f
}

func TestEndToEnd_DemoReport(t *testing.T) {
	// Config: test/workbooks/demo_report.yaml
	// VView: test/vviews/user_view.yaml
	// Data: test/data_csv/user_view.csv

	// 1. Setup Config
	wbConfig := &config.WorkbookConfig{
		Id:         "wb_demo",
		Name:       "部门人员示例报表",
		Template:   "demo_template.xlsx",
		OutputDir:  "reports",
		Parameters: map[string]string{"report_date": "2023-10-27"},
		Sheets: []config.SheetConfig{
			{
				Name:                "Sheet1",
				Dynamic:             false,
				AllowOverlap:        false,
				VerticalArrangement: true,
				Blocks: []config.BlockConfig{
					{
						Name:      "TitleBlock",
						Type:      config.BlockTypeTag,
						Range:     config.CellRange{Ref: "A1:B1"},
						VViewName: "user_view",
					},
				},
			},
		},
	}

	vViews := map[string]*config.VirtualViewConfig{
		"user_view": {
			Name: "user_view",
			Tags: []config.TagConfig{
				{Name: "dept_code", Column: "DEPT_CD"},
				{Name: "user_name", Column: "USER_NAME"},
			},
		},
	}

	// 2. Setup Data
	mockData := map[string][]map[string]interface{}{
		"user_view": {
			{"DEPT_CD": "D001", "USER_NAME": "Alice"},
			{"DEPT_CD": "D001", "USER_NAME": "Bob"},
			{"DEPT_CD": "D002", "USER_NAME": "Charlie"},
		},
	}

	// 3. Run
	fetcher := &MockFetcher{Data: mockData}
	provider := config.NewMemoryConfigRegistry(vViews)
	ctx := NewGenerationContext(wbConfig, provider, fetcher, nil)
	gen := NewGenerator(ctx)

	f := setupTemplate_Demo(t)
	adapter := &ExcelizeFile{file: f}

	// Process Sheet1
	sheetConf := wbConfig.Sheets[0]
	// Manually process blocks as per Generator.Generate logic (simplified)
	for _, block := range sheetConf.Blocks {
		if err := gen.processBlock(adapter, sheetConf.Name, &block); err != nil {
			t.Fatalf("processBlock failed: %v", err)
		}
	}

	// 4. Verify
	// TitleBlock is a TagBlock, usually takes the first row if multiple?
	// TagBlock implementation iterates all rows if it's a list?
	// TagBlock description in yaml says "TagBlock".
	// If direction is not specified, it might just fill the first one or expand?
	// In demo_report.yaml, direction is NOT specified for TitleBlock.
	// But it has VView.
	// Let's check TagBlock logic in generator.go.
	// If no direction, it defaults to filling template placeholders with first row?
	// Or does it iterate?
	// Wait, standard TagBlock (Type T) usually iterates if there's a direction.
	// If no direction, it might just be a single replacement.

	val, _ := f.GetCellValue("Sheet1", "B1")
	// Expecting "Alice" (first row)
	if val != "Alice" {
		t.Errorf("Expected B1 to be 'Alice', got '%s'", val)
	}
}

// Helper to create TagBlock Template
func setupTemplate_TagBlock(t *testing.T) *excelize.File {
	f := excelize.NewFile()
	sheet := "Sheet1"
	idx, _ := f.GetSheetIndex("Sheet1")
	if idx == -1 {
		f.NewSheet(sheet)
	}
	// EmployeeList A2:C2
	f.SetCellValue(sheet, "A2", "{dept}")
	f.SetCellValue(sheet, "B2", "{name}")
	f.SetCellValue(sheet, "C2", "{salary}")
	return f
}

func TestEndToEnd_TagBlock(t *testing.T) {
	// Config: test/workbooks/tagblock_test.yaml

	wbConfig := &config.WorkbookConfig{
		Id:        "wb_tagblock_test",
		Name:      "TagBlockTest",
		Template:  "tagblock_template.xlsx",
		OutputDir: "tests",
		Sheets: []config.SheetConfig{
			{
				Name:    "Sheet1",
				Dynamic: false,
				Blocks: []config.BlockConfig{
					{
						Name:      "EmployeeList",
						Type:      config.BlockTypeTag,
						Range:     config.CellRange{Ref: "A2:C2"},
						VViewName: "employee_view",
						Direction: config.DirectionVertical,
					},
				},
			},
		},
	}

	vViews := map[string]*config.VirtualViewConfig{
		"employee_view": {
			Name: "employee_view",
			Tags: []config.TagConfig{
				{Name: "dept", Column: "DEPT_CD"},
				{Name: "name", Column: "USER_NAME"},
				{Name: "salary", Column: "SALARY"},
			},
		},
	}

	mockData := map[string][]map[string]interface{}{
		"employee_view": {
			{"DEPT_CD": "D001", "USER_NAME": "Alice", "SALARY": 5000},
			{"DEPT_CD": "D001", "USER_NAME": "Bob", "SALARY": 6000},
			{"DEPT_CD": "D002", "USER_NAME": "Charlie", "SALARY": 7000},
		},
	}

	fetcher := &MockFetcher{Data: mockData}
	provider := config.NewMemoryConfigRegistry(vViews)
	ctx := NewGenerationContext(wbConfig, provider, fetcher, nil)
	gen := NewGenerator(ctx)

	f := setupTemplate_TagBlock(t)
	adapter := &ExcelizeFile{file: f}

	block := &wbConfig.Sheets[0].Blocks[0]
	if err := gen.processBlock(adapter, "Sheet1", block); err != nil {
		t.Fatalf("processBlock failed: %v", err)
	}

	// Verify Expansion
	// Row 2: Alice
	// Row 3: Bob
	// Row 4: Charlie

	val, _ := f.GetCellValue("Sheet1", "B2")
	if val != "Alice" {
		t.Errorf("Row 2 Name: want Alice, got %s", val)
	}
	val, _ = f.GetCellValue("Sheet1", "B3")
	if val != "Bob" {
		t.Errorf("Row 3 Name: want Bob, got %s", val)
	}
	val, _ = f.GetCellValue("Sheet1", "B4")
	if val != "Charlie" {
		t.Errorf("Row 4 Name: want Charlie, got %s", val)
	}
}

// Helper to create Cross Template
func setupTemplate_Cross(t *testing.T) *excelize.File {
	f := excelize.NewFile()
	sheet := "Sheet1"
	idx, _ := f.GetSheetIndex("Sheet1")
	if idx == -1 {
		f.NewSheet(sheet)
	}
	// ExpandBlock A2:B3
	// MonthAxis B2 (Horizontal)
	f.SetCellValue(sheet, "B2", "{month_label}")

	// EmpAxis A3 (Vertical)
	f.SetCellValue(sheet, "A3", "{emp_name}")

	// ScoreData B3 (Intersection)
	f.SetCellValue(sheet, "B3", "{score}")

	return f
}

func TestEndToEnd_CrossTest(t *testing.T) {
	// Config: test/workbooks/cross_test.yaml

	// SubBlocks
	monthAxis := config.BlockConfig{
		Name:        "MonthAxis",
		Type:        config.BlockTypeAxis,
		Direction:   config.DirectionHorizontal,
		InsertAfter: false,
		Range:       config.CellRange{Ref: "B2:B2"},
		VViewName:   "v_full_perf",
		TagVariable: "month_id",
	}

	empAxis := config.BlockConfig{
		Name:        "EmpAxis",
		Type:        config.BlockTypeAxis,
		Direction:   config.DirectionVertical,
		InsertAfter: true,
		Range:       config.CellRange{Ref: "A3:A3"},
		VViewName:   "v_full_perf",
		TagVariable: "emp_id",
	}

	scoreData := config.BlockConfig{
		Name:      "ScoreData",
		Type:      config.BlockTypeTag,
		Range:     config.CellRange{Ref: "B3:B3"},
		VViewName: "v_full_perf",
		Direction: config.DirectionVertical,
		RowLimit:  1, // Important for single cell filling per intersection
	}

	expandBlock := config.BlockConfig{
		Name:      "PerformanceMatrix",
		Type:      config.BlockTypeExpand,
		Range:     config.CellRange{Ref: "A2:B3"},
		SubBlocks: []config.BlockConfig{monthAxis, empAxis, scoreData},
	}

	wbConfig := &config.WorkbookConfig{
		Id:        "wb_cross_test",
		Name:      "CrossTest",
		Template:  "cross_template.xlsx",
		OutputDir: "tests",
		Sheets: []config.SheetConfig{
			{
				Name:    "Sheet1",
				Dynamic: false,
				Blocks:  []config.BlockConfig{expandBlock},
			},
		},
	}

	vViews := map[string]*config.VirtualViewConfig{
		"v_full_perf": {
			Name: "v_full_perf",
			Tags: []config.TagConfig{
				{Name: "emp_id", Column: "EMP_ID"},
				{Name: "emp_name", Column: "EMP_NAME"},
				{Name: "month_id", Column: "MONTH_ID"},
				{Name: "month_label", Column: "MONTH_LABEL"},
				{Name: "score", Column: "SCORE"},
			},
		},
	}

	mockData := map[string][]map[string]interface{}{
		"v_full_perf": {
			{"EMP_ID": "E001", "EMP_NAME": "Alice", "MONTH_ID": "M01", "MONTH_LABEL": "Jan", "SCORE": 85},
			{"EMP_ID": "E001", "EMP_NAME": "Alice", "MONTH_ID": "M02", "MONTH_LABEL": "Feb", "SCORE": 88},
			{"EMP_ID": "E002", "EMP_NAME": "Bob", "MONTH_ID": "M01", "MONTH_LABEL": "Jan", "SCORE": 75},
			{"EMP_ID": "E002", "EMP_NAME": "Bob", "MONTH_ID": "M02", "MONTH_LABEL": "Feb", "SCORE": 78},
		},
	}

	fetcher := &MockFetcher{Data: mockData}
	provider := config.NewMemoryConfigRegistry(vViews)
	ctx := NewGenerationContext(wbConfig, provider, fetcher, nil)
	gen := NewGenerator(ctx)

	f := setupTemplate_Cross(t)
	adapter := &ExcelizeFile{file: f}

	if err := gen.processBlock(adapter, "Sheet1", &expandBlock); err != nil {
		t.Fatalf("processBlock failed: %v", err)
	}

	// Verify
	// Axis H (Month): Jan (B2), Feb (C2)
	// Axis V (Emp): Alice (A3), Bob (A4)
	// Data:
	// B3 (Alice, Jan): 85
	// C3 (Alice, Feb): 88
	// B4 (Bob, Jan): 75
	// C4 (Bob, Feb): 78

	val, _ := f.GetCellValue("Sheet1", "B2")
	if val != "Jan" {
		t.Errorf("B2: want Jan, got %s", val)
	}
	val, _ = f.GetCellValue("Sheet1", "C2")
	if val != "Feb" {
		t.Errorf("C2: want Feb, got %s", val)
	}

	val, _ = f.GetCellValue("Sheet1", "A3")
	if val != "Alice" {
		t.Errorf("A3: want Alice, got %s", val)
	}
	val, _ = f.GetCellValue("Sheet1", "A4")
	if val != "Bob" {
		t.Errorf("A4: want Bob, got %s", val)
	}

	val, _ = f.GetCellValue("Sheet1", "B3")
	if val != "85" {
		t.Errorf("B3: want 85, got %s", val)
	}
	val, _ = f.GetCellValue("Sheet1", "C3")
	if val != "88" {
		t.Errorf("C3: want 88, got %s", val)
	}
	val, _ = f.GetCellValue("Sheet1", "B4")
	if val != "75" {
		t.Errorf("B4: want 75, got %s", val)
	}
	val, _ = f.GetCellValue("Sheet1", "C4")
	if val != "78" {
		t.Errorf("C4: want 78, got %s", val)
	}
}

// Helper to create Archive Date Template
func setupTemplate_ArchiveDate(t *testing.T) *excelize.File {
	f := excelize.NewFile()
	sheet := "Sheet1"
	idx, _ := f.GetSheetIndex("Sheet1")
	if idx == -1 {
		f.NewSheet(sheet)
	}
	// DailyList A2:C2
	f.SetCellValue(sheet, "A2", "{id}")
	f.SetCellValue(sheet, "B2", "{content}")
	f.SetCellValue(sheet, "C2", "{archivedate}")
	return f
}

func TestEndToEnd_ArchiveDate(t *testing.T) {
	// Config: test/workbooks/archivedate_filter_test.yaml
	// Rule: $date:day:day:0 (Today)

	wbConfig := &config.WorkbookConfig{
		Id:          "wb_archivedate_filter_test",
		Name:        "FilterTest",
		Template:    "filter_template.xlsx",
		OutputDir:   "tests",
		ArchiveRule: "$date:day:day:0", // Today
		Sheets: []config.SheetConfig{
			{
				Name:    "Sheet1",
				Dynamic: false,
				Blocks: []config.BlockConfig{
					{
						Name:      "DailyList",
						Type:      config.BlockTypeTag,
						Range:     config.CellRange{Ref: "A2:C2"},
						VViewName: "daily_view",
						Direction: config.DirectionVertical,
					},
				},
			},
		},
	}

	// NOTE: To filter by 'archivedate' parameter, the VView MUST have a tag named 'archivedate'.
	// The system parameter generated from ArchiveRule is "archivedate".
	vViews := map[string]*config.VirtualViewConfig{
		"daily_view": {
			Name: "daily_view",
			Tags: []config.TagConfig{
				{Name: "id", Column: "ID"},
				{Name: "content", Column: "CONTENT"},
				{Name: "archivedate", Column: "archivedate"}, // Tag name must match param name for auto-filter
			},
		},
	}

	// Dynamic Dates
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	mockData := map[string][]map[string]interface{}{
		"daily_view": {
			{"ID": "1", "CONTENT": "Report for today", "archivedate": today},
			{"ID": "2", "CONTENT": "Report for yesterday", "archivedate": yesterday},
			{"ID": "3", "CONTENT": "Report for tomorrow", "archivedate": tomorrow},
		},
	}

	fetcher := &MockFetcher{Data: mockData}
	provider := config.NewMemoryConfigRegistry(vViews)
	ctx := NewGenerationContext(wbConfig, provider, fetcher, nil)
	gen := NewGenerator(ctx)

	f := setupTemplate_ArchiveDate(t)
	adapter := &ExcelizeFile{file: f}

	block := &wbConfig.Sheets[0].Blocks[0]
	if err := gen.processBlock(adapter, "Sheet1", block); err != nil {
		t.Fatalf("processBlock failed: %v", err)
	}

	// Verify
	// Expect only 1 row (Today)
	// A2: 1, B2: Report for today, C2: <today>
	// A3: Empty (or whatever was there, assuming clean sheet)

	val, _ := f.GetCellValue("Sheet1", "B2")
	if val != "Report for today" {
		t.Errorf("Row 2 Content: want 'Report for today', got '%s'", val)
	}

	val, _ = f.GetCellValue("Sheet1", "C2")
	if val != today {
		t.Errorf("Row 2 Date: want '%s', got '%s'", today, val)
	}

	val, _ = f.GetCellValue("Sheet1", "B3")
	if val != "" {
		t.Errorf("Row 3 Content: want empty, got '%s'", val)
	}
}
