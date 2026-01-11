package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCsvDataFetcher_Fetch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "view1.csv")
	content := "dept,name\nD1,Alice\nD2,Bob\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	fetcher := NewCsvDataFetcher(dir)
	rows, err := fetcher.Fetch("view1", map[string]string{"dept": "D2"})
	if err != nil {
		t.Fatalf("Fetch error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if rows[0]["name"] != "Bob" {
		t.Fatalf("name = %v, want Bob", rows[0]["name"])
	}
}
