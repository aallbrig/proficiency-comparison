package downloaders

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/aallbrig/proficiency-comparison/internal/database"
)

type WorldBankDownloader struct {
	db *sql.DB
}

func NewWorldBankDownloader(db *sql.DB) *WorldBankDownloader {
	return &WorldBankDownloader{db: db}
}

func (w *WorldBankDownloader) Download(startYear, endYear int, dryRun bool) error {
	sourceName := "world_bank_literacy"
	
	if dryRun {
		fmt.Printf("  [DRY RUN] Would download World Bank literacy data for %d-%d\n", startYear, endYear)
		return nil
	}

	fmt.Println("  Downloading World Bank literacy data...")

	indicators := []struct {
		code     string
		ageGroup string
	}{
		{"SE.ADT.LITR.ZS", "adult_15plus"},
		{"SE.ADT.1524.LT.ZS", "youth_15-24"},
	}

	totalRows := 0
	for _, indicator := range indicators {
		url := fmt.Sprintf(
			"https://api.worldbank.org/v2/country/USA/indicator/%s?date=%d:%d&format=csv",
			indicator.code, startYear, endYear,
		)

		resp, err := http.Get(url)
		if err != nil {
			database.UpdateSourceMetadata(w.db, sourceName, "", 0, "failed", err.Error())
			return fmt.Errorf("failed to download %s: %w", indicator.code, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			errMsg := fmt.Sprintf("HTTP %d", resp.StatusCode)
			database.UpdateSourceMetadata(w.db, sourceName, "", 0, "failed", errMsg)
			return fmt.Errorf("download failed with status %d", resp.StatusCode)
		}

		// Parse CSV
		reader := csv.NewReader(resp.Body)
		reader.LazyQuotes = true
		reader.TrimLeadingSpace = true

		// Skip metadata rows (World Bank CSVs have metadata in first few rows)
		for i := 0; i < 5; i++ {
			_, err := reader.Read()
			if err == io.EOF {
				break
			}
		}

		rowCount := 0
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}

			if len(record) < 5 {
				continue
			}

			// Parse year and value
			year, err := strconv.Atoi(strings.TrimSpace(record[3]))
			if err != nil {
				continue
			}

			valueStr := strings.TrimSpace(record[4])
			if valueStr == "" || valueStr == ".." {
				continue
			}

			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				continue
			}

			// Insert into database
			_, err = w.db.Exec(`
				INSERT INTO literacy_rates (year, age_group, rate, source)
				VALUES (?, ?, ?, ?)
				ON CONFLICT(year, age_group, gender, source) DO UPDATE SET
					rate = excluded.rate
			`, year, indicator.ageGroup, value, sourceName)

			if err != nil {
				fmt.Printf("    Warning: failed to insert row: %v\n", err)
				continue
			}

			rowCount++
		}

		fmt.Printf("    ✓ Imported %d rows for %s\n", rowCount, indicator.ageGroup)
		totalRows += rowCount
	}

	yearsRange := fmt.Sprintf("%d-%d", startYear, endYear)
	database.UpdateSourceMetadata(w.db, sourceName, yearsRange, totalRows, "success", "")
	
	fmt.Printf("  ✓ World Bank download complete: %d total rows\n", totalRows)
	return nil
}
