// Timeline functionality for cohort comparison with generational labels
// State management
let markers = []; // Array of {year: number, id: string}
let selectedStats = ['literacy', 'attainment', 'proficiency', 'graduation']; // Default enabled stats
let statData = {}; // Cached stat data
let availableStats = []; // Stats with actual data
let dragState = null; // Current drag operation

// Constants
// Timeline configuration - spans full historical range
const MIN_YEAR = 1870;  // Start of historical education data
const MAX_YEAR = 2025;  // Current year + projection
const MARKER_COLORS = ['red', 'blue'];

// Stat metadata
const statMetadata = {
    literacy: { name: 'Literacy Rate', description: 'Adult literacy rates (15+)', unit: '%' },
    attainment: { name: 'Bachelor\'s+', description: 'Percentage with bachelor\'s degree or higher (25+)', unit: '%' },
    graduation: { name: 'HS Graduation', description: 'High school graduation rate', unit: '%' },
    enrollment: { name: 'Enrollment', description: 'School enrollment rates', unit: '%' },
    proficiency: { name: 'Test Proficiency', description: 'NAEP Reading scores (Grade 8)', unit: 'pts' },
    early_childhood: { name: 'Early Childhood', description: 'Early literacy and readiness', unit: 'score' }
};

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
});

async function initializeApp() {
    await loadAvailableStats();
    setupEventListeners();
    loadURLParams();
    renderTimeline();
    renderComparisonTables();
    updateNoTablesMessage();
}

// Load and detect available stats
async function loadAvailableStats() {
    const statsToTry = ['literacy', 'attainment', 'proficiency', 'graduation', 'enrollment', 'early_childhood'];
    
    for (const stat of statsToTry) {
        try {
            const response = await fetch(`/data/${stat}.json`);
            if (response.ok) {
                const data = await response.json();
                if (data && data.data && data.data.length > 0) {
                    statData[stat] = data;
                    availableStats.push(stat);
                }
            }
        } catch (error) {
            console.log(`Stat ${stat} not available`);
        }
    }
    
    if (availableStats.length === 0) {
        showNoDataWarning();
    } else {
        hideNoDataWarning();
        // Filter selected stats to only available ones
        selectedStats = selectedStats.filter(s => availableStats.includes(s));
        if (selectedStats.length === 0) {
            selectedStats = [availableStats[0]];
        }
    }
}

// Event listeners setup
function setupEventListeners() {
    // Settings button
    const settingsBtn = document.getElementById('settingsBtn');
    if (settingsBtn) {
        settingsBtn.addEventListener('click', openSettings);
    }
    
    // Add table button
    const addTableBtn = document.getElementById('addTableBtn');
    if (addTableBtn) {
        addTableBtn.addEventListener('click', addNewComparison);
    }
    
    // Add marker from timeline button
    const addMarkerBtn = document.getElementById('addMarkerFromTimelineBtn');
    if (addMarkerBtn) {
        addMarkerBtn.addEventListener('click', addMarkerFromInput);
    }
    
    // Timeline click to add marker
    const timeline = document.getElementById('timeline');
    if (timeline) {
        timeline.addEventListener('click', handleTimelineClick);
    }
    
    // Save settings
    const saveSettingsBtn = document.getElementById('saveSettingsBtn');
    if (saveSettingsBtn) {
        saveSettingsBtn.addEventListener('click', saveSettings);
    }
}

// Generational label calculation
function getGenerationalLabel(year) {
    if (year >= 1946 && year <= 1964) return { name: 'Baby Boomer', class: 'baby-boomer' };
    if (year >= 1965 && year <= 1980) return { name: 'Generation X', class: 'x' };
    if (year >= 1981 && year <= 1996) return { name: 'Millennial', class: 'millennial' };
    if (year >= 1997 && year <= 2012) return { name: 'Generation Z', class: 'z' };
    if (year >= 2013) return { name: 'Generation Alpha', class: 'alpha' };
    return { name: 'Other', class: 'other' };
}

// Generate unique ID
function generateId() {
    return 'marker_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
}

// Add new comparison (defaults to middle of range)
function addNewComparison() {
    const newYear = Math.floor((MIN_YEAR + MAX_YEAR) / 2);
    addMarker(newYear);
}

