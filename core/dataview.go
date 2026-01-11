package core

import (
	"fibr-gen/config"
	"fmt"
	"sort"
)

// DataView represents a data view with label mapping capabilities.
type DataView struct {
	Config       *config.DataViewConfig
	Data         []map[string]interface{} // The actual data table (DataTable in C#)
	LabelMapping map[string]string        // label Name -> Column Name
}

// NewDataView creates a new DataView instance.
func NewDataView(conf *config.DataViewConfig, data []map[string]interface{}) *DataView {
	mapping := make(map[string]string)
	for _, t := range conf.Labels {
		mapping[t.Name] = t.Column
	}
	return &DataView{
		Config:       conf,
		Data:         data,
		LabelMapping: mapping,
	}
}

// GetDistinctLabelValues returns all unique values for a given label.
func (v *DataView) GetDistinctLabelValues(label string) ([]string, error) {
	colName, ok := v.LabelMapping[label]
	if !ok {
		// Fallback: if label not found, check if it's the column name itself?
		// Or try to find if label matches any column directly?
		// For strict compliance, we should error or return empty.
		// Let's try to match by name first if mapping fails (for flexibility)
		colName = label
		// Check if colName exists in data keys?
		// Actually, let's stick to strict mapping first.
		return nil, fmt.Errorf("label '%s' not found in view '%s'", label, v.Config.Name)
	}

	seen := make(map[string]struct{})
	var result []string

	for _, row := range v.Data {
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
func (v *DataView) Filter(params map[string]string) {
	if len(params) == 0 {
		return
	}

	var filtered []map[string]interface{}

	for _, row := range v.Data {
		match := true
		for paramKey, paramVal := range params {
			// Find column for this paramKey (which acts as a label)
			colName, ok := v.LabelMapping[paramKey]
			if !ok {
				continue // Parameter not mapped to a label in this view, ignore
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

	v.Data = filtered
}

// GetRowCount returns the number of rows.
func (v *DataView) GetRowCount() int {
	return len(v.Data)
}

// Copy creates a deep copy of the DataView.
// The Data slice is duplicated so that modifications (like Filter) on the copy
// do not affect the original. Config and LabelMapping are shared as they are read-only.
func (v *DataView) Copy() *DataView {
	// Deep copy data
	newData := make([]map[string]interface{}, len(v.Data))
	for i, row := range v.Data {
		// Map is a reference type, so we need to copy the map too if we modify values inside it.
		// However, usually we only filter *rows* (subset of slice), not modify *cell values*.
		// But for safety, let's copy the map.
		newRow := make(map[string]interface{}, len(row))
		for k, v := range row {
			newRow[k] = v
		}
		newData[i] = newRow
	}

	return &DataView{
		Config:       v.Config,
		Data:         newData,
		LabelMapping: v.LabelMapping, // Shared reference is fine for read-only map
	}
}
