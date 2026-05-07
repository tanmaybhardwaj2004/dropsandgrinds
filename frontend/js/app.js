document.addEventListener('DOMContentLoaded', () => {
    console.log("DropsAndGrinds App Loaded");
    initApp();
    registerServiceWorker();
});

// Register Service Worker for PWA
function registerServiceWorker() {
    // Chromium/Brave can get stuck with stale service-worker caches during rapid
    // frontend changes. For local/dev reliability, we disable SW and unregister old ones.
    if (!('serviceWorker' in navigator)) return;
    window.addEventListener('load', async () => {
        try {
            const regs = await navigator.serviceWorker.getRegistrations();
            await Promise.all(regs.map((reg) => reg.unregister()));
            if ('caches' in window) {
                const keys = await caches.keys();
                await Promise.all(keys.map((key) => caches.delete(key)));
            }
            console.log('Service Workers disabled for local reliability.');
        } catch (error) {
            console.log('Service Worker cleanup failed:', error);
        }
    });
}

function normalizePlatformName(storeName) {
    const raw = String(storeName || '').trim().toLowerCase();
    if (raw === '') return 'Store';
    if (raw.includes('steam')) return 'Steam';
    if (raw.includes('epic')) return 'Epic Games';
    if (raw.includes('gog')) return 'GOG';
    if (raw.includes('humble')) return 'Humble Store';
    if (raw.includes('fanatical')) return 'Fanatical';
    if (raw.includes('ubisoft')) return 'Ubisoft Store';
    if (raw.includes('origin') || raw.includes('ea')) return 'EA App';
    return storeName;
}

function normalizeRegionName(region) {
    const value = String(region || '').trim();
    if (!value) return 'Global';
    if (value.toLowerCase() === 'global') return 'Global';
    if (value.toLowerCase() === 'india') return 'India';
    return value;
}

function mapDealPayload(deal) {
    return {
        id: deal.id,
        title: deal.title,
        cover: getProxiedImageUrl(deal.cover_url) || '/images/game-placeholder.svg',
        store: normalizePlatformName(deal.platform),
        price: deal.price_inr || 0,
        lowestPrice: deal.lowest_price_inr || 0,
        original: deal.original_inr || 0,
        discount: deal.discount_percent || 0,
        score: deal.review_score || 0,
        status: deal.deal_status || '',
        quality: deal.deal_quality || deal.deal_status || '',
        savings: deal.potential_savings_inr || 0,
        cheapestRegion: normalizeRegionName(deal.cheapest_region),
        paymentMethods: deal.payment_methods || [],
        isGSTAdded: true
    };
}

function mapGamePayloadAsDeal(game) {
    return {
        id: game.id,
        title: game.title,
        cover: getProxiedImageUrl(game.cover_url) || '/images/game-placeholder.svg',
        store: normalizePlatformName(game.platform),
        price: game.price_inr || 0,
        lowestPrice: game.lowest_price_inr || 0,
        original: game.original_inr || 0,
        discount: game.discount_percent || 0,
        score: game.review_score || 0,
        status: '',
        quality: game.discount_percent >= 70 || game.is_all_time_low ? 'hot' : game.discount_percent >= 30 ? 'good' : 'meh',
        savings: Math.max(0, (game.original_inr || 0) - (game.price_inr || 0)),
        cheapestRegion: normalizeRegionName(game.cheapest_region),
        paymentMethods: game.payment_methods || [],
        isGSTAdded: true
    };
}

function updatePlatformFilterOptionsFromDeals(deals) {
    const filters = [
        { id: 'store-steam', labels: ['steam'] },
        { id: 'store-epic', labels: ['epic'] },
        { id: 'store-gog', labels: ['gog'] }
    ];
    for (const filter of filters) {
        const input = document.getElementById(filter.id);
        if (!input) continue;
        const hasDeals = deals.some((deal) => {
            const store = String(deal.store || '').toLowerCase();
            return filter.labels.some((label) => store.includes(label));
        });
        input.disabled = !hasDeals;
    }
}

