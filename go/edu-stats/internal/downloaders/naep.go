package downloaders

import (
	"database/sql"
	"fmt"

	"github.com/aallbrig/proficiency-comparison/internal/database"
)

type NAEPDownloader struct {
	db *sql.DB
}

func NewNAEPDownloader(db *sql.DB) *NAEPDownloader {
	return &NAEPDownloader{db: db}
}

type NAEPResponse struct {
	Result []struct {
		Value string `json:"value"`
		Year  string `json:"year"`
	} `json:"result"`
}

func (n *NAEPDownloader) Download(startYear, endYear int, dryRun bool) error {
	sourceName := "naep_proficiency"
	
	if dryRun {
		fmt.Printf("  [DRY RUN] Would download NAEP test proficiency for %d-%d\n", startYear, endYear)
		return nil
	}

	fmt.Println("  Downloading NAEP test proficiency data...")
	fmt.Println("    ⚠ Note: NAEP API requires data export from NAEP Data Explorer")
	fmt.Println("    ℹ Visit: https://nces.ed.gov/nationsreportcard/data/")
	fmt.Println("    ℹ Use the Data Explorer to export CSV data for reading/math")
	fmt.Println("    ℹ Grades: 4, 8, 12 | Subjects: reading, mathematics")
	fmt.Println("    ℹ Years available: 1990-present (varies by subject/grade)")
	
	yearsRange := fmt.Sprintf("%d-%d", startYear, endYear)
	database.UpdateSourceMetadata(n.db, sourceName, yearsRange, 0, "partial", 
		"NAEP data requires manual export from Data Explorer")
	
	fmt.Printf("  ℹ NAEP download: 0 rows (manual export required)\n")
	return nil
}
