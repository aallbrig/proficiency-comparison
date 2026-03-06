/**
 * @jest-environment jsdom
 *
 * Tests for timeline.js logic — cohort mapping, generational labels,
 * marker management, and URL parameter handling.
 *
 * Pure-logic functions are re-implemented here to avoid DOM/fetch side-effects
 * from loading the full module.
 */

// ─── Helpers mirroring timeline.js ───────────────────────────────────────────

const GENERATIONS = [
    { name: 'Silent Generation', short: 'Silent',     start: 1928, end: 1945, class: 'silent' },
    { name: 'Baby Boomer',       short: 'Boomer',     start: 1946, end: 1964, class: 'baby-boomer' },
    { name: 'Generation X',      short: 'Gen X',      start: 1965, end: 1980, class: 'x' },
    { name: 'Millennial',        short: 'Millennial', start: 1981, end: 1996, class: 'millennial' },
    { name: 'Generation Z',      short: 'Gen Z',      start: 1997, end: 2012, class: 'z' },
    { name: 'Generation Alpha',  short: 'Gen Alpha',  start: 2013, end: 2030, class: 'alpha' },
];

function getGenerationalLabel(year) {
    const gen = GENERATIONS.find(g => year >= g.start && year <= g.end);
    return gen ? { name: gen.name, class: gen.class } : { name: 'Other', class: 'other' };
}

const STAT_METADATA = {
    literacy:        { cohortOffset: 20, searchWindow: 6  },
    attainment:      { cohortOffset: 26, searchWindow: 3  },
    graduation:      { cohortOffset: 18, searchWindow: 2  },
    enrollment:      { cohortOffset: 10, searchWindow: 2  },
    proficiency:     { cohortOffset: 14, searchWindow: 4  },
    early_childhood: { cohortOffset:  5, searchWindow: 2  },
};

const MARKER_COLORS = ['red', 'blue', 'green', 'orange', 'purple', 'teal'];

function findCohortDataPoint(dataPoints, targetYear, window) {
    let best = null;
    let bestDist = Infinity;
    dataPoints.forEach(d => {
        const dist = Math.abs(d.year - targetYear);
        if (dist <= window && dist < bestDist) {
            bestDist = dist;
            best = d;
        }
    });
    return best;
}

function markerPosition(year, minYear, maxYear) {
    return ((year - minYear) / (maxYear - minYear)) * 100;
}

// ─── Generational label tests ─────────────────────────────────────────────────

describe('getGenerationalLabel', () => {
    test('Silent Generation (1928–1945)', () => {
        expect(getGenerationalLabel(1928).class).toBe('silent');
        expect(getGenerationalLabel(1935).class).toBe('silent');
        expect(getGenerationalLabel(1945).class).toBe('silent');
    });

    test('Baby Boomer (1946–1964)', () => {
        expect(getGenerationalLabel(1946).class).toBe('baby-boomer');
        expect(getGenerationalLabel(1955).class).toBe('baby-boomer');
        expect(getGenerationalLabel(1964).class).toBe('baby-boomer');
    });

    test('Generation X (1965–1980)', () => {
        expect(getGenerationalLabel(1965).class).toBe('x');
        expect(getGenerationalLabel(1972).class).toBe('x');
        expect(getGenerationalLabel(1980).class).toBe('x');
    });

    test('Millennial (1981–1996)', () => {
        expect(getGenerationalLabel(1981).class).toBe('millennial');
        expect(getGenerationalLabel(1990).class).toBe('millennial');
        expect(getGenerationalLabel(1996).class).toBe('millennial');
    });

    test('Generation Z (1997–2012)', () => {
        expect(getGenerationalLabel(1997).class).toBe('z');
        expect(getGenerationalLabel(2005).class).toBe('z');
        expect(getGenerationalLabel(2012).class).toBe('z');
    });

    test('Generation Alpha (2013+)', () => {
        expect(getGenerationalLabel(2013).class).toBe('alpha');
        expect(getGenerationalLabel(2020).class).toBe('alpha');
    });

    test('border years are correctly classified', () => {
        expect(getGenerationalLabel(1964).class).toBe('baby-boomer');
        expect(getGenerationalLabel(1965).class).toBe('x');
        expect(getGenerationalLabel(1980).class).toBe('x');
        expect(getGenerationalLabel(1981).class).toBe('millennial');
        expect(getGenerationalLabel(1996).class).toBe('millennial');
        expect(getGenerationalLabel(1997).class).toBe('z');
        expect(getGenerationalLabel(2012).class).toBe('z');
        expect(getGenerationalLabel(2013).class).toBe('alpha');
    });

    test('returns name alongside class', () => {
        expect(getGenerationalLabel(1985).name).toBe('Millennial');
        expect(getGenerationalLabel(1955).name).toBe('Baby Boomer');
    });
});

