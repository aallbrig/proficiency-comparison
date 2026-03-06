package generators

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// setupGeneratorTestDB creates an in-memory SQLite database with the tables
// needed to exercise the Hugo generator.
func setupGeneratorTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}

	schema := `
	CREATE TABLE educational_attainment (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL,
		education_level TEXT NOT NULL,
		percentage REAL,
		source TEXT NOT NULL,
		UNIQUE(year, education_level, source)
	);
	CREATE TABLE literacy_rates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL,
		age_group TEXT NOT NULL,
		rate REAL,
		gender TEXT,
		source TEXT NOT NULL,
		UNIQUE(year, age_group, gender, source)
	);
	CREATE TABLE graduation_rates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL,
		rate REAL,
		state TEXT,
		source TEXT NOT NULL,
		UNIQUE(year, state, source)
	);
	CREATE TABLE enrollment_rates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL,
		enrollment_rate REAL,
		state TEXT,
		source TEXT NOT NULL,
		UNIQUE(year, state, source)
	);
	CREATE TABLE test_proficiency (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL,
		subject TEXT NOT NULL,
		grade INTEGER NOT NULL,
		avg_score REAL,
		source TEXT NOT NULL,
		UNIQUE(year, subject, grade, source)
	);
	CREATE TABLE early_childhood (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL,
		metric_value REAL,
		source TEXT NOT NULL,
		UNIQUE(year, source)
	);`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return db
}

func TestGenerateAllEmptyDB(t *testing.T) {
	db := setupGeneratorTestDB(t)
	defer db.Close()

	dir := t.TempDir()
	gen := &HugoGenerator{db: db}

	err := gen.generateToDir(dir)
	if err != nil {
		t.Fatalf("generateToDir (empty DB) returned error: %v", err)
	}

	// index.json should always be written even with no data
	indexPath := filepath.Join(dir, "index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("index.json not written: %v", err)
	}

	var idx struct {
		Generated string                 `json:"generated"`
		Stats     map[string]interface{} `json:"stats"`
	}
	if err := json.Unmarshal(data, &idx); err != nil {
		t.Fatalf("index.json is not valid JSON: %v", err)
	}

	if idx.Generated == "" {
		t.Error("index.json missing 'generated' field")
	}

	// Generated date should parse as a valid date
	if _, err := time.Parse("2006-01-02", idx.Generated); err != nil {
		t.Errorf("index.json 'generated' is not YYYY-MM-DD: %s", idx.Generated)
	}

	if len(idx.Stats) == 0 {
		t.Error("index.json has no 'stats' entries")
	}
}

func TestGenerateAttainmentData(t *testing.T) {
	db := setupGeneratorTestDB(t)
	defer db.Close()

	// Insert test attainment rows
	_, err := db.Exec(`
		INSERT INTO educational_attainment (year, education_level, percentage, source)
		VALUES (2020, 'bachelors_plus', 21.5, 'test'),
		       (2021, 'bachelors_plus', 22.0, 'test'),
		       (2022, 'bachelors_plus', 22.8, 'test')
	`)
	if err != nil {
		t.Fatalf("insert test data: %v", err)
	}

	gen := &HugoGenerator{db: db}
	data, err := gen.generateAttainmentData()
	if err != nil {
		t.Fatalf("generateAttainmentData: %v", err)
	}

	if len(data.Years) != 3 {
		t.Errorf("expected 3 data points, got %d", len(data.Years))
	}

	if data.Name == "" {
		t.Error("StatData.Name should not be empty")
	}

	if data.Source == "" {
		t.Error("StatData.Source should not be empty")
	}

	// Verify data points are ordered and correct
	if data.Years[0].Year != 2020 {
		t.Errorf("first year should be 2020, got %d", data.Years[0].Year)
	}
	if data.Years[0].Value != 21.5 {
		t.Errorf("first value should be 21.5, got %f", data.Years[0].Value)
	}
}

func TestGenerateProficiencyData(t *testing.T) {
	db := setupGeneratorTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO test_proficiency (year, subject, grade, avg_score, source)
		VALUES (1984, 'reading', 8, 257.0, 'naep'),
		       (1988, 'reading', 8, 258.0, 'naep'),
		       (1992, 'reading', 8, 260.0, 'naep')
	`)
	if err != nil {
		t.Fatalf("insert test data: %v", err)
	}

	gen := &HugoGenerator{db: db}
	data, err := gen.generateProficiencyData()
	if err != nil {
		t.Fatalf("generateProficiencyData: %v", err)
	}

	if len(data.Years) != 3 {
		t.Errorf("expected 3 data points, got %d", len(data.Years))
	}
}

func TestGenerateStatsIndexMarksUnavailable(t *testing.T) {
	dir := t.TempDir()

	// Write only attainment.json — other stat files absent
	attainment := StatData{
		Name:   "Educational Attainment",
		Source: "test",
		Years:  []DataPoint{{Year: 2020, Value: 21.5}, {Year: 2021, Value: 22.0}},
	}
	writeJSON(t, filepath.Join(dir, "attainment.json"), attainment)

	db := setupGeneratorTestDB(t)
	defer db.Close()
	gen := &HugoGenerator{db: db}

	if err := gen.generateStatsIndex(dir); err != nil {
		t.Fatalf("generateStatsIndex: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "index.json"))
	if err != nil {
		t.Fatalf("read index.json: %v", err)
	}

	var idx struct {
		Stats map[string]struct {
			Available  bool `json:"available"`
			DataPoints int  `json:"dataPoints"`
		} `json:"stats"`
	}
	if err := json.Unmarshal(data, &idx); err != nil {
		t.Fatalf("parse index.json: %v", err)
	}

	att := idx.Stats["attainment"]
	if !att.Available {
		t.Error("attainment should be marked available")
	}
	if att.DataPoints != 2 {
		t.Errorf("attainment should have 2 data points, got %d", att.DataPoints)
	}

	grad := idx.Stats["graduation"]
	if grad.Available {
		t.Error("graduation has no file, should be marked unavailable")
	}
}

// writeJSON is a test helper that writes v as JSON to path.
func writeJSON(t *testing.T, path string, v interface{}) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
}
