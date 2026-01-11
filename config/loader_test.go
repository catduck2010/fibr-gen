package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigBundle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bundle.yaml")

	content := `workbook:
  id: "wb1"
  name: "Report"
  template: "report.xlsx"
  outputDir: "out"
dataViews:
  - name: "view1"
    labels:
      - name: "id"
        column: "ID"
dataSources:
  - name: "source1"
    driver: "mysql"
    dsn: "root:pass@tcp(localhost:3306)/db"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write bundle: %v", err)
	}

	wb, views, dataSources, err := LoadConfigBundle(path)
	if err != nil {
		t.Fatalf("LoadConfigBundle error: %v", err)
	}

	if wb.Id != "wb1" {
		t.Fatalf("workbook id = %s, want wb1", wb.Id)
	}
	if _, ok := views["view1"]; !ok {
		t.Fatalf("expected virtual view view1 to be loaded")
	}
	if _, ok := dataSources["source1"]; !ok {
		t.Fatalf("expected data source source1 to be loaded")
	}
}

func TestLoadAllConfigs(t *testing.T) {
	dir := t.TempDir()

	workbookDir := filepath.Join(dir, "workbooks")
	dataViewDir := filepath.Join(dir, "dataViews")
	dataSourceDir := filepath.Join(dir, "datasources")
	for _, d := range []string{workbookDir, dataViewDir, dataSourceDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	wb := `id: "wb1"
name: "Report"
template: "report.xlsx"
outputDir: "out"
`
	if err := os.WriteFile(filepath.Join(workbookDir, "wb1.yaml"), []byte(wb), 0644); err != nil {
		t.Fatalf("write workbook: %v", err)
	}

	view := `name: "view1"
labels:
  - name: "id"
    column: "ID"
`
	if err := os.WriteFile(filepath.Join(dataViewDir, "view1.yaml"), []byte(view), 0644); err != nil {
		t.Fatalf("write data view: %v", err)
	}

	ds := `name: "source1"
driver: "mysql"
dsn: "root:pass@tcp(localhost:3306)/db"
`
	if err := os.WriteFile(filepath.Join(dataSourceDir, "source1.yaml"), []byte(ds), 0644); err != nil {
		t.Fatalf("write data source: %v", err)
	}

	wbs, views, dataSources, err := LoadAllConfigs(dir)
	if err != nil {
		t.Fatalf("LoadAllConfigs error: %v", err)
	}
	if _, ok := wbs["wb1"]; !ok {
		t.Fatalf("expected workbook wb1 to be loaded")
	}
	if _, ok := views["view1"]; !ok {
		t.Fatalf("expected virtual view view1 to be loaded")
	}
	if _, ok := dataSources["source1"]; !ok {
		t.Fatalf("expected data source source1 to be loaded")
	}
}

func TestLoadDataSourcesBundle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "datasources.yaml")

	content := `dataSources:
  - name: "source1"
    driver: "mysql"
    dsn: "root:pass@tcp(localhost:3306)/db"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write bundle: %v", err)
	}

	dataSources, err := LoadDataSourcesBundle(path)
	if err != nil {
		t.Fatalf("LoadDataSourcesBundle error: %v", err)
	}
	if _, ok := dataSources["source1"]; !ok {
		t.Fatalf("expected data source source1 to be loaded")
	}
}
