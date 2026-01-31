package downloaders

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aallbrig/proficiency-comparison/internal/database"
	"github.com/xuri/excelize/v2"
)

type NCESDownloader struct {
	db *sql.DB
}

func NewNCESDownloader(db *sql.DB) *NCESDownloader {
	return &NCESDownloader{db: db}
}

func (n *NCESDownloader) Download(startYear, endYear int, dryRun bool) error {
	sourceName := "nces_digest"
	
	if dryRun {
		fmt.Printf("  [DRY RUN] Would download NCES graduation/enrollment data for %d-%d\n", startYear, endYear)
		return nil
	}

	fmt.Println("  Downloading NCES graduation and enrollment data...")
	
	// NCES Digest tables we want to download
	tables := []struct {
		name     string
		url      string
		sheetIdx int
		dataType string
	}{
		{
			name:     "Graduation rates (Table 219.46)",
			url:      "https://nces.ed.gov/programs/digest/d22/tables/xls/tabn219.46.xls",
			sheetIdx: 0,
			dataType: "graduation",
		},
		{
			name:     "Enrollment rates (Table 103.20)",
			url:      "https://nces.ed.gov/programs/digest/d22/tables/xls/tabn103.20.xls",
			sheetIdx: 0,
			dataType: "enrollment",
		},
	}
	
	totalRows := 0
	
	for _, table := range tables {
		fmt.Printf("    Downloading %s...\n", table.name)
		
		// Download file
		filePath, err := n.downloadFile(table.url, sourceName)
		if err != nil {
			fmt.Printf("    ⚠ Failed to download %s: %v\n", table.name, err)
			continue
		}
		
		// Parse Excel file
		rows, err := n.parseExcelFile(filePath, table.sheetIdx, table.dataType, startYear, endYear)
		if err != nil {
			fmt.Printf("    ⚠ Failed to parse %s: %v\n", table.name, err)
			// Mark as parse error but don't fail
			if fileID, _ := n.getFileID(sourceName, table.url); fileID > 0 {
				database.MarkFileParseError(n.db, fileID, err.Error())
			}
			continue
		}
		
		// Mark file as parsed
		if fileID, _ := n.getFileID(sourceName, table.url); fileID > 0 {
			database.MarkFileParsed(n.db, fileID)
		}
		
		fmt.Printf("    ✓ Parsed %d rows from %s\n", rows, table.name)
		totalRows += rows
	}

	yearsRange := fmt.Sprintf("%d-%d", startYear, endYear)
	if totalRows > 0 {
		database.UpdateSourceMetadata(n.db, sourceName, yearsRange, totalRows, "success", "")
	} else {
		database.UpdateSourceMetadata(n.db, sourceName, yearsRange, 0, "partial", 
			"Unable to parse Excel files")
	}
	
	fmt.Printf("  ✓ NCES download complete: %d rows\n", totalRows)
	return nil
}

func (n *NCESDownloader) downloadFile(url, sourceName string) (string, error) {
	// Check if file already downloaded
	hash := fmt.Sprintf("%x", []byte(url)) // Simple hash for now
	
	exists, err := database.FileExists(n.db, sourceName, url, hash)
	if err == nil && exists {
		// File already exists, get path
		files, err := database.GetUnparsedFiles(n.db, sourceName)
		if err == nil {
			for _, f := range files {
				if f.FileURL == url && f.FilePath != "" {
					if _, err := os.Stat(f.FilePath); err == nil {
						fmt.Printf("    ✓ Using cached file\n")
						return f.FilePath, nil
					}
				}
			}
		}
	}
	
	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	
	// Create downloads directory
	downloadDir := "data/downloads"
	os.MkdirAll(downloadDir, 0755)
	
	// Save file
	filename := filepath.Base(url)
	if !strings.HasSuffix(filename, ".xls") && !strings.HasSuffix(filename, ".xlsx") {
		filename = filename + ".xls"
	}
	filePath := filepath.Join(downloadDir, filename)
	
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()
	
	size, err := io.Copy(outFile, resp.Body)
	if err != nil {
		return "", err
	}
	
	// Compute hash
	contentHash, _ := database.ComputeFileHash(filePath)
	
	// Save metadata
	database.SaveRawFile(n.db, sourceName, url, filePath, "xls", size, contentHash)
	
	return filePath, nil
}