// Add marker from input
function addMarkerFromInput() {
    const year = prompt(`Enter birth year (${MIN_YEAR}-${MAX_YEAR}):`);
    if (year) {
        const yearNum = parseInt(year);
        if (yearNum >= MIN_YEAR && yearNum <= MAX_YEAR) {
            addMarker(yearNum);
        } else {
            alert(`Please enter a year between ${MIN_YEAR} and ${MAX_YEAR}`);
        }
    }
}

// Add marker
function addMarker(year) {
    // Check if year already exists
    if (markers.some(m => m.year === year)) {
        alert(`A marker for ${year} already exists`);
        return;
    }
    
    const marker = {
        year: year,
        id: generateId()
    };
    
    markers.push(marker);
    markers.sort((a, b) => a.year - b.year);
    
    renderTimeline();
    renderComparisonTables();
    updateNoTablesMessage();
    updateURL();
}

// Remove marker
function removeMarker(id) {
    markers = markers.filter(m => m.id !== id);
    renderTimeline();
    renderComparisonTables();
    updateNoTablesMessage();
    updateURL();
}

// Update marker year
function updateMarkerYear(id, newYear) {
    const marker = markers.find(m => m.id === id);
    if (marker) {
        marker.year = Math.max(MIN_YEAR, Math.min(MAX_YEAR, newYear));
        markers.sort((a, b) => a.year - b.year);
        renderTimeline();
        renderComparisonTables();
        updateURL();
    }
}

// Highlight marker and table
function highlightMarker(id) {
    // Remove all highlights
    document.querySelectorAll('.timeline-marker').forEach(m => m.classList.remove('highlighted'));
    document.querySelectorAll('.comparison-table-card').forEach(t => t.classList.remove('highlighted'));
    
    // Add highlight
    const markerEl = document.querySelector(`.timeline-marker[data-id="${id}"]`);
    const tableEl = document.querySelector(`.comparison-table-card[data-id="${id}"]`);
    
    if (markerEl) markerEl.classList.add('highlighted');
    if (tableEl) {
        tableEl.classList.add('highlighted');
        tableEl.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }
}

// Render timeline
function renderTimeline() {
    const timeline = document.getElementById('timeline');
    const markersContainer = document.getElementById('timelineMarkers');
    const labelsContainer = document.getElementById('timelineLabels');
    
    if (!timeline || !markersContainer || !labelsContainer) return;
    
    // Render year labels
    labelsContainer.innerHTML = `
        <span>${MIN_YEAR}</span>
        <span>${Math.floor((MIN_YEAR + MAX_YEAR) / 2)}</span>
        <span>${MAX_YEAR}</span>
    `;
    
    // Render markers
    markersContainer.innerHTML = '';
    markers.forEach((marker, index) => {
        const position = ((marker.year - MIN_YEAR) / (MAX_YEAR - MIN_YEAR)) * 100;
        const color = MARKER_COLORS[index % MARKER_COLORS.length];
        
        const markerEl = document.createElement('div');
        markerEl.className = `timeline-marker ${color}`;
        markerEl.dataset.id = marker.id;
        markerEl.dataset.year = marker.year;
        markerEl.style.left = `${position}%`;
        markerEl.textContent = String(marker.year).slice(-2);
        markerEl.title = `Birth Year: ${marker.year}`;
        
        markerEl.addEventListener('click', (e) => {
            e.stopPropagation();
            highlightMarker(marker.id);
        });
        
        markerEl.addEventListener('mousedown', (e) => startDrag(e, marker.id));
        
        markersContainer.appendChild(markerEl);
    });
    
    // Update year range display
    const yearRangeEl = document.getElementById('yearRange');
    if (yearRangeEl) {
        yearRangeEl.textContent = `${MIN_YEAR}-${MAX_YEAR}`;
    }
}

// Drag and drop for markers
function startDrag(e, markerId) {
    e.preventDefault();
    const timeline = document.getElementById('timeline');
    const markerEl = e.target;
    
    markerEl.classList.add('dragging');
    
    dragState = {
        markerId: markerId,
        timelineRect: timeline.getBoundingClientRect()
    };
    
    document.addEventListener('mousemove', handleDrag);
    document.addEventListener('mouseup', endDrag);
}

