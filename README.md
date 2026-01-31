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
wget https://github.com/aallbrig/proficiency-comparison/releases/latest/download/edu-stats-cli-linux-amd64
chmod +x edu-stats-cli-linux-amd64
sudo mv edu-stats-cli-linux-amd64 /usr/local/bin/edu-stats-cli

# Or for macOS ARM
wget https://github.com/aallbrig/proficiency-comparison/releases/latest/download/edu-stats-cli-darwin-arm64
chmod +x edu-stats-cli-darwin-arm64
sudo mv edu-stats-cli-darwin-arm64 /usr/local/bin/edu-stats-cli
```

#### Option 2: Install from Source

```bash
git clone https://github.com/aallbrig/proficiency-comparison.git
cd proficiency-comparison/go
go install
```

### Usage

#### CLI Commands

**Check Status**
```bash
edu-stats-cli status
```
Shows database status, last download times, row counts, and data source connectivity.

**Download All Data**
```bash
edu-stats-cli all --years=1970-2025
```
Runs the complete pipeline: schema setup, data downloads, processing, and Hugo asset generation.

**Dry Run**
```bash
edu-stats-cli all --years=1970-2025 --dry-run
```
Simulates the download without fetching data, showing estimates.

**Check Version and Updates**
```bash
edu-stats-cli version
edu-stats-cli upgrade
```

**Individual Pipeline Steps**
```bash
edu-stats-cli all check-schema
edu-stats-cli all download-worldbank
edu-stats-cli all download-census
edu-stats-cli all download-nces
edu-stats-cli all download-naep
edu-stats-cli all download-ecls
```

#### Website

**Run Locally**
```bash
cd hugo
hugo server -D
```
Visit http://localhost:1313

**Build for Production**
```bash
cd hugo
hugo
```
Output in `hugo/public/`

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
│   ├── cmd/                 # Command implementations
│   ├── internal/            # Internal packages
│   │   ├── database/        # SQLite operations
│   │   ├── downloaders/     # Data source downloaders
│   │   ├── generators/      # Hugo asset generators
│   │   └── utils/           # Utilities
│   ├── go.mod
│   └── main.go
├── hugo/                    # Hugo website source
│   ├── content/             # Content and data files
│   ├── layouts/             # HTML templates
│   ├── static/              # CSS, JS, images
│   └── config.toml
├── schema.sql               # Database schema
├── .github/workflows/       # CI/CD pipelines
└── README.md
```

## Development

### Running Tests

**Go Tests**
```bash
cd go
go test ./... -v
```

**JavaScript Tests**
```bash
cd hugo
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

1. **Schema Check**: Validates/creates SQLite database with schema.sql
2. **Download**: Fetches data from each source (parallel where safe)
3. **Parse**: Processes CSV/JSON/Excel files
4. **Store**: Inserts into SQLite (idempotent, skips duplicates)
5. **Generate**: Creates JSON assets for Hugo in `/hugo/content/data/`

### Website Features

- **Interactive Timeline**: Add/remove birth year markers to compare cohorts
- **Stat Selection**: Dropdown to choose different educational metrics
- **Data Visualization**: Chart.js graphs showing trends over time
- **URL Sharing**: Shareable URLs with current configuration
- **QR Codes**: Generate QR codes for easy mobile sharing
- **Responsive Design**: Bootstrap-based mobile-friendly interface

## License

MIT License - see [LICENSE](LICENSE) for details

## Acknowledgments

Data provided by:
- World Bank / UNESCO Institute for Statistics
- US Census Bureau
- National Center for Education Statistics (NCES)
- National Assessment of Educational Progress (NAEP)
- Early Childhood Longitudinal Study (ECLS)
