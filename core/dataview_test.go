package core

import (
	"fibr-gen/config"
	"reflect"
	"testing"
)

func TestDataView_GetDistinctLabelValues(t *testing.T) {
	// Setup
	conf := &config.DataViewConfig{
		Name: "test_view",
		Labels: []config.LabelConfig{
			{Name: "label_dept", Column: "DEPT"},
			{Name: "label_name", Column: "NAME"},
		},
	}
	data := []map[string]interface{}{
		{"DEPT": "D1", "NAME": "Alice"},
		{"DEPT": "D1", "NAME": "Bob"},
		{"DEPT": "D2", "NAME": "Charlie"},
		{"DEPT": "D2", "NAME": "David"},
		{"DEPT": "D3", "NAME": ""}, // Empty value
	}
	vv := NewDataView(conf, data)

	tests := []struct {
		name    string
		label   string
		want    []string
		wantErr bool
	}{
		{
			name:    "Distinct Depts",
			label:   "label_dept",
			want:    []string{"D1", "D2", "D3"},
			wantErr: false,
		},
		{
			name:    "Distinct Names (skip empty)",
			label:   "label_name",
			want:    []string{"Alice", "Bob", "Charlie", "David"},
			wantErr: false,
		},
		{
			name:    "Unknown Tag",
			label:   "unknown_tag",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vv.GetDistinctLabelValues(tt.label)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDistinctLabelValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDistinctLabelValues() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDataView_Filter(t *testing.T) {
	// Setup
	conf := &config.DataViewConfig{
		Name: "test_view",
		Labels: []config.LabelConfig{
			{Name: "label_dept", Column: "DEPT"},
			{Name: "label_age", Column: "AGE"},
		},
	}
	// Initial Data: 4 rows
	data := []map[string]interface{}{
		{"DEPT": "D1", "AGE": 20, "ID": 1},
		{"DEPT": "D1", "AGE": 30, "ID": 2},
		{"DEPT": "D2", "AGE": 20, "ID": 3},
		{"DEPT": "D2", "AGE": 40, "ID": 4},
	}

	tests := []struct {
		name      string
		params    map[string]string
		wantCount int
		checkID   int // Optional check for a specific ID in result to verify correctness
	}{
		{
			name:      "Filter by Dept D1",
			params:    map[string]string{"label_dept": "D1"},
			wantCount: 2,
		},
		{
			name:      "Filter by Dept D1 AND Age 20",
			params:    map[string]string{"label_dept": "D1", "label_age": "20"},
			wantCount: 1,
			checkID:   1,
		},
		{
			name:      "Filter by Age 20 (Mixed Depts)",
			params:    map[string]string{"label_age": "20"},
			wantCount: 2,
		},
		{
			name:      "No Match",
			params:    map[string]string{"label_dept": "D3"},
			wantCount: 0,
		},
		{
			name:      "Ignore Unmapped Param",
			params:    map[string]string{"label_dept": "D1", "unmapped_param": "xyz"},
			wantCount: 2, // Should still match D1, ignoring unmapped
		},
		{
			name:      "Empty Params",
			params:    map[string]string{},
			wantCount: 4, // No filtering
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a FRESH view for each test because Filter modifies data in-place
			v := NewDataView(conf, make([]map[string]interface{}, len(data)))
			copy(v.Data, data) // Shallow copy slice, elements are maps (ref).
			// Since Filter creates a new slice for v.Data, the original 'data' slice is safe structure-wise,
			// but maps are shared. We don't modify map content, so it's fine.

			v.Filter(tt.params)

			if len(v.Data) != tt.wantCount {
				t.Errorf("Filter() count = %d, want %d", len(v.Data), tt.wantCount)
			}

			if tt.checkID != 0 && len(v.Data) == 1 {
				if id, ok := v.Data[0]["ID"].(int); ok {
					if id != tt.checkID {
						t.Errorf("Filter() matched ID = %d, want %d", id, tt.checkID)
					}
				}
			}
		})
	}
}
