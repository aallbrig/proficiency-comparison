package downloaders

import (
	"database/sql"
	"encoding/json"
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
	
	// Example API call for reading at grade 8
	url := fmt.Sprintf(
		"https://www.nationsreportcard.gov/api/ndecore/v2/analyze?type=national&subject=reading&grade=8&measure=average&jurisdiction=nation&stattype=avgscore&years=%d:%d",
		startYear, endYear,
	)
	
	resp, err := http.Get(url)
	if err != nil {
		database.UpdateSourceMetadata(n.db, sourceName, "", 0, "failed", err.Error())
		return fmt.Errorf("failed to download NAEP data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("    ⚠ NAEP API returned status %d, marking as partial\n", resp.StatusCode)
		database.UpdateSourceMetadata(n.db, sourceName, "", 0, "partial", 
			fmt.Sprintf("API unavailable: HTTP %d", resp.StatusCode))
		return nil
	}

	var data NAEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		database.UpdateSourceMetadata(n.db, sourceName, "", 0, "failed", err.Error())
		return fmt.Errorf("failed to parse NAEP data: %w", err)
	}

	totalRows := 0
	for _, item := range data.Result {
		// Parse year and score
		var year int
		var score float64
		fmt.Sscanf(item.Year, "%d", &year)
		fmt.Sscanf(item.Value, "%f", &score)

		if year == 0 || score == 0 {
			continue
		}

		// Insert into database
		_, err = n.db.Exec(`
			INSERT INTO test_proficiency (year, subject, grade, avg_score, source)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(year, subject, grade, proficiency_level, state, demographics, source) DO UPDATE SET
				avg_score = excluded.avg_score
		`, year, "reading", 8, score, sourceName)

		if err != nil {
			fmt.Printf("    Warning: failed to insert row: %v\n", err)
			continue
		}

		totalRows++
	}

	fmt.Printf("    ✓ Imported %d rows for reading grade 8\n", totalRows)
	fmt.Println("    ℹ Note: Additional subjects/grades would require multiple API calls")

	yearsRange := fmt.Sprintf("%d-%d", startYear, endYear)
	database.UpdateSourceMetadata(n.db, sourceName, yearsRange, totalRows, "success", "")
	
	fmt.Printf("  ✓ NAEP download complete: %d rows\n", totalRows)
	return nil
}
