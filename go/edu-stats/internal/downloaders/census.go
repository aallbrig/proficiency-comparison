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
	
	// Add historical Census data from published tables (1940-2009)
	// Source: Census Historical Tables on Educational Attainment
	fmt.Println("    Adding historical educational attainment data...")
	
	historicalData := map[int]float64{
		1940: 4.6, 1950: 6.2, 1960: 7.7, 1970: 10.7, 1975: 13.9,
		1980: 16.2, 1985: 19.4, 1990: 21.3, 1995: 23.0, 2000: 25.6,
		2005: 27.7, 2006: 28.0, 2007: 28.7, 2008: 29.4, 2009: 29.5,
	}
	
	for year := startYear; year < 2010; year++ {
		if percent, ok := historicalData[year]; ok {
			_, err := c.db.Exec(`
				INSERT INTO educational_attainment (year, age_group, education_level, percentage, source)
				VALUES (?, ?, ?, ?, ?)
				ON CONFLICT(year, age_group, education_level, state, demographics, source) DO UPDATE SET
					percentage = excluded.percentage
			`, year, "25plus", "bachelors_plus", percent, sourceName+"_historical")
			
			if err == nil {
				totalRows++
				fmt.Printf("    ✓ Added historical year %d\n", year)
			}
		}
	}
	
	if totalRows > 0 {
		database.UpdateSourceMetadata(c.db, sourceName, yearsRange, totalRows, "success", 
			"Includes historical data from Census tables")
	} else {
		database.UpdateSourceMetadata(c.db, sourceName, yearsRange, 0, "partial", 
			"No data available for requested range")
	}
	
	fmt.Printf("  ✓ Census download complete: %d rows (1940-present)\n", totalRows)
	fmt.Println("    ℹ Historical data sourced from Census Bureau published tables")
	
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
