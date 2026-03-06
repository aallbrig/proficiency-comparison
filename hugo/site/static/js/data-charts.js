// Data visualization charts - Multi-section layout

const charts = {};
let availableStats = {}; // Will be populated from index.json
const dataRanges = {}; // Store year ranges for each dataset

document.addEventListener('DOMContentLoaded', function() {
    // Load index first, then load individual stat files
    loadIndexAndRenderCharts();
});

async function loadIndexAndRenderCharts() {
    try {
        const response = await fetch('/data/index.json');
        if (!response.ok) {
            console.warn('Could not load index.json, falling back to individual file checks');
            loadChartsLegacyWay();
            return;
        }
        
        const indexData = await response.json();
        availableStats = indexData.stats || {};
        
        console.log('Loaded index:', Object.keys(availableStats).length, 'stats');
        
        // Generate sections dynamically
        const container = document.getElementById('dataSections');
        if (container) {
            for (const [statName, statInfo] of Object.entries(availableStats)) {
                if (statInfo.available) {
                    createStatSection(container, statName, statInfo);
                    await loadAndRenderChart(statName, statInfo);
                }
            }
        }
        
        // Update summary after all data loads
        updateDataSummary();
        
    } catch (error) {
        console.error('Error loading index:', error);
        loadChartsLegacyWay();
    }
}

function createStatSection(container, statName, statInfo) {
    const colors = {
        literacy: 'primary',
        attainment: 'success',
        graduation: 'info',
        enrollment: 'warning',
        proficiency: 'danger',
        early_childhood: 'secondary'
    };
    
    const icons = {
        literacy: 'book',
        attainment: 'mortarboard',
        graduation: 'award',
        enrollment: 'people',
        proficiency: 'clipboard-data',
        early_childhood: 'stars'
    };
    
    const color = colors[statName] || 'secondary';
    const icon = icons[statName] || 'graph-up';
    const yearRange = `(${statInfo.yearMin}-${statInfo.yearMax})`;
    
    const col = document.createElement('div');
    col.className = 'col';
    
    col.innerHTML = `
        <div class="card shadow-sm h-100" id="${statName}">
            <div class="card-header bg-${color} text-white">
                <h5 class="mb-0">
                    <i class="bi bi-${icon}"></i> ${statInfo.name}
                    <small class="d-block mt-1" style="font-size: 0.85rem;">${yearRange}</small>
                </h5>
            </div>
            <div class="card-body">
                <p class="small text-muted">${statInfo.description}</p>
                <div style="height: 250px;">
                    <canvas id="${statName}-chart"></canvas>
                </div>
                <div class="mt-3">
                    <button class="btn btn-sm btn-outline-secondary" type="button" data-bs-toggle="collapse" data-bs-target="#${statName}-data">
                        <i class="bi bi-table"></i> Show Data (${statInfo.dataPoints} points)
                    </button>
                    <div class="collapse mt-2" id="${statName}-data">
                        <div class="table-responsive" style="max-height: 300px; overflow-y: auto;">
                            <table class="table table-sm table-striped" id="${statName}-table">
                                <thead class="sticky-top bg-light">
                                    <tr><th>Year</th><th>Value</th></tr>
                                </thead>
                                <tbody></tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `;
    
    container.appendChild(col);
}

// Fallback method if index.json doesn't exist
function loadChartsLegacyWay() {
    const datasets = ['literacy', 'attainment', 'graduation', 'enrollment', 'proficiency', 'early_childhood'];
    datasets.forEach(stat => {
        loadAndRenderChart(stat, null);
    });
    setTimeout(updateDataSummary, 1000);
}

async function loadAndRenderChart(statName, statInfo) {
    try {
        const filename = statInfo ? statInfo.filename : `${statName}.json`;
        const response = await fetch(`/data/${filename}`);
        if (!response.ok) {
            console.warn(`Could not load ${filename}`);
            hideSection(statName);
            return;
        }
        
        const statData = await response.json();
        
        // Store year range (use from index if available, otherwise calculate)
        if (statInfo && statInfo.available) {
            dataRanges[statName] = {
                min: statInfo.yearMin,
                max: statInfo.yearMax,
                count: statInfo.dataPoints
            };
        } else if (statData.data && statData.data.length > 0) {
            const years = statData.data.map(d => d.year);
            dataRanges[statName] = {
                min: Math.min(...years),
                max: Math.max(...years),
                count: years.length
            };
        }
        
        // Update the section header with actual year range
        if (dataRanges[statName]) {
            updateSectionHeader(statName, dataRanges[statName]);
        }
        
        const chartId = `${statName}-chart`;
        const tableId = `${statName}-table`;
        
        // Render chart
        renderChart(chartId, statData, statName);
        
        // Populate table
        populateTable(tableId, statData);
        
    } catch (error) {
        console.error(`Error loading ${statName}:`, error);
        hideSection(statName);
    }
}

