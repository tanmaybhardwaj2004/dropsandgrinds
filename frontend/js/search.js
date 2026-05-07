document.addEventListener('DOMContentLoaded', () => {
    console.log('Search page loaded');
    initAuthButton();
    initSearch();
    initFilters();
    
    // Get query from URL
    const urlParams = new URLSearchParams(window.location.search);
    const query = urlParams.get('q');
    
    if (query) {
        document.getElementById('search-input').value = query;
    }
    performSearch();
});

let currentPage = 0;
const limit = 30;
let totalResults = 0;
let searchDebounceTimer = null;

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

function initSearch() {
    const searchInput = document.getElementById('search-input');
    const searchBtn = document.getElementById('search-btn');

    if (!searchInput || !searchBtn) return;

    const performSearchOnEnter = (e) => {
        if (e.key === 'Enter') {
            currentPage = 0;
            performSearch();
        }
    };

    searchBtn.addEventListener('click', () => {
        currentPage = 0;
        performSearch();
    });
    searchInput.addEventListener('keypress', performSearchOnEnter);
    searchInput.addEventListener('input', () => {
        clearTimeout(searchDebounceTimer);
        searchDebounceTimer = setTimeout(() => {
            currentPage = 0;
            performSearch();
        }, 250);
    });
}

function initFilters() {
    const applyBtn = document.getElementById('apply-filters-btn');
    const clearBtn = document.getElementById('clear-filters-btn');
    const prevBtn = document.getElementById('prev-page-btn');
    const nextBtn = document.getElementById('next-page-btn');

    if (applyBtn) {
        applyBtn.addEventListener('click', () => {
            currentPage = 0;
            performSearch();
        });
    }

    if (clearBtn) {
        clearBtn.addEventListener('click', clearFilters);
    }

    if (prevBtn) {
        // Spec requires "Load More" UX; keep Prev hidden/disabled.
        prevBtn.style.display = 'none';
    }

    if (nextBtn) {
        nextBtn.addEventListener('click', () => {
            const maxPage = Math.ceil(totalResults / limit) - 1;
            if (currentPage < maxPage) {
                currentPage++;
                performSearch({ append: true });
            }
        });
    }

    const sortSelect = document.getElementById('sort-select');
    if (sortSelect) {
        sortSelect.addEventListener('change', () => {
            currentPage = 0;
            performSearch();
        });
    }

    const scoreSlider = document.getElementById('filter-min-score');
    const scoreDisplay = document.getElementById('score-display');
    if (scoreSlider && scoreDisplay) {
        scoreSlider.addEventListener('input', () => {
            scoreDisplay.textContent = `${scoreSlider.value}+`;
        });
    }
}

function clearFilters() {
    document.getElementById('filter-platform').value = '';
    document.getElementById('filter-min-price').value = '';
    document.getElementById('filter-max-price').value = '';
    document.getElementById('filter-min-discount').value = '';
    document.getElementById('filter-max-discount').value = '';
    document.getElementById('filter-min-score').value = '';
    const maxScore = document.getElementById('filter-max-score');
    if (maxScore) maxScore.value = '';
    const payment = document.getElementById('filter-payment');
    if (payment) payment.value = '';
    currentPage = 0;
    performSearch();
}

function quickSearch(label) {
    const searchInput = document.getElementById('search-input');
    const maxPrice = document.getElementById('filter-max-price');
    const minDiscount = document.getElementById('filter-min-discount');

    if (label === 'Under ₹500') {
        if (searchInput) searchInput.value = '';
        if (maxPrice) maxPrice.value = '500';
    } else if (label === '90% off') {
        if (searchInput) searchInput.value = '';
        if (minDiscount) minDiscount.value = '90';
    } else if (searchInput) {
        searchInput.value = label;
    }

    currentPage = 0;
    performSearch();
}

