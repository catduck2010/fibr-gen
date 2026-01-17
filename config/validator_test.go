package config

import (
	"strings"
	"testing"
)

func TestValidator_ValidateWorkbook(t *testing.T) {
	// Setup Provider
	views := map[string]*DataViewConfig{
		"view1": {Name: "view1"},
	}
	provider := NewMemoryConfigRegistry(views, nil)
	validator := NewValidator(provider)

	tests := []struct {
		name    string
		wb      *WorkbookConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid Workbook",
			wb: &WorkbookConfig{
				Id:        "wb1",
				Name:      "Report",
				Template:  "tpl.xlsx",
				OutputDir: "out",
				Sheets: []SheetConfig{
					{
						Name: "Sheet1",
						Blocks: []BlockConfig{
							{
								Name:         "Block1",
								Type:         BlockTypeValue,
								Range:        CellRange{Ref: "A1"},
								DataViewName: "view1",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Missing Name",
			wb: &WorkbookConfig{
				Template:  "tpl.xlsx",
				OutputDir: "out",
				Sheets:    []SheetConfig{{Name: "S1"}},
			},
			wantErr: true,
			errMsg:  "workbook name is required",
		},
		{
			name: "Missing Sheets",
			wb: &WorkbookConfig{
				Name:      "Report",
				Template:  "tpl.xlsx",
				OutputDir: "out",
				Sheets:    []SheetConfig{},
			},
			wantErr: true,
			errMsg:  "workbook must have at least one sheet",
		},
		{
			name: "Invalid Sheet (Dynamic missing view)",
			wb: &WorkbookConfig{
				Name:      "Report",
				Template:  "tpl.xlsx",
				OutputDir: "out",
				Sheets: []SheetConfig{
					{
						Name:    "DynamicSheet",
						Dynamic: true,
						// Missing DataViewName
						ParamLabel: "p1",
					},
				},
			},
			wantErr: true,
			errMsg:  "requires a DataViewName",
		},
		{
			name: "Invalid Block (Unknown DataView)",
			wb: &WorkbookConfig{
				Name:      "Report",
				Template:  "tpl.xlsx",
				OutputDir: "out",
				Sheets: []SheetConfig{
					{
						Name: "Sheet1",
						Blocks: []BlockConfig{
							{
								Name:         "Block1",
								Type:         BlockTypeValue,
								Range:        CellRange{Ref: "A1"},
								DataViewName: "unknown_view",
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "unknown DataView",
		},
		{
			name: "Invalid Matrix Block (Missing Axes)",
			wb: &WorkbookConfig{
				Name:      "Report",
				Template:  "tpl.xlsx",
				OutputDir: "out",
				Sheets: []SheetConfig{
					{
						Name: "Sheet1",
						Blocks: []BlockConfig{
							{
								Name:  "Matrix1",
								Type:  BlockTypeMatrix,
								Range: CellRange{Ref: "A1:C3"},
								SubBlocks: []BlockConfig{
									// Only one axis
									{
										Name:      "VAxis",
										Type:      BlockTypeHeader,
										Direction: DirectionVertical,
										Range:     CellRange{Ref: "A2"},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "must have both vertical and horizontal header blocks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateWorkbook(tt.wb)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWorkbook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateWorkbook() error = %v, want error containing %s", err, tt.errMsg)
			}
		})
	}
}

func TestValidator_ValidateDataView(t *testing.T) {
	vSources := map[string]*DataSourceConfig{
		"ds1": {Name: "ds1"},
	}
	provider := NewMemoryConfigRegistry(nil, vSources)
	validator := NewValidator(provider)
	tests := []struct {
		name    string
		dv      *DataViewConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid DataView",
			dv: &DataViewConfig{
				Name:       "view1",
				DataSource: "ds1",
				Labels: []LabelConfig{
					{Name: "l1", Column: "c1"},
				},
			},
			wantErr: false,
		},
		{
			name:    "Missing Name",
			dv:      &DataViewConfig{DataSource: "ds1"},
			wantErr: true,
			errMsg:  "data view name is required",
		},
		{
			name:    "Missing DataSource",
			dv:      &DataViewConfig{Name: "view1"},
			wantErr: true,
			errMsg:  "requires a DataSource",
		},
		{
			name: "Unknown DataSource",
			dv: &DataViewConfig{
				Name:       "view1",
				DataSource: "unknown_ds",
				Labels: []LabelConfig{
					{Name: "l1", Column: "c1"},
				},
			},
			wantErr: true,
			errMsg:  "unknown DataSource",
		},
		{
			name: "Invalid Label",
			dv: &DataViewConfig{
				Name:       "view1",
				DataSource: "ds1",
				Labels: []LabelConfig{
					{Name: "", Column: "c1"},
				},
			},
			wantErr: true,
			errMsg:  "label 0 name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDataView(tt.dv)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDataView() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateDataView() error = %v, want error containing %s", err, tt.errMsg)
			}
		})
	}
}

func TestValidator_ValidateDataSource(t *testing.T) {
	validator := NewValidator(nil)
	tests := []struct {
		name    string
		ds      *DataSourceConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid DataSource",
			ds: &DataSourceConfig{
				Name:   "ds1",
				Driver: "mysql",
				DSN:    "user:pass@tcp(localhost:3306)/db",
			},
			wantErr: false,
		},
		{
			name:    "Missing Name",
			ds:      &DataSourceConfig{Driver: "mysql", DSN: "dsn"},
			wantErr: true,
			errMsg:  "data source name is required",
		},
		{
			name:    "Missing Driver",
			ds:      &DataSourceConfig{Name: "ds1", DSN: "dsn"},
			wantErr: true,
			errMsg:  "driver is required",
		},
		{
			name:    "Missing DSN",
			ds:      &DataSourceConfig{Name: "ds1", Driver: "mysql"},
			wantErr: true,
			errMsg:  "DSN is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDataSource(tt.ds)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDataSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateDataSource() error = %v, want error containing %s", err, tt.errMsg)
			}
		})
	}
}
