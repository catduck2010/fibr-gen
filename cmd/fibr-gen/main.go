package main

import (
	"fibr-gen/config"
	"fibr-gen/core"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
)

func main() {
	if err := run(os.Stdout, os.Args[1:]); err != nil {
		slog.Error("Generation failed", "error", err)
		os.Exit(1)
	}
}

func run(output io.Writer, args []string) error {
	flags := flag.NewFlagSet("fibr-gen", flag.ContinueOnError)
	flags.SetOutput(output)

	configFile := flags.String("config", "./test/config.yaml", "Path to configuration bundle")
	dataSourceFile := flags.String("datasources", "", "Path to data source bundle (optional)")
	templateDir := flags.String("templates", "./test/templates", "Template group directory")
	outputDir := flags.String("output", "./test/output", "Directory for output files")
	if err := flags.Parse(args); err != nil {
		return err
	}

	// Initialize structured logger
	logger := slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// 1. Load Config Bundle
	slog.Info("Loading configuration bundle", "file", *configFile)
	wbConf, vViews, dataSources, err := config.LoadConfigBundle(*configFile)
	if err != nil {
		return err
	}
	if *dataSourceFile != "" {
		slog.Info("Loading data source bundle", "file", *dataSourceFile)
		dataSources, err = config.LoadDataSourcesBundle(*dataSourceFile)
		if err != nil {
			return err
		}
	}
	if len(dataSources) > 0 {
		slog.Info("Loaded data sources", "count", len(dataSources))
	}

	// 2. Prepare Data Fetcher
	// In real app, we would init DB connection based on DataSources
	// Here we use CsvDataFetcher for testing if configured, or Mock

	// Check if data directory exists for CSVs
	csvFetcher := core.NewCsvDataFetcher("./test/data_csv")

	// Combined Fetcher? Or just use CsvFetcher for now as we want to test CSV.
	// Let's wrap it or use it directly.
	fetcher := csvFetcher

	// 3. Process Workbook
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
		return fmt.Errorf("generate workbook %s: %w", wbConf.Name, err)
	}

	slog.Info("Successfully generated", "name", wbConf.Name)
	return nil
}
