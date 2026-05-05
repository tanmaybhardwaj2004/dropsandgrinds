document.addEventListener('DOMContentLoaded', () => {
    console.log('Game Tracking Metrics UI Loaded');
    initGamePage();
});

let priceChart = null;

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

async function initGamePage() {
    const gameID = getGameIDFromURL();
    if (!gameID) {
        renderGameError('Invalid game id in URL');
        return;
    }

    initAuthButton();

    // Try cache first for instant render
    const cached = await getCachedGameMeta(gameID);
    if (cached) {
        populateGameDetails(cached, { fromCache: true });
    }

    // Fetch fresh data in parallel
    await Promise.all([
        loadGameDetails(gameID),
        loadBuyAdvice(gameID),
        loadPriceHistory(gameID),
        loadReviewScores(gameID)
    ]);
    initWishlistButton(gameID);
}

let priceChart = null;
let currentGameID = null;

async function loadPriceHistory(gameID, days = 30) {
    currentGameID = gameID;
    const loadingEl = document.getElementById('price-chart-loading');
    const errorEl = document.getElementById('price-chart-error');
    const canvas = document.getElementById('priceChart');
    
    // Show loading, hide error
    if (loadingEl) loadingEl.style.display = 'flex';
    if (errorEl) errorEl.style.display = 'none';
    if (canvas) canvas.style.opacity = '0.3';
    
    try {
        const response = await fetch(`/api/prices/${gameID}/history?limit=${days}`);
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || 'Failed to load price history');
        }
        
        const data = await response.json();
        renderPriceChart(data.prices || [], days);
        updatePriceStats(data.prices || []);
        
        // Hide loading
        if (loadingEl) loadingEl.style.display = 'none';
        if (canvas) canvas.style.opacity = '1';
    } catch (error) {
        console.error('Failed to load price history:', error);
        if (loadingEl) loadingEl.style.display = 'none';
        if (errorEl) errorEl.style.display = 'flex';
        if (canvas) canvas.style.opacity = '0';
    }
}

function updatePriceStats(prices) {
    if (!prices || prices.length === 0) return;
    
    const priceValues = prices.map(p => p.price_inr);
    const lowest = Math.min(...priceValues);
    const highest = Math.max(...priceValues);
    const average = Math.round(priceValues.reduce((a, b) => a + b, 0) / priceValues.length);
    const drop = highest - lowest;
    
    document.getElementById('stat-lowest').textContent = `₹${lowest.toLocaleString()}`;
    document.getElementById('stat-highest').textContent = `₹${highest.toLocaleString()}`;
    document.getElementById('stat-average').textContent = `₹${average.toLocaleString()}`;
    document.getElementById('stat-drop').textContent = `₹${drop.toLocaleString()}`;
}

function initChartPeriodButtons() {
    const buttons = document.querySelectorAll('.chart-period-btn');
    buttons.forEach(btn => {
        btn.addEventListener('click', () => {
            buttons.forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            const days = parseInt(btn.dataset.days);
            if (currentGameID) {
                loadPriceHistory(currentGameID, days);
            }
        });
    });
}

function renderPriceChart(prices) {
    const ctx = document.getElementById('priceChart');
    if (!ctx) return;

    const labels = prices.map(p => new Date(p.fetched_at).toLocaleDateString());
    const dataPoints = prices.map(p => p.price_inr);

    // Destroy existing chart if present
    if (priceChart) {
        priceChart.destroy();
    }

    priceChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'Price (₹)',
                data: dataPoints,
                borderColor: '#58a6ff',
                backgroundColor: 'rgba(88, 166, 255, 0.1)',
                borderWidth: 2,
                fill: true,
                tension: 0.4,
                pointRadius: 4,
                pointBackgroundColor: '#58a6ff',
                pointBorderColor: '#fff',
                pointBorderWidth: 2
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: { display: false },
                tooltip: {
                    backgroundColor: '#161b22',
                    titleColor: '#c9d1d9',
                    bodyColor: '#c9d1d9',
                    borderColor: '#30363d',
                    borderWidth: 1,
                    callbacks: {
                        label: (context) => `₹${context.parsed.y}`
                    }
                }
            },
            scales: {
                x: {
                    grid: { color: '#21262d' },
                    ticks: { color: '#8b949e', maxTicksLimit: 6 }
                },
                y: {
                    grid: { color: '#21262d' },
                    ticks: { 
                        color: '#8b949e',
                        callback: (value) => `₹${value}`
                    }
                }
            }
        }
    });
}

async function loadReviewScores(gameID) {
    try {
        const response = await fetch(`/api/games/${gameID}/reviews`);
        const data = await response.json();
        if (!response.ok) {
            throw new Error(data.error || 'Failed to load review scores');
        }

        renderReviewScores(data);
    } catch (error) {
        console.error('Failed to load review scores:', error);
        document.getElementById('review-aggregate').innerHTML = `
            <div style="color:var(--text-muted);padding:32px;">Review scores unavailable</div>
        `;
    }
}

function getScoreColor(score) {
    if (score >= 85) return 'green';
    if (score >= 70) return 'amber';
    if (score >= 50) return 'orange';
    return 'red';
}