function handleDrag(e) {
    if (!dragState) return;
    
    const { markerId, timelineRect } = dragState;
    const x = e.clientX - timelineRect.left;
    const percent = Math.max(0, Math.min(100, (x / timelineRect.width) * 100));
    const year = Math.round(MIN_YEAR + (percent / 100) * (MAX_YEAR - MIN_YEAR));
    
    // Update marker position visually
    const markerEl = document.querySelector(`.timeline-marker[data-id="${markerId}"]`);
    if (markerEl) {
        markerEl.style.left = `${percent}%`;
        markerEl.dataset.year = year;
        markerEl.textContent = String(year).slice(-2);
    }
}

function endDrag(e) {
    if (!dragState) return;
    
    const { markerId, timelineRect } = dragState;
    const x = e.clientX - timelineRect.left;
    const percent = Math.max(0, Math.min(100, (x / timelineRect.width) * 100));
    const year = Math.round(MIN_YEAR + (percent / 100) * (MAX_YEAR - MIN_YEAR));
    
    updateMarkerYear(markerId, year);
    
    const markerEl = document.querySelector(`.timeline-marker[data-id="${markerId}"]`);
    if (markerEl) {
        markerEl.classList.remove('dragging');
    }
    
    document.removeEventListener('mousemove', handleDrag);
    document.removeEventListener('mouseup', endDrag);
    dragState = null;
}

// Handle timeline click to add marker
function handleTimelineClick(e) {
    if (dragState) return; // Don't add during drag
    
    const timeline = e.currentTarget;
    const rect = timeline.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const percent = (x / rect.width) * 100;
    const year = Math.round(MIN_YEAR + (percent / 100) * (MAX_YEAR - MIN_YEAR));
    
    addMarker(year);
}

// Render comparison tables
function renderComparisonTables() {
    const container = document.getElementById('comparisonTables');
    if (!container) return;
    
    container.innerHTML = '';
    
    markers.forEach((marker, index) => {
        const generation = getGenerationalLabel(marker.year);
        const color = MARKER_COLORS[index % MARKER_COLORS.length];
        
        const card = document.createElement('div');
        card.className = 'col-md-6 col-lg-4';
        card.innerHTML = `
            <div class="comparison-table-card" data-id="${marker.id}">
                <div class="card-header">
                    <div>
                        <label class="small text-muted d-block mb-1">Birth Year</label>
                        <input type="number" 
                               class="birth-year-input" 
                               value="${marker.year}" 
                               min="${MIN_YEAR}" 
                               max="${MAX_YEAR}"
                               data-id="${marker.id}">
                    </div>
                    <div class="text-end">
                        <span class="generation-badge generation-${generation.class}">${generation.name}</span>
                        <button class="btn btn-sm btn-outline-danger ms-2" onclick="removeMarker('${marker.id}')">
                            <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" fill="currentColor" viewBox="0 0 16 16">
                                <path d="M2.146 2.854a.5.5 0 1 1 .708-.708L8 7.293l5.146-5.147a.5.5 0 0 1 .708.708L8.707 8l5.147 5.146a.5.5 0 0 1-.708.708L8 8.707l-5.146 5.147a.5.5 0 0 1-.708-.708L7.293 8 2.146 2.854Z"/>
                            </svg>
                        </button>
                    </div>
                </div>
                <div class="card-body">
                    <div class="timeline-marker ${color} d-inline-flex mb-3" style="position: static; transform: none;">
                        ${String(marker.year).slice(-2)}
                    </div>
                    ${renderStatsForMarker(marker)}
                </div>
            </div>
        `;
        
        container.appendChild(card);
        
        // Add event listener for birth year input
        const input = card.querySelector('.birth-year-input');
        input.addEventListener('change', (e) => {
            const newYear = parseInt(e.target.value);
            if (newYear >= MIN_YEAR && newYear <= MAX_YEAR) {
                updateMarkerYear(marker.id, newYear);
            } else {
                e.target.value = marker.year;
            }
        });
        
        // Click to highlight
        card.querySelector('.comparison-table-card').addEventListener('click', () => {
            highlightMarker(marker.id);
        });
    });
}

