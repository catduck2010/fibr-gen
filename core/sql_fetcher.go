package core

import (
	"database/sql"
	"fmt"
	"strings"
)

// SQLDataFetcher implements DataFetcher using a generic SQL database (MySQL, PostgreSQL).
// It maps viewName to a Table Name.
type SQLDataFetcher struct {
	DB         *sql.DB
	DriverName string // "mysql" or "postgres"
}

// NewSQLDataFetcher creates a new fetcher.
func NewSQLDataFetcher(db *sql.DB, driverName string) *SQLDataFetcher {
	return &SQLDataFetcher{
		DB:         db,
		DriverName: driverName,
	}
}

// Fetch executes a SELECT query on the table specified by viewName.
// It applies simple equality filtering based on params.
func (f *SQLDataFetcher) Fetch(viewName string, params map[string]string) ([]map[string]interface{}, error) {
	tableName := viewName
	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	var args []interface{}

	if len(params) > 0 {
		var conditions []string
		i := 1
		for k, v := range params {
			if f.DriverName == "postgres" {
				conditions = append(conditions, fmt.Sprintf("%s = $%d", k, i))
			} else {
				// MySQL and others usually use ?
				conditions = append(conditions, fmt.Sprintf("%s = ?", k))
			}
			args = append(args, v)
			i++
		}
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := f.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var result []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Handle []byte (MySQL driver often returns strings as []byte)
			if b, ok := val.([]byte); ok {
				entry[col] = string(b)
			} else {
				entry[col] = val
			}
		}
		result = append(result, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return result, nil
}
