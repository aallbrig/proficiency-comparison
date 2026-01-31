// Data visualization charts

let chart = null;
let currentDataStat = 'literacy';

document.addEventListener('DOMContentLoaded', function() {
    setupDataChart();
    loadDataForChart(currentDataStat);
    
    const select = document.getElementById('dataStatSelect');
    if (select) {
        select.addEventListener('change', function() {
            currentDataStat = this.value;
            loadDataForChart(currentDataStat);
        });
    }
});

function setupDataChart() {
    const ctx = document.getElementById('dataChart');
    if (!ctx) return;
    
    chart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                label: 'Value',
                data: [],
                borderColor: 'rgb(75, 192, 192)',
                backgroundColor: 'rgba(75, 192, 192, 0.2)',
                tension: 0.1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                title: {
                    display: true,
                    text: 'Educational Statistics Over Time'
                },
                legend: {
                    display: true
                }
            },
            scales: {
                y: {
                    beginAtZero: false,
                    title: {
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
