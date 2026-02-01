package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/aallbrig/proficiency-comparison/internal/database"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database with schema.sql",
	Long: `Initialize or sync the database schema from schema.sql.
	
This command will:
  - Create the database file if it doesn't exist
  - Apply all tables, indexes, and constraints from schema.sql
  - Safe to run multiple times (uses CREATE TABLE IF NOT EXISTS)
  - Reports which tables were created or already existed`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("Initializing database from schema.sql...")
	fmt.Printf("Database location: %s\n", database.GetDatabasePath())
	fmt.Println()

	db, err := database.Open()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Apply schema
	if err := database.ApplySchema(db); err != nil {
		return fmt.Errorf("failed to apply schema: %w", err)
	}

	// Verify schema was applied
	fmt.Println()
	fmt.Println("Verifying database structure...")
	
	dbInfo, err := database.GetDatabaseInfo(db)
	if err != nil {
		return fmt.Errorf("failed to verify database: %w", err)
	}

	fmt.Printf("✓ Database file: %s\n", database.GetDatabasePath())
	fmt.Printf("✓ Total tables: %d\n", dbInfo.TableCount)
	fmt.Printf("✓ Schema status: %s\n", dbInfo.SchemaStatus)
	
	// List all tables
	fmt.Println()
	fmt.Println("Tables created:")
	tables := []string{
		"pipeline_metadata",
		"source_metadata",
		"raw_files",
		"literacy_rates",
		"educational_attainment",
		"graduation_rates",
		"enrollment_rates",
		"test_proficiency",
		"early_childhood",
	}
	
	for _, table := range tables {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='%s'", table)).Scan(&count)
		if err == nil && count > 0 {
			var rowCount int
			db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&rowCount)
			fmt.Printf("  ✓ %-30s (%d rows)\n", table, rowCount)
		}
	}
	
	fmt.Println()
	fmt.Println("✅ Database initialization complete!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Check status: edu-stats status")
	fmt.Println("  2. Download data: edu-stats all --years=1970-2025")
	fmt.Println()
	fmt.Println("Note: Database location can be changed by setting EDU_STATS_DATA_DIR environment variable")
	
	return nil
}
