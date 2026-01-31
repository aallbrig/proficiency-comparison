package downloaders

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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
	
	totalRows := 0
	
	// Try ACS 1-year estimates for recent years (2010-present)
	for year := max(2010, startYear); year <= min(endYear, 2023); year++ {
		// ACS API format: /data/{year}/acs/acs1
		url := fmt.Sprintf("https://api.census.gov/data/%d/acs/acs1?get=NAME,B15003_022E,B15003_001E&for=us:*", year)
		
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("    ⚠ Failed to fetch year %d: %v\n", year, err)
			continue
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			fmt.Printf("    ⚠ Year %d unavailable (HTTP %d)\n", year, resp.StatusCode)
			continue
		}

		var data CensusResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			resp.Body.Close()
			fmt.Printf("    ⚠ Failed to parse year %d: %v\n", year, err)
			continue
		}
		resp.Body.Close()

		// Process data (skip header row)
		if len(data) < 2 {
			fmt.Printf("    ⚠ No data for year %d\n", year)
			continue
		}

		for i := 1; i < len(data); i++ {
			row := data[i]
			if len(row) < 3 {
				continue
			}

			// Parse bachelor's degree count and total
			var bachelors, total float64
			
			switch v := row[1].(type) {
			case float64:
				bachelors = v
			case string:
				bachelors, _ = strconv.ParseFloat(v, 64)
			}
			
			switch v := row[2].(type) {
			case float64:
				total = v
			case string:
				total, _ = strconv.ParseFloat(v, 64)
			}
			
			if total == 0 {
				continue
			}

			percentage := (bachelors / total) * 100

			// Insert into database
			_, err = c.db.Exec(`
				INSERT INTO educational_attainment (year, age_group, education_level, percentage, source)
				VALUES (?, ?, ?, ?, ?)
				ON CONFLICT(year, age_group, education_level, gender, race, source) DO UPDATE SET
					percentage = excluded.percentage
			`, year, "25plus", "bachelors_plus", percentage, sourceName)

			if err != nil {
				fmt.Printf("    Warning: failed to insert year %d: %v\n", year, err)
				continue
			}

			totalRows++
		}

		fmt.Printf("    ✓ Imported year %d\n", year)
	}

	yearsRange := fmt.Sprintf("%d-%d", startYear, endYear)
	if totalRows > 0 {
		database.UpdateSourceMetadata(c.db, sourceName, yearsRange, totalRows, "success", "")
	} else {
		database.UpdateSourceMetadata(c.db, sourceName, yearsRange, 0, "partial", 
			"No recent data available from ACS API")
	}
	
	fmt.Printf("  ✓ Census download complete: %d rows\n", totalRows)
	fmt.Println("    ℹ Note: Historical data (pre-2010) requires Excel file downloads")
	fmt.Println("    ℹ Visit: https://www.census.gov/topics/education/educational-attainment/data/tables.html")
	
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
