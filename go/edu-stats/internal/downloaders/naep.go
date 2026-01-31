package downloaders

import (
	"database/sql"
	"fmt"
	"net/http"

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
	
	// NAEP data is available via their data explorer
	// The API structure: fetch average scores for reading/math at grades 4, 8, 12
	subjects := []string{"reading", "mathematics"}
	grades := []int{4, 8}
	
	totalRows := 0
	
	for _, subject := range subjects {
		for _, grade := range grades {
			// Build years string (NAEP specific years only - typically every 2-4 years)
			// We'll try to fetch and handle 404s gracefully
			yearsToTry := []int{1990, 1992, 1994, 1996, 1998, 2000, 2002, 2003, 2005, 2007, 2009, 2011, 2013, 2015, 2017, 2019, 2022}
			
			for _, year := range yearsToTry {
				if year < startYear || year > endYear {
					continue
				}
				
				// NAEP Data Explorer API (simpler endpoint)
				// Note: The actual NAEP API requires more complex queries
				// This is a placeholder - real implementation would use NAEP's data export
				url := fmt.Sprintf(
					"https://nces.ed.gov/nationsreportcard/api/indicator/%s/grade%d/year/%d",
					subject, grade, year,
				)
				
				resp, err := http.Get(url)
				if err != nil {
					continue
				}
				
				if resp.StatusCode == 404 {
					resp.Body.Close()
					continue // Year not available
				}
				
				if resp.StatusCode != 200 {
					resp.Body.Close()
					continue
				}
				
				// For now, we'll mark this as a placeholder
				// Real implementation needs actual NAEP data format
				resp.Body.Close()
			}
		}
	}

	// Alternative: Use sample/historical data points
	// Insert known NAEP average scores from published reports
	knownData := []struct {
		year    int
		subject string
		grade   int
		score   float64
	}{
		{2019, "reading", 4, 220},
		{2019, "reading", 8, 263},
		{2019, "mathematics", 4, 241},
		{2019, "mathematics", 8, 282},
		{2022, "reading", 4, 217},
		{2022, "reading", 8, 260},
		{2022, "mathematics", 4, 236},
		{2022, "mathematics", 8, 274},
	}
	
	for _, data := range knownData {
		if data.year < startYear || data.year > endYear {
			continue
		}
		
		_, err := n.db.Exec(`
			INSERT INTO test_proficiency (year, subject, grade, avg_score, source)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(year, subject, grade, proficiency_level, state, demographics, source) DO UPDATE SET
				avg_score = excluded.avg_score
		`, data.year, data.subject, data.grade, data.score, sourceName)

		if err != nil {
			fmt.Printf("    Warning: failed to insert %s grade %d year %d: %v\n", 
				data.subject, data.grade, data.year, err)
			continue
		}

		totalRows++
	}

	fmt.Printf("    ✓ Imported %d sample NAEP data points\n", totalRows)
	fmt.Println("    ℹ Note: Full NAEP data requires data export from NAEP Data Explorer")
	fmt.Println("    ℹ Visit: https://nces.ed.gov/nationsreportcard/data/")

	yearsRange := fmt.Sprintf("%d-%d", startYear, endYear)
	if totalRows > 0 {
		database.UpdateSourceMetadata(n.db, sourceName, yearsRange, totalRows, "partial", 
			"Sample data only - full export needed")
	} else {
		database.UpdateSourceMetadata(n.db, sourceName, yearsRange, 0, "partial", 
			"No data in requested year range")
	}
	
	fmt.Printf("  ✓ NAEP download complete: %d rows\n", totalRows)
	return nil
}
