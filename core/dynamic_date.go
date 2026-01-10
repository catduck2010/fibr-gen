package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDynamicDate parses a dynamic date string in the format "$date:format:unit:offset".
// Example: "$date:day:day:-1" -> Yesterday's date in "2006-01-02" format.
func ParseDynamicDate(expression string, baseTime time.Time) (string, error) {
	if !strings.HasPrefix(expression, "$date:") {
		return expression, nil
	}

	parts := strings.Split(expression, ":")
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid dynamic date format: %s", expression)
	}

	format := parts[1]
	unit := parts[2]
	offsetStr := parts[3]

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return "", fmt.Errorf("invalid offset in dynamic date: %s", expression)
	}

	targetTime := baseTime

	switch unit {
	case "day":
		targetTime = targetTime.AddDate(0, 0, offset)
	case "month":
		targetTime = targetTime.AddDate(0, offset, 0)
	case "year":
		targetTime = targetTime.AddDate(offset, 0, 0)
	default:
		return "", fmt.Errorf("unsupported unit in dynamic date: %s", unit)
	}

	return formatTime(targetTime, format), nil
}

func formatTime(t time.Time, format string) string {
	switch format {
	case "day":
		return t.Format("2006-01-02")
	case "month":
		return t.Format("2006-01")
	case "year":
		return t.Format("2006")
	case "datetime":
		return t.Format("2006-01-02 15:04:05")
	default:
		// Attempt to use the format string directly if it's a valid Go layout
		// But C# formats are different (yyyy-MM-dd), so we might need more mapping if we want full compatibility.
		// For now, let's support the standard keywords.
		return t.Format("2006-01-02")
	}
}
