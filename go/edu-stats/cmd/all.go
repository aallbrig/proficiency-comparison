package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/aallbrig/proficiency-comparison/internal/database"
	"github.com/aallbrig/proficiency-comparison/internal/downloaders"
	"github.com/aallbrig/proficiency-comparison/internal/generators"
)

var (
	years  string
	dryRun bool
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Run the complete data pipeline",
	Long: `Run all pipeline steps: schema check, data downloads, processing, and Hugo asset generation.
	
Supports resumability - if interrupted, will continue from last successful step.`,
	RunE: runAll,
}

func init() {
	allCmd.Flags().StringVar(&years, "years", "1970-2025", "Year range to download (format: YYYY-YYYY)")
	allCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate without downloading data")
	
	// Add subcommands for individual steps
	allCmd.AddCommand(checkSchemaCmd)
	allCmd.AddCommand(downloadWorldBankCmd)
	allCmd.AddCommand(downloadCensusCmd)
	allCmd.AddCommand(downloadNCESCmd)
	allCmd.AddCommand(downloadNAEPCmd)
	allCmd.AddCommand(downloadECLSCmd)
	allCmd.AddCommand(generateAssetsCmd)
}

func parseYears(yearRange string) (int, int, error) {
	parts := strings.Split(yearRange, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid year range format, use YYYY-YYYY")
	}
	var start, end int
	_, err := fmt.Sscanf(parts[0], "%d", &start)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start year: %w", err)
	}
	_, err = fmt.Sscanf(parts[1], "%d", &end)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end year: %w", err)
	}
	if start > end {
		return 0, 0, fmt.Errorf("start year must be before end year")
	}
	return start, end, nil
}

func runAll(cmd *cobra.Command, args []string) error {
	startYear, endYear, err := parseYears(years)
	if err != nil {
		return err
	}

	fmt.Printf("Educational Stats CLI - Full Pipeline\n")
	fmt.Printf("=====================================\n")
	fmt.Printf("Year range: %d-%d\n", startYear, endYear)
	fmt.Printf("Dry run: %v\n\n", dryRun)

	// Open database
	db, err := database.Open()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// First, always run status
	fmt.Println("📊 Running status check...")
	if err := runStatus(cmd, args); err != nil {
		fmt.Printf("Warning: status check failed: %v\n\n", err)
	}

	steps := []struct {
		name string
		fn   func() error
	}{
		{"check-schema", func() error { return runCheckSchema(cmd, args) }},
		{"download-worldbank", func() error { return runDownloadWorldBank(startYear, endYear, dryRun) }},
		{"download-census", func() error { return runDownloadCensus(startYear, endYear, dryRun) }},
		{"download-nces", func() error { return runDownloadNCES(startYear, endYear, dryRun) }},
		{"download-naep", func() error { return runDownloadNAEP(startYear, endYear, dryRun) }},
		{"download-ecls", func() error { return runDownloadECLS(startYear, endYear, dryRun) }},
		{"generate-assets", func() error { return runGenerateAssets(cmd, args) }},
	}

	// Check for resumability
	lastStep, err := database.GetLastCompletedStep(db)
	if err != nil {
		fmt.Printf("Warning: could not check last completed step: %v\n", err)
		lastStep = ""
	}

	startIndex := 0
	if lastStep != "" {
		fmt.Printf("📌 Resuming from last completed step: %s\n\n", lastStep)
		for i, step := range steps {
			if step.name == lastStep {
				startIndex = i + 1
				break
			}
		}
	}

	// Execute steps
	for i := startIndex; i < len(steps); i++ {
		step := steps[i]
		fmt.Printf("\n🔄 Step %d/%d: %s\n", i+1, len(steps), step.name)
		fmt.Println(strings.Repeat("-", 50))
		
		startTime := time.Now()
		
		// Record step start
		if !dryRun {
			database.RecordPipelineStep(db, step.name, "started", "", nil)
		}
		
		err := step.fn()
		elapsed := time.Since(startTime)
		
		if err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
			if !dryRun {
				database.RecordPipelineStep(db, step.name, "failed", "", &err)
			}
			return fmt.Errorf("pipeline failed at step %s: %w", step.name, err)
		}
		
		fmt.Printf("✓ Completed in %.2fs\n", elapsed.Seconds())
		if !dryRun {
			database.RecordPipelineStep(db, step.name, "completed", 
				fmt.Sprintf("%d-%d", startYear, endYear), nil)
		}
	}

	fmt.Printf("\n✅ Pipeline completed successfully!\n")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. View data: edu-stats status\n")
	fmt.Printf("  2. Run website: cd hugo/site && hugo server\n")
	
	return nil
}

