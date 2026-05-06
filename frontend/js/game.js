document.addEventListener('DOMContentLoaded', () => {
    console.log('Game Tracking Metrics UI Loaded');
    initGamePage();
});

let priceChart = null;
let currentGame = null;

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
    await Promise.all([
        loadGameDetails(gameID), 
        loadBuyAdvice(gameID),
        loadPriceHistory(gameID),
        loadReviewScores(gameID)
    ]);
    initWishlistButton(gameID);
}

async function loadPriceHistory(gameID) {
    try {
        const response = await fetch(`/api/prices/${gameID}/history?limit=30`);
        const data = await response.json();
        if (!response.ok) {
            throw new Error(data.error || 'Failed to load price history');
        }

        renderPriceChart(data.prices || []);
    } catch (error) {
        console.error('Failed to load price history:', error);
        document.getElementById('price-chart-container').innerHTML = `
            <div style="display:flex;align-items:center;justify-content:center;height:100%;color:var(--text-muted);">
                Price history unavailable
            </div>
        `;
    }
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

        document.getElementById('game-title').textContent = game.title;
        document.title = `${game.title} - DropsAndGrinds`;
        document.getElementById('game-cover').src = getProxiedImageUrl(game.cover_url) || '';
        document.getElementById('game-cover').alt = `${game.title} cover`;
        currentGame = game;

        document.getElementById('main-price').textContent = formatINR(game.price_inr);
        document.getElementById('original-price').textContent = formatINR(game.original_inr);
        const discountBadge = document.getElementById('discount-badge');
        if (discountBadge) discountBadge.textContent = game.discount_percent > 0 ? `-${game.discount_percent}%` : 'Current price';
        const savingsBadge = document.getElementById('savings-badge');
        if (savingsBadge) {
            const savings = Math.max(0, (game.original_inr || 0) - (game.price_inr || 0));
            savingsBadge.textContent = savings > 0 ? `Save ${formatINR(savings)}` : 'Live price';
        }
        document.getElementById('store-badge').textContent = game.platform || 'Store';
        const desc = document.querySelector('.description');
        if (desc) {
            desc.textContent = `${game.title} is tracked on ${game.platform || 'this store'} with current price, historical low, India pricing, and review data from DropsAndGrinds.`;
        }
        const targetInput = document.getElementById('target-price-input');
        if (targetInput && game.lowest_price_inr > 0) {
            targetInput.value = game.lowest_price_inr;
        }
        initStoreLink(gameID, game.platform || 'steam');

        await loadIndiaArbitrage(gameID);
    } catch (error) {
        renderGameError(error.message);
    }
}

function initStoreLink(gameID, platform) {
    const link = document.getElementById('store-link');
    if (!link) return;
    link.textContent = '';
    link.innerHTML = `<i data-lucide="external-link"></i> Open in ${platform || 'Store'}`;
    link.onclick = async (event) => {
        event.preventDefault();
        try {
            const token = getAccessToken();
            const headers = token ? { Authorization: `Bearer ${token}` } : {};
            const response = await fetch(`/api/games/${gameID}/redirect?platform=${encodeURIComponent(platform || '')}`, { headers });
            const payload = await response.json();
            if (!response.ok) {
                throw new Error(payload.error || 'Store URL unavailable');
            }
            window.location.href = payload.url;
        } catch (error) {
            alert(error.message || 'Store URL unavailable');
        }
    };
    if (window.lucide) window.lucide.createIcons();
}

async function loadIndiaArbitrage(gameID) {
    try {
        const response = await fetch(`/api/prices/${gameID}/india`);
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || 'Failed to load India pricing');
        const rows = document.querySelectorAll('.receipt-row span:last-child');
        if (rows[0]) rows[0].textContent = formatINR(data.steam_global_inr);
        if (rows[1]) rows[1].textContent = `+ ${formatINR(data.gst_amount)}`;
        if (rows[2]) rows[2].textContent = formatINR(data.total_with_gst);
        const arbitrage = document.getElementById('arbitrage-cost');
        if (arbitrage) arbitrage.textContent = formatINR(data.steam_india_price);
        const verdict = document.getElementById('arbitrage-verdict');
        if (verdict) {
            verdict.textContent = `${data.cheapest_region}: ${data.verdict}`;
            verdict.classList.toggle('green', data.cheapest_region === 'India');
        }
    } catch (error) {
        const verdict = document.getElementById('arbitrage-verdict');
        if (verdict) verdict.textContent = error.message;
    }
}

async function loadBuyAdvice(gameID) {
    try {
        const response = await fetch(`/api/games/${gameID}/buy-timing`);
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || 'Failed to load buy advice');
        }
        const advice = await response.json();

        const verdict = document.getElementById('timeline-verdict');
        const recommendation = (advice.recommendation || 'unknown').toUpperCase();
        verdict.textContent = `${recommendation}: ${advice.reason || 'No recommendation available.'}`;

        const activeTitle = document.querySelector('.time-node.active h4');
        const activeCost = document.querySelector('.current-node-cost');
        if (activeTitle) activeTitle.textContent = advice.active_sale ? advice.active_sale.name : 'Current price';
        if (activeCost) activeCost.textContent = currentGame ? formatINR(currentGame.price_inr) : 'Live API';

        const futureTitle = document.querySelector('.time-node.future h4');
        const futureNode = document.querySelector('.time-node.future .node-cost');
        if (futureTitle) futureTitle.textContent = advice.next_sale ? advice.next_sale.name : 'Next sale window';
        if (futureNode) futureNode.textContent = advice.days_until_sale ? `Starts in ${advice.days_until_sale} days` : 'No scheduled sale';

        const confidence = document.querySelector('.time-node.future .confidence');
        if (confidence) confidence.textContent = 'Sale calendar signal';
    } catch (error) {
        const verdict = document.getElementById('timeline-verdict');
        verdict.textContent = `UNKNOWN: ${error.message}`;
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
                btn.textContent = 'Already Wishlisted';
                return;
            }

            const payload = await response.json();
            if (!response.ok) {
                throw new Error(payload.error || 'Failed to add wishlist item');
            }

            btn.classList.add('active');
            btn.textContent = 'Wishlisted';
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
