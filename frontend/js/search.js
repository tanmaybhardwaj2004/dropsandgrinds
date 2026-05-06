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

async function performSearch() {
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
    if (paymentMethod) params.append('payment_method', paymentMethod);
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
            renderResults(sortGames(data.games));
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

function renderResults(games) {
    const resultsGrid = document.getElementById('results-grid');
    if (!resultsGrid) return;

    resultsGrid.innerHTML = games.map(game => createGameCard(game)).join('');
    if (window.lucide) window.lucide.createIcons();
}

function createGameCard(game) {
    const discountBadge = game.discount_percent > 0 
        ? `<div class="discount">-${game.discount_percent}%</div>` 
        : '';
    
    const reviewBadge = game.review_score > 0
        ? `<div class="score-badge">${game.review_score.toFixed(0)}</div>`
        : '';

    const lowestBadge = game.is_all_time_low
        ? `<div class="lowest-badge">All-time low</div>`
        : '';

    return `
        <div class="deal-card" onclick="window.location.href='game.html?id=${game.id}'">
            <img src="${getProxiedImageUrl(game.cover_url) || 'data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22 width=%22200%22 height=%22150%22%3E%3Crect fill=%22%23333%22 width=%22200%22 height=%22150%22/%3E%3Ctext fill=%22%23666%22 x=%2250%25%22 y=%2250%25%22 text-anchor=%22middle%22%3ENo Image%3C/text%3E%3C/svg%3E'}" 
                 alt="${game.title}" 
                 class="deal-cover"
                 onerror="this.src='/images/game-placeholder.svg'">
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
                <div class="meta-row">
                    <span>${game.cheapest_region || 'India'}</span>
                    <span>${(game.payment_methods || ['Card']).slice(0, 2).join(' / ')}</span>
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
