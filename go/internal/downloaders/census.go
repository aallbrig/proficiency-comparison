package downloaders

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aallbrig/proficiency-comparison/internal/database"
)

type CensusDownloader struct {
	db *sql.DB
}

func NewCensusDownloader(db *sql.DB) *CensusDownloader {
	return &CensusDownloader{db: db}
}

type CensusResponse [][]interface{}

func (c *CensusDownloader) Download(startYear, endYear int, dryRun bool) error {
	sourceName := "census_attainment"
	
	if dryRun {
		fmt.Printf("  [DRY RUN] Would download Census educational attainment for %d-%d\n", startYear, endYear)
		return nil
	}

	fmt.Println("  Downloading Census educational attainment data...")
	
	// Note: Census API only has recent years. For historical data, would need to download Excel files.
	// This is a simplified version that demonstrates the API approach for recent data.
	
	totalRows := 0
	
	// Try to get recent year data from ACS
	if endYear >= 2010 {
		url := "https://api.census.gov/data/2022/acs/acs1?get=NAME,B15003_022E,B15003_001E&for=us:*"
		
		resp, err := http.Get(url)
		if err != nil {
			database.UpdateSourceMetadata(c.db, sourceName, "", 0, "failed", err.Error())
			return fmt.Errorf("failed to download Census data: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			// Census API might be unavailable, mark as partial success
			fmt.Printf("    ⚠ Census API returned status %d, marking as partial\n", resp.StatusCode)
			database.UpdateSourceMetadata(c.db, sourceName, "", 0, "partial", 
				fmt.Sprintf("API unavailable: HTTP %d", resp.StatusCode))
			return nil
		}

		var data CensusResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			database.UpdateSourceMetadata(c.db, sourceName, "", 0, "failed", err.Error())
			return fmt.Errorf("failed to parse Census data: %w", err)
		}

		// Process data (skip header row)
		for i := 1; i < len(data); i++ {
			row := data[i]
			if len(row) < 3 {
				continue
			}

			// Calculate percentage with bachelor's degree or higher
			bachelorsFloat, ok1 := row[1].(float64)
			totalFloat, ok2 := row[2].(float64)
			
			if !ok1 || !ok2 || totalFloat == 0 {
				continue
			}

			percentage := (bachelorsFloat / totalFloat) * 100

			// Insert into database (using 2022 as year)
			_, err = c.db.Exec(`
				INSERT INTO educational_attainment (year, age_group, education_level, percentage, source)
				VALUES (?, ?, ?, ?, ?)
				ON CONFLICT(year, age_group, education_level, gender, race, source) DO UPDATE SET
					percentage = excluded.percentage
			`, 2022, "25plus", "bachelors_plus", percentage, sourceName)

			if err != nil {
				fmt.Printf("    Warning: failed to insert row: %v\n", err)
				continue
			}

			totalRows++
		}

		fmt.Printf("    ✓ Imported %d rows from Census API\n", totalRows)
	}

	// For full historical data, note that Excel/CSV downloads would be needed
	if totalRows == 0 {
		fmt.Println("    ℹ Note: Full historical data requires Excel file downloads")
		fmt.Println("    ℹ Visit: https://www.census.gov/topics/education/educational-attainment/data/tables.html")
		database.UpdateSourceMetadata(c.db, sourceName, "", 0, "partial", 
			"API data only, historical requires manual download")
	} else {
		yearsRange := fmt.Sprintf("%d-%d", startYear, endYear)
		database.UpdateSourceMetadata(c.db, sourceName, yearsRange, totalRows, "success", "")
	}
	
	fmt.Printf("  ✓ Census download complete: %d rows\n", totalRows)
	return nil
}
