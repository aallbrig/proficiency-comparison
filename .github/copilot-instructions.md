# Copilot Instructions for Educational Stats Project

This file provides key guidelines for implementing the project. Refer to it throughout development. Add unit/integration tests for all components (e.g., Go CLI commands, data parsing, DB interactions, JS timeline logic). Use testing frameworks like Go's `testing` package and Jest for JS.

## Project Overview
- Build a Go CLI for downloading educational stats data into SQLite.
- Generate JSON assets for a Hugo website.
- Website: Interactive timeline for comparing stats by birth year cohorts.
- Focus stats: Literacy rates, educational attainment (% high school/bachelor's+), high school graduation rates, enrollment rates, standardized test proficiency (e.g., NAEP reading/math), early childhood metrics.
- Handle cohorts: Map birth years to age-appropriate data (e.g., adult literacy at 15+, proficiency at grade levels). For recent years (e.g., 2016 births), use proxies like current grade-level data.

## Data Sources and Stats
Use these sources for downloading data. Prioritize bulk CSV/Excel/JSON downloads or APIs. Store raw/processed data in SQLite. Track download metadata (e.g., last run time, year coverage, success/failure) in DB for idempotency.

1. **Literacy Rates** (Youth/Adult, by year/age group):
   - Source: World Bank DataBank (mirrors UNESCO UIS). API: `https://api.worldbank.org/v2/`.
   - Indicators: SE.ADT.LITR.ZS (adult 15+), SE.ADT.1524.LT.ZS (youth 15-24).
   - Download: CSV via API (e.g., `curl "https://api.worldbank.org/v2/country/USA/indicator/SE.ADT.LITR.ZS?date=1970:2024&downloadformat=csv"`).
   - Historical: 1970–present. Granularity: Nationwide (US), gender/age.
   - Alternative: NCES historical (1870–1979) via tables at https://nces.ed.gov/programs/digest/d23/tables/dt23_603.10.asp (download XLS).

2. **Educational Attainment** (% with high school, bachelor's+, by age group/year):
   - Source: US Census Bureau CPS/ASEC. Tables: https://www.census.gov/topics/education/educational-attainment/data/tables.html.
   - Download: Excel/CSV historical series (e.g., Table A-2, 1940–present).
   - API: `https://api.census.gov/data/` (e.g., `curl "https://api.census.gov/data/2023/acs/acs1?get=NAME,B15003_022E,B15003_001E&for=us:*"` for bachelor's %).
   - Historical: 1940–present. Granularity: Nationwide, state, age/sex/race.

3. **High School Graduation and Enrollment Rates** (By year/cohort):
   - Source: NCES Digest of Education Statistics. Tables: https://nces.ed.gov/programs/digest/current_tables.asp (e.g., Table 219.10 for enrollment, Table 104.10 for graduation).
   - Download: Machine-readable CSV/Excel (e.g., `curl -O https://nces.ed.gov/programs/digest/d23/tables/dt23_104.10.xls`).
   - Historical: 1869+ for some. Granularity: Nationwide, state, demographics.

4. **Standardized Test Proficiency** (Reading/math at grades 4/8/12, ties to literacy):
   - Source: NAEP. Data Explorer: https://nces.ed.gov/nationsreportcard/data/.
   - API: `https://www.nationsreportcard.gov/api/` (e.g., `curl "https://www.nationsreportcard.gov/api/ndecore/v2/analyze?type=national&subject=reading&grade=8&measure=average&jurisdiction=nation&stattype=avgscore&years=1992:2022"`).
   - Download: Export CSV from explorer or API JSON.
   - Historical: 1969–present. Granularity: Nationwide, state, demographics.

5. **Early Childhood Metrics** (Readiness, early literacy for recent cohorts):
   - Source: NCES ECLS. Reports/Data: https://nces.ed.gov/ecls/.
   - Download: Public summaries in Excel/PDF (apply for restricted data if needed, but start with public).
   - Historical: 1998+. Granularity: Nationwide, demographics.

## General Data Guidelines
- Focus: US nationwide data; add global if time.
- Year Range: Default 1970–present; user-arg like 1971-2025.
- Re-download Logic: Check DB for last download timestamp per source. If >1 month old or data missing for requested years, re-download. Compare response headers (e.g., Last-Modified) or hash content.
- Store in SQLite: Schema in `schema.sql` (tables for each stat, metadata table for pipeline steps/last runs).
- Generate Hugo Assets: After download, export DB data to JSON (e.g., `/content/data/literacy.json`) for website use.
- Tests: Add for parsing responses, DB inserts, JSON generation.

## Additional Instructions
- Add tests for all code.
- Ensure CLI/website handle no-data case (e.g., prompt to run CLI).
- Use research above for stats comparisons (e.g., align cohorts by life stage).
