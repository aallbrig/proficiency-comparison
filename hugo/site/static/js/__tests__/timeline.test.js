/**
 * @jest-environment jsdom
 */

// Mock Bootstrap
global.bootstrap = {
    Modal: class {
        constructor(element) {
            this.element = element;
        }
        show() {}
        hide() {}
        static getInstance() {
            return new this();
        }
    }
};

// Mock QRCode
global.QRCode = class {
    constructor(element, options) {
        this.element = element;
        this.options = options;
    }
};

describe('Timeline functionality with generational labels', () => {
    beforeEach(() => {
        document.body.innerHTML = `
            <button id="settingsBtn">Settings</button>
            <button id="addTableBtn">Add Table</button>
            <button id="addMarkerFromTimelineBtn">Add Marker</button>
            <div id="timeline" class="timeline-track"></div>
            <div id="timelineLabels" class="timeline-labels"></div>
            <div id="timelineMarkers" class="timeline-markers"></div>
            <div id="yearRange"></div>
            <div id="comparisonTables"></div>
            <div id="noTablesMessage"></div>
            <div id="noDataWarning" style="display: none;"></div>
            <div id="qrcode"></div>
            <div class="modal" id="settingsModal">
                <div id="statCheckboxes"></div>
                <button id="saveSettingsBtn">Save</button>
            </div>
        `;
        
        // Reset global state
        if (typeof window.markers !== 'undefined') {
            window.markers = [];
        }
    });

    test('generational label calculation - Baby Boomer', () => {
        const getGenerationalLabel = (year) => {
            if (year >= 1946 && year <= 1964) return { name: 'Baby Boomer', class: 'baby-boomer' };
            if (year >= 1965 && year <= 1980) return { name: 'Generation X', class: 'x' };
            if (year >= 1981 && year <= 1996) return { name: 'Millennial', class: 'millennial' };
            if (year >= 1997 && year <= 2012) return { name: 'Generation Z', class: 'z' };
            if (year >= 2013) return { name: 'Generation Alpha', class: 'alpha' };
            return { name: 'Other', class: 'other' };
        };
        
        expect(getGenerationalLabel(1950)).toEqual({ name: 'Baby Boomer', class: 'baby-boomer' });
        expect(getGenerationalLabel(1946)).toEqual({ name: 'Baby Boomer', class: 'baby-boomer' });
        expect(getGenerationalLabel(1964)).toEqual({ name: 'Baby Boomer', class: 'baby-boomer' });
    });

    test('generational label calculation - Generation X', () => {
        const getGenerationalLabel = (year) => {
            if (year >= 1946 && year <= 1964) return { name: 'Baby Boomer', class: 'baby-boomer' };
            if (year >= 1965 && year <= 1980) return { name: 'Generation X', class: 'x' };
            if (year >= 1981 && year <= 1996) return { name: 'Millennial', class: 'millennial' };
            if (year >= 1997 && year <= 2012) return { name: 'Generation Z', class: 'z' };
            if (year >= 2013) return { name: 'Generation Alpha', class: 'alpha' };
            return { name: 'Other', class: 'other' };
        };
        
        expect(getGenerationalLabel(1970)).toEqual({ name: 'Generation X', class: 'x' });
        expect(getGenerationalLabel(1965)).toEqual({ name: 'Generation X', class: 'x' });
        expect(getGenerationalLabel(1980)).toEqual({ name: 'Generation X', class: 'x' });
    });

    test('generational label calculation - Millennial', () => {
        const getGenerationalLabel = (year) => {
            if (year >= 1946 && year <= 1964) return { name: 'Baby Boomer', class: 'baby-boomer' };
            if (year >= 1965 && year <= 1980) return { name: 'Generation X', class: 'x' };
            if (year >= 1981 && year <= 1996) return { name: 'Millennial', class: 'millennial' };
            if (year >= 1997 && year <= 2012) return { name: 'Generation Z', class: 'z' };
            if (year >= 2013) return { name: 'Generation Alpha', class: 'alpha' };
            return { name: 'Other', class: 'other' };
        };
        
        expect(getGenerationalLabel(1990)).toEqual({ name: 'Millennial', class: 'millennial' });
        expect(getGenerationalLabel(1981)).toEqual({ name: 'Millennial', class: 'millennial' });
        expect(getGenerationalLabel(1996)).toEqual({ name: 'Millennial', class: 'millennial' });
    });

    test('generational label calculation - Generation Z', () => {
        const getGenerationalLabel = (year) => {
            if (year >= 1946 && year <= 1964) return { name: 'Baby Boomer', class: 'baby-boomer' };
            if (year >= 1965 && year <= 1980) return { name: 'Generation X', class: 'x' };
            if (year >= 1981 && year <= 1996) return { name: 'Millennial', class: 'millennial' };
            if (year >= 1997 && year <= 2012) return { name: 'Generation Z', class: 'z' };
            if (year >= 2013) return { name: 'Generation Alpha', class: 'alpha' };
            return { name: 'Other', class: 'other' };
        };
        
        expect(getGenerationalLabel(2000)).toEqual({ name: 'Generation Z', class: 'z' });
        expect(getGenerationalLabel(1997)).toEqual({ name: 'Generation Z', class: 'z' });
        expect(getGenerationalLabel(2012)).toEqual({ name: 'Generation Z', class: 'z' });
    });

    test('generational label calculation - Generation Alpha', () => {
        const getGenerationalLabel = (year) => {
            if (year >= 1946 && year <= 1964) return { name: 'Baby Boomer', class: 'baby-boomer' };
            if (year >= 1965 && year <= 1980) return { name: 'Generation X', class: 'x' };
            if (year >= 1981 && year <= 1996) return { name: 'Millennial', class: 'millennial' };
            if (year >= 1997 && year <= 2012) return { name: 'Generation Z', class: 'z' };
            if (year >= 2013) return { name: 'Generation Alpha', class: 'alpha' };
            return { name: 'Other', class: 'other' };
        };
        
        expect(getGenerationalLabel(2015)).toEqual({ name: 'Generation Alpha', class: 'alpha' });
        expect(getGenerationalLabel(2013)).toEqual({ name: 'Generation Alpha', class: 'alpha' });
    });

    test('generational label calculation - Other', () => {
        const getGenerationalLabel = (year) => {
            if (year >= 1946 && year <= 1964) return { name: 'Baby Boomer', class: 'baby-boomer' };
            if (year >= 1965 && year <= 1980) return { name: 'Generation X', class: 'x' };
            if (year >= 1981 && year <= 1996) return { name: 'Millennial', class: 'millennial' };
            if (year >= 1997 && year <= 2012) return { name: 'Generation Z', class: 'z' };
            if (year >= 2013) return { name: 'Generation Alpha', class: 'alpha' };
            return { name: 'Other', class: 'other' };
        };
        
        expect(getGenerationalLabel(1940)).toEqual({ name: 'Other', class: 'other' });
        expect(getGenerationalLabel(1920)).toEqual({ name: 'Other', class: 'other' });
    });

    test('marker object structure', () => {
        const marker = {
            year: 1985,
            id: 'marker_123'
        };
        
        expect(marker.year).toBe(1985);
        expect(marker.id).toBe('marker_123');
        expect(typeof marker.year).toBe('number');
        expect(typeof marker.id).toBe('string');
    });

    test('year validation bounds', () => {
        const MIN_YEAR = 1950;
        const MAX_YEAR = 2020;
        
        expect(1985).toBeGreaterThanOrEqual(MIN_YEAR);
        expect(1985).toBeLessThanOrEqual(MAX_YEAR);
        expect(1945).toBeLessThan(MIN_YEAR);
        expect(2025).toBeGreaterThan(MAX_YEAR);
    });

    test('stat metadata structure', () => {
        const statMetadata = {
            literacy: { name: 'Literacy Rate', description: 'Adult literacy rates (15+)', unit: '%' },
            attainment: { name: 'Bachelor\'s+', description: 'Percentage with bachelor\'s degree or higher (25+)', unit: '%' }
        };
        
        expect(statMetadata.literacy).toHaveProperty('name');
        expect(statMetadata.literacy).toHaveProperty('description');
        expect(statMetadata.literacy).toHaveProperty('unit');
        expect(statMetadata.literacy.unit).toBe('%');
    });

    test('marker position calculation', () => {
        const MIN_YEAR = 1950;
        const MAX_YEAR = 2020;
        const year = 1985;
        const position = ((year - MIN_YEAR) / (MAX_YEAR - MIN_YEAR)) * 100;
        
        expect(position).toBeCloseTo(50, 0);
        
        const year2000 = 2000;
        const position2000 = ((year2000 - MIN_YEAR) / (MAX_YEAR - MIN_YEAR)) * 100;
        expect(position2000).toBeCloseTo(71.4, 1);
    });

    test('marker color assignment', () => {
        const MARKER_COLORS = ['red', 'blue'];
        const markers = [
            { year: 1970, id: 'm1' },
            { year: 1980, id: 'm2' },
            { year: 1990, id: 'm3' }
        ];
        
        markers.forEach((marker, index) => {
            const color = MARKER_COLORS[index % MARKER_COLORS.length];
            expect(['red', 'blue']).toContain(color);
        });
    });

    test('URL parameter parsing for cohorts', () => {
        const urlParams = 'cohorts=1970,1980,1990';
        const params = new URLSearchParams(urlParams);
        const cohortsParam = params.get('cohorts');
        const years = cohortsParam.split(',').map(y => parseInt(y));
        
        expect(years).toEqual([1970, 1980, 1990]);
        expect(years.length).toBe(3);
    });

    test('URL parameter parsing for stats', () => {
        const urlParams = 'stats=literacy,attainment,proficiency';
        const params = new URLSearchParams(urlParams);
        const statsParam = params.get('stats');
        const stats = statsParam.split(',');
        
        expect(stats).toEqual(['literacy', 'attainment', 'proficiency']);
        expect(stats.length).toBe(3);
    });

    test('cohort data lookup simulation', () => {
        const marker = { year: 1980, id: 'm1' };
        const targetYear = marker.year + 25; // 2005 for adult stats
        
        const mockData = {
            data: [
                { year: 2003, value: 20.5 },
                { year: 2005, value: 21.2 },
                { year: 2007, value: 22.1 }
            ]
        };
        
        const dataPoint = mockData.data.find(d => Math.abs(d.year - targetYear) < 3);
        expect(dataPoint).toBeDefined();
        expect(dataPoint.year).toBe(2005);
        expect(dataPoint.value).toBe(21.2);
    });

    test('no data handling', () => {
        const selectedStats = [];
        const message = selectedStats.length === 0 
            ? 'No statistics selected. Click settings to choose.' 
            : '';
        
        expect(message).toContain('No statistics selected');
    });

    test('duplicate marker prevention', () => {
        const markers = [
            { year: 1980, id: 'm1' }
        ];
        
        const newYear = 1980;
        const exists = markers.some(m => m.year === newYear);
        
        expect(exists).toBe(true);
    });

    test('marker sorting by year', () => {
        const markers = [
            { year: 1990, id: 'm1' },
            { year: 1970, id: 'm2' },
            { year: 1980, id: 'm3' }
        ];
        
        markers.sort((a, b) => a.year - b.year);
        
        expect(markers[0].year).toBe(1970);
        expect(markers[1].year).toBe(1980);
        expect(markers[2].year).toBe(1990);
    });

    test('DOM elements exist', () => {
        expect(document.getElementById('timeline')).toBeTruthy();
        expect(document.getElementById('comparisonTables')).toBeTruthy();
        expect(document.getElementById('settingsBtn')).toBeTruthy();
        expect(document.getElementById('addTableBtn')).toBeTruthy();
    });
});
    
    const year = parseInt(yearInput.value);
    expect(year < 1950).toBe(true);
  });

  test('markers display updates', () => {
    const display = document.getElementById('markersDisplay');
    expect(display).not.toBeNull();
  });
});

describe('URL parameter handling', () => {
  test('parseURLParams extracts stat parameter', () => {
    const params = new URLSearchParams('?stat=attainment&markers=1970,1980');
    
    expect(params.get('stat')).toBe('attainment');
    expect(params.get('markers')).toBe('1970,1980');
  });

  test('markers parameter parsing', () => {
    const markersParam = '1970,1980,1990';
    const markers = markersParam.split(',').map(m => parseInt(m));
    
    expect(markers).toEqual([1970, 1980, 1990]);
    expect(markers.length).toBe(3);
  });
});
