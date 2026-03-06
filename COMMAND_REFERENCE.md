# CLI Command Reference

## Quick Start

### Generate Hugo Assets (Most Common)
```bash
# Regenerate Hugo JSON files from current database
edu-stats step generate-assets
```

### Check Status
```bash
# Basic status (no reset history)
edu-stats status

# Verbose status (includes reset history)
edu-stats status -v
# or
edu-stats status --verbose
```

## Individual Steps

The `step` command lets you run any pipeline step independently:

```bash
# Check/apply database schema
edu-stats step check-schema

# Download from specific sources
edu-stats step download-worldbank --years 2010-2024
edu-stats step download-census --years 2010-2024
edu-stats step download-nces --years 2010-2024
edu-stats step download-naep --years 2010-2024
edu-stats step download-ecls --years 2010-2024

# Generate Hugo assets
edu-stats step generate-assets

# Dry run (test without downloading)
edu-stats step download-census --years 2010-2024 --dry-run
```

## Full Pipeline

```bash
# Run entire pipeline
edu-stats all --years 2010-2024

# Dry run (test without downloading)
edu-stats all --years 2010-2024 --dry-run

# Force re-download (ignore resume)
edu-stats all --years 2010-2024 --force
```

## Data Management

### Reset Data
```bash
# Reset specific year range
edu-stats reset 2010 2024

# Reset all data
edu-stats reset 1870 2025
```

### Check Status
```bash
# Basic status
edu-stats status

# With reset history
edu-stats status -v
```

## Common Workflows

### After Manually Adding Data
If you manually insert data (via Python scripts or direct SQL):

```bash
# Regenerate Hugo assets
edu-stats step generate-assets
```

### Fresh Start
```bash
# 1. Reset all data
edu-stats reset 1870 2025

# 2. Download fresh data
edu-stats all --years 2010-2024

# Hugo assets are generated automatically at end of pipeline
```

### Update Hugo Site Only
```bash
# Just regenerate Hugo JSON files
edu-stats step generate-assets

# Then run Hugo server
cd hugo/site
hugo server
```

### Test Downloads
```bash
# Test what would be downloaded without actually downloading
edu-stats step download-census --years 2010-2024 --dry-run
```

## Hugo Output

Generated files are placed in: `hugo/site/static/data/`

Files generated:
- `literacy.json` - Literacy/completion rates
- `attainment.json` - Educational attainment (degrees)
- `graduation.json` - High school graduation rates
- `enrollment.json` - School enrollment rates
- `proficiency.json` - Test proficiency (NAEP)
- `early_childhood.json` - Early childhood metrics
- `stats_index.json` - Index of all stats

## Database Location

Default: `~/.local/share/edu-stats/edu_stats.db`

Override with environment variable:
```bash
export EDU_STATS_DATA_DIR=/path/to/custom/dir
```

## Status Flags

- Default: Clean output without reset history
- `-v` or `--verbose`: Show reset operation history

## Step Flags

- `--years YYYY-YYYY`: Year range (default: 1970-2025)
- `--dry-run`: Simulate without downloading

## All Command Flags

- `--years YYYY-YYYY`: Year range (default: 1970-2025)
- `--dry-run`: Simulate without downloading
- `--force`: Force re-download (ignore resume checkpoint)

## Examples

### Daily Workflow
```bash
# Check current status
edu-stats status

# Regenerate Hugo assets after data changes
edu-stats step generate-assets

# Run Hugo site
cd hugo/site && hugo server
```

### Weekly Update
```bash
# Download latest data
edu-stats all --years 2010-2024

# Site is automatically updated
```

### Troubleshooting
```bash
# Check status with verbose output
edu-stats status -v

# Check database location and connectivity
edu-stats status

# Test specific downloader
edu-stats step download-census --years 2020-2024 --dry-run
```
