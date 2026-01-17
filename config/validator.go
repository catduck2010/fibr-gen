package config

import (
	"fmt"
)

// Validator validates the configuration objects.
type Validator struct {
	Provider Provider
}

// NewValidator creates a new Validator.
func NewValidator(provider Provider) *Validator {
	return &Validator{Provider: provider}
}

// ValidateWorkbook validates the WorkbookConfig.
func (v *Validator) ValidateWorkbook(wb *WorkbookConfig) error {
	if wb.Name == "" {
		return fmt.Errorf("workbook name is required")
	}
	if wb.Template == "" {
		return fmt.Errorf("workbook template is required")
	}
	if wb.OutputDir == "" {
		return fmt.Errorf("workbook output directory is required")
	}
	if len(wb.Sheets) == 0 {
		return fmt.Errorf("workbook must have at least one sheet")
	}

	for i := range wb.Sheets {
		if err := v.ValidateSheet(&wb.Sheets[i]); err != nil {
			return fmt.Errorf("sheet %d error: %w", i, err)
		}
	}
	return nil
}

// ValidateSheet validates the SheetConfig.
func (v *Validator) ValidateSheet(sheet *SheetConfig) error {
	if sheet.Name == "" {
		return fmt.Errorf("sheet name is required")
	}
	if sheet.Dynamic {
		if sheet.DataViewName == "" {
			return fmt.Errorf("dynamic sheet '%s' requires a DataViewName", sheet.Name)
		}
		if sheet.ParamLabel == "" {
			return fmt.Errorf("dynamic sheet '%s' requires a ParamLabel", sheet.Name)
		}
		// Verify DataView exists
		if v.Provider != nil {
			if _, err := v.Provider.GetDataViewConfig(sheet.DataViewName); err != nil {
				return fmt.Errorf("sheet '%s' references unknown DataView '%s'", sheet.Name, sheet.DataViewName)
			}
		}
	}

	for i := range sheet.Blocks {
		if err := v.ValidateBlock(&sheet.Blocks[i]); err != nil {
			return fmt.Errorf("block %d error: %w", i, err)
		}
	}
	return nil
}

// ValidateBlock validates the BlockConfig.
func (v *Validator) ValidateBlock(block *BlockConfig) error {
	if block.Name == "" {
		return fmt.Errorf("block name is required")
	}
	if block.Type == "" {
		return fmt.Errorf("block '%s' type is required", block.Name)
	}
	// Check valid type enum
	switch block.Type {
	case BlockTypeValue, BlockTypeHeader, BlockTypeMatrix:
		// OK
	default:
		return fmt.Errorf("block '%s' has invalid type '%s'", block.Name, block.Type)
	}

	if block.Range.Ref == "" {
		return fmt.Errorf("block '%s' range is required", block.Name)
	}

	if block.DataViewName != "" && v.Provider != nil {
		if _, err := v.Provider.GetDataViewConfig(block.DataViewName); err != nil {
			return fmt.Errorf("block '%s' references unknown DataView '%s'", block.Name, block.DataViewName)
		}
	}

	if block.Type == BlockTypeMatrix {
		if len(block.SubBlocks) == 0 {
			return fmt.Errorf("matrix block '%s' must have sub-blocks", block.Name)
		}
		hasVertical := false
		hasHorizontal := false
		for i := range block.SubBlocks {
			sb := &block.SubBlocks[i]
			if err := v.ValidateBlock(sb); err != nil {
				return fmt.Errorf("matrix block '%s' sub-block %d error: %w", block.Name, i, err)
			}
			if sb.Type == BlockTypeHeader {
				switch sb.Direction {
				case DirectionVertical, "":
					hasVertical = true
				case DirectionHorizontal:
					hasHorizontal = true
				}
			}
		}
		if !hasVertical || !hasHorizontal {
			return fmt.Errorf("matrix block '%s' must have both vertical and horizontal header blocks", block.Name)
		}
	} else {
		// Non-matrix blocks can also have sub-blocks?
		// Currently structure allows it, but validation usually checks recursion if needed.
		// For now, only recursive validation for Matrix or if SubBlocks exist.
		for i := range block.SubBlocks {
			if err := v.ValidateBlock(&block.SubBlocks[i]); err != nil {
				return fmt.Errorf("block '%s' sub-block %d error: %w", block.Name, i, err)
			}
		}
	}

	return nil
}

// ValidateDataView validates the DataViewConfig.
func (v *Validator) ValidateDataView(dv *DataViewConfig) error {
	if dv.Name == "" {
		return fmt.Errorf("data view name is required")
	}
	if dv.DataSource == "" {
		return fmt.Errorf("data view '%s' requires a DataSource", dv.Name)
	}
	if v.Provider != nil {
		if _, err := v.Provider.GetDataSourceConfig(dv.DataSource); err != nil {
			return fmt.Errorf("data view '%s' references unknown DataSource '%s'", dv.Name, dv.DataSource)
		}
	}
	for i, label := range dv.Labels {
		if label.Name == "" {
			return fmt.Errorf("data view '%s' label %d name is required", dv.Name, i)
		}
		if label.Column == "" {
			return fmt.Errorf("data view '%s' label %d column is required", dv.Name, i)
		}
	}
	return nil
}

// ValidateDataSource validates the DataSourceConfig.
func (v *Validator) ValidateDataSource(ds *DataSourceConfig) error {
	if ds.Name == "" {
		return fmt.Errorf("data source name is required")
	}
	if ds.Driver == "" {
		return fmt.Errorf("data source '%s' driver is required", ds.Name)
	}
	if ds.DSN == "" {
		return fmt.Errorf("data source '%s' DSN is required", ds.Name)
	}
	return nil
}
