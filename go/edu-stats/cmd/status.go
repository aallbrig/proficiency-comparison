package cmd

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/aallbrig/proficiency-comparison/internal/database"
	"github.com/aallbrig/proficiency-comparison/internal/utils"
)

var (
	verbose bool
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show database and data source status",
	Long:  `Display database status, last download times, row counts, and data source connectivity.`,
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output including reset history")
}

func runStatus(cmd *cobra.Command, args []string) error {
	fmt.Println("Educational Stats CLI - Status Report")
	fmt.Println("=====================================")
	fmt.Println()
	fmt.Printf("📍 Database location: %s\n", database.GetDatabasePath())
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
		fmt.Println("  ⚠ No data downloaded yet. Run 'edu-stats all' to download data.")
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
	fmt.Println()

	// Show reset history (if verbose flag is set)
	if verbose {
		resetHistory, err := getResetHistory(db)
		if err == nil && len(resetHistory) > 0 {
			fmt.Println("🔄 Recent Reset Operations:")
			for _, reset := range resetHistory {
				fmt.Printf("  • %s: Years %d-%d (%d rows deleted, %.2fs)\n",
					reset.Timestamp.Format("2006-01-02 15:04:05"),
					reset.StartYear, reset.EndYear, reset.RowsDeleted, reset.ExecutionTime)
				if reset.DeletionSummary != "" {
					fmt.Printf("    Summary: %s\n", reset.DeletionSummary)
				}
			}
			fmt.Println()
		}
	}

	return nil
}

type ResetRecord struct {
	Timestamp       time.Time
	StartYear       int
	EndYear         int
	RowsDeleted     int
	ExecutionTime   float64
	DeletionSummary string
}

func getResetHistory(db *sql.DB) ([]ResetRecord, error) {
	// Check if reset_audit table exists
	var tableExists int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='reset_audit'").Scan(&tableExists)
	if err != nil || tableExists == 0 {
		return nil, nil
	}

	rows, err := db.Query(`
		SELECT reset_timestamp, start_year, end_year, rows_deleted, execution_time_seconds, deletion_summary
		FROM reset_audit
		ORDER BY reset_timestamp DESC
		LIMIT 5
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ResetRecord
	for rows.Next() {
		var r ResetRecord
		err := rows.Scan(&r.Timestamp, &r.StartYear, &r.EndYear, &r.RowsDeleted, &r.ExecutionTime, &r.DeletionSummary)
		if err != nil {
			continue
		}
		records = append(records, r)
	}

	return records, nil
}