function getSelectedStores() {
    const stores = [];
    const steam = document.getElementById('store-steam');
    const epic = document.getElementById('store-epic');
    const gog = document.getElementById('store-gog');
    if (steam?.checked) stores.push('steam');
    if (epic?.checked) stores.push('epic');
    if (gog?.checked) stores.push('gog');
    return stores;
}

async function fetchDealsFromApi(limit, offset) {
    const selectedStores = getSelectedStores();
    const query = (document.getElementById('sidebar-search')?.value || '').trim();
    const maxPrice = parseInt(document.getElementById('price-slider')?.value || '0', 10);

    // If user applied search/store filters, use live search endpoint so results are not
    // limited to only what was previously loaded on the page.
    if (query || selectedStores.length > 0 || maxPrice > 0) {
        const aggregate = [];
        const storesToQuery = selectedStores.length > 0 ? selectedStores : ['steam', 'epic', 'gog'];
        for (const store of storesToQuery) {
            const params = new URLSearchParams();
            if (query) params.set('q', query);
            params.set('platform', store);
            if (maxPrice > 0) params.set('max_price', String(maxPrice));
            params.set('limit', String(limit));
            params.set('offset', String(offset));
            const response = await fetch(`/api/games/search?${params.toString()}`);
            const payload = await response.json();
            if (!response.ok) {
                throw new Error(payload.error || 'Failed to fetch live search deals');
            }
            for (const game of payload.games || []) {
                aggregate.push(mapGamePayloadAsDeal(game));
            }
        }
        // Deduplicate by game id + store label
        const seen = new Set();
        const unique = [];
        for (const row of aggregate) {
            const key = `${row.id}:${String(row.store).toLowerCase()}`;
            if (seen.has(key)) continue;
            seen.add(key);
            unique.push(row);
        }
        return { rows: unique, total: unique.length };
    }

    const response = await fetch(`/api/deals?limit=${limit}&offset=${offset}`);
    const payload = await response.json();
    if (!response.ok) {
        throw new Error(payload.error || 'Failed to fetch deals');
    }
    return {
        rows: (payload.deals || []).map(mapDealPayload),
        total: payload.total || 0
    };
}

let allDeals = [];
let dealsRefreshTimer = null;
let dealsOffset = 0;
let dealsTotal = 0;
const dealsPageSize = 24;

// Transform external image URLs to use local proxy (bypasses hotlink protection)
function getProxiedImageUrl(originalUrl) {
    if (!originalUrl) return '';
    let nextUrl = originalUrl;
    if (nextUrl.includes('/header.jpg')) {
        nextUrl = nextUrl.replace('/header.jpg', '/library_600x900.jpg');
    }
    if (nextUrl.includes('/capsule_231x87.jpg')) {
        nextUrl = nextUrl.replace('/capsule_231x87.jpg', '/library_600x900.jpg');
    }
    
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

async function loadActiveSales() {
    try {
        const response = await fetch('/api/sales/active');
        const sales = await response.json();

        if (!response.ok) {
            throw new Error('Failed to load active sales');
        }

        const banner = document.getElementById('live-sale-banner');
        const title = document.getElementById('sale-banner-title');
        const message = document.getElementById('sale-banner-message');

        if (sales && sales.length > 0) {
            // Show the first active sale
            const sale = sales[0];
            const endDate = new Date(sale.end_date);
            const daysRemaining = Math.ceil((endDate - new Date()) / (1000 * 60 * 60 * 24));
            
            title.textContent = `LIVE: ${sale.name}`;
            message.textContent = `Ending in ${daysRemaining} days. Don't miss out on great deals!`;
            banner.style.display = 'block';
        } else {
            // Hide banner if no active sales
            banner.style.display = 'none';
        }
    } catch (error) {
        console.error('Failed to load active sales:', error);
        // Hide banner on error
        const banner = document.getElementById('live-sale-banner');
        if (banner) banner.style.display = 'none';
    }
}

function initSearch() {
    const searchInput = document.getElementById('sidebar-search');

    if (!searchInput) return;

    const performSearch = () => {
        const query = searchInput.value.trim();
        if (query) {
            window.location.href = `search.html?q=${encodeURIComponent(query)}`;
        }
    };

    searchInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            performSearch();
        }
    });
    searchInput.addEventListener('input', updateFilters);
}

