package core

import (
	"os"
	"path/filepath"
	"testing"

	"fibr-gen/config"

	"github.com/xuri/excelize/v2"
)

func TestReplacePlaceholders(t *testing.T) {
	params := map[string]string{
		"month": "jan",
		"day":   "01",
	}

	got := replacePlaceholders("reports/${month}/report-${day}", params)
	if got != "reports/jan/report-01" {
		t.Fatalf("expected placeholder replacements, got %q", got)
	}
}

func TestGenerateOutputPathWithParams(t *testing.T) {
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	outputDir := filepath.Join(tmpDir, "outputs")

	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	templatePath := filepath.Join(templateDir, "template.xlsx")
	workbook := excelize.NewFile()
	if err := workbook.SaveAs(templatePath); err != nil {
		t.Fatalf("failed to save template: %v", err)
	}

	wbConf := &config.WorkbookConfig{
		Name:      "report-${day}",
		Template:  "template.xlsx",
		OutputDir: "reports/${month}",
	}

	ctx := NewGenerationContext(
		wbConf,
		config.NewMemoryConfigRegistry(map[string]*config.VirtualViewConfig{}),
		&MockDataFetcher{Data: map[string][]map[string]interface{}{}},
		map[string]string{
			"month": "jan",
			"day":   "01",
		},
	)

	gen := NewGenerator(ctx)
	if err := gen.Generate(templateDir, outputDir); err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	expectedPath := filepath.Join(outputDir, "reports", "jan", "report-01.xlsx")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected output file at %s: %v", expectedPath, err)
	}
}
