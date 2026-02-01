// Data visualization charts - Multi-section layout

const charts = {};
const datasets = ['literacy', 'attainment', 'graduation', 'enrollment', 'proficiency', 'early_childhood'];

document.addEventListener('DOMContentLoaded', function() {
    // Load all datasets and render charts
    datasets.forEach(stat => {
        loadAndRenderChart(stat);
    });
});

async function loadAndRenderChart(statName) {
    try {
        const response = await fetch(`/data/${statName}.json`);
        if (!response.ok) {
            console.warn(`Could not load ${statName}.json`);
            return;
        }
        
        const statData = await response.json();
        const chartId = `${statName}-chart`;
        const tableId = `${statName}-table`;
        
        // Render chart
        renderChart(chartId, statData, statName);
        
        // Populate table
        populateTable(tableId, statData);
        
    } catch (error) {
        console.error(`Error loading ${statName}:`, error);
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
            plugins: {
                title: {
                    display: false
                },
                legend: {
                    display: false
                },
                tooltip: {
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

                        display: true,
                        text: 'Value'
                    }
                },
                x: {
                    title: {
                        display: true,
                        text: 'Year'
                    }
                }
            }
        }
    });
}

function loadDataForChart(stat) {
    fetch(`/data/${stat}.json`)
        .then(response => {
            if (!response.ok) {
                throw new Error(`Failed to load ${stat} data`);
            }
            return response.json();
        })
        .then(data => {
            updateChart(data);
            updateDataTable(data);
        })
        .catch(error => {
            console.error(`Error loading ${stat} data:`, error);
            showError('Unable to load data. Please ensure the CLI tool has been run to download data.');
        });
}

function updateChart(data) {
    if (!chart || !data || !data.data) return;
    
    const years = data.data.map(d => d.year);
    const values = data.data.map(d => d.value);
    
    chart.data.labels = years;
    chart.data.datasets[0].data = values;
    chart.data.datasets[0].label = data.name;
    chart.options.plugins.title.text = `${data.name} Over Time`;
    chart.update();
}

function updateDataTable(data) {
    const tbody = document.querySelector('#dataTable tbody');
    if (!tbody || !data || !data.data) return;
    
    tbody.innerHTML = '';
    
    data.data.forEach(point => {
        const row = tbody.insertRow();
        row.insertCell(0).textContent = point.year;
        row.insertCell(1).textContent = point.value.toFixed(2);
    });
}

function showError(message) {
    const chartContainer = document.querySelector('#dataChart').parentElement;
    if (chartContainer) {
        chartContainer.innerHTML = `
            <div class="alert alert-warning" role="alert">
                <h4 class="alert-heading">Data Not Available</h4>
                <p>${message}</p>
                <hr>
                <p class="mb-0">Run: <code>edu-stats all --years=1970-2025</code></p>
            </div>
        `;
    }
}
