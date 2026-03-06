# Dynamic Data Display

## How It Works

The `/data/` page now displays year ranges and data statistics **dynamically** based on the actual JSON data files, not hardcoded values.

### JavaScript Updates (data-charts.js)

1. **Loads each JSON file** from `/static/data/*.json`
2. **Calculates year ranges** (min/max) from actual data points
3. **Updates section headers** with accurate year ranges `(YYYY-YYYY)`
4. **Updates summary paragraph** with total year span and data point count
5. **Hides sections** with no data (instead of showing empty charts)

### HTML Template (layouts/data/list.html)

- Headers no longer have hardcoded year ranges
- Summary paragraph is generic and gets populated by JS
- Data notes section is updated to reflect current data sources

## Example Output

**Before (hardcoded):**
```
Literacy Rates (1870-2025)
Explore 150+ years of data from 1870 to 2025 across six key metrics.
```

**After (dynamic):**
```
Literacy Rates (2010-2020)
Explore 10+ years of US educational statistics from 2010 to 2020 across 4 key metrics with 174 total data points.
```

## Updating Data

When you regenerate Hugo assets:

```bash
# 1. Update database (add new data)
edu-stats step download-census --years 2010-2024

# 2. Regenerate JSON files
edu-stats step generate-assets

# 3. Rebuild Hugo site
cd hugo/site
hugo

# 4. Year ranges automatically update on page load!
```

## Graceful Handling

- **No data:** Section is hidden entirely
- **Partial data:** Shows actual year range (e.g., 2015-2023)
- **Multiple metrics:** Summary shows aggregate statistics
- **Last updated:** Timestamp shows when page was loaded

## Files Modified

- `layouts/data/list.html` - Removed hardcoded year ranges
- `static/js/data-charts.js` - Added dynamic year range detection
- All section headers now show `(YYYY-YYYY)` based on actual data

## Future Enhancements

Potential improvements:
- Add data quality indicators (API vs estimated)
- Show gaps in data coverage
- Display per-metric data source information
- Add data refresh timestamp from JSON metadata
