document.addEventListener('DOMContentLoaded', () => {
    console.log("DropsAndGrinds App Loaded");
    initApp();
    registerServiceWorker();
});

// Register Service Worker for PWA
function registerServiceWorker() {
    if ('serviceWorker' in navigator) {
        window.addEventListener('load', () => {
            navigator.serviceWorker.register('/sw.js')
                .then((registration) => {
                    console.log('Service Worker registered:', registration);
                })
                .catch((error) => {
                    console.log('Service Worker registration failed:', error);
                });
        });
    }
}

let allDeals = [];

// Transform external image URLs to use local proxy (bypasses hotlink protection)
function getProxiedImageUrl(originalUrl) {
    if (!originalUrl) return '';
    
    if (originalUrl.includes('shared.cloudflare.steamstatic.com')) {
        return originalUrl.replace('https://shared.cloudflare.steamstatic.com/', '/img/steam/');
    }
    if (originalUrl.includes('images.gog-statics.com')) {
        return originalUrl.replace('https://images.gog-statics.com/', '/img/gog/');
    }
    if (originalUrl.includes('cdn2.unrealengine.com')) {
        return originalUrl.replace('https://cdn2.unrealengine.com/', '/img/epic/');
    }
    
    return originalUrl;
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
    const searchBtn = document.getElementById('search-btn');

    if (!searchInput || !searchBtn) return;

    const performSearch = () => {
        const query = searchInput.value.trim();
        if (query) {
            window.location.href = `search.html?q=${encodeURIComponent(query)}`;
        }
    };

    searchBtn.addEventListener('click', performSearch);
    searchInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            performSearch();
        }
    });
}

async function initApp() {
    initAuthButton();
    initSearch();
    await checkHealth();
    await loadActiveSales();
    await loadDeals();
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
    
    const searchInput = document.getElementById('search-input');
    if (searchInput) searchInput.addEventListener('input', updateFilters);

    const hideOwnedCheckbox = document.getElementById('hide-owned');
    if (hideOwnedCheckbox) {
        hideOwnedCheckbox.addEventListener('change', loadDeals);
    }

    const sortSelect = document.getElementById('sort-select');
    if (sortSelect) {
        sortSelect.addEventListener('change', updateFilters);
    }
}

function initAuthButton() {
    const btn = document.getElementById('auth-btn');
    if (!btn) return;

    const token = getAccessToken();
    if (!token) {
        btn.textContent = 'Login';
        btn.onclick = () => {
            window.location.href = 'login.html';
        };
        return;
    }

    btn.textContent = 'Logout';
    btn.onclick = () => {
        sessionStorage.removeItem('dropsandgrinds_access_token');
        sessionStorage.removeItem('dropsandgrinds_refresh_token');
        sessionStorage.removeItem('dropsandgrinds_user_id');
        sessionStorage.removeItem('dropsandgrinds_is_authenticated');
        window.location.href = 'login.html';
    };
}

