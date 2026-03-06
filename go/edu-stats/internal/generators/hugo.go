package generators

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type HugoGenerator struct {
	db *sql.DB
}

func NewHugoGenerator(db *sql.DB) *HugoGenerator {
	return &HugoGenerator{db: db}
}

type DataPoint struct {
	Year  int     `json:"year"`
	Value float64 `json:"value"`
	Label string  `json:"label,omitempty"`
}

type StatData struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Source      string      `json:"source"`
	Years       []DataPoint `json:"data"`
}

func (h *HugoGenerator) GenerateAll() error {
	fmt.Println("  Generating Hugo JSON assets...")

	// Try multiple locations for Hugo site
	hugoOutputLocations := []string{
		filepath.Join("hugo", "site", "static", "data"),                    // From repo root
		filepath.Join("..", "..", "hugo", "site", "static", "data"),         // From go/edu-stats/
		filepath.Join("..", "..", "..", "hugo", "site", "static", "data"),   // From go/edu-stats/cmd/
		filepath.Join(os.Getenv("HOME"), "src", "proficiency-comparison", "hugo", "site", "static", "data"), // Absolute fallback
	}
	
	var outputDir string
	for _, loc := range hugoOutputLocations {
		// Check if the parent hugo/site/ directory has either hugo.toml or config.toml
		hugoDir := filepath.Dir(filepath.Dir(loc)) // Go up to hugo/site/
		hasHugo := false
		for _, cfgFile := range []string{"hugo.toml", "config.toml", "config.yaml", "hugo.yaml"} {
			if _, err := os.Stat(filepath.Join(hugoDir, cfgFile)); err == nil {
				hasHugo = true
				break
			}
		}
		if hasHugo {
			outputDir = loc
			break
		}
	}
	
	if outputDir == "" {
		// Fall back to first option and create it
		outputDir = hugoOutputLocations[0]
	}
	
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	fmt.Printf("    Output directory: %s\n", outputDir)
	return h.generateToDir(outputDir)
}

// generateToDir writes all stat JSON files and index.json into outputDir.
// It is separated from GenerateAll to allow testing with a temp directory.
func (h *HugoGenerator) generateToDir(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	generators := []struct {
		name     string
		filename string
		fn       func() (StatData, error)
	}{
		{"Literacy Rates", "literacy.json", h.generateLiteracyData},
		{"Educational Attainment", "attainment.json", h.generateAttainmentData},
		{"Graduation Rates", "graduation.json", h.generateGraduationData},
		{"Enrollment Rates", "enrollment.json", h.generateEnrollmentData},
		{"Test Proficiency", "proficiency.json", h.generateProficiencyData},
		{"Early Childhood", "early_childhood.json", h.generateEarlyChildhoodData},
	}

	for _, gen := range generators {
		data, err := gen.fn()
		if err != nil {
			fmt.Printf("    Warning: failed to generate %s: %v\n", gen.name, err)
			continue
		}

		if len(data.Years) == 0 {
			fmt.Printf("    ⚠ %s: no data available\n", gen.name)
			continue
		}

		outputPath := filepath.Join(outputDir, gen.filename)
		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", gen.filename, err)
		}

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(data); err != nil {
			file.Close()
			return fmt.Errorf("failed to write %s: %w", gen.filename, err)
		}
		file.Close()

		fmt.Printf("    ✓ Generated %s (%d data points)\n", gen.filename, len(data.Years))
	}

	// Generate stats index
	if err := h.generateStatsIndex(outputDir); err != nil {
		return fmt.Errorf("failed to generate stats index: %w", err)
	}

	fmt.Println("  ✓ Hugo asset generation complete")
	return nil
}

func (h *HugoGenerator) generateLiteracyData() (StatData, error) {
	rows, err := h.db.Query(`
		SELECT year, AVG(rate) as avg_rate
		FROM literacy_rates
		GROUP BY year
		ORDER BY year
	`)
	if err != nil {
		return StatData{}, err
	}
	defer rows.Close()

	var data StatData
	data.Name = "Literacy Rates"
	data.Description = "Adult literacy and high school completion rates"
	data.Source = "NCES Digest"

	for rows.Next() {
		var dp DataPoint
		if err := rows.Scan(&dp.Year, &dp.Value); err != nil {
			continue
		}
		data.Years = append(data.Years, dp)
	}

	return data, nil
}

func (h *HugoGenerator) generateAttainmentData() (StatData, error) {
	rows, err := h.db.Query(`
		SELECT year, AVG(percentage) as avg_pct
		FROM educational_attainment
		WHERE education_level = 'bachelors_plus'
		GROUP BY year
		ORDER BY year
	`)
	if err != nil {
		return StatData{}, err
	}
	defer rows.Close()

	var data StatData
	data.Name = "Educational Attainment"
	data.Description = "Percentage with bachelor's degree or higher (25+)"
	data.Source = "US Census Bureau"

	for rows.Next() {
		var dp DataPoint
		if err := rows.Scan(&dp.Year, &dp.Value); err != nil {
			continue
		}
		data.Years = append(data.Years, dp)
	}

	return data, nil
}

