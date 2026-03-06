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

// historicalAttainment contains Census Bureau CPS historical bachelor's degree
// attainment (% of population 25+) from 1940–2009. Source: Census CPS
// Historical Time Series Table A-2.
// https://www.census.gov/data/tables/time-series/demo/educational-attainment/cps-historical-time-series.html
var historicalAttainment = []struct {
	year int
	pct  float64
}{
	{1940, 4.6}, {1950, 6.2}, {1960, 7.7}, {1965, 9.4},
	{1970, 11.0}, {1975, 13.9}, {1980, 17.0}, {1985, 19.4},
	{1990, 21.3}, {1995, 23.0}, {2000, 25.6}, {2001, 25.9},
	{2002, 26.7}, {2003, 27.2}, {2004, 27.7}, {2005, 27.7},
	{2006, 28.0}, {2007, 29.4}, {2008, 29.4}, {2009, 29.9},
}

func (c *CensusDownloader) Download(startYear, endYear int, dryRun bool) error {
	sourceName := "census_attainment"

	if dryRun {
		fmt.Printf("  [DRY RUN] Would download Census educational attainment for %d-%d\n", startYear, endYear)
		return nil
	}

	fmt.Println("  Downloading Census educational attainment data...")

	// Clear existing data for this source to avoid duplicates on re-run.
	if _, err := c.db.Exec(`DELETE FROM educational_attainment WHERE source = ?`, sourceName); err != nil {
		return fmt.Errorf("failed to clear existing attainment data: %w", err)
	}

	totalRows := 0

	// Seed historical data (pre-2010) from embedded Census CPS series.
	for _, h := range historicalAttainment {
		if h.year < startYear || h.year > endYear {
			continue
		}
		_, err := c.db.Exec(`
			INSERT INTO educational_attainment (year, age_group, education_level, percentage, source)
			VALUES (?, ?, ?, ?, ?)
		`, h.year, "25plus", "bachelors_plus", h.pct, sourceName)
		if err != nil {
			fmt.Printf("    Warning: failed to insert historical year %d: %v\n", h.year, err)
			continue
		}
		totalRows++
	}
	fmt.Printf("    ✓ Inserted %d historical attainment rows (1940–2009)\n", totalRows)

	// Fetch live ACS 1-year estimates for 2010–present.
	apiRows := 0
	for year := max(2010, startYear); year <= min(endYear, 2024); year++ {
		// B15003_022E = Bachelor's degree count, B15003_001E = Total population 25+
		url := fmt.Sprintf(
			"https://api.census.gov/data/%d/acs/acs1?get=NAME,B15003_022E,B15003_001E&for=us:*",
			year,
		)

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

		if len(data) < 2 {
			fmt.Printf("    ⚠ No data for year %d\n", year)
			continue
		}

		row := data[1]
		if len(row) < 3 {
			continue
		}

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
		_, err = c.db.Exec(`
			INSERT INTO educational_attainment (year, age_group, education_level, percentage, source)
			VALUES (?, ?, ?, ?, ?)
		`, year, "25plus", "bachelors_plus", percentage, sourceName)
		if err != nil {
			fmt.Printf("    Warning: failed to insert year %d: %v\n", year, err)
			continue
		}
		apiRows++
		totalRows++
		fmt.Printf("    ✓ Year %d: %.1f%%\n", year, percentage)
	}
	fmt.Printf("    ✓ Fetched %d years from Census ACS API (2010–present)\n", apiRows)

	yearsRange := fmt.Sprintf("%d-%d", startYear, endYear)
	status := "success"
	if totalRows == 0 {
		status = "partial"
	}
	database.UpdateSourceMetadata(c.db, sourceName, yearsRange, totalRows, status,
		fmt.Sprintf("Historical (1940–2009) + Census ACS API (2010+): %d total rows", totalRows))

	fmt.Printf("  ✓ Census download complete: %d rows\n", totalRows)
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