function updateSectionHeader(statName, range) {
    const headerElement = document.querySelector(`#${statName} .card-header h2`);
    if (headerElement) {
        // Update year range in header
        const yearRangeText = `(${range.min}-${range.max})`;
        const headerText = headerElement.textContent;
        
        // Replace any existing year range or add new one
        const updatedText = headerText.replace(/\(\d{4}-\d{4}\)/, yearRangeText);
        if (!updatedText.includes(yearRangeText)) {
            headerElement.textContent = headerText.trim() + ' ' + yearRangeText;
        } else {
            headerElement.textContent = updatedText;
        }
    }
}

function hideSection(statName) {
    const section = document.getElementById(statName);
    if (section) {
        section.style.display = 'none';
    }
}

function updateDataSummary() {
    // Update the lead paragraph with actual data range
    const leadPara = document.querySelector('.lead');
    if (leadPara && Object.keys(dataRanges).length > 0) {
        const allYears = [];
        let totalPoints = 0;
        let metricsWithData = 0;
        
        Object.values(dataRanges).forEach(range => {
            allYears.push(range.min, range.max);
            totalPoints += range.count;
            metricsWithData++;
        });
        
        const minYear = Math.min(...allYears);
        const maxYear = Math.max(...allYears);
        const yearSpan = maxYear - minYear;
        
        leadPara.textContent = `Explore ${yearSpan}+ years of US educational statistics with interactive charts. Data spans from ${minYear} to ${maxYear} across ${metricsWithData} key metrics with ${totalPoints} total data points.`;
    }
    
    // Update last updated timestamp
    const lastUpdated = document.getElementById('data-last-updated');
    if (lastUpdated) {
        const now = new Date();
        lastUpdated.textContent = `Page loaded ${now.toLocaleString()}`;
    }
}

function renderChart(canvasId, statData, statName) {
    const ctx = document.getElementById(canvasId);
    if (!ctx) return;
    
    const years = statData.data.map(d => d.year);
    const values = statData.data.map(d => d.value);
    
    // Color schemes for different stats
    const colors = {
        literacy: { border: 'rgb(13, 110, 253)', bg: 'rgba(13, 110, 253, 0.1)' },
        attainment: { border: 'rgb(25, 135, 84)', bg: 'rgba(25, 135, 84, 0.1)' },
        graduation: { border: 'rgb(13, 202, 240)', bg: 'rgba(13, 202, 240, 0.1)' },
        enrollment: { border: 'rgb(255, 193, 7)', bg: 'rgba(255, 193, 7, 0.1)' },
        proficiency: { border: 'rgb(220, 53, 69)', bg: 'rgba(220, 53, 69, 0.1)' },
        early_childhood: { border: 'rgb(108, 117, 125)', bg: 'rgba(108, 117, 125, 0.1)' }
    };
    
    const color = colors[statName] || { border: 'rgb(75, 192, 192)', bg: 'rgba(75, 192, 192, 0.2)' };
    
    charts[statName] = new Chart(ctx, {
        type: 'line',
        data: {
            labels: years,
            datasets: [{
                label: statData.name,
                data: values,
                borderColor: color.border,
                backgroundColor: color.bg,
                borderWidth: 2,
                tension: 0.4,
                pointRadius: years.length > 50 ? 0 : 3,
                pointHoverRadius: 5
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false,
            },
            plugins: {
                title: {
                    display: false
                },
                legend: {
                    display: false
                },
                tooltip: {
                    mode: 'index',
                    intersect: false,
                    callbacks: {
                        label: function(context) {
                            let label = context.dataset.label || '';
                            if (label) {
                                label += ': ';
                            }
                            label += context.parsed.y.toFixed(1);
                            if (statName !== 'proficiency' && statName !== 'early_childhood') {
                                label += '%';
                            }
                            return label;
                        }
                    }
                }
            },
            scales: {
                x: {
                    title: {
                        display: true,
                        text: 'Year'
                    },
                    ticks: {
                        maxTicksLimit: 10
                    }
                },
                y: {
                    beginAtZero: statName === 'graduation' || statName === 'enrollment',
                    title: {
                        display: true,
                        text: getYAxisLabel(statName)
                    },
                    ticks: {
                        callback: function(value) {
                            if (statName !== 'proficiency' && statName !== 'early_childhood') {
                                return value + '%';
                            }
                            return value;
                        }
                    }
                }
            },
            hover: {
                mode: 'index',
                intersect: false
            }
        }
    });
}

function populateTable(tableId, statData) {
    const table = document.getElementById(tableId);
    if (!table) return;
    
    const tbody = table.querySelector('tbody');
    tbody.innerHTML = '';
    
    statData.data.forEach(point => {
        const row = tbody.insertRow();
        row.insertCell(0).textContent = point.year;
        const valueCell = row.insertCell(1);
        valueCell.textContent = point.value.toFixed(1);
    });
}

function getYAxisLabel(statName) {
    const labels = {
        literacy: 'Literacy Rate (%)',
        attainment: 'Bachelor\'s Degree or Higher (%)',
        graduation: 'Graduation Rate (%)',
        enrollment: 'Enrollment Rate (%)',
        proficiency: 'Average Score (0-500 scale)',
        early_childhood: 'Average Score (0-100 scale)'
    };
    return labels[statName] || 'Value';
}