async function performSearch(options = {}) {
    const append = Boolean(options.append);
    const query = document.getElementById('search-input').value.trim();
    const platform = document.getElementById('filter-platform').value;
    const minPrice = parseFloat(document.getElementById('filter-min-price').value) || 0;
    const maxPrice = parseFloat(document.getElementById('filter-max-price').value) || 0;
    const minDiscount = parseInt(document.getElementById('filter-min-discount').value) || 0;
    const maxDiscount = parseInt(document.getElementById('filter-max-discount').value) || 0;
    const minReviewScore = parseFloat(document.getElementById('filter-min-score').value) || 0;
    const maxScoreEl = document.getElementById('filter-max-score');
    const maxReviewScore = maxScoreEl ? parseFloat(maxScoreEl.value) || 0 : 0;
    const paymentMethod = document.getElementById('filter-payment')?.value || '';

    const loadingState = document.getElementById('loading-state');
    const emptyState = document.getElementById('empty-state');
    const errorState = document.getElementById('error-state');
    const resultsGrid = document.getElementById('results-grid');
    const pagination = document.getElementById('pagination');

    // Show loading state
    if (!append) {
        if (loadingState) loadingState.style.display = 'flex';
        if (emptyState) emptyState.style.display = 'none';
        if (errorState) errorState.style.display = 'none';
        if (resultsGrid) resultsGrid.innerHTML = '';
        if (pagination) pagination.style.display = 'none';
    }

    // Build query params
    const params = new URLSearchParams();
    if (query) params.append('q', query);
    if (platform) params.append('platform', platform);
    if (minPrice > 0) params.append('min_price', minPrice);
    if (maxPrice > 0) params.append('max_price', maxPrice);
    if (minDiscount > 0) params.append('min_discount', minDiscount);
    if (maxDiscount > 0) params.append('max_discount', maxDiscount);
    if (minReviewScore > 0) params.append('min_review_score', minReviewScore);
    if (maxReviewScore > 0) params.append('max_review_score', maxReviewScore);
    if (paymentMethod) params.append('payment_method', paymentMethod);
    params.append('limit', limit);
    params.append('offset', currentPage * limit);

    try {
        const endpoint = query ? '/api/games/search' : '/api/games';
        const response = await fetch(`${endpoint}?${params.toString()}`);
        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Search failed');
        }

        // Hide loading state
        if (loadingState) loadingState.style.display = 'none';

        totalResults = data.total || 0;

        // Update results count
        const resultsCount = document.getElementById('results-count');
        if (resultsCount) {
            resultsCount.textContent = `${totalResults} results`;
        }

        // Update results title
        const resultsTitle = document.getElementById('results-title');
        if (resultsTitle) {
            resultsTitle.textContent = query ? `Results for "${query}"` : 'All Games';
        }

        if (data.games && data.games.length > 0) {
            const sorted = sortGames(data.games);
            renderResults(sorted, { append });
            updatePagination();
        } else {
            if (!append && emptyState) emptyState.style.display = 'flex';
        }

    } catch (error) {
        console.error('Search failed:', error);
        if (loadingState) loadingState.style.display = 'none';
        if (!append && errorState) errorState.style.display = 'flex';
    }
}

function sortGames(games) {
    const sort = document.getElementById('sort-select')?.value || 'relevance';
    return [...games].sort((a, b) => {
        switch (sort) {
            case 'price-low':
                return (a.price_inr || 0) - (b.price_inr || 0);
            case 'price-high':
                return (b.price_inr || 0) - (a.price_inr || 0);
            case 'discount':
                return (b.discount_percent || 0) - (a.discount_percent || 0);
            case 'rating':
                return (b.review_score || 0) - (a.review_score || 0);
            default:
                return 0;
        }
    });
}

function renderResults(games, options = {}) {
    const resultsGrid = document.getElementById('results-grid');
    if (!resultsGrid) return;

    const append = Boolean(options.append);
    const html = games.map((game, index) => createGameCard(game, index)).join('');
    if (append) {
        resultsGrid.insertAdjacentHTML('beforeend', html);
    } else {
        resultsGrid.innerHTML = html;
    }
    if (window.lucide) window.lucide.createIcons();
}

