document.addEventListener('DOMContentLoaded', () => {
    console.log("DropsAndGrinds App Loaded");
    initApp();
});

let allDeals = [];

async function initApp() {
    initAuthButton();
    await checkHealth();
    await loadDeals();
    await loadWishlistPreview();

    // Attach Event Listeners to Filters
    const filters = ['store-steam', 'store-epic', 'store-gog'];
    filters.forEach(id => {
        document.getElementById(id).addEventListener('change', updateFilters);
    });

    const priceSlider = document.getElementById('price-slider');
    const priceDisplay = document.getElementById('price-display');
    priceSlider.addEventListener('input', (e) => {
        priceDisplay.textContent = `₹${e.target.value}`;
        updateFilters();
    });
    
    const searchInput = document.getElementById('search-input');
    searchInput.addEventListener('input', updateFilters);
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

async function loadDeals() {
    const container = document.getElementById('deals-container');
    container.innerHTML = '<p style="color: var(--text-muted); grid-column: 1/-1;">Loading live deals...</p>';

    try {
        const response = await fetch('/api/deals?limit=100&offset=0');
        const payload = await response.json();
        if (!response.ok) {
            throw new Error(payload.error || 'Failed to fetch deals');
        }

        allDeals = (payload.deals || []).map((deal) => ({
            id: deal.id,
            title: deal.title,
            cover: deal.cover_url || '',
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

        renderDeals(allDeals);
    } catch (error) {
        container.innerHTML = '<p style="color: #ff7b72; grid-column: 1/-1;">Failed to load deals from API. Please try again.</p>';
        console.error(error);
    }
}

function updateFilters() {
    const steamChecked = document.getElementById('store-steam').checked;
    const epicChecked = document.getElementById('store-epic').checked;
    const gogChecked = document.getElementById('store-gog').checked;
    const maxPrice = parseInt(document.getElementById('price-slider').value);
    const searchTerm = document.getElementById('search-input').value.toLowerCase();

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

    renderDeals(filtered);
}

function renderDeals(dealsArray) {
    const container = document.getElementById('deals-container');
    container.innerHTML = ''; // clear grid

    if(dealsArray.length === 0) {
        container.innerHTML = '<p style="color: var(--text-muted); grid-column: 1/-1;">No deals found matching criteria.</p>';
        return;
    }

    dealsArray.forEach(deal => {
        const card = document.createElement('div');
        card.className = 'deal-card';
        card.addEventListener('click', () => {
            window.location.href = `game.html?id=${deal.id}`;
        });
        
        card.innerHTML = `
            <img src="${deal.cover}" class="deal-cover" alt="${deal.title} cover">
            <div class="deal-info">
                <div class="meta-row">
                    <span>${deal.store} ${deal.isGSTAdded ? '(Inc. GST)' : ''}</span>
                    <span class="score-badge">★ ${deal.score}</span>
                </div>
                <div class="deal-title">${deal.title}</div>
                <div class="deal-price-row">
                    <span class="discount">-${deal.discount}%</span>
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
        `;
        container.appendChild(card);
    });
}
