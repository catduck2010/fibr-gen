package core

import (
	"testing"
	"time"
)

func TestParseDynamicDate(t *testing.T) {
	// Fixed base times for deterministic testing
	baseNormal := time.Date(2023, 5, 15, 10, 0, 0, 0, time.UTC) // 2023-05-15
	baseLeap := time.Date(2024, 2, 29, 10, 0, 0, 0, time.UTC)   // 2024-02-29 (Leap Day)
	baseEOY := time.Date(2023, 12, 31, 10, 0, 0, 0, time.UTC)   // 2023-12-31

	tests := []struct {
		name       string
		expression string
		baseTime   time.Time
		want       string
		wantErr    bool
	}{
		// --- Standard Scenarios ---
		{
			name:       "Static String (No Prefix)",
			expression: "2023-01-01",
			baseTime:   baseNormal,
			want:       "2023-01-01",
			wantErr:    false,
		},
		{
			name:       "Today",
			expression: "$date:day:day:0",
			baseTime:   baseNormal,
			want:       "2023-05-15",
			wantErr:    false,
		},
		{
			name:       "Yesterday",
			expression: "$date:day:day:-1",
			baseTime:   baseNormal,
			want:       "2023-05-14",
			wantErr:    false,
		},
		{
			name:       "Tomorrow",
			expression: "$date:day:day:1",
			baseTime:   baseNormal,
			want:       "2023-05-16",
			wantErr:    false,
		},
		{
			name:       "Next Month",
			expression: "$date:day:month:1",
			baseTime:   baseNormal,
			want:       "2023-06-15",
			wantErr:    false,
		},
		{
			name:       "Last Year",
			expression: "$date:day:year:-1",
			baseTime:   baseNormal,
			want:       "2022-05-15",
			wantErr:    false,
		},

		// --- Formatting Scenarios ---
		{
			name:       "Format Month",
			expression: "$date:month:day:0",
			baseTime:   baseNormal,
			want:       "2023-05",
			wantErr:    false,
		},
		{
			name:       "Format Year",
			expression: "$date:year:day:0",
			baseTime:   baseNormal,
			want:       "2023",
			wantErr:    false,
		},
		{
			name:       "Format DateTime",
			expression: "$date:datetime:day:0",
			baseTime:   baseNormal,
			want:       "2023-05-15 10:00:00",
			wantErr:    false,
		},

		// --- Edge Cases (Leap Years & Boundaries) ---
		{
			name:       "Leap Day + 1 Year",
			expression: "$date:day:year:1",
			baseTime:   baseLeap,
			want:       "2025-03-01", // Go normalizes Feb 29 2025 to Mar 1
			wantErr:    false,
		},
		{
			name:       "Leap Day - 1 Year",
			expression: "$date:day:year:-1",
			baseTime:   baseLeap,
			want:       "2023-03-01", // Go normalizes Feb 29 2023 to Mar 1
			wantErr:    false,
		},
		{
			name:       "End of Year + 1 Day",
			expression: "$date:day:day:1",
			baseTime:   baseEOY,
			want:       "2024-01-01",
			wantErr:    false,
		},
		{
			name:       "Month Boundary Normalization (Jan 31 + 1 Month)",
			expression: "$date:day:month:1",
			baseTime:   time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC),
			want:       "2023-03-03", // Jan 31 + 1 month = Feb 31 -> Mar 3 (28+3)
			wantErr:    false,
		},
		// --- 1900 Problem (Excel Compatibility) ---
		// Excel incorrectly treats 1900 as a leap year (1900-02-29 exists in Excel).
		// Go's time package correctly treats 1900 as a common year.
		// We verify that Go does NOT generate 1900-02-29.
		{
			name:       "1900 Feb 28 + 1 Day",
			expression: "$date:day:day:1",
			baseTime:   time.Date(1900, 2, 28, 10, 0, 0, 0, time.UTC),
			want:       "1900-03-01", // Correct Gregorian behavior (Excel would think this is Feb 29)
			wantErr:    false,
		},
		{
			name:       "1900 Mar 1 - 1 Day",
			expression: "$date:day:day:-1",
			baseTime:   time.Date(1900, 3, 1, 10, 0, 0, 0, time.UTC),
			want:       "1900-02-28", // Correct Gregorian behavior
			wantErr:    false,
		},

		// --- Error Scenarios ---
		{
			name:       "Invalid Format (Too Short)",
			expression: "$date:day:day",
			baseTime:   baseNormal,
			want:       "",
			wantErr:    true,
		},
		{
			name:       "Invalid Offset (Not a Number)",
			expression: "$date:day:day:abc",
			baseTime:   baseNormal,
			want:       "",
			wantErr:    true,
		},
		{
			name:       "Unsupported Unit",
			expression: "$date:day:century:1",
			baseTime:   baseNormal,
			want:       "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDynamicDate(tt.expression, tt.baseTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDynamicDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDynamicDate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
