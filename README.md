# Educational Stats Comparison Project

A comprehensive tool for downloading, analyzing, and visualizing US educational statistics by birth year cohorts. Compare literacy rates, educational attainment, graduation rates, standardized test proficiency, and early childhood metrics across different generations.

## Features

- **Go CLI Tool**: Download and manage educational statistics from multiple authoritative sources
- **SQLite Database**: Efficient local storage with automatic schema management
- **Hugo Website**: Interactive timeline for comparing cohorts with responsive design
- **Data Sources**: World Bank, US Census, NCES (NAEP, ECLS), and more
- **Idempotent Pipeline**: Resumable downloads with intelligent caching
- **Auto-Updates**: Check for and install CLI updates from GitHub releases

## Quick Start

### Prerequisites

- Go 1.21+ (for building from source)
- Hugo 0.120+ (for website development)
- SQLite3 (usually pre-installed)

### Installation

#### Option 1: Download Pre-built Binary (Recommended)

Visit the [Releases](https://github.com/aallbrig/proficiency-comparison/releases) page and download the binary for your platform:

```bash
# Linux/macOS
wget https://github.com/aallbrig/proficiency-comparison/releases/latest/download/edu-stats-linux-amd64
chmod +x edu-stats-linux-amd64
sudo mv edu-stats-linux-amd64 /usr/local/bin/edu-stats

# Or for macOS ARM
wget https://github.com/aallbrig/proficiency-comparison/releases/latest/download/edu-stats-darwin-arm64
chmod +x edu-stats-darwin-arm64
sudo mv edu-stats-darwin-arm64 /usr/local/bin/edu-stats
```

#### Option 2: Install from Source

```bash
git clone https://github.com/aallbrig/proficiency-comparison.git
cd proficiency-comparison/go/edu-stats
go install
```

### Usage

#### CLI Commands

**Initialize Database**
```bash
edu-stats init
```
Creates the database and applies schema.sql. Safe to run multiple times.

**Sync Schema**
```bash
edu-stats sync
```
Applies schema changes from schema.sql (for migrations). Safe to run multiple times.

**Check Status**
```bash
edu-stats status
```
Shows database status, last download times, row counts, and data source connectivity.

**Download All Data**
```bash
edu-stats all --years=1970-2025
```
Runs the complete pipeline: schema setup, data downloads, processing, and Hugo asset generation.

**Dry Run**
```bash
edu-stats all --years=1970-2025 --dry-run
```
Simulates the download without fetching data, showing estimates.

**Check Version and Updates**
```bash
edu-stats version
edu-stats upgrade
```

**Individual Pipeline Steps**
```bash
edu-stats init                    # Initialize database schema
edu-stats all check-schema
edu-stats all download-worldbank
edu-stats all download-census
edu-stats all download-nces
edu-stats all download-naep
edu-stats all download-ecls
```

#### Website

**Run Locally**
```bash
cd hugo/site
hugo server -D
```
Visit http://localhost:1313

**Build for Production**
```bash
cd hugo/site
hugo
```
Output in `hugo/site/public/`

## Data Sources

The project integrates data from:

1. **World Bank DataBank** - Literacy rates (1970-present)
2. **US Census Bureau** - Educational attainment (1940-present)
3. **NCES Digest** - Graduation and enrollment rates (1869-present)
4. **NAEP** - Standardized test proficiency (1969-present)
5. **NCES ECLS** - Early childhood metrics (1998-present)

## Project Structure

```
proficiency-comparison/
├── go/                      # Go CLI source code
│   └── edu-stats/          # CLI application
│       ├── cmd/            # Command implementations
│       ├── internal/       # Internal packages
│       │   ├── database/   # SQLite operations
│       │   ├── downloaders/ # Data source downloaders
│       │   ├── generators/ # Hugo asset generators
│       │   └── utils/      # Utilities
│       ├── data/           # Database storage (gitignored)
│       ├── go.mod
│       ├── main.go
│       └── .gitignore
├── hugo/                    # Hugo website source
│   └── site/               # Website root
│       ├── content/        # Content and data files
│       ├── layouts/        # HTML templates
│       ├── static/         # CSS, JS, images
│       ├── config.toml
│       └── .gitignore
├── schema.sql              # Database schema
├── .github/workflows/      # CI/CD pipelines
└── README.md
```

## Development

### Running Tests

**Go Tests**
```bash
cd go/edu-stats
go test ./... -v
```

**JavaScript Tests**
```bash
cd hugo/site
npm install
npm test
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Architecture

### CLI Pipeline

1. **Schema Check**: Validates/creates SQLite database with schema.sql in `go/edu-stats/data/`
   - Can be run standalone with `edu-stats init`
2. **Download**: Fetches data from each source (parallel where safe)
3. **Parse**: Processes CSV/JSON/Excel files
4. **Store**: Inserts into SQLite (idempotent, skips duplicates)
5. **Generate**: Creates JSON assets for Hugo in `/hugo/site/static/data/`

### Website Features

- **Comparison Tables**: Visual cards showing stats for each birth year cohort with generational labels (Baby Boomer, Generation X, Millennial, Generation Z, Generation Alpha)
- **Interactive Timeline**: Click to add markers, drag to adjust birth years
- **Linked Highlighting**: Click markers or tables to highlight both simultaneously
- **Settings Modal**: Select which statistics to display in comparisons
- **Editable Birth Years**: Change cohort years directly in comparison tables
- **Data Visualization**: Chart.js graphs showing trends over time (on data page)
- **URL Sharing**: Shareable URLs with current configuration (cohorts and stats)
  - Example: `/?cohorts=1970,1980,1990&stats=literacy,attainment`
- **QR Codes**: Generate QR codes for easy mobile sharing
- **Responsive Design**: Bootstrap-based mobile-friendly interface
- **Dynamic Data Detection**: Only shows statistics with available data

### Using the Website

**Add Comparisons:**
1. Click "Add Comparison" button to add a new cohort table
2. Click on the timeline to add a marker at a specific year
3. Use "Add Marker" button to enter a specific year

**Edit Birth Years:**
1. Click on the birth year input field in any comparison table
2. Type a new year (1950-2020)
3. The marker and data will update automatically

**Highlight Linked Items:**
- Click any marker on the timeline to highlight its comparison table
- Click any comparison table to highlight its timeline marker
- Tables will scroll into view when highlighted

**Customize Statistics:**
1. Click the gear/settings icon in the top right
2. Check/uncheck statistics to show/hide in comparison tables
3. Click "Save Changes"

**Share Your Configuration:**
- The URL updates automatically with your selections
- Copy the URL or scan the QR code in the footer to share

## License

MIT License - see [LICENSE](LICENSE) for details

## Acknowledgments

Data provided by:
- World Bank / UNESCO Institute for Statistics
- US Census Bureau
- National Center for Education Statistics (NCES)
- National Assessment of Educational Progress (NAEP)
- Early Childhood Longitudinal Study (ECLS)
