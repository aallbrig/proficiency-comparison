# Hugo Site Dynamic Data Display

## Summary

Both the main timeline tool and the `/data/` page now **dynamically detect** year ranges and data availability from the JSON files, eliminating all hardcoded "1870-2025" references.

## Changes Made

### 1. Main Page Timeline (`layouts/index.html` + `js/timeline.js`)

**Before:**
```javascript
const MIN_YEAR = 1870;  // Hardcoded
const MAX_YEAR = 2025;  // Hardcoded
```

**After:**
```javascript
let MIN_YEAR = 1950;  // Default fallback
let MAX_YEAR = 2024;  // Default fallback

// Dynamically calculated from data files
MIN_YEAR = Math.min(...allYears);  // e.g., 2010
MAX_YEAR = Math.max(...allYears);  // e.g., 2024
```

**Display Updates:**
- Initial: "Birth Years: Loading..."
- After load: "Birth Years: 2010-2024" (based on actual data)
- Timeline range: Automatically set to data coverage

### 2. Data Page (`layouts/data/list.html` + `js/data-charts.js`)

**Before:**
```html
<h2>Literacy Rates (1870-2025)</h2>
<p>Explore 150+ years of data from 1870 to 2025...</p>
```

**After:**
```html
<h2>Literacy Rates</h2> <!-- Updated by JS -->
<p>Explore US educational statistics...</p> <!-- Updated by JS -->
```

**JavaScript Updates:**
```javascript
// Updates header: "Literacy Rates (2010-2020)"
updateSectionHeader(statName, {min: 2010, max: 2020});

// Updates summary: "Explore 14+ years from 2010 to 2024..."
updateDataSummary();

// Hides sections with no data
hideSection('proficiency'); // No NAEP data
```

## How to Update Data

When you add new data, the site automatically adjusts:

```bash
# 1. Download new data
edu-stats all --years 2010-2024

# 2. Generate Hugo JSON files
edu-stats step generate-assets

# 3. Rebuild Hugo site
cd hugo/site && hugo

# 4. Year ranges update automatically on page load!
```

## Current Behavior

### With Current Data (2010-2024):

**Main Page:**
- Timeline: 2010-2024 range
- Markers: Can only be placed 2010-2024
- Display: "Birth Years: 2010-2024"

**Data Page:**
- Summary: "Explore 14+ years... from 2010 to 2024 across 4 key metrics with 174 total data points"
- Literacy Rates (2010-2020)
- Educational Attainment (2010-2024)
- Graduation Rates (2011-2020)
- Enrollment Rates (2010-2020)
- Test Proficiency: [HIDDEN]
- Early Childhood: [HIDDEN]

### With No Data:

**Main Page:**
- Shows "No Data Available" warning
- Provides CLI commands to populate data
- Timeline uses fallback range (1950-2024)

**Data Page:**
- All sections hidden
- Message: "Year ranges automatically detected from available data files"

## Benefits

✅ **Always accurate** - Shows what you have, not aspirational ranges
✅ **Self-documenting** - Users see exact data coverage
✅ **Graceful handling** - Missing data sections are hidden
✅ **Maintenance-free** - No manual year range updates needed
✅ **Informative** - Shows data point counts and spans

## Files Modified

### Main Timeline:
- `layouts/index.html` - Changed default year display, updated text
- `static/js/timeline.js` - Made MIN_YEAR/MAX_YEAR dynamic

### Data Page:
- `layouts/data/list.html` - Removed hardcoded year ranges
- `static/js/data-charts.js` - Added dynamic year range detection
- Added section hiding for missing data
- Added dynamic summary updates

## Testing

The site will now correctly display:

**Scenario 1: Only 2010-2020 data**
- Timeline: 2010-2020
- All sections show (2010-2020)

**Scenario 2: Historical data added (1970-2024)**
- Timeline: 1970-2024
- Sections show individual ranges (e.g., Literacy 1970-2000, Attainment 2010-2024)
- Summary: "54+ years from 1970 to 2024..."

**Scenario 3: No data**
- Timeline: Uses fallback 1950-2024
- Data page: All sections hidden
- Warning message displayed

## Future Enhancements

Potential additions:
- Show data gaps in timeline (e.g., 1970-1980, then 2010-2024)
- Add data quality indicators (API vs manual)
- Display last generated timestamp in JSON metadata
- Show per-metric data sources dynamically
