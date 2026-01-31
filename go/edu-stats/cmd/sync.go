package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/aallbrig/proficiency-comparison/internal/database"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync database schema from schema.sql",
	Long: `Apply schema changes from schema.sql to the database.
	
This command will:
  - Read schema.sql from the project root
  - Apply all CREATE TABLE IF NOT EXISTS statements
  - Add any new tables, indexes, or columns
  - Safe to run multiple times (idempotent)
  - Use this after updating schema.sql to migrate the database`,
	RunE: runSync,
}

func runSync(cmd *cobra.Command, args []string) error {
	fmt.Println("Syncing database schema from schema.sql...")
	fmt.Println()

	db, err := database.Open()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Apply schema
	if err := database.ApplySchema(db); err != nil {
		return fmt.Errorf("failed to sync schema: %w", err)
	}

	// Verify
	fmt.Println()
	fmt.Println("Verifying schema...")
	
	dbInfo, err := database.GetDatabaseInfo(db)
	if err != nil {
		return fmt.Errorf("failed to verify database: %w", err)
	}

	fmt.Printf("✓ Total tables: %d\n", dbInfo.TableCount)
	fmt.Printf("✓ Schema status: %s\n", dbInfo.SchemaStatus)
	
	fmt.Println()
	fmt.Println("✅ Schema sync complete!")
	
	return nil
}
