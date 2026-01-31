package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/aallbrig/proficiency-comparison/internal/database"
	"github.com/aallbrig/proficiency-comparison/internal/utils"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show database and data source status",
	Long:  `Display database status, last download times, row counts, and data source connectivity.`,
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	fmt.Println("Educational Stats CLI - Status Report")
	fmt.Println("=====================================")
	fmt.Println()

	db, err := database.Open()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Check database and schema
	fmt.Println("📊 Database Status:")
	dbInfo, err := database.GetDatabaseInfo(db)
	if err != nil {
		fmt.Printf("  ❌ Error reading database: %v\n", err)
	} else {
		fmt.Printf("  ✓ Database: edu_stats.db (%.2f MB)\n", float64(dbInfo.SizeBytes)/1024/1024)
		fmt.Printf("  ✓ Schema: %s\n", dbInfo.SchemaStatus)
		fmt.Printf("  ✓ Total tables: %d\n", dbInfo.TableCount)
	}
	fmt.Println()

	// Last data retrieval
	fmt.Println("🕒 Last Data Retrieval:")
	sources, err := database.GetSourceMetadata(db)
	if err != nil {
		fmt.Printf("  ❌ Error reading source metadata: %v\n", err)
	} else if len(sources) == 0 {
		fmt.Println("  ⚠ No data downloaded yet. Run 'edu-stats-cli all' to download data.")
	} else {
		for _, source := range sources {
			status := "✓"
			if source.Status == "failed" {
				status = "❌"
			} else if source.Status == "partial" {
				status = "⚠"
			}
			
			lastDownload := "never"
			if source.LastDownload != nil {
				lastDownload = source.LastDownload.Format("2006-01-02 15:04:05")
				elapsed := time.Since(*source.LastDownload)
				if elapsed > 30*24*time.Hour {
					lastDownload += " (stale, >30 days old)"
				}
			}
			
			fmt.Printf("  %s %s: %s (%d rows, years: %s)\n", 
				status, source.Name, lastDownload, source.RowCount, source.YearsAvailable)
		}
	}
	fmt.Println()

	// Row counts per table
	fmt.Println("📈 Data Summary:")
	tableCounts, err := database.GetTableRowCounts(db)
	if err != nil {
		fmt.Printf("  ❌ Error reading row counts: %v\n", err)
	} else {
		totalRows := 0
		for table, count := range tableCounts {
			fmt.Printf("  • %-25s %d rows\n", table+":", count)
			totalRows += count
		}
		fmt.Printf("  • %-25s %d rows\n", "TOTAL:", totalRows)
	}
	fmt.Println()

	// Data source connectivity
	fmt.Println("🌐 Data Source Connectivity:")
	sources_to_check := []struct {
		name string
		url  string
	}{
		{"World Bank API", "https://api.worldbank.org/v2/country/USA"},
		{"Census Bureau API", "https://api.census.gov/data.json"},
		{"NCES Website", "https://nces.ed.gov/programs/digest/"},
		{"NAEP API", "https://www.nationsreportcard.gov/"},
		{"NCES ECLS", "https://nces.ed.gov/ecls/"},
	}

	for _, source := range sources_to_check {
		if utils.CheckConnectivity(source.url) {
			fmt.Printf("  ✓ %s: accessible\n", source.name)
		} else {
			fmt.Printf("  ❌ %s: unreachable\n", source.name)
		}
	}

	return nil
}
