package core

import (
	"fibr-gen/config"
	"fmt"
	"sort"
)

// VirtualView represents a data view with tag mapping capabilities.
// It mimics the VView class in C#.
type VirtualView struct {
	Config     *config.VirtualViewConfig
	Data       []map[string]interface{} // The actual data table (DataTable in C#)
	TagMapping map[string]string        // Tag Name -> Column Name
}

// NewVirtualView creates a new VirtualView instance.
func NewVirtualView(conf *config.VirtualViewConfig, data []map[string]interface{}) *VirtualView {
	mapping := make(map[string]string)
	for _, t := range conf.Tags {
		mapping[t.Name] = t.Column
	}
	return &VirtualView{
		Config:     conf,
		Data:       data,
		TagMapping: mapping,
	}
}

// GetDistinctTagValues returns all unique values for a given tag.
// Corresponds to VView.GetDistinctTagValues in C#.
func (vv *VirtualView) GetDistinctTagValues(tagName string) ([]string, error) {
	colName, ok := vv.TagMapping[tagName]
	if !ok {
		// Fallback: if tag not found, check if it's the column name itself?
		// Or try to find if tagName matches any column directly?
		// For strict compliance, we should error or return empty.
		// Let's try to match by name first if mapping fails (for flexibility)
		colName = tagName
		// Check if colName exists in data keys?
		// Actually, let's stick to strict mapping first.
		return nil, fmt.Errorf("tag '%s' not found in view '%s'", tagName, vv.Config.Name)
	}

	seen := make(map[string]struct{})
	var result []string

	for _, row := range vv.Data {
		if val, ok := row[colName]; ok {
			strVal := fmt.Sprintf("%v", val)
			if strVal == "" {
				continue
			}
			if _, exists := seen[strVal]; !exists {
				seen[strVal] = struct{}{}
				result = append(result, strVal)
			}
		}
	}

	// Sort for consistency
	sort.Strings(result)
	return result, nil
}

// Filter filters the internal data based on a parameter dictionary.
// Corresponds to VView.Filter in C#.
func (vv *VirtualView) Filter(params map[string]string) {
	if len(params) == 0 {
		return
	}

	var filtered []map[string]interface{}

	for _, row := range vv.Data {
		match := true
		for paramKey, paramVal := range params {
			// Find column for this paramKey (which acts as a tag)
			colName, ok := vv.TagMapping[paramKey]
			if !ok {
				continue // Parameter not mapped to a tag in this view, ignore
			}

			if rowVal, hasCol := row[colName]; hasCol {
				if fmt.Sprintf("%v", rowVal) != paramVal {
					match = false
					break
				}
			}
		}
		if match {
			filtered = append(filtered, row)
		}
	}

	vv.Data = filtered
}

// GetRowCount returns the number of rows.
func (vv *VirtualView) GetRowCount() int {
	return len(vv.Data)
}

// Copy creates a deep copy of the VirtualView.
// The Data slice is duplicated so that modifications (like Filter) on the copy
// do not affect the original. Config and TagMapping are shared as they are read-only.
func (vv *VirtualView) Copy() *VirtualView {
	// Deep copy data
	newData := make([]map[string]interface{}, len(vv.Data))
	for i, row := range vv.Data {
		// Map is a reference type, so we need to copy the map too if we modify values inside it.
		// However, usually we only filter *rows* (subset of slice), not modify *cell values*.
		// But for safety, let's copy the map.
		newRow := make(map[string]interface{}, len(row))
		for k, v := range row {
			newRow[k] = v
		}
		newData[i] = newRow
	}

	return &VirtualView{
		Config:     vv.Config,
		Data:       newData,
		TagMapping: vv.TagMapping, // Shared reference is fine for read-only map
	}
}