function getAccessToken() {
    if (window.authState?.accessToken) {
        return window.authState.accessToken;
    }
    return sessionStorage.getItem('dropsandgrinds_access_token');
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

async function loadDeals() {
    const container = document.getElementById('deals-container');
    renderSkeletons();

    try {
        const hideOwned = document.getElementById('hide-owned')?.checked || false;
        const url = hideOwned
            ? '/api/games?limit=100&offset=0&exclude_owned=true'
            : '/api/deals?limit=100&offset=0';

        const headers = {};
        const token = typeof getAccessToken === 'function' ? getAccessToken() : null;
        if (token) {
            headers.Authorization = `Bearer ${token}`;
        }
        const response = await fetch(url, { headers });
        const payload = await response.json();
        if (!response.ok) {
            throw new Error(payload.error || 'Failed to fetch deals');
        }

        const rows = hideOwned ? (payload.games || []) : (payload.deals || []);
        allDeals = rows.map((deal) => ({
            id: deal.id,
            title: deal.title,
            cover: getProxiedImageUrl(deal.cover_url) || '',
            store: deal.platform || 'Store',
            price: deal.price_inr || 0,
            lowestPrice: deal.lowest_price_inr || 0,
            original: deal.original_inr || 0,
            discount: deal.discount_percent || 0,
            score: deal.review_score || 0,
            status: deal.deal_status || '',
            savings: deal.potential_savings_inr || 0,
            isGSTAdded: true
        }));

        updateDealStats(allDeals, payload.total || allDeals.length);
        updateDealsHeading(hideOwned, payload.total || allDeals.length);
        renderDeals(allDeals);
    } catch (error) {
        container.innerHTML = renderErrorState('Failed to load deals from API. Please try again.', 'loadDeals');
        console.error(error);
    }
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
    
    if (!token || !section) return; // Don't show section if not logged in
    
    section.style.display = 'block';

    const container = document.getElementById('deals-for-you-container');
    if (!container) return;

    // Show loading skeletons
    container.innerHTML = '<div class="deals-for-you-grid">' + 
        Array(4).fill(`<div class="skeleton-card"><div class="skeleton-cover" style="height:180px"></div><div class="skeleton-info"><div class="skeleton-line short"></div></div></div>`).join('') + 
        '</div>';

    try {
        const response = await fetch('/api/deals/for-you?limit=4&offset=0', {
            headers: {
                Authorization: `Bearer ${token}`,
            },
        });

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
            cover: getProxiedImageUrl(deal.cover_url) || '',
            store: deal.platform || 'Store',
            price: deal.price_inr || 0,
            original: deal.original_inr || 0,
            discount: deal.discount_percent || 0,
            score: deal.review_score || 0,
            reason: deal.personalized_reason || 'Recommended for you'
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

    container.innerHTML = '<div class="deals-for-you-grid">' + deals.map(deal => `
        <div class="deal-card-small" onclick="window.location.href='game.html?id=${deal.id}'">
            <img src="${deal.cover}" class="deal-cover" alt="${deal.title} cover" onerror="this.src='data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22 width=%22200%22 height=%22150%22%3E%3Crect fill=%22%23333%22 width=%22200%22 height=%22150%22/%3E%3Ctext fill=%22%23666%22 x=%2250%25%22 y=%2250%25%22 text-anchor=%22middle%22%3ENo Image%3C/text%3E%3C/svg%3E'">
            <div class="deal-info">
                <span class="personalized-badge">${deal.reason}</span>
                <div class="deal-title">${deal.title}</div>
                <div class="deal-price-row" style="margin-top:8px;">
                    <span class="discount">-${deal.discount}%</span>
                    <span class="price">₹${deal.price}</span>
                </div>
            </div>
        </div>
    `).join('') + '</div>';
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
        if (deal.store === "Steam" && !steamChecked) return false;
        if (deal.store === "Epic Games" && !epicChecked) return false;
        if (deal.store === "GOG" && !gogChecked) return false;

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

function renderDeals(dealsArray) {
    const container = document.getElementById('deals-container');
    container.innerHTML = ''; // clear grid

    if(dealsArray.length === 0) {
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
        const savingsAmount = deal.original - deal.price;
        
        card.innerHTML = `
            <img src="${deal.cover}" class="deal-cover" alt="${deal.title} cover" onerror="this.src='data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22 width=%22200%22 height=%22150%22%3E%3Crect fill=%22%23333%22 width=%22200%22 height=%22150%22/%3E%3Ctext fill=%22%23666%22 x=%2250%25%22 y=%2250%25%22 text-anchor=%22middle%22%3ENo Image%3C/text%3E%3C/svg%3E'">
            <div class="deal-info">
                <div class="meta-row">
                    <span>${deal.store} ${deal.isGSTAdded ? '(Inc. GST)' : ''}</span>
                    ${deal.score > 0 ? `<span class="score-badge">${deal.score}</span>` : '<span class="score-badge muted">No score</span>'}
                </div>
                <div class="deal-title">${deal.title}</div>
                <div class="deal-price-row">
                    ${deal.discount > 0 ? `<span class="discount">-${deal.discount}%</span>` : '<span class="discount muted">Deal</span>'}
                    <div style="text-align: right;">
                        <span style="text-decoration: line-through; color: var(--text-muted); font-size: 0.8rem; display: block;">₹${deal.original}</span>
                        <span class="price">₹${deal.price}</span>
                    </div>
                </div>
                <div class="meta-row" style="margin-top: 10px;">
                    <span>Best: ₹${deal.lowestPrice}</span>
                    <span>${deal.status ? deal.status.toUpperCase() : 'DEAL'}</span>
                </div>
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
