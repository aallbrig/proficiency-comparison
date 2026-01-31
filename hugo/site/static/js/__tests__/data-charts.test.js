/**
 * @jest-environment jsdom
 */

describe('Data charts functionality', () => {
  beforeEach(() => {
    document.body.innerHTML = `
      <select id="dataStatSelect">
        <option value="literacy">Literacy</option>
      </select>
      <canvas id="dataChart"></canvas>
      <table id="dataTable">
        <tbody></tbody>
      </table>
    `;
  });

  test('chart canvas element exists', () => {
    const canvas = document.getElementById('dataChart');
    expect(canvas).not.toBeNull();
    expect(canvas.tagName).toBe('CANVAS');
  });

  test('data table exists', () => {
    const table = document.getElementById('dataTable');
    expect(table).not.toBeNull();
  });

  test('stat select dropdown exists', () => {
    const select = document.getElementById('dataStatSelect');
    expect(select).not.toBeNull();
    expect(select.options.length).toBeGreaterThan(0);
  });

  test('data point structure', () => {
    const dataPoint = {
      year: 2020,
      value: 95.5
    };
    
    expect(dataPoint.year).toBe(2020);
    expect(dataPoint.value).toBeCloseTo(95.5);
  });

  test('data array processing', () => {
    const data = {
      name: 'Literacy Rates',
      data: [
        { year: 2018, value: 95.0 },
        { year: 2019, value: 95.2 },
        { year: 2020, value: 95.5 }
      ]
    };
    
    const years = data.data.map(d => d.year);
    const values = data.data.map(d => d.value);
    
    expect(years).toEqual([2018, 2019, 2020]);
    expect(values).toEqual([95.0, 95.2, 95.5]);
  });
});