async function initApp() {
    initAuthButton();
    initSearch();
    await checkHealth();
    await loadActiveSales();
    await loadDeals();
    startDealAutoRefresh();
    await loadWishlistPreview();
    await loadDealsForYou();

    // Attach Event Listeners to Filters
    const filters = ['store-steam', 'store-epic', 'store-gog'];
    filters.forEach(id => {
        const el = document.getElementById(id);
        if (el) el.addEventListener('change', updateFilters);
    });

    const priceSlider = document.getElementById('price-slider');
    const priceDisplay = document.getElementById('price-display');
    if (priceSlider) {
        priceSlider.addEventListener('input', (e) => {
            priceDisplay.textContent = `₹${e.target.value}`;
            updateFilters();
        });
    }
    
    const hideOwnedCheckbox = document.getElementById('hide-owned');
    if (hideOwnedCheckbox) {
        hideOwnedCheckbox.addEventListener('change', () => loadDeals({ reset: true }));
    }

    const sortSelect = document.getElementById('sort-select');
    if (sortSelect) {
        sortSelect.addEventListener('change', updateFilters);
    }

    const loadMoreBtn = document.getElementById('load-more-btn');
    if (loadMoreBtn) {
        loadMoreBtn.addEventListener('click', () => loadDeals({ append: true }));
    }
}

function startDealAutoRefresh() {
    if (dealsRefreshTimer) clearInterval(dealsRefreshTimer);
    dealsRefreshTimer = setInterval(() => {
        const active = document.visibilityState === 'visible';
        if (active) loadDeals({ silent: true, rotate: true });
    }, 60000);
}

async function loadWishlistPreview() {
    const host = document.getElementById('wishlist-preview');
    if (!host) return;

    const token = getAccessToken();
    if (!token) {
        host.textContent = 'Log in to view your wishlist.';
        return;
    }

    try {
        const response = await fetch('/api/wishlist?limit=5&offset=0', {
            headers: {
                Authorization: `Bearer ${token}`,
            },
        });

        if (response.status === 401) {
            host.textContent = 'Session expired. Please log in again.';
            return;
        }

        const payload = await response.json();
        if (!response.ok) {
            throw new Error(payload.error || 'Failed to load wishlist');
        }

        const items = payload.items || [];
        if (items.length === 0) {
            host.textContent = 'No games in wishlist yet.';
            return;
        }

        host.innerHTML = `<ul>${items
            .map((item) => `<li><span>${item.title}</span><span>₹${item.target_price_inr}</span></li>`)
            .join('')}</ul>`;
    } catch (error) {
        host.textContent = 'Could not load wishlist.';
        console.error(error);
    }
}

async function checkHealth() {
    try {
        const response = await fetch('/health');
        if (!response.ok) {
            console.warn('Health check returned non-200 status');
        }
    } catch (err) {
        console.error('Backend not reachable:', err);
    }
}

function renderSkeletons(count = 6) {
    const container = document.getElementById('deals-container');
    if (!container) return;
    container.innerHTML = '';
    for (let i = 0; i < count; i++) {
        const skeleton = document.createElement('div');
        skeleton.className = 'skeleton-card';
        skeleton.innerHTML = `
            <div class="skeleton-cover"></div>
            <div class="skeleton-info">
                <div class="skeleton-line short"></div>
                <div class="skeleton-line medium"></div>
                <div class="skeleton-line"></div>
            </div>
        `;
        container.appendChild(skeleton);
    }
}

function renderEmptyState(message = 'No deals found matching your criteria.') {
    return `
        <div class="state-container state-empty">
            <div class="state-icon"><i data-lucide="inbox"></i></div>
            <div class="state-title">No Deals Found</div>
            <div class="state-message">${message}</div>
        </div>
    `;
}

function renderErrorState(message = 'Failed to load deals.', onRetry) {
    return `
        <div class="state-container state-error">
            <div class="state-icon"><i data-lucide="triangle-alert"></i></div>
            <div class="state-title">Oops!</div>
            <div class="state-message">${message}</div>
            <button class="retry-btn" onclick="${onRetry}()">Try Again</button>
        </div>
    `;
}

