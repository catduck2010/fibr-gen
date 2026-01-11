package core

import (
	"fibr-gen/config"
	"testing"
	"time"
)

type countingFetcher struct {
	calls int
	data  map[string][]map[string]interface{}
}

func (f *countingFetcher) Fetch(viewName string, params map[string]string) ([]map[string]interface{}, error) {
	f.calls++
	return f.data[viewName], nil
}

func TestNewGenerationContext_MergeParams(t *testing.T) {
	wb := &config.WorkbookConfig{
		Parameters: map[string]string{
			"env":    "prod",
			"region": "us",
		},
	}

	ctx := NewGenerationContext(wb, nil, nil, map[string]string{
		"env":   "dev",
		"extra": "1",
	})

	if ctx.Parameters["env"] != "dev" {
		t.Fatalf("env = %s, want dev", ctx.Parameters["env"])
	}
	if ctx.Parameters["region"] != "us" {
		t.Fatalf("region = %s, want us", ctx.Parameters["region"])
	}
	if ctx.Parameters["extra"] != "1" {
		t.Fatalf("extra = %s, want 1", ctx.Parameters["extra"])
	}
}

func TestNewGenerationContext_ArchiveRule(t *testing.T) {
	wb := &config.WorkbookConfig{
		ArchiveRule: "$date:day:day:0",
	}
	ctx := NewGenerationContext(wb, nil, nil, nil)
	today := time.Now().Format("2006-01-02")
	if ctx.Parameters["archivedate"] != today {
		t.Fatalf("archivedate = %s, want %s", ctx.Parameters["archivedate"], today)
	}
}

func TestGenerationContext_GetVirtualViewCaching(t *testing.T) {
	vviews := map[string]*config.VirtualViewConfig{
		"view1": {
			Name: "view1",
			Tags: []config.TagConfig{{Name: "id", Column: "ID"}},
		},
	}
	registry := config.NewMemoryConfigRegistry(vviews)
	fetcher := &countingFetcher{
		data: map[string][]map[string]interface{}{
			"view1": {{"ID": "1"}},
		},
	}

	ctx := NewGenerationContext(&config.WorkbookConfig{}, registry, fetcher, nil)
	if _, err := ctx.GetVirtualView("view1"); err != nil {
		t.Fatalf("GetVirtualView error: %v", err)
	}
	if _, err := ctx.GetVirtualView("view1"); err != nil {
		t.Fatalf("GetVirtualView error: %v", err)
	}
	if fetcher.calls != 1 {
		t.Fatalf("fetcher calls = %d, want 1", fetcher.calls)
	}
}
