package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadWorkbookConfig loads a workbook configuration from a YAML file.
func LoadWorkbookConfig(path string) (*WorkbookConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read workbook config file: %w", err)
	}

	var cfg WorkbookConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse workbook config: %w", err)
	}

	return &cfg, nil
}

// LoadVirtualViewConfig loads a virtual view configuration from a YAML file.
func LoadVirtualViewConfig(path string) (*VirtualViewConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read virtual view config file: %w", err)
	}

	var cfg VirtualViewConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse virtual view config: %w", err)
	}

	return &cfg, nil
}

// LoadDataSourceConfig loads a data source configuration from a YAML file.
func LoadDataSourceConfig(path string) (*DataSourceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read data source config file: %w", err)
	}

	var cfg DataSourceConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse data source config: %w", err)
	}

	return &cfg, nil
}

// LoadAllConfigs loads all configurations from a directory.
// It expects subdirectories or naming conventions to distinguish types,
// or it tries to parse into different structs.
// For simplicity, let's assume a structure like:
// test/
//
//	workbooks/
//	  wb1.yaml
//	vviews/
//	  vv1.yaml
//	datasources/
//	  ds1.yaml
func LoadAllConfigs(rootDir string) (map[string]*WorkbookConfig, map[string]*VirtualViewConfig, map[string]*DataSourceConfig, error) {
	workbooks := make(map[string]*WorkbookConfig)
	vViews := make(map[string]*VirtualViewConfig)
	dataSources := make(map[string]*DataSourceConfig)

	// Helper to walk and load
	walkDir := func(subDir string, loader func(string) error) error {
		path := filepath.Join(rootDir, subDir)
		entries, err := os.ReadDir(path)
		if os.IsNotExist(err) {
			return nil // Optional directory
		}
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if !entry.IsDir() && (filepath.Ext(entry.Name()) == ".yaml" || filepath.Ext(entry.Name()) == ".yml") {
				if err := loader(filepath.Join(path, entry.Name())); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Load DataSources
	err := walkDir("datasources", func(f string) error {
		cfg, err := LoadDataSourceConfig(f)
		if err != nil {
			return err
		}
		dataSources[cfg.Name] = cfg
		return nil
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("loading datasources: %w", err)
	}

	// Load VirtualViews
	err = walkDir("vviews", func(f string) error {
		cfg, err := LoadVirtualViewConfig(f)
		if err != nil {
			return err
		}
		vViews[cfg.Name] = cfg
		return nil
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("loading vviews: %w", err)
	}

	// Load Workbooks
	err = walkDir("workbooks", func(f string) error {
		cfg, err := LoadWorkbookConfig(f)
		if err != nil {
			return err
		}
		workbooks[cfg.Id] = cfg
		return nil
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("loading workbooks: %w", err)
	}

	return workbooks, vViews, dataSources, nil
}
