// Timeline functionality for cohort comparison

let currentStat = 'attainment'; // Default to attainment which usually has data
let markers = [];
let statData = {};

// Load stat descriptions
const statDescriptions = {
    literacy: '<strong>Literacy Rates:</strong> Adult literacy rates (15+)',
    attainment: '<strong>Educational Attainment:</strong> Percentage with bachelor\'s degree or higher (25+)',
    graduation: '<strong>High School Graduation:</strong> Percentage graduating from high school',
    enrollment: '<strong>Enrollment Rates:</strong> School enrollment rates by level',
    proficiency: '<strong>Test Proficiency:</strong> NAEP Reading scores (Grade 8)',
    early_childhood: '<strong>Early Childhood Metrics:</strong> Early literacy and readiness indicators'
};

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    loadStatsIndex();
    setupEventListeners();
    loadURLParams();
});

function setupEventListeners() {
    const statSelect = document.getElementById('statSelect');
    const addMarkerBtn = document.getElementById('addMarkerBtn');
    const yearInput = document.getElementById('yearInput');
    
    if (statSelect) {
        statSelect.addEventListener('change', function() {
            currentStat = this.value;
            updateStatDescription();
            loadStatData(currentStat);
            updateURL();
        });
    }
    
    if (addMarkerBtn) {
        addMarkerBtn.addEventListener('click', addMarker);
    }
    
    if (yearInput) {
        yearInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                addMarker();
            }
        });
    }
}

function loadStatsIndex() {
    fetch('/data/stats_index.json')
        .then(response => {
            if (!response.ok) {
                showNoDataWarning();
                return null;
            }
            return response.json();
        })
        .then(data => {
            if (data) {
                // Try to find a stat with actual data
                tryLoadAvailableStat();
            }
        })
        .catch(error => {
            console.error('Error loading stats index:', error);
            showNoDataWarning();
        });
}

function tryLoadAvailableStat() {
    // Try stats in order of priority
    const statsToTry = ['attainment', 'proficiency', 'literacy', 'graduation', 'enrollment', 'early_childhood'];
    
    function tryNext(index) {
        if (index >= statsToTry.length) {
            showNoDataWarning();
            return;
        }
        
        const stat = statsToTry[index];
        fetch(`/data/${stat}.json`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Not found');
                }
                return response.json();
            })
            .then(data => {
                if (data && data.data && data.data.length > 0) {
                    // Found data! Use this stat
                    currentStat = stat;
                    const statSelect = document.getElementById('statSelect');
                    if (statSelect) {
                        statSelect.value = stat;
                    }
                    updateStatDescription();
                    // Store the data and hide warning
                    statData[stat] = data;
                    hideNoDataWarning();
                    updateComparison();
                } else {
                    // No data, try next
                    tryNext(index + 1);
                }
            })
            .catch(() => {
                // Failed to load, try next
                tryNext(index + 1);
            });
    }
    
    tryNext(0);
}

function loadStatData(stat) {
    fetch(`/data/${stat}.json`)
        .then(response => {
            if (!response.ok) {
                throw new Error(`Failed to load ${stat} data`);
            }
            return response.json();
        })
        .then(data => {
            statData[stat] = data;
            hideNoDataWarning();
            updateComparison();
        })
        .catch(error => {
            console.error(`Error loading ${stat} data:`, error);
            showNoDataWarning();
        });
}

function updateStatDescription() {
    const descEl = document.getElementById('statDescription');
    if (descEl) {
        descEl.innerHTML = statDescriptions[currentStat];
    }
}

function addMarker() {
    const yearInput = document.getElementById('yearInput');
    const year = parseInt(yearInput.value);
    
    if (isNaN(year) || year < 1950 || year > 2020) {
        alert('Please enter a valid year between 1950 and 2020');
        return;
    }
    
    if (markers.includes(year)) {
        alert('This marker already exists');
        return;
    }
    
    markers.push(year);
    markers.sort((a, b) => a - b);
    yearInput.value = '';
    
    updateMarkersDisplay();
    updateComparison();
    updateURL();
}

function removeMarker(year) {
    markers = markers.filter(m => m !== year);
    updateMarkersDisplay();
    updateComparison();
    updateURL();
}

function updateMarkersDisplay() {
    const display = document.getElementById('markersDisplay');
    if (!display) return;
    
    if (markers.length === 0) {
        display.innerHTML = '<span class="text-muted">No markers added</span>';
        return;
    }
    
    display.innerHTML = markers.map(year => `
        <span class="badge bg-primary" style="cursor: pointer;" onclick="removeMarker(${year})">
            ${year} ✕
        </span>
    `).join('');
}

function updateComparison() {
    const resultsDiv = document.getElementById('comparisonResults');
    if (!resultsDiv) return;
    
    if (markers.length === 0) {
        resultsDiv.innerHTML = '<p class="text-muted">Add markers to see comparisons</p>';
        return;
    }
    
    const data = statData[currentStat];
    if (!data || !data.data) {
        resultsDiv.innerHTML = '<p class="text-warning">No data available for this statistic</p>';
        return;
    }
    
    // Build comparison table
    let html = '<table class="table table-bordered mt-3"><thead><tr><th>Birth Year</th>';
    
    // Add column headers for ages (example: 15, 18, 25)
    html += '<th>Current Age</th><th>Value</th></tr></thead><tbody>';
    
    markers.forEach(year => {
        const currentAge = new Date().getFullYear() - year;
        
        // Find appropriate data point (simplified - would need cohort mapping logic)
        const dataPoint = data.data.find(d => Math.abs(d.year - (year + 25)) < 3);
        const value = dataPoint ? dataPoint.value.toFixed(2) : 'N/A';
        
        html += `<tr>
            <td><strong>${year}</strong></td>
            <td>${currentAge} years old</td>
            <td>${value}</td>
        </tr>`;
    });
    
    html += '</tbody></table>';
    html += `<p class="small text-muted mt-2"><strong>Source:</strong> ${data.source}</p>`;
    
    resultsDiv.innerHTML = html;
}

function showNoDataWarning() {
    const warning = document.getElementById('noDataWarning');
    if (warning) {
        warning.style.display = 'block';
    }
}

function hideNoDataWarning() {
    const warning = document.getElementById('noDataWarning');
    if (warning) {
        warning.style.display = 'none';
    }
}

function updateURL() {
    const params = new URLSearchParams();
    params.set('stat', currentStat);
    if (markers.length > 0) {
        params.set('markers', markers.join(','));
    }
    const newURL = `${window.location.pathname}?${params.toString()}`;
    window.history.replaceState({}, '', newURL);
}

function loadURLParams() {
    const params = new URLSearchParams(window.location.search);
    
    const stat = params.get('stat');
    if (stat) {
        currentStat = stat;
        const statSelect = document.getElementById('statSelect');
        if (statSelect) {
            statSelect.value = stat;
        }
        updateStatDescription();
    }
    
    const markersParam = params.get('markers');
    if (markersParam) {
        markers = markersParam.split(',').map(m => parseInt(m)).filter(m => !isNaN(m));
        updateMarkersDisplay();
        updateComparison();
    }
}
