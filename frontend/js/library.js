document.addEventListener('DOMContentLoaded', () => {
    console.log('Library page loaded');
    initAuthButton();
    checkAuthAndLoadLibrary();
    initImportForm();
});

function checkAuthAndLoadLibrary() {
    const token = getAccessToken();
    if (!token) {
        // Redirect to login if not authenticated
        window.location.href = 'login.html';
        return;
    }
    loadLibraryStats();
    loadFlaggedDLCs();
}

function initImportForm() {
    const form = document.getElementById('import-form');
    if (!form) return;

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const steamID = document.getElementById('steam-id').value;
        const consentAnalytics = document.getElementById('consent-analytics').checked;
        const importBtn = document.getElementById('import-btn');
        
        // Validate SteamID
        if (!steamID || steamID.length < 16) {
            showImportError('Invalid Steam ID', 'Please enter a valid 64-bit Steam ID.');
            return;
        }

        // Disable button and show loading state
        importBtn.disabled = true;
        importBtn.textContent = 'Importing...';

        try {
            const token = getAccessToken();
            const response = await fetch('/api/library/import', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({
                    steam_id: steamID
                })
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || 'Failed to import library');
            }

            // Show success
            showImportResult('Import Successful', data.message || 'Your library has been imported successfully.');
            
            // Reload library stats
            loadLibraryStats();
            loadFlaggedDLCs();
            
            // Reset form
            form.reset();

        } catch (error) {
            console.error('Import failed:', error);
            showImportError('Import Failed', error.message || 'Failed to import your library. Please try again.');
        } finally {
            importBtn.disabled = false;
            importBtn.textContent = 'Import Library';
        }
    });
}

async function loadFlaggedDLCs() {
    try {
        const token = getAccessToken();
        const response = await fetch('/api/library/dlc', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        const data = await response.json();
        if (!response.ok) {
            throw new Error(data.error || 'Failed to load DLCs');
        }
        renderFlaggedDLCs(data.dlcs || []);
    } catch (error) {
        console.error('Failed to load DLC recommendations:', error);
    }
}

function renderFlaggedDLCs(dlcs) {
    const host = document.getElementById('dlc-list');
    const section = document.getElementById('dlc-section');
    if (!host || !section) return;

    if (dlcs.length === 0) {
        section.style.display = 'none';
        host.innerHTML = '';
        return;
    }

    section.style.display = 'block';
    host.innerHTML = dlcs.map(dlc => `
        <div class="dlc-item">
            <div>
                <div class="dlc-title">${dlc.title}</div>
                <div class="dlc-meta">${dlc.platform || 'Store'} · ₹${dlc.price_inr || 0}</div>
            </div>
            <a class="btn btn-secondary btn-sm" href="game.html?id=${dlc.id}">View</a>
        </div>
    `).join('');
}

async function loadLibraryStats() {
    try {
        const token = getAccessToken();
        const response = await fetch('/api/library', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Failed to load library');
        }

        // Update stats
        const statsSection = document.getElementById('library-stats');
        const ownedCount = document.getElementById('owned-count');
        const hiddenCount = document.getElementById('hidden-count');

        if (statsSection && ownedCount && hiddenCount) {
            ownedCount.textContent = data.count || 0;
            hiddenCount.textContent = data.count || 0; // For now, same as owned
            statsSection.style.display = 'grid';
        }

    } catch (error) {
        console.error('Failed to load library stats:', error);
    }
}

function showImportResult(title, message) {
    const resultDiv = document.getElementById('import-result');
    const errorDiv = document.getElementById('import-error');
    const form = document.getElementById('import-form');

    if (resultDiv) {
        document.getElementById('result-title').textContent = title;
        document.getElementById('result-message').textContent = message;
        resultDiv.style.display = 'block';
    }

    if (errorDiv) {
        errorDiv.style.display = 'none';
    }

    if (form) {
        form.style.display = 'none';
    }
}

function showImportError(title, message) {
    const resultDiv = document.getElementById('import-result');
    const errorDiv = document.getElementById('import-error');
    const form = document.getElementById('import-form');

    if (errorDiv) {
        document.getElementById('error-title').textContent = title;
        document.getElementById('error-message').textContent = message;
        errorDiv.style.display = 'block';
    }

    if (resultDiv) {
        resultDiv.style.display = 'none';
    }

    if (form) {
        form.style.display = 'flex';
    }
}