// Render stats for a specific marker
function renderStatsForMarker(marker) {
    if (selectedStats.length === 0) {
        return '<p class="text-muted small">No statistics selected. Click settings to choose.</p>';
    }
    
    let html = '';
    selectedStats.forEach(stat => {
        const data = statData[stat];
        const metadata = statMetadata[stat];
        
        if (!data || !metadata) return;
        
        // Find appropriate data point (simplified cohort mapping)
        // For adult stats (25+), look at year + 25
        const targetYear = marker.year + 25;
        const dataPoint = data.data.find(d => Math.abs(d.year - targetYear) < 3);
        const value = dataPoint ? `${dataPoint.value.toFixed(1)}${metadata.unit}` : 'N/A';
        
        html += `
            <div class="stat-row">
                <span class="stat-label" title="${metadata.description}">${metadata.name}</span>
                <span class="stat-value">${value}</span>
            </div>
        `;
    });
    
    return html || '<p class="text-muted small">No data available</p>';
}

// Update no tables message
function updateNoTablesMessage() {
    const message = document.getElementById('noTablesMessage');
    if (message) {
        message.style.display = markers.length === 0 ? 'block' : 'none';
    }
}

// Settings modal
function openSettings() {
    const modal = new bootstrap.Modal(document.getElementById('settingsModal'));
    renderSettingsCheckboxes();
    modal.show();
}

function renderSettingsCheckboxes() {
    const container = document.getElementById('statCheckboxes');
    if (!container) return;
    
    container.innerHTML = '';
    
    availableStats.forEach(stat => {
        const metadata = statMetadata[stat];
        const checked = selectedStats.includes(stat);
        
        const div = document.createElement('div');
        div.className = 'form-check';
        div.innerHTML = `
            <input class="form-check-input" type="checkbox" value="${stat}" id="stat_${stat}" ${checked ? 'checked' : ''}>
            <label class="form-check-label" for="stat_${stat}">
                <strong>${metadata.name}</strong><br>
                <small class="text-muted">${metadata.description}</small>
            </label>
        `;
        container.appendChild(div);
    });
}

function saveSettings() {
    const checkboxes = document.querySelectorAll('#statCheckboxes input[type="checkbox"]');
    selectedStats = Array.from(checkboxes)
        .filter(cb => cb.checked)
        .map(cb => cb.value);
    
    renderComparisonTables();
    updateURL();
    
    const modal = bootstrap.Modal.getInstance(document.getElementById('settingsModal'));
    modal.hide();
}

// URL management
function updateURL() {
    const params = new URLSearchParams();
    
    if (markers.length > 0) {
        params.set('cohorts', markers.map(m => m.year).join(','));
    }
    
    if (selectedStats.length > 0) {
        params.set('stats', selectedStats.join(','));
    }
    
    const newURL = params.toString() ? `${window.location.pathname}?${params.toString()}` : window.location.pathname;
    window.history.replaceState({}, '', newURL);
    
    // Update QR code
    updateQRCode();
}

function loadURLParams() {
    const params = new URLSearchParams(window.location.search);
    
    const cohortsParam = params.get('cohorts');
    if (cohortsParam) {
        const years = cohortsParam.split(',').map(y => parseInt(y)).filter(y => !isNaN(y) && y >= MIN_YEAR && y <= MAX_YEAR);
        years.forEach(year => {
            markers.push({ year, id: generateId() });
        });
        markers.sort((a, b) => a.year - b.year);
    }
    
    const statsParam = params.get('stats');
    if (statsParam) {
        const stats = statsParam.split(',').filter(s => availableStats.includes(s));
        if (stats.length > 0) {
            selectedStats = stats;
        }
    }
}

function updateQRCode() {
    const qrContainer = document.getElementById('qrcode');
    if (qrContainer && typeof QRCode !== 'undefined') {
        qrContainer.innerHTML = '';
        new QRCode(qrContainer, {
            text: window.location.href,
            width: 128,
            height: 128
        });
    }
}

// No data warning
function showNoDataWarning() {
    const warning = document.getElementById('noDataWarning');
    if (warning) warning.style.display = 'block';
}

function hideNoDataWarning() {
    const warning = document.getElementById('noDataWarning');
    if (warning) warning.style.display = 'none';
}

// Expose functions to global scope for inline event handlers
window.removeMarker = removeMarker;
window.highlightMarker = highlightMarker;
