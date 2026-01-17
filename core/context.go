package core

import (
	"fibr-gen/config"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// DataFetcher defines the interface for fetching data for Data Views.
type DataFetcher interface {
	// Fetch returns a list of rows (maps) for a given view name and parameters.
	Fetch(viewName string, params map[string]string) ([]map[string]interface{}, error)
}

// GenerationContext holds the state for the current generation process.
type GenerationContext struct {
	WorkbookConfig *config.WorkbookConfig
	Parameters     map[string]string
	Fetcher        DataFetcher
	ConfigProvider config.Provider
	// Cache for loaded DataViews (to avoid re-fetching or re-creating)
	LoadedViews map[string]*DataView
}

// NewGenerationContext creates a new context.
func NewGenerationContext(wb *config.WorkbookConfig, provider config.Provider, fetcher DataFetcher, params map[string]string) *GenerationContext {
	// Merge params
	mergedParams := make(map[string]string)
	if wb.Parameters != nil {
		for k, v := range wb.Parameters {
			mergedParams[k] = v
		}
	}
	for k, v := range params {
		mergedParams[k] = v
	}

	// Handle archive_date special rule
	if wb.ArchiveRule != "" {
		if val, err := ParseDynamicDate(wb.ArchiveRule, time.Now()); err == nil {
			mergedParams["archive_date"] = val
		}
	}

	// Process all dynamic parameters
	for k, v := range mergedParams {
		if strings.HasPrefix(v, "$date:") {
			if val, err := ParseDynamicDate(v, time.Now()); err == nil {
				mergedParams[k] = val
			}
		}
	}

	return &GenerationContext{
		WorkbookConfig: wb,
		Parameters:     mergedParams,
		Fetcher:        fetcher,
		ConfigProvider: provider,
		LoadedViews:    make(map[string]*DataView),
	}
}

// GetDataView resolves and loads a DataView by name.
// It uses caching to ensure the same view (and its data) is reused if needed,
// but note that DataView is mutable (can be filtered).
// For distinct operations, we typically need a fresh copy or the raw data.
// In this design, GetDataView returns a NEW instance populated with data,
func (ctx *GenerationContext) GetDataView(viewName string) (*DataView, error) {
	// 1. Check Cache
	if cachedView, ok := ctx.LoadedViews[viewName]; ok {
		// Return a DEEP COPY to ensure isolation
		// The caller can filter/modify the copy without affecting the cache
		return cachedView.Copy(), nil
	}

	// 2. Resolve Config
	conf, err := ctx.ConfigProvider.GetDataViewConfig(viewName)
	if err != nil {
		return nil, err
	}

	// 3. Fetch Data (Full Load)
	// Note: Fetcher.Fetch might already filter based on global params?
	// Currently Fetcher.Fetch takes params.
	data, err := ctx.Fetcher.Fetch(conf.Name, ctx.Parameters)
	if err != nil {
		return nil, err
	}

	// 4. Create and Cache DataView
	// This instance holds the ORIGINAL full data
	vv := NewDataView(conf, data)
	ctx.LoadedViews[viewName] = vv

	// 5. Return Copy
	return vv.Copy(), nil
}

// GetBlockData fetches data for a specific block based on its DataView.
func (ctx *GenerationContext) GetBlockData(block *config.BlockConfig) ([]map[string]interface{}, error) {
	return ctx.GetBlockDataWithParams(block, ctx.Parameters)
}

// GetBlockDataWithParams fetches data with custom parameters (for MatrixBlock iteration).
func (ctx *GenerationContext) GetBlockDataWithParams(block *config.BlockConfig, params map[string]string) ([]map[string]interface{}, error) {
	if block.DataViewName == "" {
		return nil, nil // No data source
	}

	// Load DataView
	vv, err := ctx.GetDataView(block.DataViewName)
	if err != nil {
		return nil, err
	}

	// Apply Params Filter (if any specific to this block/context)
	// The Fetcher might have already filtered by GLOBAL params.
	// But `params` here might contain loop variables (e.g. emp_id=E001).
	// So we should filter the DataView memory data.
	vv.Filter(params)

	// Apply Distinct logic if it's an Header Block
	var finalData []map[string]interface{}
	if block.Type == config.BlockTypeHeader {
		result, err := ctx.distinctData(vv.Data, block, vv)
		if err != nil {
			return nil, err
		}
		finalData = result
	} else {
		finalData = vv.Data
	}

	// Apply RowLimit if configured
	if block.RowLimit > 0 && len(finalData) > block.RowLimit {
		finalData = finalData[:block.RowLimit]
	}

	// Debug Log
	blockTypeStr := ""
	if block.Type == config.BlockTypeHeader {
		blockTypeStr = " (Header)"
	}

	slog.Debug("Block Fetched",
		"block", block.Name+blockTypeStr,
		"DataView", block.DataViewName,
		"params", params,
		"rows", len(finalData),
	)

	if len(finalData) > 0 {
		slog.Debug("Sample Row", "row", finalData[0])
	}

	return finalData, nil
}

// distinctData filters the data to unique values based on the block's label configuration.
func (ctx *GenerationContext) distinctData(data []map[string]interface{}, block *config.BlockConfig, v *DataView) ([]map[string]interface{}, error) {
	// Identify the key label for this header block.
	keyLabel := block.LabelVariable
	if keyLabel == "" {
		if len(v.Config.Labels) > 0 {
			keyLabel = v.Config.Labels[0].Name
		} else {
			return data, nil // No labels to distinct by
		}
	}

	// Use DataView's mapping
	colName, ok := v.LabelMapping[keyLabel]
	if !ok {
		return data, nil
	}

	// Perform Distinct
	seen := make(map[string]struct{})
	var result []map[string]interface{}

	for _, row := range data {
		val, ok := row[colName]
		if !ok {
			continue
		}
		strVal := fmt.Sprintf("%v", val)

		if _, exists := seen[strVal]; !exists {
			seen[strVal] = struct{}{}
			result = append(result, row)
		}
	}

	return result, nil
}

// MockDataFetcher is a simple implementation for testing.
type MockDataFetcher struct {
	Data map[string][]map[string]interface{}
}

func (m *MockDataFetcher) Fetch(viewName string, params map[string]string) ([]map[string]interface{}, error) {
	if data, ok := m.Data[viewName]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("view not found: %s", viewName)
}