function getScoreColorClass(score) {
    if (score >= 85) return 'green';
    if (score >= 70) return 'amber';
    if (score >= 50) return 'orange';
    return 'red';
}

async function loadDeals(options = {}) {
    const container = document.getElementById('deals-container');
    const loadMoreBtn = document.getElementById('load-more-btn');
    const append = Boolean(options.append);
    if (options.reset || !append) {
        dealsOffset = 0;
    }
    if (!options.silent && !append) renderSkeletons();
    if (loadMoreBtn) {
        loadMoreBtn.disabled = true;
        loadMoreBtn.textContent = append ? 'Loading...' : 'Load More Deals';
    }

    try {
        const hideOwned = document.getElementById('hide-owned')?.checked || false;
        const offset = append ? dealsOffset : 0;
        let nextDeals = [];
        let computedTotal = 0;
        if (hideOwned) {
            const token = typeof getAccessToken === 'function' ? getAccessToken() : null;
            const headers = token ? { Authorization: `Bearer ${token}` } : {};
            const response = await fetch(`/api/games?limit=${dealsPageSize}&offset=${offset}&exclude_owned=true`, { headers });
            const payload = await response.json();
            if (!response.ok) {
                throw new Error(payload.error || 'Failed to fetch games');
            }
            nextDeals = (payload.games || []).map(mapGamePayloadAsDeal);
            computedTotal = payload.total || nextDeals.length;
        } else {
            const live = await fetchDealsFromApi(dealsPageSize, offset);
            nextDeals = live.rows;
            computedTotal = live.total;
        }

        allDeals = append ? allDeals.concat(nextDeals) : nextDeals;
        dealsOffset = offset + nextDeals.length;
        dealsTotal = computedTotal || allDeals.length;
        updatePlatformFilterOptionsFromDeals(allDeals);

        updateDealStats(allDeals, dealsTotal);
        updateDealsHeading(hideOwned, dealsTotal);
        if (append) {
            // Requirement: "Load More" must append, not replace.
            renderDeals(nextDeals, { append: true });
        } else {
            renderDeals(options.rotate ? rotateDeals(allDeals) : allDeals);
        }
        updateLoadMoreButton();
    } catch (error) {
        container.innerHTML = renderErrorState('Failed to load deals from API. Please try again.', 'loadDeals');
        console.error(error);
    } finally {
        if (loadMoreBtn) {
            loadMoreBtn.disabled = false;
            loadMoreBtn.innerHTML = '<i data-lucide="plus"></i> Load More Deals';
        }
        updateLoadMoreButton();
        if (window.lucide) window.lucide.createIcons();
    }
}

function updateLoadMoreButton() {
    const loadMoreBtn = document.getElementById('load-more-btn');
    if (!loadMoreBtn) return;
    loadMoreBtn.style.display = dealsOffset < dealsTotal ? 'inline-flex' : 'none';
}

function rotateDeals(deals) {
    if (deals.length <= 1) return deals;
    const offset = Math.floor(Math.random() * deals.length);
    return deals.slice(offset).concat(deals.slice(0, offset));
}

function updateDealStats(deals, total) {
    const statDeals = document.getElementById('stat-deals');
    const statSavings = document.getElementById('stat-savings');
    if (statDeals) statDeals.textContent = total;
    if (statSavings) {
        const discounted = deals.filter((deal) => deal.discount > 0);
        const avg = discounted.length
            ? Math.round(discounted.reduce((sum, deal) => sum + deal.discount, 0) / discounted.length)
            : 0;
        statSavings.textContent = `${avg}%`;
    }
}

function updateDealsHeading(hideOwned, total) {
    const heading = document.getElementById('deals-heading');
    if (!heading) return;
    heading.textContent = hideOwned ? `Games You Do Not Own (${total})` : `Live Deals (${total})`;
}

