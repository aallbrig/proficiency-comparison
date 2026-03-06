package downloaders

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupDownloaderTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	schema := `
	CREATE TABLE educational_attainment (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL, age_group TEXT, education_level TEXT,
		percentage REAL, gender TEXT, race TEXT, source TEXT NOT NULL,
		UNIQUE(year, age_group, education_level, gender, race, source)
	);
	CREATE TABLE literacy_rates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL, age_group TEXT, rate REAL, gender TEXT, source TEXT NOT NULL,
		UNIQUE(year, age_group, gender, source)
	);
	CREATE TABLE graduation_rates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL, rate REAL, cohort_year INTEGER,
		state TEXT, demographics TEXT, source TEXT NOT NULL,
		UNIQUE(year, cohort_year, state, demographics, source)
	);
	CREATE TABLE enrollment_rates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL, age_group TEXT, enrollment_rate REAL,
		level TEXT, state TEXT, demographics TEXT, source TEXT NOT NULL,
		UNIQUE(year, age_group, level, state, demographics, source)
	);
	CREATE TABLE test_proficiency (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL, subject TEXT, grade INTEGER, avg_score REAL,
		proficiency_level TEXT, percentage_proficient REAL, state TEXT,
		demographics TEXT, source TEXT NOT NULL
	);
	CREATE TABLE source_metadata (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_name TEXT UNIQUE NOT NULL, years_range TEXT,
		rows_downloaded INTEGER, status TEXT, notes TEXT, last_run DATETIME
	);`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return db
}

