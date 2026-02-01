package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DatabaseFile string

func init() {
	// Get user config/data directory (follows XDG Base Directory Specification)
	configDir := os.Getenv("EDU_STATS_DATA_DIR")
	
	if configDir == "" {
		// Use XDG_DATA_HOME or fallback to ~/.local/share
		if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
			configDir = filepath.Join(xdgDataHome, "edu-stats")
		} else if homeDir, err := os.UserHomeDir(); err == nil {
			configDir = filepath.Join(homeDir, ".local", "share", "edu-stats")
		} else {
			// Last resort fallback to current directory
			configDir = "data"
		}
	}
	
	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create data directory %s: %v\n", configDir, err)
		configDir = "data"
		os.MkdirAll(configDir, 0755)
	}
	
	DatabaseFile = filepath.Join(configDir, "edu_stats.db")
}

func GetDatabasePath() string {
	return DatabaseFile
}

func GetDataDir() string {
	return filepath.Dir(DatabaseFile)
}

func Open() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", DatabaseFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return db, nil
}

func ApplySchema(db *sql.DB) error {
	schemaPath := "schema.sql"
	
	// Try multiple locations in order:
	// 1. Current directory
	// 2. Repository root (../../ from binary location)
	// 3. /usr/share/edu-stats/ (system install)
	// 4. Same directory as executable
	
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	
	locations := []string{
		schemaPath,                                    // Current directory
		filepath.Join("..", "..", schemaPath),         // From go/edu-stats/
		filepath.Join("..", "..", "..", schemaPath),   // From go/edu-stats/cmd/
		filepath.Join(execDir, schemaPath),            // Same as executable
		filepath.Join(execDir, "..", "..", schemaPath), // From installed location
		"/usr/share/edu-stats/schema.sql",            // System install
	}
	
	var schemaBytes []byte
	var err error
	var foundLocation string
	
	for _, loc := range locations {
		schemaBytes, err = os.ReadFile(loc)
		if err == nil {
			foundLocation = loc
			break
		}
	}
	
	if err != nil {
		return fmt.Errorf("failed to find schema.sql (searched %d locations): %w", len(locations), err)
	}

	if _, err := db.Exec(string(schemaBytes)); err != nil {
		return fmt.Errorf("failed to apply schema from %s: %w", foundLocation, err)
	}

	fmt.Printf("✓ Database schema applied successfully (from %s)\n", foundLocation)
	return nil
}

type DatabaseInfo struct {
	SizeBytes    int64
	SchemaStatus string
	TableCount   int
}

func GetDatabaseInfo(db *sql.DB) (DatabaseInfo, error) {
	var info DatabaseInfo

	// Get file size
	if stat, err := os.Stat(DatabaseFile); err == nil {
		info.SizeBytes = stat.Size()
	}

	// Count tables
	row := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err := row.Scan(&info.TableCount); err != nil {
		return info, err
	}

	if info.TableCount > 0 {
		info.SchemaStatus = "present"
	} else {
		info.SchemaStatus = "missing"
	}

	return info, nil
}

type SourceMetadata struct {
	Name           string
	LastDownload   *time.Time
	YearsAvailable string
	RowCount       int
	Status         string
}

func GetSourceMetadata(db *sql.DB) ([]SourceMetadata, error) {
	rows, err := db.Query(`
		SELECT source_name, last_download, years_available, row_count, status 
		FROM source_metadata 
		ORDER BY source_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []SourceMetadata
	for rows.Next() {
		var s SourceMetadata
		var lastDownload sql.NullTime
		var status sql.NullString
		
		err := rows.Scan(&s.Name, &lastDownload, &s.YearsAvailable, &s.RowCount, &status)
		if err != nil {
			return nil, err
		}
		
		if lastDownload.Valid {
			s.LastDownload = &lastDownload.Time
		}
		if status.Valid {
			s.Status = status.String
		} else {
			s.Status = "unknown"
		}
		
		sources = append(sources, s)
	}

	return sources, rows.Err()
}

func GetTableRowCounts(db *sql.DB) (map[string]int, error) {
	tables := []string{
		"literacy_rates",
		"educational_attainment",
		"graduation_rates",
		"enrollment_rates",
		"test_proficiency",
		"early_childhood",
	}

	counts := make(map[string]int)
	for _, table := range tables {
		var count int
		row := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table))
		if err := row.Scan(&count); err != nil {
			return nil, err
		}
		counts[table] = count
	}

	return counts, nil
}

func RecordPipelineStep(db *sql.DB, stepName, status, yearsCovered string, err *error) error {
	errorMsg := ""
	if err != nil && *err != nil {
		errorMsg = (*err).Error()
	}

	_, execErr := db.Exec(`
		INSERT INTO pipeline_metadata (step_name, status, years_covered, error_message)
		VALUES (?, ?, ?, ?)
	`, stepName, status, yearsCovered, errorMsg)

	return execErr
}

func GetLastCompletedStep(db *sql.DB) (string, error) {
	var stepName string
	err := db.QueryRow(`
		SELECT step_name 
		FROM pipeline_metadata 
		WHERE status = 'completed'
		ORDER BY timestamp DESC 
		LIMIT 1
	`).Scan(&stepName)

	if err == sql.ErrNoRows {
		return "", nil
	}

	return stepName, err
}

func UpdateSourceMetadata(db *sql.DB, sourceName, yearsAvailable string, rowCount int, status string, errorMsg string) error {
	_, err := db.Exec(`
		INSERT INTO source_metadata (source_name, last_download, years_available, row_count, status, error_message)
		VALUES (?, CURRENT_TIMESTAMP, ?, ?, ?, ?)
		ON CONFLICT(source_name) DO UPDATE SET
			last_download = CURRENT_TIMESTAMP,
			years_available = excluded.years_available,
			row_count = excluded.row_count,
			status = excluded.status,
			error_message = excluded.error_message
	`, sourceName, yearsAvailable, rowCount, status, errorMsg)

	return err
}