function renderReviewScores(data) {
    const scoreRing = document.getElementById('review-score-ring');
    const scoreEl = document.getElementById('review-score');
    const labelEl = document.getElementById('review-label');
    const countEl = document.getElementById('review-source-count');
    const sourcesContainer = document.getElementById('review-sources');

    if (!data.score || data.score === 0) {
        scoreRing.className = 'score-ring gray';
        scoreEl.textContent = '--';
        labelEl.textContent = 'No reviews yet';
        countEl.textContent = '';
        sourcesContainer.innerHTML = '';
        return;
    }

    const color = getScoreColor(data.score);
    scoreRing.className = `score-ring ${color}`;
    scoreEl.textContent = data.score;
    labelEl.textContent = data.label || 'Mixed';
    countEl.textContent = `Based on ${data.source_count || 0} review sources`;

    if (data.sources && data.sources.length > 0) {
        sourcesContainer.innerHTML = data.sources.map(source => {
            const sourceColor = getScoreColor(source.score);
            return `
                <div class="source-item">
                    <span class="source-name">${capitalizeFirst(source.source)}</span>
                    <div class="source-score">
                        <div class="source-bar">
                            <div class="source-bar-fill ${sourceColor}" style="width: ${source.score}%"></div>
                        </div>
                        <span class="source-value">${source.score}</span>
                    </div>
                </div>
            `;
        }).join('');
    }
}

function capitalizeFirst(str) {
    return str.replace(/\b\w/g, l => l.toUpperCase());
}

function getGameIDFromURL() {
    const url = new URL(window.location.href);
    const raw = url.searchParams.get('id');
    const parsed = Number(raw);
    if (!raw || Number.isNaN(parsed) || parsed <= 0) {
        return null;
    }
    return parsed;
}

function formatINR(value) {
    return new Intl.NumberFormat('en-IN', {
        style: 'currency',
        currency: 'INR',
        maximumFractionDigits: 0,
    }).format(value || 0);
}

async function loadGameDetails(gameID) {
    try {
        const response = await fetch(`/api/games/${gameID}`);
        const game = await response.json();
        if (!response.ok) {
            throw new Error(game.error || 'Failed to load game');
        }

        // Cache the fetched data
        if (window.imageCache) {
            window.imageCache.cacheGameMeta(gameID, game);
        }

        populateGameDetails(game, { fromCache: false });
    } catch (error) {
        renderGameError(error.message);
    }
}

function populateGameDetails(game, opts = {}) {
    // Update page title
    document.title = `${game.title || 'Game'} - DropsAndGrinds`;

    // Title
    const titleEl = document.getElementById('game-title');
    if (titleEl) titleEl.textContent = game.title || 'Unknown Game';

    // Cover image (use cache if available)
    const coverEl = document.getElementById('game-cover');
    const coverSkeleton = document.getElementById('cover-skeleton');
    if (coverEl && game.cover_url) {
        const proxiedUrl = getProxiedImageUrl(game.cover_url);
        coverEl.style.display = 'block';
        if (coverSkeleton) coverSkeleton.style.display = 'none';

        if (window.imageCache) {
            window.imageCache.fetchCachedImage(proxiedUrl || game.cover_url).then(objectUrl => {
                coverEl.src = objectUrl;
                coverEl.alt = `${game.title} cover`;
            });
        } else {
            coverEl.src = proxiedUrl || game.cover_url;
            coverEl.alt = `${game.title} cover`;
        }
    }

    // Genre badges (if genres exist in API response)
    const genreContainer = document.getElementById('genre-badges');
    if (genreContainer && game.genres && game.genres.length > 0) {
        genreContainer.innerHTML = game.genres.map(g =>
            `<span class="badge badge-primary">${g}</span>`
        ).join('');
    }

    // Discount badge
    const discountBadge = document.getElementById('discount-badge');
    if (discountBadge) {
        if (game.discount_percent > 0) {
            discountBadge.textContent = `-${game.discount_percent}%`;
            discountBadge.style.display = 'inline-flex';
        } else {
            discountBadge.style.display = 'none';
        }
    }

    // Description
    const descEl = document.getElementById('game-description');
    if (descEl && game.description) {
        descEl.textContent = game.description;
    }

    // Store badge
    const storeBadge = document.getElementById('store-badge');
    if (storeBadge) storeBadge.textContent = game.platform || 'Store';

    // Prices
    const mainPrice = document.getElementById('main-price');
    if (mainPrice) mainPrice.textContent = formatINR(game.price_inr);

    const originalPrice = document.getElementById('original-price');
    if (originalPrice) originalPrice.textContent = formatINR(game.original_inr);

    // Savings badge
    const savingsBadge = document.getElementById('savings-badge');
    if (savingsBadge) {
        const savings = (game.original_inr || 0) - (game.price_inr || 0);
        if (savings > 0) {
            savingsBadge.textContent = `Save ${formatINR(savings)}`;
            savingsBadge.style.display = 'inline-flex';
        } else {
            savingsBadge.style.display = 'none';
        }
    }

    // Arbitrage: switch from skeleton to loaded view
    const arbLoading = document.getElementById('arbitrage-loading');
    const arbLoaded = document.getElementById('arbitrage-loaded');
    if (arbLoading) arbLoading.style.display = 'none';
    if (arbLoaded) arbLoaded.style.display = 'block';

    // Populate arbitrage loaded view
    const arbCostFinal = document.getElementById('arbitrage-cost-final');
    if (arbCostFinal) arbCostFinal.textContent = formatINR(game.price_inr);

    const arbBase = document.getElementById('arbitrage-base');
    if (arbBase && game.global_base_inr) arbBase.textContent = formatINR(game.global_base_inr);

    const arbGst = document.getElementById('arbitrage-gst');
    if (arbGst && game.gst_amount) arbGst.textContent = `+ ${formatINR(game.gst_amount)}`;

    const arbGlobal = document.getElementById('arbitrage-global');
    if (arbGlobal && game.global_total_inr) arbGlobal.textContent = formatINR(game.global_total_inr);

    // Arbitrage verdict
    const verdict = document.getElementById('arbitrage-verdict');
    if (verdict) {
        verdict.innerHTML = '';
        if (game.is_all_time_low) {
            verdict.textContent = 'VERDICT: All-time low detected. Excellent buy window.';
            verdict.className = 'arbitrage-verdict green';
        } else {
            verdict.textContent = `VERDICT: Current lowest seen: ${formatINR(game.lowest_price_inr || game.price_inr)}.`;
            verdict.className = 'arbitrage-verdict';
        }
    }
}

