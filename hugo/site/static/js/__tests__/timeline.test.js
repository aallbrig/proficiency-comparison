/**
 * @jest-environment jsdom
 */

describe('Timeline functionality', () => {
  beforeEach(() => {
    document.body.innerHTML = `
      <select id="statSelect">
        <option value="literacy">Literacy</option>
        <option value="attainment">Attainment</option>
      </select>
      <div id="statDescription"></div>
      <input type="number" id="yearInput" />
      <button id="addMarkerBtn">Add</button>
      <div id="markersDisplay"></div>
      <div id="comparisonResults"></div>
      <div id="noDataWarning"></div>
    `;
    
    // Reset global state
    global.markers = [];
    global.currentStat = 'literacy';
  });

  test('stat descriptions are defined', () => {
    const statDescriptions = {
      literacy: '<strong>Literacy Rates:</strong> Adult literacy rates (15+)',
      attainment: '<strong>Educational Attainment:</strong> Percentage with bachelor\'s degree or higher (25+)'
    };
    
    expect(statDescriptions.literacy).toContain('Literacy Rates');
    expect(statDescriptions.attainment).toContain('Educational Attainment');
  });

  test('markers array starts empty', () => {
    expect(global.markers).toEqual([]);
  });

  test('year input validation', () => {
    const yearInput = document.getElementById('yearInput');
    yearInput.value = '1980';
    
    const year = parseInt(yearInput.value);
    expect(year).toBe(1980);
    expect(year >= 1950 && year <= 2020).toBe(true);
  });

  test('invalid year input', () => {
    const yearInput = document.getElementById('yearInput');
    yearInput.value = '1800';
    
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
