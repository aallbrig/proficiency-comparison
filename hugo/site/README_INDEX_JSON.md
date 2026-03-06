# Index.json - Dynamic Data Discovery

## Overview

The site now uses `index.json` as a single source of truth for discovering available data files. This eliminates hardcoding and enables fully dynamic behavior.

## Index.json Structure

```json
{
  "generated": "2026-02-03",
  "stats": {
    "literacy": {
      "name": "Literacy Rates",
      "description": "Adult literacy and high school completion rates",
      "filename": "literacy.json",
      "available": true,
      "yearMin": 2010,
      "yearMax": 2020,
      "dataPoints": 4
    },
    ...
  }
}
```

## How It Works

### Generation (Go CLI)

When you run `edu-stats step generate-assets`:

1. Generates individual JSON files (literacy.json, attainment.json, etc.)
2. Scans all generated files
3. Creates `index.json` with metadata about each file
4. Includes: name, description, filename, availability, year range, data points

### Consumption (JavaScript)

Both pages load `index.json` first:

```javascript
// 1. Load index
const response = await fetch('/data/index.json');
const indexData = await response.json();

// 2. Iterate through available stats
for (const [statName, statInfo] of Object.entries(indexData.stats)) {
  if (statInfo.available) {
    // 3. Load individual stat file
    await fetch(`/data/${statInfo.filename}`);
    
    // 4. Use metadata from index
    console.log(`${statInfo.name}: ${statInfo.yearMin}-${statInfo.yearMax}`);
  }
}
```

## Benefits

### ✅ Single Request for Discovery
- **Before:** 6 fetch requests to discover data (try each file)
- **After:** 1 fetch request (index.json), then only load available files

### ✅ Rich Metadata
- Year ranges known before loading data
- Data point counts available immediately
- Descriptions and display names centralized

### ✅ Performance
- Skip unavailable data files entirely
- No 404 errors for missing data
- Faster page load (fewer requests)

### ✅ Dynamic Rendering
- Generate UI elements based on available data
- Show correct year ranges immediately
- Hide unavailable sections proactively

## Page Implementations

### /data/ Page (Data Visualization)

**Responsive Grid:**
- Mobile: 1 column
- Desktop (lg+): 3 columns
- Cards auto-generated from index

**Dynamic Behavior:**
```javascript
// Create section only if available
if (statInfo.available) {
  createStatSection(container, statName, statInfo);
  loadAndRenderChart(statName, statInfo);
}
```

**Result:**
- Only available stats appear
- Year ranges in headers
- Data point counts in "Show Data" button

### / Page (Timeline Tool)

**Available Stats:**
- Settings modal only shows available stats
- Stats list populated from index
- Unavailable stats filtered out

**Timeline Range:**
```javascript
// Automatically set from all available data
MIN_YEAR = Math.min(...allYearMins);
MAX_YEAR = Math.max(...allYearMaxs);
```

## Updating Data

When you add new data, everything auto-updates:

```bash
# 1. Add data to database
edu-stats all --years 2010-2024

# 2. Regenerate assets (includes index.json)
edu-stats step generate-assets

# 3. Rebuild Hugo
cd hugo/site && hugo

# 4. Site automatically shows new data range!
```

## Fallback Behavior

If `index.json` doesn't exist (backward compatibility):

```javascript
if (!index) {
  // Fall back to trying individual files
  for (const stat of ['literacy', 'attainment', ...]) {
    try { await fetch(`/data/${stat}.json`); } catch {}
  }
}
```

## File Locations

- **Source:** `go/edu-stats/internal/generators/hugo.go`
- **Output:** `hugo/site/static/data/index.json`
- **Consumed by:**
  - `hugo/site/static/js/data-charts.js` (/data/ page)
  - `hugo/site/static/js/timeline.js` (/ page)

## Example Use Cases

### Show Data Coverage Summary
```javascript
const stats = indexData.stats;
const available = Object.values(stats).filter(s => s.available);
console.log(`${available.length} metrics available`);
console.log(`Total data points: ${available.reduce((sum, s) => sum + s.dataPoints, 0)}`);
```

### Dynamic Year Range
```javascript
const years = available.map(s => [s.yearMin, s.yearMax]).flat();
const range = `${Math.min(...years)}-${Math.max(...years)}`;
```

### Filter Stats
```javascript
// Only show stats with 2020+ data
const recentStats = Object.entries(stats)
  .filter(([_, s]) => s.available && s.yearMax >= 2020);
```

## Migration Notes

### Old Approach (Hardcoded)
```javascript
const datasets = ['literacy', 'attainment', ...]; // Hardcoded list
for (const stat of datasets) {
  try { await fetch(`/data/${stat}.json`); } // Try blindly
  catch { console.log('Not found'); } // Handle 404s
}
```

### New Approach (Dynamic)
```javascript
const index = await fetch('/data/index.json'); // One request
for (const [stat, info] of Object.entries(index.stats)) {
  if (info.available) { // Only load if available
    await fetch(`/data/${info.filename}`); // No 404s
  }
}
```

## Future Enhancements

Potential additions to index.json:
- **lastUpdated**: Timestamp of last generation
- **dataQuality**: API vs manual vs estimated
- **gaps**: Year gaps in data (e.g., [1970-1980, 2010-2024])
- **sources**: List of data sources per stat
- **aggregations**: Pre-calculated summaries