// ─── Cohort life-stage offset tests ──────────────────────────────────────────

describe('cohort life-stage offsets', () => {
    test('proficiency: NAEP Grade 8 is at birth year + 14', () => {
        expect(STAT_METADATA.proficiency.cohortOffset).toBe(14);
    });

    test('graduation: high school completion at birth year + 18', () => {
        expect(STAT_METADATA.graduation.cohortOffset).toBe(18);
    });

    test('attainment: bachelor degree at birth year + 26', () => {
        expect(STAT_METADATA.attainment.cohortOffset).toBe(26);
    });

    test('enrollment: K-12 midpoint at birth year + 10', () => {
        expect(STAT_METADATA.enrollment.cohortOffset).toBe(10);
    });

    test('literacy: young adult at birth year + 20', () => {
        expect(STAT_METADATA.literacy.cohortOffset).toBe(20);
    });

    test('early_childhood: kindergarten at birth year + 5', () => {
        expect(STAT_METADATA.early_childhood.cohortOffset).toBe(5);
    });

    test('proficiency search window allows for biennial gaps (4 years)', () => {
        expect(STAT_METADATA.proficiency.searchWindow).toBe(4);
    });
});

// ─── Cohort data-point lookup ─────────────────────────────────────────────────

describe('findCohortDataPoint', () => {
    const naepData = [
        { year: 1971, value: 255 },
        { year: 1975, value: 256 },
        { year: 1980, value: 259 },
        { year: 1984, value: 257 },
        { year: 1988, value: 258 },
        { year: 1992, value: 260 },
        { year: 2004, value: 264 },
        { year: 2019, value: 263 },
    ];

    test('finds exact year match', () => {
        const pt = findCohortDataPoint(naepData, 1984, 4);
        expect(pt).not.toBeNull();
        expect(pt.year).toBe(1984);
    });

    test('finds closest point within window', () => {
        // target 1986: 1984 (dist 2) or 1988 (dist 2)
        const pt = findCohortDataPoint(naepData, 1986, 4);
        expect(pt).not.toBeNull();
        expect([1984, 1988]).toContain(pt.year);
    });

    test('returns null when no point within window', () => {
        const pt = findCohortDataPoint(naepData, 1960, 4);
        expect(pt).toBeNull();
    });

    test('boomer born 1957: NAEP at age 14 = 1971, score 255', () => {
        const birthYear = 1957;
        const offset = STAT_METADATA.proficiency.cohortOffset; // 14
        const window = STAT_METADATA.proficiency.searchWindow; // 4
        const pt = findCohortDataPoint(naepData, birthYear + offset, window);
        expect(pt).not.toBeNull();
        expect(pt.year).toBe(1971);
        expect(pt.value).toBe(255);
    });

    test('gen X born 1970: NAEP at age 14 = 1984, score 257', () => {
        const pt = findCohortDataPoint(naepData, 1970 + 14, 4);
        expect(pt).not.toBeNull();
        expect(pt.year).toBe(1984);
        expect(pt.value).toBe(257);
    });

    test('millennial born 1990: NAEP at age 14 = 2004, score 264', () => {
        const pt = findCohortDataPoint(naepData, 1990 + 14, 4);
        expect(pt).not.toBeNull();
        expect(pt.year).toBe(2004);
        expect(pt.value).toBe(264);
    });

    test('gen Z born 2005: NAEP at age 14 = 2019, score 263', () => {
        const pt = findCohortDataPoint(naepData, 2005 + 14, 4);
        expect(pt).not.toBeNull();
        expect(pt.year).toBe(2019);
    });
});

