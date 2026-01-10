package core

import (
	"fibr-gen/config"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Generator struct {
	Context *GenerationContext
}

func NewGenerator(ctx *GenerationContext) *Generator {
	return &Generator{Context: ctx}
}

func replacePlaceholders(input string, params map[string]string) string {
	output := input
	for k, v := range params {
		placeholder := fmt.Sprintf("${%s}", k)
		output = strings.ReplaceAll(output, placeholder, v)
	}
	return output
}

func cloneParams(params map[string]string) map[string]string {
	copied := make(map[string]string, len(params))
	for k, v := range params {
		copied[k] = v
	}
	return copied
}

// Generate executes the workbook generation process.
func (g *Generator) Generate(templateRoot, outputRoot string) error {
	wbConf := g.Context.WorkbookConfig
	templatePath := filepath.Join(templateRoot, wbConf.Template)

	// Replace parameters in output path (e.g. ${archivedate})
	outputPathStr := replacePlaceholders(wbConf.OutputDir, g.Context.Parameters)

	outputPath := filepath.Join(outputRoot, outputPathStr)
	if filepath.Ext(outputPath) == "" {
		// Replace parameters in workbook name
		name := replacePlaceholders(wbConf.Name, g.Context.Parameters)
		outputPath = filepath.Join(outputPath, name+".xlsx")
	}

	f, err := excelize.OpenFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to open template: %w", err)
	}
	defer f.Close()

	for _, sheetConf := range wbConf.Sheets {
		if err := g.processSheet(f, &sheetConf); err != nil {
			return fmt.Errorf("processing sheet %s: %w", sheetConf.Name, err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save output: %w", err)
	}
	return nil
}

func (g *Generator) processSheet(f *excelize.File, sheetConf *config.SheetConfig) error {
	if sheetConf.Dynamic {
		return g.processDynamicSheet(f, sheetConf)
	}

	for _, block := range sheetConf.Blocks {
		if err := g.processBlock(f, sheetConf.Name, &block); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) processBlock(f *excelize.File, sheetName string, block *config.BlockConfig) error {
	return g.processBlockWithParams(f, sheetName, block, g.Context.Parameters)
}

func (g *Generator) processDynamicSheet(f *excelize.File, sheetConf *config.SheetConfig) error {
	// 1. Get Distinct Tag Values (e.g. Month Names)
	// Need to find VView Config
	vViewConf, err := g.Context.ConfigProvider.GetVirtualViewConfig(sheetConf.VViewName)
	if err != nil {
		return fmt.Errorf("vview not found for dynamic sheet: %s", sheetConf.VViewName)
	}

	// Fetch data to get distinct values for ParamTag
	// We need all rows first
	data, err := g.Context.Fetcher.Fetch(vViewConf.Name, g.Context.Parameters)
	if err != nil {
		return fmt.Errorf("failed to fetch dynamic sheet data: %w", err)
	}

	// Distinct values
	distinctValues := make(map[string]struct{})
	var values []string

	// Find which column maps to ParamTag
	var paramColumn string
	for _, tag := range vViewConf.Tags {
		if tag.Name == sheetConf.ParamTag {
			paramColumn = tag.Column
			break
		}
	}
	if paramColumn == "" {
		return fmt.Errorf("param tag %s not found in vview %s", sheetConf.ParamTag, sheetConf.VViewName)
	}

	for _, row := range data {
		if val, ok := row[paramColumn]; ok {
			strVal := fmt.Sprintf("%v", val)
			if _, exists := distinctValues[strVal]; !exists {
				distinctValues[strVal] = struct{}{}
				values = append(values, strVal)
			}
		}
	}

	// 2. Clone Sheets
	templateSheetName := sheetConf.Name
	templateIdx, err := f.GetSheetIndex(templateSheetName)
	if err != nil {
		return fmt.Errorf("template sheet not found: %s", templateSheetName)
	}

	for _, val := range values {
		// Validate sheet name (simple version)
		newSheetName := val
		// Handle conflict if exists? Excelize NewSheet might not handle it?
		// Assuming unique values for now.

		newIdx, err := f.NewSheet(newSheetName)
		if err != nil {
			return fmt.Errorf("failed to create sheet %s: %w", newSheetName, err)
		}

		// Copy content
		if err := f.CopySheet(templateIdx, newIdx); err != nil {
			return fmt.Errorf("failed to copy sheet: %w", err)
		}

		// 3. Process Blocks for this new sheet
		// We need to inject the parameter for this sheet (e.g. month=January)
		sheetParams := cloneParams(g.Context.Parameters)
		sheetParams[sheetConf.ParamTag] = val

		// Process each block in the NEW sheet
		for _, block := range sheetConf.Blocks {
			// We need to pass the sheetParams down.
			// But processBlock calls processTagBlock which uses g.Context.Parameters.
			// We should refactor processBlock to accept params too, or temporarily modify context?
			// Modifying context is not thread-safe if we parallelize.
			// Passing params is better.

			// We need a helper processBlockWithParams
			if err := g.processBlockWithParams(f, newSheetName, &block, sheetParams); err != nil {
				return err
			}
		}
	}

	// Delete Template Sheet if we generated others?
	if len(values) > 0 {
		f.DeleteSheet(templateSheetName)
	}

	return nil
}

func (g *Generator) processBlockWithParams(f *excelize.File, sheetName string, block *config.BlockConfig, params map[string]string) error {
	switch block.Type {
	case config.BlockTypeTag:
		return g.processTagBlockWithParams(f, sheetName, block, params)
	case config.BlockTypeExpand:
		// ExpandableBlock also needs to accept params!
		// But processExpandableBlock signature is fixed currently.
		// Let's assume we can change it or Context has the params?
		// Ideally we update Context for this operation?
		// Or refactor processExpandableBlock to take params.
		// For now, let's update Context (since Generator is per-request, single threaded usually)
		// But wait, we are inside a loop.

		// Refactoring processExpandableBlock to accept params is the right way.
		return g.processExpandableBlockWithParams(f, sheetName, block, params)
	default:
		return fmt.Errorf("unsupported block type: %s", block.Type)
	}
}

// Rename processExpandableBlock to processExpandableBlockWithParams and update usage
func (g *Generator) processExpandableBlockWithParams(f *excelize.File, sheetName string, block *config.BlockConfig, params map[string]string) error {
	// ... (content of processExpandableBlock but using params instead of g.Context.Parameters)
	// We need to change g.Context.GetBlockData(vAxis) to GetBlockDataWithParams(vAxis, params)

	// 1. Identify Axes
	var vAxis, hAxis *config.BlockConfig
	for i := range block.SubBlocks {
		sb := &block.SubBlocks[i]
		if sb.Type == config.BlockTypeAxis {
			if sb.Direction == config.DirectionVertical || sb.Direction == "" {
				vAxis = sb
			} else if sb.Direction == config.DirectionHorizontal {
				hAxis = sb
			}
		}
	}

	if vAxis == nil || hAxis == nil {
		return fmt.Errorf("ExpandableBlock %s must have both vertical and horizontal axes", block.Name)
	}

	// 2. Determine Expansion Mode
	isVerticalExpand := vAxis.InsertAfter

	var axisData []map[string]interface{}
	var staticData []map[string]interface{}
	var err error

	// 3. Process Axes
	if isVerticalExpand {
		// Vertical Expand Mode
		axisData, err = g.Context.GetBlockDataWithParams(vAxis, params)
		if err != nil {
			return err
		}

		// Insert Rows logic (same as before)
		dataCount := len(axisData)
		if dataCount > 1 {
			_, _, _, endRow, err := parseRange(vAxis.Range.Ref)
			if err != nil {
				return err
			}
			_, startRow, _, _, err := parseRange(vAxis.Range.Ref)
			if err != nil {
				return err
			}
			axisHeight := endRow - startRow + 1
			insertCount := (dataCount - 1) * axisHeight
			if err := f.InsertRows(sheetName, endRow+1, insertCount); err != nil {
				return err
			}
		}

		staticData, err = g.Context.GetBlockDataWithParams(hAxis, params)
		if err != nil {
			return err
		}

		// If Horizontal Axis has multiple items, we need to expand columns too, even if Vertical Axis expanded rows
		if len(staticData) > 1 {
			startCol, _, endCol, _, err := parseRange(hAxis.Range.Ref)
			if err != nil {
				return err
			}
			axisWidth := endCol - startCol + 1
			insertCount := (len(staticData) - 1) * axisWidth
			colName, _ := excelize.ColumnNumberToName(endCol + 1)
			if err := f.InsertCols(sheetName, colName, insertCount); err != nil {
				return err
			}

			// Copy Template Columns
			// Identify template cols
			minCol, maxCol := 999999, -1
			for i := range block.SubBlocks {
				sb := &block.SubBlocks[i]
				// Skip Vertical Axis when expanding horizontally
				if sb.Name == vAxis.Name {
					continue
				}
				c1, _, c2, _, err := parseRange(sb.Range.Ref)
				if err == nil {
					if c1 < minCol {
						minCol = c1
					}
					if c2 > maxCol {
						maxCol = c2
					}
				}
			}

			if minCol <= maxCol {
				if err := g.copyCols(f, sheetName, minCol, maxCol, endCol+1, insertCount); err != nil {
					return fmt.Errorf("failed to copy template cols: %w", err)
				}
			}
		}

		// Copy Template Rows to the new inserted rows (AFTER cols are inserted to ensure correct width?)
		// Actually, InsertCols might shift rows? No.
		// But InsertRows must happen before we fill data.
		// And we already inserted rows above.
		// BUT we haven't copied the template rows yet.
		// We should copy template rows NOW, after horizontal expansion is done (if any),
		// because horizontal expansion might have widened the template row.
		if dataCount > 1 {
			// Calculate bounding box rows
			// Re-parse ranges because InsertCols might have shifted things?
			// InsertCols shifts columns to the right. It doesn't affect row indices.
			// BUT it affects the CONTENT of the row.
			// So yes, we should copy rows AFTER columns are expanded.

			_, startRow, _, endRow, err := parseRange(vAxis.Range.Ref)
			if err != nil {
				return err
			}
			axisHeight := endRow - startRow + 1
			insertCount := (dataCount - 1) * axisHeight

			minRow, maxRow := 999999, -1
			for i := range block.SubBlocks {
				sb := &block.SubBlocks[i]
				// Skip Horizontal Axis when expanding vertically
				if sb.Name == hAxis.Name {
					continue
				}
				_, r1, _, r2, err := parseRange(sb.Range.Ref)
				if err == nil {
					if r1 < minRow {
						minRow = r1
					}
					if r2 > maxRow {
						maxRow = r2
					}
				}
			}

			if minRow <= maxRow {
				if err := g.copyRows(f, sheetName, minRow, maxRow, endRow+1, insertCount); err != nil {
					return fmt.Errorf("failed to copy template rows: %w", err)
				}
			}
		}

	} else {
		// Horizontal Expand Mode
		axisData, err = g.Context.GetBlockDataWithParams(hAxis, params)
		if err != nil {
			return err
		}

		// Insert Cols logic
		dataCount := len(axisData)
		if dataCount > 1 {
			startCol, _, endCol, _, err := parseRange(hAxis.Range.Ref)
			if err != nil {
				return err
			}
			axisWidth := endCol - startCol + 1
			insertCount := (dataCount - 1) * axisWidth
			colName, _ := excelize.ColumnNumberToName(endCol + 1)
			if err := f.InsertCols(sheetName, colName, insertCount); err != nil {
				return err
			}

			// Copy Template Columns
			minCol, maxCol := 999999, -1
			for i := range block.SubBlocks {
				sb := &block.SubBlocks[i]
				// Skip Vertical Axis when expanding horizontally
				if sb.Name == vAxis.Name {
					continue
				}
				c1, _, c2, _, err := parseRange(sb.Range.Ref)
				if err == nil {
					if c1 < minCol {
						minCol = c1
					}
					if c2 > maxCol {
						maxCol = c2
					}
				}
			}

			if minCol <= maxCol {
				if err := g.copyCols(f, sheetName, minCol, maxCol, endCol+1, insertCount); err != nil {
					return fmt.Errorf("failed to copy template cols: %w", err)
				}
			}
		}

		staticData, err = g.Context.GetBlockDataWithParams(vAxis, params)
		if err != nil {
			return err
		}
	}

	// Helper to fill a block without expansion
	fillBlockSimple := func(b *config.BlockConfig, data []map[string]interface{}) error {
		// Resolve VView for tags
		vv, err := g.Context.ConfigProvider.GetVirtualViewConfig(b.VViewName)
		if err != nil {
			return fmt.Errorf("vview not found: %s", b.VViewName)
		}

		// Determine direction
		isVert := b.Direction == config.DirectionVertical || b.Direction == ""

		// Parse Range
		c1, r1, c2, r2, err := parseRange(b.Range.Ref)
		if err != nil {
			return err
		}
		bH := r2 - r1 + 1
		bW := c2 - c1 + 1

		// Cache Template
		tmplPat := make([][]string, bH)
		tmplSty := make([][]int, bH)
		for r := 0; r < bH; r++ {
			tmplPat[r] = make([]string, bW)
			tmplSty[r] = make([]int, bW)
			for c := 0; c < bW; c++ {
				cn, _ := excelize.CoordinatesToCellName(c1+c, r1+r)
				val, _ := f.GetCellValue(sheetName, cn)
				sty, _ := f.GetCellStyle(sheetName, cn)
				tmplPat[r][c] = val
				tmplSty[r][c] = sty
			}
		}

		// Iterate
		for i, row := range data {
			rOff, cOff := 0, 0
			if isVert {
				rOff = i * bH
			} else {
				cOff = i * bW
			}

			// Replacement map
			rep := make(map[string]interface{})
			for _, t := range vv.Tags {
				if v, ok := row[t.Column]; ok {
					rep[t.Name] = v
				}
			}

			// Fill
			for r := 0; r < bH; r++ {
				for c := 0; c < bW; c++ {
					// Coords
					tc, tr := c1+c+cOff, r1+r+rOff
					tcn, _ := excelize.CoordinatesToCellName(tc, tr)

					val := tmplPat[r][c]
					sty := tmplSty[r][c]

					// Replace
					for t, v := range rep {
						ph := fmt.Sprintf("{%s}", t)
						if strings.Contains(val, ph) {
							val = strings.ReplaceAll(val, ph, fmt.Sprintf("%v", v))
						}
					}

					f.SetCellValue(sheetName, tcn, val)
					f.SetCellStyle(sheetName, tcn, tcn, sty)
				}
			}
		}
		return nil
	}

	// Fill vAxis Headers
	if err := fillBlockSimple(vAxis, axisData); err != nil {
		return fmt.Errorf("failed to fill vAxis headers: %w", err)
	}

	// Fill hAxis Headers
	if err := fillBlockSimple(hAxis, staticData); err != nil {
		return fmt.Errorf("failed to fill hAxis headers: %w", err)
	}

	// 5. Fill Intersection Data (Template Blocks)
	// Iterate over the grid defined by axisData x staticData
	// For each cell in the grid, find the corresponding TemplateBlock and fill it.

	// Collect Template Blocks (SubBlocks that are NOT Axis)
	var templateBlocks []*config.BlockConfig
	for i := range block.SubBlocks {
		sb := &block.SubBlocks[i]
		// Use Template flag if available, otherwise fallback to Type != Axis
		if sb.Template || sb.Type != config.BlockTypeAxis {
			templateBlocks = append(templateBlocks, sb)
		}
	}

	// Pre-cache Template Content (Read-Once)
	// We read the template values and styles once from the original position.
	// Then we use this cache to fill all target cells.
	type CellData struct {
		Val   string
		Style int
	}
	type TemplateCache struct {
		Block    *config.BlockConfig
		Cells    [][]CellData // [row][col]
		StartCol int
		StartRow int
		Width    int
		Height   int
	}

	var cachedTemplates []TemplateCache
	for _, tb := range templateBlocks {
		c1, r1, c2, r2, err := parseRange(tb.Range.Ref)
		if err != nil {
			return err
		}
		w := c2 - c1 + 1
		h := r2 - r1 + 1
		cells := make([][]CellData, h)
		for r := 0; r < h; r++ {
			cells[r] = make([]CellData, w)
			for c := 0; c < w; c++ {
				cn, _ := excelize.CoordinatesToCellName(c1+c, r1+r)
				val, _ := f.GetCellValue(sheetName, cn)
				sty, _ := f.GetCellStyle(sheetName, cn)
				cells[r][c] = CellData{Val: val, Style: sty}
			}
		}
		cachedTemplates = append(cachedTemplates, TemplateCache{
			Block:    tb,
			Cells:    cells,
			StartCol: c1,
			StartRow: r1,
			Width:    w,
			Height:   h,
		})
	}

	rows := axisData
	cols := staticData
	if !isVerticalExpand {
		rows = staticData
		cols = axisData
	}

	// ... (Axis Param Key Logic) ...
	getAxisParamKey := func(axis *config.BlockConfig) (string, error) {
		if axis.TagVariable != "" {
			return axis.TagVariable, nil
		}
		conf, err := g.Context.ConfigProvider.GetVirtualViewConfig(axis.VViewName)
		if err != nil {
			return "", err
		}
		if len(conf.Tags) > 0 {
			return conf.Tags[0].Name, nil
		}
		return "", fmt.Errorf("cannot determine parameter key for axis %s", axis.Name)
	}

	vKey, err := getAxisParamKey(vAxis)
	if err != nil {
		return err
	}
	hKey, err := getAxisParamKey(hAxis)
	if err != nil {
		return err
	}

	// Re-parse ranges to get dimensions for stepping
	_, _, _, vEndRow, _ := parseRange(vAxis.Range.Ref)
	_, vStartRow, _, _, _ := parseRange(vAxis.Range.Ref)
	vStep := vEndRow - vStartRow + 1

	hStartCol, _, hEndCol, _, _ := parseRange(hAxis.Range.Ref)
	hStep := hEndCol - hStartCol + 1

	// Iterate Grid & Fill (Write-Many)
	for r, rowItem := range rows {
		for c, colItem := range cols {
			// Construct parameters for this cell
			cellParams := cloneParams(g.Context.Parameters)

			// Resolve Tag Name -> Column Name first!
			getColName := func(vViewName, tagName string) string {
				conf, err := g.Context.ConfigProvider.GetVirtualViewConfig(vViewName)
				if err != nil {
					return ""
				}
				for _, t := range conf.Tags {
					if t.Name == tagName {
						return t.Column
					}
				}
				return ""
			}

			vCol := getColName(vAxis.VViewName, vKey)
			if vCol != "" {
				if val, ok := rowItem[vCol]; ok {
					cellParams[vKey] = fmt.Sprintf("%v", val)
				}
			}

			hCol := getColName(hAxis.VViewName, hKey)
			if hCol != "" {
				if val, ok := colItem[hCol]; ok {
					cellParams[hKey] = fmt.Sprintf("%v", val)
				}
			}

			// Calculate Offsets
			rowOffset := r * vStep
			colOffset := c * hStep

			// Process each Template using Cache
			for _, cache := range cachedTemplates {
				// Fetch Data for this cell
				cellDataList, err := g.Context.GetBlockDataWithParams(cache.Block, cellParams)
				if err != nil {
					return err
				}

				rep := make(map[string]interface{})
				if len(cellDataList) > 0 {
					// Load tag mapping for this block
					conf, _ := g.Context.ConfigProvider.GetVirtualViewConfig(cache.Block.VViewName)
					if conf != nil {
						for _, t := range conf.Tags {
							if v, ok := cellDataList[0][t.Column]; ok {
								rep[t.Name] = v
							}
						}
					}
				}

				// Fill Cells
				for tr := 0; tr < cache.Height; tr++ {
					for tc := 0; tc < cache.Width; tc++ {
						// Original Template
						cell := cache.Cells[tr][tc]
						val := cell.Val
						style := cell.Style

						// Replace
						for tag, v := range rep {
							ph := fmt.Sprintf("{%s}", tag)
							if strings.Contains(val, ph) {
								val = strings.ReplaceAll(val, ph, fmt.Sprintf("%v", v))
							}
						}

						// Calculate Target Coord
						targetC := cache.StartCol + colOffset + tc
						targetR := cache.StartRow + rowOffset + tr

						cn, _ := excelize.CoordinatesToCellName(targetC, targetR)
						f.SetCellValue(sheetName, cn, val)
						if style != 0 {
							f.SetCellStyle(sheetName, cn, cn, style)
						}
					}
				}
			}
		}
	}

	return nil
}

// copyRows copies a range of rows to a new location, replicating them count times.
func (g *Generator) copyRows(f *excelize.File, sheet string, srcStartRow, srcEndRow, destStartRow, insertHeight int) error {
	// 1. Read Source Data
	srcHeight := srcEndRow - srcStartRow + 1
	type cellData struct {
		val   string
		style int
	}
	// Map: ColIndex -> cellData
	srcMap := make(map[int]cellData)

	// Determine max column used in the sheet
	dims, err := f.GetSheetDimension(sheet)
	if err != nil {
		return err
	}
	_, _, maxC, _, err := parseRange(dims)
	if err != nil {
		maxC = 100 // Fallback
	}

	for r := srcStartRow; r <= srcEndRow; r++ {
		for c := 1; c <= maxC; c++ {
			cn, _ := excelize.CoordinatesToCellName(c, r)
			val, _ := f.GetCellValue(sheet, cn)
			style, _ := f.GetCellStyle(sheet, cn)
			srcMap[(r-srcStartRow)*10000+c] = cellData{val, style}
		}
	}

	// 2. Write to Dest
	for i := 0; i < insertHeight; i++ {
		srcOffset := i % srcHeight
		destRow := destStartRow + i

		for c := 1; c <= maxC; c++ {
			key := srcOffset*10000 + c
			if data, ok := srcMap[key]; ok {
				cn, _ := excelize.CoordinatesToCellName(c, destRow)
				f.SetCellValue(sheet, cn, data.val)
				if data.style != 0 {
					f.SetCellStyle(sheet, cn, cn, data.style)
				}
			}
		}
	}
	return nil
}

// copyCols copies a range of columns to a new location.
func (g *Generator) copyCols(f *excelize.File, sheet string, srcStartCol, srcEndCol, destStartCol, insertWidth int) error {
	srcWidth := srcEndCol - srcStartCol + 1

	type cellData struct {
		val   string
		style int
	}
	srcMap := make(map[int]cellData)

	dims, err := f.GetSheetDimension(sheet)
	if err != nil {
		return err
	}
	_, _, _, maxR, err := parseRange(dims)
	if err != nil {
		maxR = 1000
	}

	for c := srcStartCol; c <= srcEndCol; c++ {
		for r := 1; r <= maxR; r++ {
			cn, _ := excelize.CoordinatesToCellName(c, r)
			val, _ := f.GetCellValue(sheet, cn)
			style, _ := f.GetCellStyle(sheet, cn)
			srcMap[(c-srcStartCol)*10000+r] = cellData{val, style}
		}
	}

	for i := 0; i < insertWidth; i++ {
		srcOffset := i % srcWidth
		destCol := destStartCol + i

		for r := 1; r <= maxR; r++ {
			key := srcOffset*10000 + r
			if data, ok := srcMap[key]; ok {
				cn, _ := excelize.CoordinatesToCellName(destCol, r)
				f.SetCellValue(sheet, cn, data.val)
				if data.style != 0 {
					f.SetCellStyle(sheet, cn, cn, data.style)
				}
			}
		}
	}
	return nil
}

// Helper to shift range "A1:B2" by (colOffset, rowOffset)
// func shiftRange(ref string, colOffset, rowOffset int) (string, error) {
// 	c1, r1, c2, r2, err := parseRange(ref)
// 	if err != nil {
// 		return "", err
// 	}

// 	c1 += colOffset
// 	c2 += colOffset
// 	r1 += rowOffset
// 	r2 += rowOffset

// 	n1, _ := excelize.CoordinatesToCellName(c1, r1)
// 	n2, _ := excelize.CoordinatesToCellName(c2, r2)
// 	return n1 + ":" + n2, nil
// }

func (g *Generator) processTagBlock(f *excelize.File, sheetName string, block *config.BlockConfig) error {
	return g.processTagBlockWithParams(f, sheetName, block, g.Context.Parameters)
}

func (g *Generator) processTagBlockWithParams(f *excelize.File, sheetName string, block *config.BlockConfig, params map[string]string) error {
	data, err := g.Context.GetBlockDataWithParams(block, params)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil // No data
	}

	// Resolve tags
	// vview, ok := g.Context.VViewConfigs[block.VViewName]
	// if !ok {
	// 	return fmt.Errorf("vview not found: %s", block.VViewName)
	// }
	// We need VView for tags.
	// But wait, GetBlockDataWithParams might have filtered it.
	// But we need the config to know the tags mapping.
	// Since we are moving to VView struct, maybe GetBlockData should return VView?
	// For now, let's keep it simple and just look up the config again.

	vView, err := g.Context.ConfigProvider.GetVirtualViewConfig(block.VViewName)
	if err != nil {
		return fmt.Errorf("vview not found: %s", block.VViewName)
	}

	// Determine direction (default: vertical)
	isVertical := block.Direction == config.DirectionVertical || block.Direction == ""

	// Parse Template Range
	startCol, startRow, endCol, endRow, err := parseRange(block.Range.Ref)
	if err != nil {
		return err
	}

	// Height/Width of the template block
	blockHeight := endRow - startRow + 1
	blockWidth := endCol - startCol + 1

	// 1. Expand (Insert Rows/Columns) if needed
	// Note: In C#, expansion happens only if block.Expand is true (which maps to InsertAfter here probably, or explicit Expand flag)
	// But TagBlock in C# also has an implicit loop if multiple rows are returned.
	// Let's assume we always expand if row count > 1, similar to C# logic.
	dataCount := len(data)
	if dataCount > 1 {
		// Calculate how many new rows/cols we need.
		// We already have 1 set of rows/cols in the template. We need (dataCount - 1) more.
		if isVertical {
			insertCount := (dataCount - 1) * blockHeight
			// Insert after the block's bottom
			if err := f.InsertRows(sheetName, endRow+1, insertCount); err != nil {
				return fmt.Errorf("failed to insert rows: %w", err)
			}
		} else {
			insertCount := (dataCount - 1) * blockWidth
			// Insert after the block's right
			// Excelize InsertCols takes column name "C"
			colName, _ := excelize.ColumnNumberToName(endCol + 1)
			if err := f.InsertCols(sheetName, colName, insertCount); err != nil {
				return fmt.Errorf("failed to insert cols: %w", err)
			}
		}
	}

	// 2. Cache Template Pattern (Values and Styles)
	// We must read the template cells BEFORE we start filling data,
	// because the first iteration (i=0) will overwrite the template cells with actual values.
	templatePattern := make([][]string, blockHeight)
	templateStyles := make([][]int, blockHeight)
	for r := 0; r < blockHeight; r++ {
		templatePattern[r] = make([]string, blockWidth)
		templateStyles[r] = make([]int, blockWidth)
		for c := 0; c < blockWidth; c++ {
			cellName, _ := excelize.CoordinatesToCellName(startCol+c, startRow+r)
			val, _ := f.GetCellValue(sheetName, cellName)
			styleID, _ := f.GetCellStyle(sheetName, cellName)
			templatePattern[r][c] = val
			templateStyles[r][c] = styleID
		}
	}

	// 3. Iterate and Fill
	for i, rowData := range data {
		// Calculate current offset
		rowOffset := 0
		colOffset := 0
		if isVertical {
			rowOffset = i * blockHeight
		} else {
			colOffset = i * blockWidth
		}

		// Build replacement map for this row
		replacements := make(map[string]interface{})
		for _, tag := range vView.Tags {
			if val, ok := rowData[tag.Column]; ok {
				replacements[tag.Name] = val
			}
		}

		// Iterate over the block cells
		for r := 0; r < blockHeight; r++ {
			for c := 0; c < blockWidth; c++ {
				// Target cell coords
				tmplC, tmplR := startCol+c, startRow+r
				targetC, targetR := tmplC+colOffset, tmplR+rowOffset

				// Use CACHED template pattern and style
				val := templatePattern[r][c]
				styleID := templateStyles[r][c]

				targetCellName, _ := excelize.CoordinatesToCellName(targetC, targetR)

				// Perform replacement
				currentVal := val
				for tag, v := range replacements {
					placeholder := fmt.Sprintf("{%s}", tag)
					if strings.Contains(currentVal, placeholder) {
						currentVal = strings.ReplaceAll(currentVal, placeholder, fmt.Sprintf("%v", v))
					}
				}

				f.SetCellValue(sheetName, targetCellName, currentVal)
				f.SetCellStyle(sheetName, targetCellName, targetCellName, styleID)
			}
		}
	}

	return nil
}

// Helper to parse "A1:B2"
func parseRange(ref string) (int, int, int, int, error) {
	parts := strings.Split(ref, ":")
	if len(parts) != 2 {
		return 0, 0, 0, 0, fmt.Errorf("invalid range: %s", ref)
	}
	c1, r1, err := excelize.CellNameToCoordinates(parts[0])
	if err != nil {
		return 0, 0, 0, 0, err
	}
	c2, r2, err := excelize.CellNameToCoordinates(parts[1])
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return c1, r1, c2, r2, nil
}