func countRows(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

// --- Census downloader ---

func TestCensusDownloaderDryRun(t *testing.T) {
	db := setupDownloaderTestDB(t)
	defer db.Close()
	d := NewCensusDownloader(db)
	if err := d.Download(1940, 2024, true); err != nil {
		t.Errorf("dry run returned error: %v", err)
	}
	if n := countRows(t, db, "educational_attainment"); n != 0 {
		t.Errorf("dry run should insert 0 rows, got %d", n)
	}
}

func TestCensusDownloaderHistoricalData(t *testing.T) {
	db := setupDownloaderTestDB(t)
	defer db.Close()
	d := NewCensusDownloader(db)
	// Use a year range that only covers historical embedded data (no API calls).
	if err := d.Download(1940, 2009, false); err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	n := countRows(t, db, "educational_attainment")
	if n == 0 {
		t.Error("expected historical attainment rows, got 0")
	}
	// Check a specific known value: 1970 = 11.0%
	var pct float64
	err := db.QueryRow(
		`SELECT percentage FROM educational_attainment WHERE year = 1970`,
	).Scan(&pct)
	if err != nil {
		t.Fatalf("row for 1970 not found: %v", err)
	}
	if pct != 11.0 {
		t.Errorf("1970 attainment: want 11.0, got %.2f", pct)
	}
}

func TestCensusDownloaderIdempotent(t *testing.T) {
	db := setupDownloaderTestDB(t)
	defer db.Close()
	d := NewCensusDownloader(db)
	if err := d.Download(1940, 2009, false); err != nil {
		t.Fatalf("first run: %v", err)
	}
	n1 := countRows(t, db, "educational_attainment")
	if err := d.Download(1940, 2009, false); err != nil {
		t.Fatalf("second run: %v", err)
	}
	n2 := countRows(t, db, "educational_attainment")
	if n1 != n2 {
		t.Errorf("re-run changed row count: %d → %d (not idempotent)", n1, n2)
	}
}

// --- NAEP downloader ---

func TestNAEPDownloaderDryRun(t *testing.T) {
	db := setupDownloaderTestDB(t)
	defer db.Close()
	d := NewNAEPDownloader(db)
	if err := d.Download(1971, 2022, true); err != nil {
		t.Errorf("dry run returned error: %v", err)
	}
	if n := countRows(t, db, "test_proficiency"); n != 0 {
		t.Errorf("dry run should insert 0 rows, got %d", n)
	}
}

func TestNAEPDownloaderSeeds(t *testing.T) {
	db := setupDownloaderTestDB(t)
	defer db.Close()
	d := NewNAEPDownloader(db)
	if err := d.Download(1971, 2022, false); err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	n := countRows(t, db, "test_proficiency")
	if n == 0 {
		t.Error("expected proficiency rows, got 0")
	}
	// Spot check: 1971 LTT score = 255
	var score float64
	err := db.QueryRow(
		`SELECT avg_score FROM test_proficiency WHERE year = 1971 AND subject = 'reading'`,
	).Scan(&score)
	if err != nil {
		t.Fatalf("row for 1971 not found: %v", err)
	}
	if score != 255 {
		t.Errorf("1971 NAEP score: want 255, got %.0f", score)
	}
}

func TestNAEPDownloaderYearFilter(t *testing.T) {
	db := setupDownloaderTestDB(t)
	defer db.Close()
	d := NewNAEPDownloader(db)
	// Only request 2000–2022: should only include entries from 2002 onward.
	if err := d.Download(2000, 2022, false); err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	var minYear int
	if err := db.QueryRow(`SELECT MIN(year) FROM test_proficiency`).Scan(&minYear); err != nil {
		t.Fatalf("min year query: %v", err)
	}
	if minYear < 2000 {
		t.Errorf("expected min year >= 2000, got %d", minYear)
	}
}

func TestNAEPDownloaderIdempotent(t *testing.T) {
	db := setupDownloaderTestDB(t)
	defer db.Close()
	d := NewNAEPDownloader(db)
	if err := d.Download(1971, 2022, false); err != nil {
		t.Fatalf("first run: %v", err)
	}
	n1 := countRows(t, db, "test_proficiency")
	if err := d.Download(1971, 2022, false); err != nil {
		t.Fatalf("second run: %v", err)
	}
	n2 := countRows(t, db, "test_proficiency")
	if n1 != n2 {
		t.Errorf("re-run changed row count: %d → %d (not idempotent)", n1, n2)
	}
}

// --- NCES downloader ---

func TestNCESDownloaderDryRun(t *testing.T) {
	db := setupDownloaderTestDB(t)
	defer db.Close()
	d := NewNCESDownloader(db)
	if err := d.Download(1960, 2020, true); err != nil {
		t.Errorf("dry run returned error: %v", err)
	}
	if n := countRows(t, db, "graduation_rates"); n != 0 {
		t.Errorf("dry run should insert 0 graduation rows, got %d", n)
	}
}

func TestNCESDownloaderGraduationSeeds(t *testing.T) {
	db := setupDownloaderTestDB(t)
	defer db.Close()
	d := NewNCESDownloader(db)
	if err := d.Download(1960, 2020, false); err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	n := countRows(t, db, "graduation_rates")
	if n == 0 {
		t.Error("expected graduation rows, got 0")
	}
	// Spot check: 1990 AFGR = 74.4%
	var rate float64
	if err := db.QueryRow(`SELECT rate FROM graduation_rates WHERE year = 1990`).Scan(&rate); err != nil {
		t.Fatalf("1990 graduation not found: %v", err)
	}
	if rate != 74.4 {
		t.Errorf("1990 graduation rate: want 74.4, got %.1f", rate)
	}
}

func TestNCESDownloaderEnrollmentSeeds(t *testing.T) {
	db := setupDownloaderTestDB(t)
	defer db.Close()
	d := NewNCESDownloader(db)
	if err := d.Download(1950, 2020, false); err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	n := countRows(t, db, "enrollment_rates")
	if n == 0 {
		t.Error("expected enrollment rows, got 0")
	}
	// Spot check: 1970 enrollment = 90.2%
	var rate float64
	if err := db.QueryRow(`SELECT enrollment_rate FROM enrollment_rates WHERE year = 1970`).Scan(&rate); err != nil {
		t.Fatalf("1970 enrollment not found: %v", err)
	}
	if rate != 90.2 {
		t.Errorf("1970 enrollment rate: want 90.2, got %.1f", rate)
	}
}

func TestNCESDownloaderIdempotent(t *testing.T) {
	db := setupDownloaderTestDB(t)
	defer db.Close()
	d := NewNCESDownloader(db)
	if err := d.Download(1960, 2020, false); err != nil {
		t.Fatalf("first run: %v", err)
	}
	g1 := countRows(t, db, "graduation_rates")
	e1 := countRows(t, db, "enrollment_rates")
	if err := d.Download(1960, 2020, false); err != nil {
		t.Fatalf("second run: %v", err)
	}
	g2 := countRows(t, db, "graduation_rates")
	e2 := countRows(t, db, "enrollment_rates")
	if g1 != g2 || e1 != e2 {
		t.Errorf("re-run not idempotent: grad %d→%d, enroll %d→%d", g1, g2, e1, e2)
	}
}

// --- World Bank (literacy) downloader ---

func TestWorldBankDownloaderDryRun(t *testing.T) {
	db := setupDownloaderTestDB(t)
	defer db.Close()
	d := NewWorldBankDownloader(db)
	if err := d.Download(1950, 2020, true); err != nil {
		t.Errorf("dry run returned error: %v", err)
	}
	if n := countRows(t, db, "literacy_rates"); n != 0 {
		t.Errorf("dry run should insert 0 rows, got %d", n)
	}
}

func TestWorldBankDownloaderSeeds(t *testing.T) {
	db := setupDownloaderTestDB(t)
	defer db.Close()
	d := NewWorldBankDownloader(db)
	if err := d.Download(1950, 2020, false); err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	n := countRows(t, db, "literacy_rates")
	if n == 0 {
		t.Error("expected literacy rows, got 0")
	}
	// Spot check: 1969 literacy = 98.9%
	var rate float64
	if err := db.QueryRow(`SELECT rate FROM literacy_rates WHERE year = 1969`).Scan(&rate); err != nil {
		t.Fatalf("1969 literacy not found: %v", err)
	}
	if rate != 98.9 {
		t.Errorf("1969 literacy rate: want 98.9, got %.1f", rate)
	}
}

func TestWorldBankDownloaderIdempotent(t *testing.T) {
	db := setupDownloaderTestDB(t)
	defer db.Close()
	d := NewWorldBankDownloader(db)
	if err := d.Download(1950, 2020, false); err != nil {
		t.Fatalf("first run: %v", err)
	}
	n1 := countRows(t, db, "literacy_rates")
	if err := d.Download(1950, 2020, false); err != nil {
		t.Fatalf("second run: %v", err)
	}
	n2 := countRows(t, db, "literacy_rates")
	if n1 != n2 {
		t.Errorf("re-run changed row count: %d → %d (not idempotent)", n1, n2)
	}
}