// Individual step commands
var checkSchemaCmd = &cobra.Command{
	Use:   "check-schema",
	Short: "Check and apply database schema",
	RunE:  runCheckSchema,
}

func runCheckSchema(cmd *cobra.Command, args []string) error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()
	
	fmt.Println("Checking database schema...")
	return database.ApplySchema(db)
}

var downloadWorldBankCmd = &cobra.Command{
	Use:   "download-worldbank",
	Short: "Download literacy data from World Bank",
	RunE: func(cmd *cobra.Command, args []string) error {
		start, end, _ := parseYears(years)
		return runDownloadWorldBank(start, end, dryRun)
	},
}

func runDownloadWorldBank(startYear, endYear int, dryRun bool) error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()
	
	downloader := downloaders.NewWorldBankDownloader(db)
	return downloader.Download(startYear, endYear, dryRun)
}

var downloadCensusCmd = &cobra.Command{
	Use:   "download-census",
	Short: "Download educational attainment from Census Bureau",
	RunE: func(cmd *cobra.Command, args []string) error {
		start, end, _ := parseYears(years)
		return runDownloadCensus(start, end, dryRun)
	},
}

func runDownloadCensus(startYear, endYear int, dryRun bool) error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()
	
	downloader := downloaders.NewCensusDownloader(db)
	return downloader.Download(startYear, endYear, dryRun)
}

var downloadNCESCmd = &cobra.Command{
	Use:   "download-nces",
	Short: "Download graduation/enrollment from NCES",
	RunE: func(cmd *cobra.Command, args []string) error {
		start, end, _ := parseYears(years)
		return runDownloadNCES(start, end, dryRun)
	},
}

func runDownloadNCES(startYear, endYear int, dryRun bool) error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()
	
	downloader := downloaders.NewNCESDownloader(db)
	return downloader.Download(startYear, endYear, dryRun)
}

var downloadNAEPCmd = &cobra.Command{
	Use:   "download-naep",
	Short: "Download test proficiency from NAEP",
	RunE: func(cmd *cobra.Command, args []string) error {
		start, end, _ := parseYears(years)
		return runDownloadNAEP(start, end, dryRun)
	},
}

func runDownloadNAEP(startYear, endYear int, dryRun bool) error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()
	
	downloader := downloaders.NewNAEPDownloader(db)
	return downloader.Download(startYear, endYear, dryRun)
}

var downloadECLSCmd = &cobra.Command{
	Use:   "download-ecls",
	Short: "Download early childhood metrics from NCES ECLS",
	RunE: func(cmd *cobra.Command, args []string) error {
		start, end, _ := parseYears(years)
		return runDownloadECLS(start, end, dryRun)
	},
}

func runDownloadECLS(startYear, endYear int, dryRun bool) error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()
	
	downloader := downloaders.NewECLSDownloader(db)
	return downloader.Download(startYear, endYear, dryRun)
}

var generateAssetsCmd = &cobra.Command{
	Use:   "generate-assets",
	Short: "Generate Hugo JSON assets from database",
	RunE:  runGenerateAssets,
}

func runGenerateAssets(cmd *cobra.Command, args []string) error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()
	
	generator := generators.NewHugoGenerator(db)
	return generator.GenerateAll()
}
