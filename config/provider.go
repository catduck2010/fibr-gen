package config

import "fmt"

// Provider defines the interface for retrieving configurations.
type Provider interface {
	GetDataViewConfig(name string) (*DataViewConfig, error)
}

// MemoryConfigRegistry implements Provider using an in-memory map.
type MemoryConfigRegistry struct {
	dataViews map[string]*DataViewConfig
}

// NewMemoryConfigRegistry creates a new registry with the given configurations.
func NewMemoryConfigRegistry(v map[string]*DataViewConfig) *MemoryConfigRegistry {
	return &MemoryConfigRegistry{
		dataViews: v,
	}
}

// GetDataViewConfig retrieves a DataViewConfig by name.
func (r *MemoryConfigRegistry) GetDataViewConfig(name string) (*DataViewConfig, error) {
	if conf, ok := r.dataViews[name]; ok {
		return conf, nil
	}
	return nil, fmt.Errorf("virtual view config not found: %s", name)
}