function createGameCard(game, index = 0) {
    // Match Home page deal card size/layout (same class names).
    const cover = getProxiedImageUrl(game.cover_url) || '/images/game-placeholder.svg';
    const store = game.platform || 'Store';
    const price = Number(game.price_inr || 0);
    const original = Number(game.original_inr || 0);
    const discount = Number(game.discount_percent || 0);
    const score = Number(game.review_score || 0);
    const lowestPrice = Number(game.lowest_price_inr || 0);
    const cheapestRegion = game.cheapest_region || 'India';
    const paymentMethods = (game.payment_methods || []);
    const quality = discount >= 70 || game.is_all_time_low ? 'HOT' : discount >= 30 ? 'GOOD' : 'MEH';

    const savingsAmount = Math.max(0, original - price);
    const scoreBadge = score > 0 ? `<span class="score-badge">${Math.round(score)}</span>` : '<span class="score-badge muted">No score</span>';
    const discountBadge = discount > 0 ? `<span class="discount">-${discount}%</span>` : '<span class="discount muted">Deal</span>';

    return `
        <div class="deal-card" onclick="window.location.href='game.html?id=${game.id}'" role="link" tabindex="0" aria-label="View ${game.title}">
            <img src="${cover}" class="deal-cover" alt="${game.title} cover" onerror="this.src='/images/game-placeholder.svg'">
            <div class="deal-info">
                <div class="meta-row">
                    <span>${store}</span>
                    ${scoreBadge}
                </div>
                <div class="deal-title">${game.title}</div>
                <div class="deal-price-row">
                    ${discountBadge}
                    <div style="text-align: right;">
                        <span style="text-decoration: line-through; color: var(--color-text-muted); font-size: 0.8rem; display: block;">₹${original || 0}</span>
                        <span class="price">₹${Math.round(price)}</span>
                    </div>
                </div>
                <div class="meta-row" style="margin-top: 10px;">
                    <span>Best: ₹${lowestPrice || Math.round(price)}</span>
                    <span>${quality}</span>
                </div>
                <div class="meta-row" style="margin-top: 8px;">
                    <span>${cheapestRegion}</span>
                    <span>${paymentMethods.slice(0, 2).join(' / ') || 'Card'}</span>
                </div>
                <button class="btn btn-secondary btn-sm result-open-btn" onclick="event.stopPropagation(); window.location.href='game.html?id=${game.id}'">View Details</button>
            </div>
            <div class="deal-overlay">
                <div class="overlay-title">${game.title}</div>
                <div class="overlay-savings">Save ₹${savingsAmount}</div>
                <a href="game.html?id=${game.id}" class="overlay-btn" onclick="event.stopPropagation()">View Details</a>
            </div>
        </div>
    `;
}

function updatePagination() {
    const pagination = document.getElementById('pagination');
    const prevBtn = document.getElementById('prev-page-btn');
    const nextBtn = document.getElementById('next-page-btn');
    const pageInfo = document.getElementById('page-info');

    if (!pagination) return;

    const maxPage = Math.ceil(totalResults / limit) - 1;

    if (totalResults > limit) {
        pagination.style.display = 'flex';
        
        if (prevBtn) {
            prevBtn.disabled = true;
            prevBtn.style.display = 'none';
        }
        if (nextBtn) {
            nextBtn.disabled = currentPage >= maxPage;
            nextBtn.innerHTML = `<i data-lucide="plus"></i> Load More`;
        }
        
        if (pageInfo) {
            const shown = Math.min((currentPage + 1) * limit, totalResults);
            pageInfo.textContent = `${shown} of ${totalResults}`;
        }
    } else {
        pagination.style.display = 'none';
    }
}
