package config

import "fmt"

// ConfigProvider defines the interface for retrieving configurations.
type ConfigProvider interface {
	GetVirtualViewConfig(name string) (*VirtualViewConfig, error)
}

// MemoryConfigRegistry implements ConfigProvider using an in-memory map.
type MemoryConfigRegistry struct {
	vviews map[string]*VirtualViewConfig
}

// NewMemoryConfigRegistry creates a new registry with the given configurations.
func NewMemoryConfigRegistry(vviews map[string]*VirtualViewConfig) *MemoryConfigRegistry {
	return &MemoryConfigRegistry{
		vviews: vviews,
	}
}

// GetVirtualViewConfig retrieves a VirtualViewConfig by name.
func (r *MemoryConfigRegistry) GetVirtualViewConfig(name string) (*VirtualViewConfig, error) {
	if conf, ok := r.vviews[name]; ok {
		return conf, nil
	}
	return nil, fmt.Errorf("virtual view config not found: %s", name)
}
