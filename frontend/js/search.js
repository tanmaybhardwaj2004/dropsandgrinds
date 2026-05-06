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
        performSearch();
    } else {
        // Load initial results when no query is provided
        performSearch();
    }
});

let currentPage = 0;
const limit = 30;
let totalResults = 0;

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

    // Add score slider event listener
    const minScoreInput = document.getElementById('filter-min-score');
    const scoreDisplay = document.getElementById('score-display');
    if (minScoreInput && scoreDisplay) {
        minScoreInput.addEventListener('input', (e) => {
            scoreDisplay.textContent = e.target.value + '+';
        });
    }

    if (prevBtn) {
        prevBtn.addEventListener('click', () => {
            if (currentPage > 0) {
                currentPage--;
                performSearch();
            }
        });
    }

    if (nextBtn) {
        nextBtn.addEventListener('click', () => {
            const maxPage = Math.ceil(totalResults / limit) - 1;
            if (currentPage < maxPage) {
                currentPage++;
                performSearch();
            }
        });
    }
}

function clearFilters() {
    const platformSelect = document.getElementById('filter-platform');
    const minPriceInput = document.getElementById('filter-min-price');
    const maxPriceInput = document.getElementById('filter-max-price');
    const minDiscountInput = document.getElementById('filter-min-discount');
    const maxDiscountInput = document.getElementById('filter-max-discount');
    const minScoreInput = document.getElementById('filter-min-score');
    const maxScoreInput = document.getElementById('filter-max-score');

    if (platformSelect) platformSelect.value = '';
    if (minPriceInput) minPriceInput.value = '';
    if (maxPriceInput) maxPriceInput.value = '';
    if (minDiscountInput) minDiscountInput.value = '';
    if (maxDiscountInput) maxDiscountInput.value = '';
    if (minScoreInput) minScoreInput.value = '';
    if (maxScoreInput) maxScoreInput.value = '';
    currentPage = 0;
    performSearch();
}

async function performSearch() {
    const searchInput = document.getElementById('search-input');
    const platformSelect = document.getElementById('filter-platform');
    const minPriceInput = document.getElementById('filter-min-price');
    const maxPriceInput = document.getElementById('filter-max-price');
    const minDiscountInput = document.getElementById('filter-min-discount');
    const maxDiscountInput = document.getElementById('filter-max-discount');
    const minScoreInput = document.getElementById('filter-min-score');
    const maxScoreInput = document.getElementById('filter-max-score');

    const query = searchInput ? searchInput.value.trim() : '';
    const platform = platformSelect ? platformSelect.value : '';
    const minPrice = minPriceInput ? parseFloat(minPriceInput.value) || 0 : 0;
    const maxPrice = maxPriceInput ? parseFloat(maxPriceInput.value) || 0 : 0;
    const minDiscount = minDiscountInput ? parseInt(minDiscountInput.value) || 0 : 0;
    const maxDiscount = maxDiscountInput ? parseInt(maxDiscountInput.value) || 0 : 0;
    const minReviewScore = minScoreInput ? parseFloat(minScoreInput.value) || 0 : 0;
    const maxReviewScore = maxScoreInput ? parseFloat(maxScoreInput.value) || 0 : 0;

    const loadingState = document.getElementById('loading-state');
    const emptyState = document.getElementById('empty-state');
    const errorState = document.getElementById('error-state');
    const resultsGrid = document.getElementById('results-grid');
    const pagination = document.getElementById('pagination');

    // Show loading state
    if (loadingState) loadingState.style.display = 'flex';
    if (emptyState) emptyState.style.display = 'none';
    if (errorState) errorState.style.display = 'none';
    if (resultsGrid) resultsGrid.innerHTML = '';
    if (pagination) pagination.style.display = 'none';

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
    params.append('limit', limit);
    params.append('offset', currentPage * limit);

    try {
        const response = await fetch(`/api/games/search?${params.toString()}`);
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
            renderResults(data.games);
            updatePagination();
        } else {
            if (emptyState) emptyState.style.display = 'flex';
        }

    } catch (error) {
        console.error('Search failed:', error);
        if (loadingState) loadingState.style.display = 'none';
        if (errorState) errorState.style.display = 'flex';
    }
}

function renderResults(games) {
    const resultsGrid = document.getElementById('results-grid');
    if (!resultsGrid) return;

    resultsGrid.innerHTML = games.map(game => createGameCard(game)).join('');
}

function createGameCard(game) {
    const discountBadge = game.discount_percent > 0 
        ? `<div class="discount">-${game.discount_percent}%</div>` 
        : '';
    
    const scoreColor = getScoreColorClass(game.review_score);
    const reviewBadge = game.review_score > 0
        ? `<span class="deal-score-badge ${scoreColor}">${game.review_score.toFixed(0)}</span>`
        : '';

    const lowestBadge = game.is_all_time_low
        ? `<div class="lowest-badge">🔥 All-time low</div>`
        : '';

    return `
        <div class="deal-card" onclick="window.location.href='game.html?id=${game.id}'">
            <img src="${getProxiedImageUrl(game.cover_url) || 'data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22 width=%22200%22 height=%22150%22%3E%3Crect fill=%22%23333%22 width=%22200%22 height=%22150%22/%3E%3Ctext fill=%22%23666%22 x=%2250%25%22 y=%2250%25%22 text-anchor=%22middle%22%3ENo Image%3C/text%3E%3C/svg%3E'}" 
                 alt="${game.title}" 
                 class="deal-cover"
                 loading="lazy"
                 onerror="this.src='https://via.placeholder.com/400x600?text=No+Cover'">
            <div class="deal-info">
                <h3 class="deal-title">${game.title}</h3>
                <div class="deal-price-row">
                    <div class="price">₹${game.price_inr.toFixed(0)}</div>
                    ${discountBadge}
                </div>
                <div class="meta-row">
                    <span>${game.platform}</span>
                    ${reviewBadge}
                </div>
                ${lowestBadge}
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
        
        if (prevBtn) prevBtn.disabled = currentPage === 0;
        if (nextBtn) nextBtn.disabled = currentPage >= maxPage;
        
        if (pageInfo) {
            pageInfo.textContent = `Page ${currentPage + 1} of ${maxPage + 1}`;
        }
    } else {
        pagination.style.display = 'none';
    }
}
