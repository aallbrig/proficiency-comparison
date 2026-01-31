package downloaders

import (
	"database/sql"
	"fmt"

	"github.com/aallbrig/proficiency-comparison/internal/database"
)

type ECLSDownloader struct {
	db *sql.DB
}

func NewECLSDownloader(db *sql.DB) *ECLSDownloader {
	return &ECLSDownloader{db: db}
}

func (e *ECLSDownloader) Download(startYear, endYear int, dryRun bool) error {
	sourceName := "ecls_early_childhood"
	
	if dryRun {
		fmt.Printf("  [DRY RUN] Would download ECLS early childhood metrics for %d-%d\n", startYear, endYear)
		return nil
	}

	fmt.Println("  Downloading ECLS early childhood metrics...")
	fmt.Println("    ℹ Note: ECLS data is primarily available through reports and restricted-use files")
	fmt.Println("    ℹ URL: https://nces.ed.gov/ecls/")
	fmt.Println("    ℹ Public summaries in Excel/PDF format")
	
	database.UpdateSourceMetadata(e.db, sourceName, "", 0, "partial", 
		"Requires report parsing or restricted data access")
	
	fmt.Println("  ⚠ ECLS download incomplete: requires manual data extraction")
	return nil
}