func (n *NCESDownloader) parseExcelFile(filePath string, sheetIdx int, dataType string, startYear, endYear int) (int, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return 0, fmt.Errorf("no sheets in file")
	}
	
	sheetName := sheets[0]
	if sheetIdx < len(sheets) {
		sheetName = sheets[sheetIdx]
	}
	
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return 0, err
	}
	
	// Parse based on data type
	rowCount := 0
	switch dataType {
	case "graduation":
		rowCount, err = n.parseGraduationData(rows, startYear, endYear)
	case "enrollment":
		rowCount, err = n.parseEnrollmentData(rows, startYear, endYear)
	default:
		return 0, fmt.Errorf("unknown data type: %s", dataType)
	}
	
	return rowCount, err
}

func (n *NCESDownloader) parseGraduationData(rows [][]string, startYear, endYear int) (int, error) {
	// NCES tables typically have header rows, then data rows with year in first column
	// This is a simplified parser - real NCES tables are complex
	
	rowCount := 0
	sourceName := "nces_digest"
	
	// Skip header rows (usually first 5-10 rows)
	for i := 10; i < len(rows); i++ {
		if len(rows[i]) < 2 {
			continue
		}
		
		// Try to parse year from first column
		yearStr := strings.TrimSpace(rows[i][0])
		var year int
		fmt.Sscanf(yearStr, "%d", &year)
		
		if year < startYear || year > endYear {
			continue
		}
		
		// Try to parse graduation rate from second column (usually a percentage)
		rateStr := strings.TrimSpace(rows[i][1])
		rateStr = strings.ReplaceAll(rateStr, "%", "")
		var rate float64
		_, err := fmt.Sscanf(rateStr, "%f", &rate)
		if err != nil {
			continue
		}
		
		// Insert into database
		_, err = n.db.Exec(`
			INSERT INTO graduation_rates (year, rate, source)
			VALUES (?, ?, ?)
			ON CONFLICT(year, cohort_year, state, demographics, source) DO UPDATE SET
				rate = excluded.rate
		`, year, rate, sourceName)
		
		if err == nil {
			rowCount++
		}
	}
	
	return rowCount, nil
}

func (n *NCESDownloader) parseEnrollmentData(rows [][]string, startYear, endYear int) (int, error) {
	// Similar structure to graduation data
	rowCount := 0
	sourceName := "nces_digest"
	
	for i := 10; i < len(rows); i++ {
		if len(rows[i]) < 3 {
			continue
		}
		
		yearStr := strings.TrimSpace(rows[i][0])
		var year int
		fmt.Sscanf(yearStr, "%d", &year)
		
		if year < startYear || year > endYear {
			continue
		}
		
		// Parse enrollment rate
		rateStr := strings.TrimSpace(rows[i][1])
		rateStr = strings.ReplaceAll(rateStr, "%", "")
		var rate float64
		_, err := fmt.Sscanf(rateStr, "%f", &rate)
		if err != nil {
			continue
		}
		
		// Insert into database
		_, err = n.db.Exec(`
			INSERT INTO enrollment_rates (year, age_group, enrollment_rate, source)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(year, age_group, level, state, demographics, source) DO UPDATE SET
				enrollment_rate = excluded.enrollment_rate
		`, year, "all", rate, sourceName)
		
		if err == nil {
			rowCount++
		}
	}
	
	return rowCount, nil
}

func (n *NCESDownloader) getFileID(sourceName, fileURL string) (int64, error) {
	var id int64
	err := n.db.QueryRow(`
		SELECT id FROM raw_files WHERE source_name = ? AND file_url = ?
	`, sourceName, fileURL).Scan(&id)
	return id, err
}