async function loadDealsForYou() {
    const token = getAccessToken();
    const section = document.getElementById('deals-for-you-section');
    
    if (!section) return;
    
    section.style.display = 'block';

    const container = document.getElementById('deals-for-you-container');
    if (!container) return;

    // Show loading skeletons
    container.innerHTML = '<div class="deals-for-you-grid">' + 
        Array(4).fill(`<div class="skeleton-card"><div class="skeleton-cover" style="height:180px"></div><div class="skeleton-info"><div class="skeleton-line short"></div></div></div>`).join('') + 
        '</div>';

    try {
        // Backend supports optional auth: returns personalized when logged in, top deals otherwise.
        const endpoint = '/api/deals/for-you?limit=4&offset=0';
        const headers = token ? { Authorization: `Bearer ${token}` } : {};
        const response = await fetch(endpoint, { headers });

        if (response.status === 401) {
            container.innerHTML = '';
            return;
        }

        const payload = await response.json();
        if (!response.ok) {
            throw new Error(payload.error || 'Failed to fetch personalized deals');
        }

        const deals = (payload.deals || []).map((deal) => ({
            id: deal.id,
            title: deal.title,
            cover: getProxiedImageUrl(deal.cover_url) || '/images/game-placeholder.svg',
            store: deal.platform || 'Store',
            price: deal.price_inr || 0,
            lowestPrice: deal.lowest_price_inr || 0,
            original: deal.original_inr || 0,
            discount: deal.discount_percent || 0,
            score: deal.review_score || 0,
            status: deal.deal_status || '',
            quality: deal.deal_quality || deal.deal_status || '',
            savings: deal.potential_savings_inr || 0,
            cheapestRegion: deal.cheapest_region || 'India',
            paymentMethods: deal.payment_methods || [],
            isGSTAdded: true,
            reason: token ? (deal.personalized_reason || 'Recommended for you') : 'Top deal',
        }));

        renderDealsForYou(deals);
    } catch (error) {
        container.innerHTML = '';
        console.error('Failed to load deals for you:', error);
    }
}

function renderDealsForYou(deals) {
    const container = document.getElementById('deals-for-you-container');
    if (!container) return;

    if (deals.length === 0) {
        container.innerHTML = `
            <div class="state-container" style="padding: 32px;">
                <div class="state-message">Add games to your wishlist to get personalized recommendations!</div>
            </div>
        `;
        return;
    }

    // Render with the same card template/size as Live Deals.
    container.innerHTML = '';
    renderDeals(deals, { append: true, targetContainerId: 'deals-for-you-container', showReasonBadge: true });
    if (window.lucide) window.lucide.createIcons();
}

function updateFilters() {
    const steamChecked = document.getElementById('store-steam')?.checked ?? true;
    const epicChecked = document.getElementById('store-epic')?.checked ?? true;
    const gogChecked = document.getElementById('store-gog')?.checked ?? true;
    const maxPrice = parseInt(document.getElementById('price-slider')?.value || '5000');
    const searchTerm = (document.getElementById('sidebar-search')?.value || '').toLowerCase();

    const filtered = allDeals.filter(deal => {
        // Store filter
        const store = String(deal.store || '').toLowerCase();
        if (store.includes("steam") && !steamChecked) return false;
        if (store.includes("epic") && !epicChecked) return false;
        if (store.includes("gog") && !gogChecked) return false;

        // Price filter
        if (deal.price > maxPrice) return false;
        
        // Search Filter
        if (searchTerm && !deal.title.toLowerCase().includes(searchTerm)) return false;

        return true;
    });

    const sort = document.getElementById('sort-select')?.value || 'discount';
    filtered.sort((a, b) => {
        switch (sort) {
            case 'price-low':
                return a.price - b.price;
            case 'price-high':
                return b.price - a.price;
            case 'rating':
                return b.score - a.score;
            case 'discount':
            default:
                return b.discount - a.discount;
        }
    });

    renderDeals(filtered);
}

