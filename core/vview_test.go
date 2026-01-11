package core

import (
	"fibr-gen/config"
	"reflect"
	"testing"
)

func TestVirtualView_GetDistinctTagValues(t *testing.T) {
	// Setup
	conf := &config.VirtualViewConfig{
		Name: "test_view",
		Tags: []config.TagConfig{
			{Name: "tag_dept", Column: "DEPT"},
			{Name: "tag_name", Column: "NAME"},
		},
	}
	data := []map[string]interface{}{
		{"DEPT": "D1", "NAME": "Alice"},
		{"DEPT": "D1", "NAME": "Bob"},
		{"DEPT": "D2", "NAME": "Charlie"},
		{"DEPT": "D2", "NAME": "David"},
		{"DEPT": "D3", "NAME": ""}, // Empty value
	}
	vv := NewVirtualView(conf, data)

	tests := []struct {
		name    string
		tagName string
		want    []string
		wantErr bool
	}{
		{
			name:    "Distinct Depts",
			tagName: "tag_dept",
			want:    []string{"D1", "D2", "D3"},
			wantErr: false,
		},
		{
			name:    "Distinct Names (skip empty)",
			tagName: "tag_name",
			want:    []string{"Alice", "Bob", "Charlie", "David"},
			wantErr: false,
		},
		{
			name:    "Unknown Tag",
			tagName: "unknown_tag",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vv.GetDistinctTagValues(tt.tagName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDistinctTagValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDistinctTagValues() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVirtualView_Filter(t *testing.T) {
	// Setup
	conf := &config.VirtualViewConfig{
		Name: "test_view",
		Tags: []config.TagConfig{
			{Name: "tag_dept", Column: "DEPT"},
			{Name: "tag_age", Column: "AGE"},
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
			params:    map[string]string{"tag_dept": "D1"},
			wantCount: 2,
		},
		{
			name:      "Filter by Dept D1 AND Age 20",
			params:    map[string]string{"tag_dept": "D1", "tag_age": "20"},
			wantCount: 1,
			checkID:   1,
		},
		{
			name:      "Filter by Age 20 (Mixed Depts)",
			params:    map[string]string{"tag_age": "20"},
			wantCount: 2,
		},
		{
			name:      "No Match",
			params:    map[string]string{"tag_dept": "D3"},
			wantCount: 0,
		},
		{
			name:      "Ignore Unmapped Param",
			params:    map[string]string{"tag_dept": "D1", "unmapped_param": "xyz"},
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
			vv := NewVirtualView(conf, make([]map[string]interface{}, len(data)))
			copy(vv.Data, data) // Shallow copy slice, elements are maps (ref). 
			// Since Filter creates a new slice for vv.Data, the original 'data' slice is safe structure-wise,
			// but maps are shared. We don't modify map content, so it's fine.
			
			vv.Filter(tt.params)

			if len(vv.Data) != tt.wantCount {
				t.Errorf("Filter() count = %d, want %d", len(vv.Data), tt.wantCount)
			}

			if tt.checkID != 0 && len(vv.Data) == 1 {
				if id, ok := vv.Data[0]["ID"].(int); ok {
					if id != tt.checkID {
						t.Errorf("Filter() matched ID = %d, want %d", id, tt.checkID)
					}
				}
			}
		})
	}
}
