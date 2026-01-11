package config

import "testing"

func TestMemoryConfigRegistry_GetDataViewConfig(t *testing.T) {
	dataViews := map[string]*DataViewConfig{
		"view1": {Name: "view1"},
	}
	registry := NewMemoryConfigRegistry(dataViews)

	conf, err := registry.GetDataViewConfig("view1")
	if err != nil {
		t.Fatalf("expected config, got error: %v", err)
	}
	if conf.Name != "view1" {
		t.Fatalf("unexpected config name: %s", conf.Name)
	}
}

func TestMemoryConfigRegistry_GetDataViewConfig_NotFound(t *testing.T) {
	registry := NewMemoryConfigRegistry(map[string]*DataViewConfig{})
	if _, err := registry.GetDataViewConfig("missing"); err == nil {
		t.Fatalf("expected error for missing config")
	}
}
