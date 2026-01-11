package main

import (
	"fibr-gen/config"
	"fibr-gen/core"
	"flag"
	"log/slog"
	"os"
)

func main() {
	configDir := flag.String("config", "./test", "Directory containing configurations")
	templateDir := flag.String("templates", "./test/templates", "Directory containing Excel templates")
	outputDir := flag.String("output", "./test/output", "Directory for output files")
	flag.Parse()

	// Initialize structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// 1. Load Configs
	slog.Info("Loading configurations", "dir", *configDir)
	wbs, vViews, _, err := config.LoadAllConfigs(*configDir)
	if err != nil {
		slog.Error("Failed to load configs", "error", err)
		os.Exit(1)
	}

	// 2. Prepare Data Fetcher
	// In real app, we would init DB connection based on DataSources
	// Here we use CsvDataFetcher for testing if configured, or Mock

	// Check if data directory exists for CSVs
	csvFetcher := core.NewCsvDataFetcher("./test/data_csv")

	// Combined Fetcher? Or just use CsvFetcher for now as we want to test CSV.
	// Let's wrap it or use it directly.
	fetcher := csvFetcher

	// 3. Process Workbooks
	for _, wbConf := range wbs {
		slog.Info("Processing Workbook", "name", wbConf.Name, "id", wbConf.Id)

		// Create Config Registry
		configRegistry := config.NewMemoryConfigRegistry(vViews)

		// Create Context
		// Pass Registry instead of raw map
		ctx := core.NewGenerationContext(wbConf, configRegistry, fetcher, map[string]string{
			"env": "dev",
		})

		generator := core.NewGenerator(ctx)
		if err := generator.Generate(*templateDir, *outputDir); err != nil {
			slog.Error("Error generating workbook", "name", wbConf.Name, "error", err)
		} else {
			slog.Info("Successfully generated", "name", wbConf.Name)
		}
	}
}