async function loadBuyAdvice(gameID) {
    try {
        const response = await fetch(`/api/games/${gameID}/buy-advice`);
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || 'Failed to load buy advice');
        }
        const advice = await response.json();

        // Switch from skeleton to loaded timeline
        const tlLoading = document.getElementById('timeline-loading');
        const tlLoaded = document.getElementById('timeline-loaded');
        if (tlLoading) tlLoading.style.display = 'none';
        if (tlLoaded) tlLoaded.style.display = 'block';

        const verdict = document.getElementById('timeline-verdict');
        if (verdict) {
            const recommendation = (advice.recommendation || 'unknown').toUpperCase();
            verdict.textContent = `${recommendation}: ${advice.reason || 'No recommendation available.'}`;
        }

        const activeCost = document.querySelector('.current-node-cost');
        if (activeCost) {
            activeCost.textContent = formatINR(advice.current_price_inr);
        }

        const pastTitle = document.getElementById('timeline-past-title');
        if (pastTitle) pastTitle.textContent = advice.past_sale_name || 'Previous Sale';

        const pastPrice = document.getElementById('timeline-past-price');
        if (pastPrice) pastPrice.textContent = formatINR(advice.past_sale_price_inr);

        const futureTitle = document.getElementById('timeline-future-title');
        if (futureTitle) futureTitle.textContent = advice.next_sale_name || 'Expected Drop';

        const futureNode = document.getElementById('timeline-future-price');
        if (futureNode) {
            futureNode.textContent = `Predicted: ${formatINR(advice.lowest_price_inr)}`;
        }

        const confidence = document.getElementById('timeline-confidence');
        if (confidence) {
            confidence.textContent = `${advice.confidence_percent || 0}% Confidence`;
        }
    } catch (error) {
        const verdict = document.getElementById('timeline-verdict');
        if (verdict) verdict.textContent = `UNKNOWN: ${error.message}`;
    }
}

function getAccessToken() {
    if (window.authState?.accessToken) {
        return window.authState.accessToken;
    }
    return sessionStorage.getItem('dropsandgrinds_access_token');
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

function initWishlistButton(gameID) {
    const btn = document.getElementById('wishlist-btn');
    const targetInput = document.getElementById('target-price-input');
    if (!btn) return;

    btn.addEventListener('click', async () => {
        const token = getAccessToken();
        if (!token) {
            window.location.href = 'login.html';
            return;
        }

        btn.disabled = true;
        const oldText = btn.textContent;
        btn.textContent = 'Adding...';
        const targetPrice = Number(targetInput?.value || 0);

        if (!Number.isInteger(targetPrice) || targetPrice <= 0) {
            btn.textContent = oldText;
            btn.disabled = false;
            alert('Please enter a valid target price (INR).');
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
                    target_price_inr: targetPrice,
                }),
            });

            if (response.status === 401) {
                window.location.href = 'login.html';
                return;
            }

            if (response.status === 409) {
                btn.classList.add('active');
                btn.textContent = '♥ Already Wishlisted';
                return;
            }

            const payload = await response.json();
            if (!response.ok) {
                throw new Error(payload.error || 'Failed to add wishlist item');
            }

            btn.classList.add('active');
            btn.textContent = '♥ Wishlisted';
        } catch (err) {
            btn.textContent = oldText;
            alert(err.message || 'Could not update wishlist');
        } finally {
            btn.disabled = false;
        }
    });
}

function renderGameError(message) {
    const title = document.getElementById('game-title');
    if (title) {
        title.textContent = 'Game unavailable';
    }
    const desc = document.querySelector('.description');
    if (desc) {
        desc.textContent = message;
    }
}