func (h *HugoGenerator) generateGraduationData() (StatData, error) {
	rows, err := h.db.Query(`
		SELECT year, AVG(rate) as avg_rate
		FROM graduation_rates
		WHERE state = 'US' OR state IS NULL
		GROUP BY year
		ORDER BY year
	`)
	if err != nil {
		return StatData{}, err
	}
	defer rows.Close()

	var data StatData
	data.Name = "High School Graduation Rates"
	data.Description = "4-year adjusted cohort graduation rate (ACGR)"
	data.Source = "NCES Digest"

	for rows.Next() {
		var dp DataPoint
		if err := rows.Scan(&dp.Year, &dp.Value); err != nil {
			continue
		}
		data.Years = append(data.Years, dp)
	}

	return data, nil
}

func (h *HugoGenerator) generateEnrollmentData() (StatData, error) {
	rows, err := h.db.Query(`
		SELECT year, AVG(enrollment_rate) as avg_rate
		FROM enrollment_rates
		WHERE state = 'US' OR state IS NULL
		GROUP BY year
		ORDER BY year
	`)
	if err != nil {
		return StatData{}, err
	}
	defer rows.Close()

	var data StatData
	data.Name = "Enrollment Rates"
	data.Description = "School enrollment rates (averaged across age groups)"
	data.Source = "NCES Digest"

	for rows.Next() {
		var dp DataPoint
		if err := rows.Scan(&dp.Year, &dp.Value); err != nil {
			continue
		}
		data.Years = append(data.Years, dp)
	}

	return data, nil
}

func (h *HugoGenerator) generateProficiencyData() (StatData, error) {
	rows, err := h.db.Query(`
		SELECT year, AVG(avg_score) as avg_score
		FROM test_proficiency
		WHERE subject = 'reading' AND grade = 8
		GROUP BY year
		ORDER BY year
	`)
	if err != nil {
		return StatData{}, err
	}
	defer rows.Close()

	var data StatData
	data.Name = "Test Proficiency"
	data.Description = "NAEP Reading scores (Grade 8)"
	data.Source = "NAEP"

	for rows.Next() {
		var dp DataPoint
		if err := rows.Scan(&dp.Year, &dp.Value); err != nil {
			continue
		}
		data.Years = append(data.Years, dp)
	}

	return data, nil
}

func (h *HugoGenerator) generateEarlyChildhoodData() (StatData, error) {
	rows, err := h.db.Query(`
		SELECT year, AVG(metric_value) as avg_value
		FROM early_childhood
		GROUP BY year
		ORDER BY year
	`)
	if err != nil {
		return StatData{}, err
	}
	defer rows.Close()

	var data StatData
	data.Name = "Early Childhood Metrics"
	data.Description = "Early literacy and readiness indicators"
	data.Source = "NCES ECLS"

	for rows.Next() {
		var dp DataPoint
		if err := rows.Scan(&dp.Year, &dp.Value); err != nil {
			continue
		}
		data.Years = append(data.Years, dp)
	}

	return data, nil
}

func (h *HugoGenerator) generateStatsIndex(outputDir string) error {
	type StatIndexEntry struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Filename    string `json:"filename"`
		Available   bool   `json:"available"`
		YearMin     int    `json:"yearMin,omitempty"`
		YearMax     int    `json:"yearMax,omitempty"`
		DataPoints  int    `json:"dataPoints,omitempty"`
	}

	type IndexData struct {
		Generated string                    `json:"generated"`
		Stats     map[string]StatIndexEntry `json:"stats"`
	}

	index := IndexData{
		Generated: time.Now().Format("2006-01-02"),
		Stats:     make(map[string]StatIndexEntry),
	}

	// Check each stat file to see if it exists and has data
	statsToCheck := []struct {
		key         string
		name        string
		description string
		filename    string
	}{
		{"literacy", "Literacy Rates", "Adult literacy and high school completion rates", "literacy.json"},
		{"attainment", "Educational Attainment", "Educational attainment levels by degree", "attainment.json"},
		{"graduation", "Graduation Rates", "High school graduation rates", "graduation.json"},
		{"enrollment", "Enrollment Rates", "School enrollment rates by age group", "enrollment.json"},
		{"proficiency", "Test Proficiency", "NAEP test proficiency scores", "proficiency.json"},
		{"early_childhood", "Early Childhood", "Early childhood readiness metrics", "early_childhood.json"},
	}

	for _, stat := range statsToCheck {
		entry := StatIndexEntry{
			Name:        stat.name,
			Description: stat.description,
			Filename:    stat.filename,
			Available:   false,
		}

		// Check if file exists and read its metadata
		filePath := filepath.Join(outputDir, stat.filename)
		if fileData, err := os.ReadFile(filePath); err == nil {
			var statData StatData
			if err := json.Unmarshal(fileData, &statData); err == nil && len(statData.Years) > 0 {
				entry.Available = true
				entry.DataPoints = len(statData.Years)
				
				// Calculate year range
				minYear := statData.Years[0].Year
				maxYear := statData.Years[0].Year
				for _, dp := range statData.Years {
					if dp.Year < minYear {
						minYear = dp.Year
					}
					if dp.Year > maxYear {
						maxYear = dp.Year
					}
				}
				entry.YearMin = minYear
				entry.YearMax = maxYear
			}
		}

		index.Stats[stat.key] = entry
	}

	outputPath := filepath.Join(outputDir, "index.json")
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(index); err != nil {
		return err
	}

	fmt.Printf("    ✓ Generated index.json (%d stats)\n", len(index.Stats))
	return nil
}
