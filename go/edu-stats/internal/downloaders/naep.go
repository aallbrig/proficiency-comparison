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

// naepReadingGrade8 contains NAEP reading scale scores for Grade 8 / Age 13.
//
// Years 1971–1999 are from the NAEP Long-Term Trend (LTT) assessment at Age 13
// (scale score 0–500). Years 2002–2022 are from the Main NAEP Grade 8 reading
// assessment. Both use comparable scales and together form the best available
// continuous US reading proficiency series.
//
// Sources:
//   - NAEP LTT: https://nces.ed.gov/nationsreportcard/ltt/
//   - Main NAEP: https://nces.ed.gov/nationsreportcard/reading/
var naepReadingGrade8 = []struct {
	year  int
	score float64
}{
	// NAEP Long-Term Trend, Age 13 reading
	{1971, 255}, {1975, 256}, {1980, 259}, {1984, 257},
	{1988, 258}, {1990, 257}, {1992, 260}, {1994, 260},
	{1996, 259}, {1999, 259},
	// Main NAEP Grade 8 reading (scale score)
	{2002, 264}, {2003, 263}, {2005, 262}, {2007, 263},
	{2009, 264}, {2011, 265}, {2013, 266}, {2015, 265},
	{2017, 267}, {2019, 263}, {2022, 260},
}

func (n *NAEPDownloader) Download(startYear, endYear int, dryRun bool) error {
	sourceName := "naep_proficiency"

	if dryRun {
		fmt.Printf("  [DRY RUN] Would seed NAEP test proficiency for %d-%d\n", startYear, endYear)
		return nil
	}

	fmt.Println("  Seeding NAEP reading proficiency data...")
	fmt.Println("    ℹ NAEP LTT Age 13 (1971–1999) + Main NAEP Grade 8 (2002–2022)")

	// Clear existing data for this source to avoid duplicates on re-run.
	if _, err := n.db.Exec(`DELETE FROM test_proficiency WHERE source = ?`, sourceName); err != nil {
		return fmt.Errorf("failed to clear existing proficiency data: %w", err)
	}

	totalRows := 0
	for _, row := range naepReadingGrade8 {
		if row.year < startYear || row.year > endYear {
			continue
		}
		_, err := n.db.Exec(`
			INSERT INTO test_proficiency (year, subject, grade, avg_score, source)
			VALUES (?, ?, ?, ?, ?)
		`, row.year, "reading", 8, row.score, sourceName)
		if err != nil {
			fmt.Printf("    Warning: failed to insert NAEP year %d: %v\n", row.year, err)
			continue
		}
		totalRows++
	}

	yearsRange := fmt.Sprintf("%d-%d", startYear, endYear)
	database.UpdateSourceMetadata(n.db, sourceName, yearsRange, totalRows, "success",
		fmt.Sprintf("NAEP LTT Age 13 (1971–1999) + Main Grade 8 (2002–2022): %d rows", totalRows))

	fmt.Printf("  ✓ NAEP data seeded: %d rows\n", totalRows)
	return nil
}