function renderDeals(dealsArray, options = {}) {
    const targetId = options.targetContainerId || 'deals-container';
    const container = document.getElementById(targetId);
    const append = Boolean(options.append);
    if (!append) {
        container.innerHTML = ''; // clear grid
    }

    if (!append && dealsArray.length === 0) {
        container.innerHTML = renderEmptyState();
        return;
    }

    dealsArray.forEach(deal => {
        const card = document.createElement('div');
        card.className = 'deal-card';
        card.tabIndex = 0;
        card.setAttribute('role', 'link');
        card.setAttribute('aria-label', `View ${deal.title}`);
        card.addEventListener('click', () => {
            window.location.href = `game.html?id=${deal.id}`;
        });
        card.addEventListener('keydown', (event) => {
            if (event.key === 'Enter' || event.key === ' ') {
                event.preventDefault();
                window.location.href = `game.html?id=${deal.id}`;
            }
        });
        
        const scoreColor = getScoreColorClass(deal.score);
        const savingsAmount = Math.max(0, deal.original - deal.price);
        
        card.innerHTML = `
            <img src="${deal.cover}" class="deal-cover" alt="${deal.title} cover" onerror="this.src='/images/game-placeholder.svg'">
            <div class="deal-info">
                ${options.showReasonBadge && deal.reason ? `<div class="badge badge-primary" style="margin-bottom: 8px; width: fit-content;">${deal.reason}</div>` : ''}
                <div class="meta-row">
                    <span>${deal.store} ${deal.isGSTAdded ? '(Inc. GST)' : ''}</span>
                    ${deal.score > 0 ? `<span class="score-badge">${deal.score}</span>` : '<span class="score-badge muted">No score</span>'}
                </div>
                <div class="deal-title">${deal.title}</div>
                <div class="deal-price-row">
                    ${deal.discount > 0 ? `<span class="discount">-${deal.discount}%</span>` : '<span class="discount muted">Deal</span>'}
                    <div style="text-align: right;">
                        <span style="text-decoration: line-through; color: var(--color-text-muted); font-size: 0.8rem; display: block;">₹${deal.original}</span>
                        <span class="price">₹${deal.price}</span>
                    </div>
                </div>
                <div class="meta-row" style="margin-top: 10px;">
                    <span>Best: ₹${deal.lowestPrice}</span>
                    <span>${deal.quality ? deal.quality.toUpperCase() : 'DEAL'}</span>
                </div>
                <div class="meta-row" style="margin-top: 8px;">
                    <span>${deal.cheapestRegion}</span>
                    <span>${deal.paymentMethods.slice(0, 2).join(' / ') || 'Card'}</span>
                </div>
                <button class="btn btn-secondary btn-sm wishlist-card-btn" onclick="event.stopPropagation(); addDealToWishlist(${deal.id}, ${Math.max(1, deal.lowestPrice || deal.price || 1)})">
                    Add to Wishlist
                </button>
            </div>
            <div class="deal-overlay">
                <div class="overlay-title">${deal.title}</div>
                <div class="overlay-score">
                    <div class="score-circle ${scoreColor}">${deal.score}</div>
                    <span>Review Score</span>
                </div>
                <div class="overlay-savings">Save ₹${savingsAmount}</div>
                <a href="game.html?id=${deal.id}" class="overlay-btn" onclick="event.stopPropagation()">View Details</a>
            </div>
        `;
        container.appendChild(card);
    });
    if (window.lucide) window.lucide.createIcons();
}

async function addDealToWishlist(gameID, targetPrice) {
    const token = getAccessToken();
    if (!token) {
        sessionStorage.setItem('dropsandgrinds_login_message', 'Sign in to add to wishlist');
        window.location.href = 'login.html';
        return;
    }

    try {
        const response = await fetch('/api/wishlist', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
                game_id: gameID,
                target_price_inr: Number(targetPrice) || 1,
            }),
        });

        if (response.status === 401) {
            sessionStorage.setItem('dropsandgrinds_login_message', 'Sign in to add to wishlist');
            window.location.href = 'login.html';
            return;
        }

        if (response.status === 409) {
            alert('This game is already in your wishlist.');
            return;
        }

        const payload = await response.json().catch(() => ({}));
        if (!response.ok) {
            throw new Error(payload.error || 'Failed to add to wishlist');
        }
        alert('Added to wishlist.');
    } catch (error) {
        alert(error.message || 'Could not add to wishlist');
    }
}
