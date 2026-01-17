package config

import "fmt"

// Provider defines the interface for retrieving configurations.
type Provider interface {
	GetDataViewConfig(name string) (*DataViewConfig, error)
	GetDataSourceConfig(name string) (*DataSourceConfig, error)
}

// MemoryConfigRegistry implements Provider using an in-memory map.
type MemoryConfigRegistry struct {
	dataViews   map[string]*DataViewConfig
	dataSources map[string]*DataSourceConfig
}

// NewMemoryConfigRegistry creates a new registry with the given configurations.
func NewMemoryConfigRegistry(views map[string]*DataViewConfig, sources map[string]*DataSourceConfig) *MemoryConfigRegistry {
	return &MemoryConfigRegistry{
		dataViews:   views,
		dataSources: sources,
	}
}

// GetDataViewConfig retrieves a DataViewConfig by name.
func (r *MemoryConfigRegistry) GetDataViewConfig(name string) (*DataViewConfig, error) {
	if conf, ok := r.dataViews[name]; ok {
		return conf, nil
	}
	return nil, fmt.Errorf("data view config not found: %s", name)
}

// GetDataSourceConfig retrieves a DataSourceConfig by name.
func (r *MemoryConfigRegistry) GetDataSourceConfig(name string) (*DataSourceConfig, error) {
	if conf, ok := r.dataSources[name]; ok {
		return conf, nil
	}
	return nil, fmt.Errorf("data source config not found: %s", name)
}
