package config

import "testing"

func TestMemoryConfigRegistry_GetVirtualViewConfig(t *testing.T) {
	vviews := map[string]*VirtualViewConfig{
		"view1": {Name: "view1"},
	}
	registry := NewMemoryConfigRegistry(vviews)

	conf, err := registry.GetVirtualViewConfig("view1")
	if err != nil {
		t.Fatalf("expected config, got error: %v", err)
	}
	if conf.Name != "view1" {
		t.Fatalf("unexpected config name: %s", conf.Name)
	}
}

func TestMemoryConfigRegistry_GetVirtualViewConfig_NotFound(t *testing.T) {
	registry := NewMemoryConfigRegistry(map[string]*VirtualViewConfig{})
	if _, err := registry.GetVirtualViewConfig("missing"); err == nil {
		t.Fatalf("expected error for missing config")
	}
}