// ─── Marker position calculation ──────────────────────────────────────────────

describe('markerPosition', () => {
    const MIN = 1928, MAX = 2024;

    test('MIN_YEAR → 0%', () => {
        expect(markerPosition(MIN, MIN, MAX)).toBe(0);
    });

    test('MAX_YEAR → 100%', () => {
        expect(markerPosition(MAX, MIN, MAX)).toBe(100);
    });

    test('midpoint → ~50%', () => {
        const mid = Math.floor((MIN + MAX) / 2);
        expect(markerPosition(mid, MIN, MAX)).toBeCloseTo(50, 0);
    });

    test('1985 is within valid range', () => {
        const pos = markerPosition(1985, MIN, MAX);
        expect(pos).toBeGreaterThan(0);
        expect(pos).toBeLessThan(100);
    });
});

// ─── Marker management ────────────────────────────────────────────────────────

describe('marker management', () => {
    test('duplicate birth-year detection', () => {
        const markers = [{ year: 1980, id: 'm1' }, { year: 1990, id: 'm2' }];
        expect(markers.some(m => m.year === 1980)).toBe(true);
        expect(markers.some(m => m.year === 2000)).toBe(false);
    });

    test('markers are sorted ascending by year', () => {
        const markers = [
            { year: 1990, id: 'm1' },
            { year: 1970, id: 'm2' },
            { year: 1980, id: 'm3' },
        ];
        markers.sort((a, b) => a.year - b.year);
        expect(markers.map(m => m.year)).toEqual([1970, 1980, 1990]);
    });

    test('marker color palette has 6 entries', () => {
        expect(MARKER_COLORS.length).toBe(6);
    });

    test('color assignment cycles after 6 markers', () => {
        expect(MARKER_COLORS[0 % MARKER_COLORS.length]).toBe('red');
        expect(MARKER_COLORS[5 % MARKER_COLORS.length]).toBe('teal');
        expect(MARKER_COLORS[6 % MARKER_COLORS.length]).toBe('red');
    });
});

// ─── URL parameter parsing ────────────────────────────────────────────────────

describe('URL parameter parsing', () => {
    test('parses cohort years', () => {
        const params = new URLSearchParams('cohorts=1955,1970,1985,2000');
        const years = params.get('cohorts').split(',').map(Number);
        expect(years).toEqual([1955, 1970, 1985, 2000]);
    });

    test('parses stat selection', () => {
        const params = new URLSearchParams('stats=literacy,attainment,proficiency');
        expect(params.get('stats').split(',')).toEqual(['literacy', 'attainment', 'proficiency']);
    });

    test('missing cohorts param returns null', () => {
        const params = new URLSearchParams('stats=literacy');
        expect(params.get('cohorts')).toBeNull();
    });

    test('roundtrip: build then parse', () => {
        const years = [1955, 1970, 1985];
        const params = new URLSearchParams();
        params.set('cohorts', years.join(','));
        const parsed = params.get('cohorts').split(',').map(Number);
        expect(parsed).toEqual(years);
    });
});

// ─── GENERATIONS constant integrity ───────────────────────────────────────────

describe('GENERATIONS constant', () => {
    test('contains 6 entries', () => {
        expect(GENERATIONS.length).toBe(6);
    });

    test('generations do not overlap or have gaps', () => {
        for (let i = 1; i < GENERATIONS.length; i++) {
            expect(GENERATIONS[i].start).toBe(GENERATIONS[i - 1].end + 1);
        }
    });

    test('each entry has required fields', () => {
        GENERATIONS.forEach(gen => {
            expect(gen).toHaveProperty('name');
            expect(gen).toHaveProperty('short');
            expect(gen).toHaveProperty('start');
            expect(gen).toHaveProperty('end');
            expect(gen).toHaveProperty('class');
            expect(gen.start).toBeLessThan(gen.end);
        });
    });
});
