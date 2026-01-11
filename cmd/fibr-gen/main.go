package main

import (
	"context"
	"database/sql"
	"fibr-gen/config"
	"fibr-gen/core"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"

	// Database drivers

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
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
	fetcherType := flags.String("fetcher", "csv", "Data fetcher type: csv, dynamodb, mysql, postgres")
	dbDSN := flags.String("db-dsn", "", "Database connection string (DSN) for mysql/postgres")
	s3Bucket := flags.String("s3-bucket", "", "S3 bucket name for uploading output")
	s3Prefix := flags.String("s3-prefix", "fibr-gen-output", "S3 prefix (folder) for uploaded files")

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
	wbConf, views, dataSources, err := config.LoadConfigBundle(*configFile)
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
	var fetcher core.DataFetcher

	switch *fetcherType {
	case "dynamodb":
		slog.Info("Initializing DynamoDB Data Fetcher")
		// Load AWS Config (handles env vars, IAM roles, etc.)
		cfg, err := awsconfig.LoadDefaultConfig(context.TODO())
		if err != nil {
			return fmt.Errorf("unable to load AWS SDK config: %w", err)
		}
		fetcher = core.NewDynamoDBDataFetcher(cfg)
	case "mysql", "postgres":
		if *dbDSN == "" {
			return fmt.Errorf("db-dsn is required for %s fetcher", *fetcherType)
		}
		slog.Info("Initializing SQL Data Fetcher", "type", *fetcherType)
		db, err := sql.Open(*fetcherType, *dbDSN)
		if err != nil {
			return fmt.Errorf("failed to open db connection: %w", err)
		}
		// Verify connection
		if err := db.Ping(); err != nil {
			return fmt.Errorf("failed to ping db: %w", err)
		}
		fetcher = core.NewSQLDataFetcher(db, *fetcherType)
	default:
		// Default to CSV
		slog.Info("Initializing CSV Data Fetcher", "dir", "./test/data_csv")
		fetcher = core.NewCsvDataFetcher("./test/data_csv")
	}

	// 3. Process Workbook
	slog.Info("Processing Workbook", "name", wbConf.Name, "id", wbConf.Id)

	// Create Config Registry
	configRegistry := config.NewMemoryConfigRegistry(views)

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

	// 4. Upload to S3 if configured
	if *s3Bucket != "" {
		slog.Info("Starting S3 upload", "bucket", *s3Bucket, "prefix", *s3Prefix)

		// Load AWS Config if not already loaded (e.g. if fetcher was CSV)
		// It's cheap to load again or we could have shared it.
		cfg, err := awsconfig.LoadDefaultConfig(context.TODO())
		if err != nil {
			return fmt.Errorf("unable to load AWS SDK config for S3: %w", err)
		}

		uploader := core.NewS3Uploader(cfg, *s3Bucket, *s3Prefix)
		if err := uploader.UploadDirectory(*outputDir); err != nil {
			return fmt.Errorf("failed to upload output to s3: %w", err)
		}
		slog.Info("Successfully uploaded to S3")
	}

	return nil
}
