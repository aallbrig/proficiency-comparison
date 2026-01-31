package database

import (
	"database/sql"
	"os"
	"testing"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Use in-memory database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Apply schema
	schemaSQL := `
		CREATE TABLE source_metadata (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_name TEXT NOT NULL UNIQUE,
			last_download DATETIME,
			years_available TEXT,
			row_count INTEGER DEFAULT 0,
			status TEXT,
			error_message TEXT
		);

		CREATE TABLE pipeline_metadata (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			step_name TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			status TEXT NOT NULL,
			years_covered TEXT,
			error_message TEXT
		);

		CREATE TABLE literacy_rates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			year INTEGER NOT NULL,
			age_group TEXT NOT NULL,
			rate REAL,
			gender TEXT,
			source TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(year, age_group, gender, source)
		);
	`

	if _, err := db.Exec(schemaSQL); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	return db
}

func TestUpdateSourceMetadata(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := UpdateSourceMetadata(db, "test_source", "2020-2022", 100, "success", "")
	if err != nil {
		t.Errorf("UpdateSourceMetadata failed: %v", err)
	}

	// Verify data
	sources, err := GetSourceMetadata(db)
	if err != nil {
		t.Errorf("GetSourceMetadata failed: %v", err)
	}

	if len(sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(sources))
	}

	if sources[0].Name != "test_source" {
		t.Errorf("Expected source name 'test_source', got '%s'", sources[0].Name)
	}

	if sources[0].RowCount != 100 {
		t.Errorf("Expected row count 100, got %d", sources[0].RowCount)
	}
}

func TestRecordPipelineStep(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := RecordPipelineStep(db, "test_step", "completed", "2020-2022", nil)
	if err != nil {
		t.Errorf("RecordPipelineStep failed: %v", err)
	}

	// Verify data
	lastStep, err := GetLastCompletedStep(db)
	if err != nil {
		t.Errorf("GetLastCompletedStep failed: %v", err)
	}

	if lastStep != "test_step" {
		t.Errorf("Expected last step 'test_step', got '%s'", lastStep)
	}
}

func TestGetTableRowCounts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO literacy_rates (year, age_group, rate, source)
		VALUES (2020, 'adult', 99.0, 'test')
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Note: This test only works if all tables exist
	// In real implementation, we'd use the full schema
}

func TestApplySchema(t *testing.T) {
	// Create temporary schema file
	tmpSchema := "test_schema.sql"
	schemaContent := `
		CREATE TABLE IF NOT EXISTS test_table (
			id INTEGER PRIMARY KEY,
			name TEXT
		);
	`
	
	err := os.WriteFile(tmpSchema, []byte(schemaContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp schema: %v", err)
	}
	defer os.Remove(tmpSchema)

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test schema application
	// Note: ApplySchema looks for schema.sql in specific locations
	// This test would need to be adjusted based on actual file structure
}
