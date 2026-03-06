package downloaders

import (
	"database/sql"
	"fmt"

	"github.com/aallbrig/proficiency-comparison/internal/database"
)

type NCESDownloader struct {
	db *sql.DB
}

func NewNCESDownloader(db *sql.DB) *NCESDownloader {
	return &NCESDownloader{db: db}
}

// historicalGraduation contains US public high school graduation rates (%).
//
// 1960–2010: AFGR (Averaged Freshman Graduation Rate) from NCES Digest of
// Education Statistics Table 219.10.
// 2011–2020: ACGR (4-year Adjusted Cohort Graduation Rate), NCES Table 219.46.
// Note: ACGR replaced AFGR in 2010–11; ACGR values are slightly higher due to
// methodological differences.
//
// Sources:
//   - https://nces.ed.gov/programs/digest/d23/tables/dt23_219.10.asp
//   - https://nces.ed.gov/programs/digest/d23/tables/dt23_219.46.asp
var historicalGraduation = []struct {
	year int
	rate float64
}{
	// AFGR series
	{1960, 69.5}, {1965, 76.5}, {1970, 76.9}, {1975, 73.9},
	{1980, 71.4}, {1985, 72.1}, {1990, 74.4}, {1995, 74.5},
	{2000, 72.6}, {2001, 73.0}, {2002, 73.9}, {2003, 74.4},
	{2004, 74.9}, {2005, 74.7}, {2006, 73.4}, {2007, 73.9},
	{2008, 74.9}, {2009, 75.5}, {2010, 78.2},
	// ACGR series (2011+)
	{2011, 79.0}, {2012, 80.0}, {2013, 81.4}, {2014, 82.3},
	{2015, 83.2}, {2016, 84.1}, {2017, 84.6}, {2018, 85.3},
	{2019, 86.8}, {2020, 86.5},
}

// historicalEnrollment contains US school enrollment rates (% of school-age
// population enrolled in any school — public, private, or home school).
//
// Values for 1950–2000 are from NCES Digest Table 103.20 (5–17 year-olds).
// Values for 2010–2020 are from Census ACS (ages 5–19 including college
// entrants, producing ~86% vs ~93% for 5–17 only).
// We use the 5–17 series for consistency.
//
// Sources:
//   - https://nces.ed.gov/programs/digest/d23/tables/dt23_103.20.asp
//   - US Census Bureau, School Enrollment Historical Tables
var historicalEnrollment = []struct {
	year int
	rate float64
}{
	{1950, 83.7}, {1955, 86.8}, {1960, 87.2}, {1965, 88.5},
	{1970, 90.2}, {1975, 91.0}, {1980, 91.3}, {1985, 91.5},
	{1990, 92.1}, {1995, 92.6}, {2000, 93.6}, {2001, 93.5},
	{2002, 93.3}, {2003, 93.1}, {2004, 93.2}, {2005, 94.1},
	{2006, 94.0}, {2007, 93.8}, {2008, 93.9}, {2009, 94.0},
	{2010, 93.7}, {2011, 93.5}, {2012, 93.5}, {2013, 93.7},
	{2014, 93.6}, {2015, 93.6}, {2016, 93.5}, {2017, 93.5},
	{2018, 93.6}, {2019, 93.7}, {2020, 92.8},
}

func (n *NCESDownloader) Download(startYear, endYear int, dryRun bool) error {
	sourceName := "nces_digest"

	if dryRun {
		fmt.Printf("  [DRY RUN] Would seed NCES graduation/enrollment data for %d-%d\n", startYear, endYear)
		return nil
	}

	fmt.Println("  Seeding NCES graduation and enrollment data...")
	fmt.Println("    ℹ AFGR series (1960–2010) + ACGR series (2011–2020)")

	// Clear existing data to avoid duplicates on re-run.
	if _, err := n.db.Exec(`DELETE FROM graduation_rates WHERE source = ?`, sourceName); err != nil {
		return fmt.Errorf("failed to clear existing graduation data: %w", err)
	}
	if _, err := n.db.Exec(`DELETE FROM enrollment_rates WHERE source = ?`, sourceName); err != nil {
		return fmt.Errorf("failed to clear existing enrollment data: %w", err)
	}

	gradRows := 0
	for _, row := range historicalGraduation {
		if row.year < startYear || row.year > endYear {
			continue
		}
		_, err := n.db.Exec(`
			INSERT INTO graduation_rates (year, rate, source)
			VALUES (?, ?, ?)
		`, row.year, row.rate, sourceName)
		if err != nil {
			fmt.Printf("    Warning: failed to insert graduation year %d: %v\n", row.year, err)
			continue
		}
		gradRows++
	}
	fmt.Printf("    ✓ Inserted %d graduation rate rows (1960–2020)\n", gradRows)

	enrollRows := 0
	for _, row := range historicalEnrollment {
		if row.year < startYear || row.year > endYear {
			continue
		}
		_, err := n.db.Exec(`
			INSERT INTO enrollment_rates (year, age_group, enrollment_rate, source)
			VALUES (?, ?, ?, ?)
		`, row.year, "5_to_17", row.rate, sourceName)
		if err != nil {
			fmt.Printf("    Warning: failed to insert enrollment year %d: %v\n", row.year, err)
			continue
		}
		enrollRows++
	}
	fmt.Printf("    ✓ Inserted %d enrollment rate rows (1950–2020)\n", enrollRows)

	totalRows := gradRows + enrollRows
	yearsRange := fmt.Sprintf("%d-%d", startYear, endYear)
	status := "success"
	if totalRows == 0 {
		status = "partial"
	}
	database.UpdateSourceMetadata(n.db, sourceName, yearsRange, totalRows, status,
		fmt.Sprintf("NCES historical series: %d graduation + %d enrollment rows", gradRows, enrollRows))

	fmt.Printf("  ✓ NCES data seeded: %d rows\n", totalRows)
	return nil
}
