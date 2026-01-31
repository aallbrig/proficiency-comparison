-- Educational Stats Database Schema

-- Pipeline metadata tracking
CREATE TABLE IF NOT EXISTS pipeline_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    step_name TEXT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL CHECK(status IN ('started', 'completed', 'failed')),
    years_covered TEXT,
    error_message TEXT,
    execution_time_seconds INTEGER
);

-- Data source download tracking
CREATE TABLE IF NOT EXISTS source_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_name TEXT NOT NULL UNIQUE,
    last_download DATETIME,
    last_modified TEXT,
    etag TEXT,
    content_hash TEXT,
    years_available TEXT,
    row_count INTEGER DEFAULT 0,
    status TEXT CHECK(status IN ('success', 'partial', 'failed')),
    error_message TEXT
);

-- Literacy rates data
CREATE TABLE IF NOT EXISTS literacy_rates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    year INTEGER NOT NULL,
    age_group TEXT NOT NULL,
    rate REAL,
    gender TEXT,
    source TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(year, age_group, gender, source)
);

-- Educational attainment data
CREATE TABLE IF NOT EXISTS educational_attainment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    year INTEGER NOT NULL,
    age_group TEXT NOT NULL,
    education_level TEXT NOT NULL CHECK(education_level IN ('high_school', 'bachelors_plus', 'associates', 'graduate')),
    percentage REAL,
    gender TEXT,
    race TEXT,
    source TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(year, age_group, education_level, gender, race, source)
);

-- High school graduation rates
CREATE TABLE IF NOT EXISTS graduation_rates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    year INTEGER NOT NULL,
    rate REAL,
    cohort_year INTEGER,
    state TEXT,
    demographics TEXT,
    source TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(year, cohort_year, state, demographics, source)
);

-- Enrollment rates
CREATE TABLE IF NOT EXISTS enrollment_rates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    year INTEGER NOT NULL,
    age_group TEXT NOT NULL,
    enrollment_rate REAL,
    level TEXT CHECK(level IN ('elementary', 'secondary', 'postsecondary')),
    state TEXT,
    demographics TEXT,
    source TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(year, age_group, level, state, demographics, source)
);

-- Standardized test proficiency (NAEP)
CREATE TABLE IF NOT EXISTS test_proficiency (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    year INTEGER NOT NULL,
    subject TEXT NOT NULL CHECK(subject IN ('reading', 'math', 'science', 'writing')),
    grade INTEGER NOT NULL CHECK(grade IN (4, 8, 12)),
    avg_score REAL,
    proficiency_level TEXT,
    percentage_proficient REAL,
    state TEXT,
    demographics TEXT,
    source TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(year, subject, grade, proficiency_level, state, demographics, source)
);

-- Early childhood metrics
CREATE TABLE IF NOT EXISTS early_childhood (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    year INTEGER NOT NULL,
    cohort_year INTEGER,
    metric_name TEXT NOT NULL,
    metric_value REAL,
    age_months INTEGER,
    demographics TEXT,
    source TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(year, cohort_year, metric_name, age_months, demographics, source)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_literacy_year ON literacy_rates(year);
CREATE INDEX IF NOT EXISTS idx_attainment_year ON educational_attainment(year);
CREATE INDEX IF NOT EXISTS idx_graduation_year ON graduation_rates(year);
CREATE INDEX IF NOT EXISTS idx_enrollment_year ON enrollment_rates(year);
CREATE INDEX IF NOT EXISTS idx_test_year ON test_proficiency(year);
CREATE INDEX IF NOT EXISTS idx_early_year ON early_childhood(year);
CREATE INDEX IF NOT EXISTS idx_pipeline_step ON pipeline_metadata(step_name, timestamp);
CREATE INDEX IF NOT EXISTS idx_source_name ON source_metadata(source_name);
