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
	fmt.Println("    ⚠ Note: ECLS data is primarily available through reports and restricted-use files")
	fmt.Println("    ℹ URL: https://nces.ed.gov/ecls/")
	fmt.Println("    ℹ Available cohorts:")
	fmt.Println("      - ECLS-K (1998-99 kindergarten cohort)")
	fmt.Println("      - ECLS-K:2011 (2010-11 kindergarten cohort)")
	fmt.Println("      - ECLS-B (2001 birth cohort)")
	fmt.Println("    ℹ Data requires application for restricted-use license")
	
	yearsRange := fmt.Sprintf("%d-%d", startYear, endYear)
	database.UpdateSourceMetadata(e.db, sourceName, yearsRange, 0, "partial", 
		"ECLS data requires restricted-use license application")
	
	fmt.Printf("  ℹ ECLS download: 0 rows (restricted data - manual access required)\n")
	return nil
}
