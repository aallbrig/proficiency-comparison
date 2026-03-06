package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/aallbrig/proficiency-comparison/internal/database"
)

var resetCmd = &cobra.Command{
	Use:   "reset [start-year] [end-year]",
	Short: "Reset (delete) data for a specific year range",
	Long: `Delete downloaded data for a specific year range from the database.
	
This command removes data from all tables for the specified year range and logs
the operation for auditing purposes. The status command will show reset history.

Examples:
  edu-stats reset 1870 2025    # Reset all data from 1870-2025
  edu-stats reset 2020 2023    # Reset only 2020-2023 data`,
	Args: cobra.ExactArgs(2),
	RunE: runReset,
}

func runReset(cmd *cobra.Command, args []string) error {
	var startYear, endYear int
	
	if _, err := fmt.Sscanf(args[0], "%d", &startYear); err != nil {
		return fmt.Errorf("invalid start year: %s", args[0])
	}
	
	if _, err := fmt.Sscanf(args[1], "%d", &endYear); err != nil {
		return fmt.Errorf("invalid end year: %s", args[1])
	}
	
	if startYear > endYear {
		return fmt.Errorf("start year (%d) cannot be greater than end year (%d)", startYear, endYear)
	}
	
	fmt.Println("Educational Stats CLI - Data Reset")
	fmt.Println("===================================")
	fmt.Printf("📅 Year range: %d - %d\n", startYear, endYear)
	fmt.Println()
	
	db, err := database.Open()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()
	
	// Confirm with user
	fmt.Printf("⚠️  WARNING: This will delete all data between %d and %d.\n", startYear, endYear)
	fmt.Print("Type 'yes' to confirm: ")
	
	var confirmation string
	fmt.Scanln(&confirmation)
	
	if confirmation != "yes" {
		fmt.Println("❌ Reset cancelled")
		return nil
	}
	
	fmt.Println()
	fmt.Println("🔄 Starting data reset...")
	
	startTime := time.Now()
	
	// Track rows deleted per table
	tablesWithYears := []string{
		"literacy_rates",
		"educational_attainment",
		"graduation_rates",
		"enrollment_rates",
		"test_proficiency",
		"early_childhood",
	}
	
	totalRowsDeleted := 0
	deletionSummary := make(map[string]int)
	
	for _, table := range tablesWithYears {
		fmt.Printf("  Resetting %s...", table)
		
		// Count rows before deletion
		var countBefore int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE year >= ? AND year <= ?", table), 
			startYear, endYear).Scan(&countBefore)
		if err != nil {
			fmt.Printf(" ⚠️  Error counting: %v\n", err)
			continue
		}
		
		// Delete rows
		result, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE year >= ? AND year <= ?", table), 
			startYear, endYear)
		if err != nil {
			fmt.Printf(" ❌ Error: %v\n", err)
			continue
		}
		
		rowsAffected, _ := result.RowsAffected()
		deletionSummary[table] = int(rowsAffected)
		totalRowsDeleted += int(rowsAffected)
		
		fmt.Printf(" ✓ Deleted %d rows\n", rowsAffected)
	}
	
	executionTime := time.Since(startTime)
	
	// Update source metadata to reflect reset
	_, err = db.Exec(`
		UPDATE source_metadata 
		SET row_count = row_count - ?,
		    last_download = NULL
		WHERE row_count > 0
	`, totalRowsDeleted)
	
	// Log the reset operation in pipeline_metadata
	err = database.RecordPipelineStep(db, "reset", "completed", 
		fmt.Sprintf("%d-%d", startYear, endYear), nil)
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to log reset operation: %v\n", err)
	}
	
	// Store detailed reset audit in a new table (create if not exists)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS reset_audit (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			reset_timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			start_year INTEGER NOT NULL,
			end_year INTEGER NOT NULL,
			rows_deleted INTEGER NOT NULL,
			execution_time_seconds REAL,
			deletion_summary TEXT
		)
	`)
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to create reset_audit table: %v\n", err)
	} else {
		// Format deletion summary as JSON-like string
		summaryStr := "{"
		first := true
		for table, count := range deletionSummary {
			if !first {
				summaryStr += ", "
			}
			summaryStr += fmt.Sprintf("\"%s\": %d", table, count)
			first = false
		}
		summaryStr += "}"
		
		_, err = db.Exec(`
			INSERT INTO reset_audit (start_year, end_year, rows_deleted, execution_time_seconds, deletion_summary)
			VALUES (?, ?, ?, ?, ?)
		`, startYear, endYear, totalRowsDeleted, executionTime.Seconds(), summaryStr)
		
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to log detailed audit: %v\n", err)
		}
	}
	
	fmt.Println()
	fmt.Println("✅ Reset complete!")
	fmt.Printf("   Total rows deleted: %d\n", totalRowsDeleted)
	fmt.Printf("   Execution time: %.2f seconds\n", executionTime.Seconds())
	fmt.Println()
	fmt.Println("📊 Deletion breakdown:")
	for table, count := range deletionSummary {
		if count > 0 {
			fmt.Printf("   • %-30s %d rows\n", table+":", count)
		}
	}
	fmt.Println()
	fmt.Println("ℹ️  Run 'edu-stats status' to see reset history")
	
	return nil
}
