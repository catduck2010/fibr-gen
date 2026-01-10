package core

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

// CsvDataFetcher implements DataFetcher using CSV files.
// It maps viewName to a CSV file path.
type CsvDataFetcher struct {
	RootDir string
}

func NewCsvDataFetcher(rootDir string) *CsvDataFetcher {
	return &CsvDataFetcher{RootDir: rootDir}
}

func (f *CsvDataFetcher) Fetch(viewName string, params map[string]string) ([]map[string]interface{}, error) {
	// Assume viewName is the filename without extension, or we can use a mapping.
	// For simplicity, let's assume viewName + ".csv" exists in RootDir.
	filePath := filepath.Join(f.RootDir, viewName+".csv")

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open csv file %s: %w", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv content: %w", err)
	}

	if len(records) < 1 {
		return nil, nil // Empty
	}

	header := records[0]
	var result []map[string]interface{}

	for i := 1; i < len(records); i++ {
		row := records[i]
		item := make(map[string]interface{})
		for j, col := range row {
			if j < len(header) {
				item[header[j]] = col
			}
		}
		// TODO: Implement filtering based on params if needed for CSV?
		// Simple filter: if param key matches a column name, filter by value.
		match := true
		for k, v := range params {
			if colVal, hasCol := item[k]; hasCol {
				// Convert both to string for comparison (already string in CSV)
				if fmt.Sprintf("%v", colVal) != v {
					match = false
					break
				}
			}
		}

		if match {
			result = append(result, item)
		}
	}

	return result, nil
}
