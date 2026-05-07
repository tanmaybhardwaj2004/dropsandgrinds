document.addEventListener('DOMContentLoaded', () => {
    console.log('Library page loaded');
    initAuthButton();
    checkAuthAndLoadLibrary();
    initImportForm();
    initLibraryPagination();
});

let libraryGameIDs = [];
let libraryPageOffset = 0;
const libraryPageSize = 200;

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

function initLibraryPagination() {
    const btn = document.getElementById('library-load-more-btn');
    if (btn) {
        btn.addEventListener('click', () => renderNextLibraryPage());
    }
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
        if (!consentAnalytics) {
            showImportError('Consent Required', 'Please check the consent box before importing your Steam library.');
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
                    steam_id: steamID,
                    consent_analytics: consentAnalytics
                })
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || 'Failed to import library');
            }

            // Show success
            const imported = data.imported_count ?? data.total_games ?? 0;
            showImportResult('Import Successful', `${data.message || 'Your library has been imported successfully.'} Imported games: ${imported}.`);
            
            // Reload full imported library and DLC recommendations.
            await loadLibraryStats();
            loadFlaggedDLCs();
            
            // Reset form
            sessionStorage.setItem('dropsandgrinds_last_steam_id', steamID);
            document.getElementById('steam-id').value = steamID;

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

        libraryGameIDs = data.game_ids || [];
        libraryPageOffset = 0;

        // Update stats
        const statsSection = document.getElementById('library-stats');
        const ownedCount = document.getElementById('owned-count');
        const hiddenCount = document.getElementById('hidden-count');

        if (statsSection && ownedCount && hiddenCount) {
            ownedCount.textContent = data.count || libraryGameIDs.length || 0;
            hiddenCount.textContent = data.count || libraryGameIDs.length || 0; // For now, same as owned
            statsSection.style.display = 'grid';
        }
        renderLibraryGames();
        const savedSteamID = sessionStorage.getItem('dropsandgrinds_last_steam_id');
        const steamInput = document.getElementById('steam-id');
        if (steamInput && savedSteamID && !steamInput.value) {
            steamInput.value = savedSteamID;
        }

    } catch (error) {
        console.error('Failed to load library stats:', error);
    }
}

async function renderLibraryGames() {
    const section = document.getElementById('library-games-section');
    const list = document.getElementById('library-games-list');
    const count = document.getElementById('library-games-count');
    if (!section || !list) return;

    if (count) count.textContent = `${libraryGameIDs.length} imported`;

    if (libraryGameIDs.length === 0) {
        section.style.display = 'none';
        list.innerHTML = '';
        return;
    }

    section.style.display = 'block';
    list.innerHTML = '';
    await renderNextLibraryPage();
}

async function renderNextLibraryPage() {
    const list = document.getElementById('library-games-list');
    const btn = document.getElementById('library-load-more-btn');
    if (!list) return;

    const nextIDs = libraryGameIDs.slice(libraryPageOffset, libraryPageOffset + libraryPageSize);
    if (nextIDs.length === 0) {
        if (btn) btn.style.display = 'none';
        return;
    }

    if (btn) {
        btn.disabled = true;
        btn.textContent = 'Loading...';
    }

    const rows = await Promise.all(nextIDs.map(async (id) => {
        try {
            const response = await fetch(`/api/games/${id}`);
            const game = await response.json();
            if (!response.ok) throw new Error(game.error || 'Game unavailable');
            return renderLibraryGame(game);
        } catch {
            return renderLibraryGame({ id, title: `Imported game #${id}`, platform: 'Steam', price_inr: 0 });
        }
    }));

    list.insertAdjacentHTML('beforeend', rows.join(''));
    libraryPageOffset += nextIDs.length;

    if (btn) {
        btn.disabled = false;
        btn.textContent = 'Load More';
        btn.style.display = libraryPageOffset < libraryGameIDs.length ? 'inline-flex' : 'none';
    }
}

function renderLibraryGame(game) {
    return `
        <a class="library-game-item" href="game.html?id=${game.id}">
            <img src="${getProxiedImageUrl(game.cover_url) || '/images/game-placeholder.svg'}" alt="${game.title} cover" onerror="this.src='/images/game-placeholder.svg'">
            <div>
                <div class="library-game-title">${game.title}</div>
                <div class="library-game-meta">${game.platform || 'Steam'}${game.price_inr ? ` · ₹${game.price_inr}` : ''}</div>
            </div>
        </a>
    `;
}

function getProxiedImageUrl(originalUrl) {
    if (!originalUrl) return '';
    let nextUrl = originalUrl.replace('/header.jpg', '/library_600x900.jpg').replace('/capsule_231x87.jpg', '/library_600x900.jpg');
    if (nextUrl.includes('shared.cloudflare.steamstatic.com') || nextUrl.includes('shared.fastly.steamstatic.com')) {
        return nextUrl
            .replace('https://shared.cloudflare.steamstatic.com/', '/img/steam/')
            .replace('https://shared.fastly.steamstatic.com/', '/img/steam/');
    }
    if (nextUrl.includes('images.gog-statics.com')) {
        return nextUrl.replace('https://images.gog-statics.com/', '/img/gog/');
    }
    if (nextUrl.includes('cdn2.unrealengine.com')) {
        return nextUrl.replace('https://cdn2.unrealengine.com/', '/img/epic/');
    }
    return nextUrl;
}

function showImportResult(title, message) {
    const resultDiv = document.getElementById('import-result');
    const errorDiv = document.getElementById('import-error');
    const form = document.getElementById('import-form');

    if (resultDiv) {
        const titleEl = resultDiv.querySelector('h3');
        const messageEl = resultDiv.querySelector('p');
        if (titleEl) titleEl.textContent = title;
        if (messageEl) messageEl.textContent = message;
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
        const icon = errorDiv.querySelector('i[data-lucide]');
        errorDiv.innerHTML = `${icon ? icon.outerHTML : ''}<span><strong>${title}</strong> — ${message}</span>`;
        errorDiv.style.display = 'block';
    }

    if (resultDiv) {
        resultDiv.style.display = 'none';
    }

    if (form) {
        form.style.display = 'flex';
    }
}
