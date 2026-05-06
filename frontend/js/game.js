document.addEventListener('DOMContentLoaded', () => {
    console.log('Game Tracking Metrics UI Loaded');
    initGamePage();
});

window.priceChart = null;

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
        loadReviewScores(gameID),
        loadEnhancedGameData(gameID)
    ]);
    initWishlistButton(gameID);
    initChartPeriodButtons();
}

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
            const linkHtml = source.url ? `
                <a href="${source.url}" target="_blank" rel="noopener noreferrer" class="source-link">
                    Read review
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
                </a>
            ` : '';
            return `
                <div class="source-item">
                    <div>
                        <span class="source-name">${capitalizeFirst(source.source)}</span>
                        <div class="source-score">
                            <div class="source-bar">
                                <div class="source-bar-fill ${sourceColor}" style="width: ${source.score}%"></div>
                            </div>
                            <span class="source-value">${source.score}</span>
                        </div>
                        ${linkHtml}
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

    // Populate arbitrage table
    const indiaBase = document.getElementById('arbitrage-india-base');
    const indiaGst = document.getElementById('arbitrage-india-gst');
    const indiaTotal = document.getElementById('arbitrage-india-total');
    const globalBase = document.getElementById('arbitrage-global-base');
    const globalGst = document.getElementById('arbitrage-global-gst');
    const globalTotal = document.getElementById('arbitrage-global-total');

    if (game.arbitrage) {
        if (indiaBase) indiaBase.textContent = formatINR(game.arbitrage.india_base_inr);
        if (indiaGst) indiaGst.textContent = `+ ${formatINR(game.arbitrage.india_gst_inr)}`;
        if (indiaTotal) indiaTotal.textContent = formatINR(game.arbitrage.india_total_inr);
        if (globalBase) globalBase.textContent = formatINR(game.arbitrage.global_base_inr);
        if (globalGst) globalGst.textContent = `+ ${formatINR(game.arbitrage.global_gst_inr)}`;
        if (globalTotal) globalTotal.textContent = formatINR(game.arbitrage.global_total_inr);

        // Highlight cheapest row
        const indiaRow = document.getElementById('arbitrage-row-india');
        const globalRow = document.getElementById('arbitrage-row-global');
        const cheapestBadge = document.getElementById('arbitrage-cheapest-badge');

        if (indiaRow) indiaRow.classList.remove('cheapest');
        if (globalRow) globalRow.classList.remove('cheapest');

        if (game.arbitrage.cheapest_region === 'india' && indiaRow) {
            indiaRow.classList.add('cheapest');
            if (cheapestBadge) cheapestBadge.innerHTML = '<span class="cheapest-badge">Cheapest Region</span>';
        } else if (game.arbitrage.cheapest_region === 'global' && globalRow) {
            globalRow.classList.add('cheapest');
            if (cheapestBadge) cheapestBadge.innerHTML = '<span class="cheapest-badge">Cheapest Region</span>';
        } else {
            if (cheapestBadge) cheapestBadge.innerHTML = '';
        }
    } else {
        // Fallback: use current price as India price, show placeholder for global
        if (indiaBase) indiaBase.textContent = formatINR(game.price_inr);
        if (indiaGst) indiaGst.textContent = '+ --';
        if (indiaTotal) indiaTotal.textContent = formatINR(game.price_inr);
        if (globalBase) globalBase.textContent = '--';
        if (globalGst) globalGst.textContent = '+ --';
        if (globalTotal) globalTotal.textContent = '--';
    }

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

async function loadEnhancedGameData(gameID) {
    try {
        const response = await fetch(`/api/games/${gameID}/enhanced`);
        if (!response.ok) {
            throw new Error('Failed to load enhanced game data');
        }
        const data = await response.json();
        
        populateScreenshots(data.screenshots || []);
        populateTrailers(data.trailers || []);
        populateSystemRequirements(data.system_requirements);
        populatePlatforms(data.platforms || []);
        populatePriceComparison(data.prices || []);
    } catch (error) {
        console.error('Failed to load enhanced game data:', error);
        // Hide skeleton loaders on error
        document.getElementById('screenshots-gallery').innerHTML = '<p style="color: var(--color-text-muted);">Screenshots unavailable</p>';
        document.getElementById('trailers-section').innerHTML = '<p style="color: var(--color-text-muted);">Trailers unavailable</p>';
        document.getElementById('system-requirements').innerHTML = '<p style="color: var(--color-text-muted);">System requirements unavailable</p>';
        document.getElementById('platform-badges').innerHTML = '<p style="color: var(--color-text-muted);">Platform info unavailable</p>';
        document.getElementById('price-comparison-body').innerHTML = '<tr><td colspan="4" style="text-align: center; padding: var(--space-4); color: var(--color-text-muted);">Price comparison unavailable</td></tr>';
    }
}

async function populateScreenshots(screenshots) {
    const gallery = document.getElementById('screenshots-gallery');
    if (!gallery) return;

    if (!screenshots || screenshots.length === 0) {
        gallery.innerHTML = '<p style="color: var(--color-text-muted);">No screenshots available</p>';
        return;
    }

    gallery.innerHTML = '';
    for (const screenshot of screenshots.slice(0, 6)) {
        const proxiedUrl = getProxiedImageUrl(screenshot);
        const img = document.createElement('img');
        img.src = proxiedUrl || screenshot;
        img.alt = 'Game screenshot';
        img.style.cssText = 'width: 100%; height: 100%; object-fit: cover; border-radius: var(--radius-md); cursor: pointer; transition: transform 0.2s;';
        img.onclick = () => openImageModal(screenshot);
        
        const container = document.createElement('div');
        container.style.cssText = 'aspect-ratio: 16/9; border-radius: var(--radius-md); overflow: hidden; background: var(--color-surface-secondary);';
        container.appendChild(img);
        gallery.appendChild(container);
    }
}

async function populateTrailers(trailers) {
    const section = document.getElementById('trailers-section');
    if (!section) return;

    if (!trailers || trailers.length === 0) {
        section.innerHTML = '<p style="color: var(--color-text-muted);">No trailers available</p>';
        return;
    }

    section.innerHTML = '';
    for (const trailer of trailers.slice(0, 3)) {
        const videoContainer = document.createElement('div');
        videoContainer.style.cssText = 'position: relative; aspect-ratio: 16/9; border-radius: var(--radius-md); overflow: hidden; background: var(--color-surface-secondary);';
        
        // Check if it's a YouTube URL
        if (trailer.includes('youtube.com') || trailer.includes('youtu.be')) {
            const videoId = extractYouTubeId(trailer);
            if (videoId) {
                const iframe = document.createElement('iframe');
                iframe.src = `https://www.youtube.com/embed/${videoId}`;
                iframe.style.cssText = 'width: 100%; height: 100%; border: none;';
                iframe.allow = 'accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture';
                iframe.allowFullscreen = true;
                videoContainer.appendChild(iframe);
            }
        } else {
            // Fallback for other video types
            const link = document.createElement('a');
            link.href = trailer;
            link.target = '_blank';
            link.style.cssText = 'display: flex; align-items: center; justify-content: center; height: 100%; text-decoration: none; color: var(--color-text-primary);';
            link.innerHTML = `
                <div style="text-align: center;">
                    <svg width="48" height="48" viewBox="0 0 24 24" fill="currentColor" style="margin-bottom: var(--space-2);"><path d="M8 5v14l11-7z"/></svg>
                    <span>Watch Trailer</span>
                </div>
            `;
            videoContainer.appendChild(link);
        }
        
        section.appendChild(videoContainer);
    }
}

function extractYouTubeId(url) {
    const regex = /(?:youtube\.com\/(?:[^\/]+\/.+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([^"&?\/\s]{11})/;
    const match = url.match(regex);
    return match ? match[1] : null;
}

function populateSystemRequirements(requirements) {
    const minContainer = document.getElementById('min-requirements');
    const recContainer = document.getElementById('rec-requirements');
    
    if (!requirements) {
        if (minContainer) minContainer.innerHTML = '<p style="color: var(--color-text-muted);">System requirements unavailable</p>';
        if (recContainer) recContainer.innerHTML = '';
        return;
    }

    if (minContainer && requirements.minimum) {
        minContainer.innerHTML = formatRequirements(requirements.minimum);
    }
    
    if (recContainer && requirements.recommended) {
        recContainer.innerHTML = formatRequirements(requirements.recommended);
    }
}

function formatRequirements(reqs) {
    if (typeof reqs === 'string') {
        return `<p style="line-height: 1.6; font-size: var(--text-sm);">${reqs}</p>`;
    }
    
    let html = '';
    if (reqs.os) html += `<p style="line-height: 1.6; font-size: var(--text-sm);"><strong>OS:</strong> ${reqs.os}</p>`;
    if (reqs.processor) html += `<p style="line-height: 1.6; font-size: var(--text-sm);"><strong>Processor:</strong> ${reqs.processor}</p>`;
    if (reqs.memory) html += `<p style="line-height: 1.6; font-size: var(--text-sm);"><strong>Memory:</strong> ${reqs.memory}</p>`;
    if (reqs.graphics) html += `<p style="line-height: 1.6; font-size: var(--text-sm);"><strong>Graphics:</strong> ${reqs.graphics}</p>`;
    if (reqs.storage) html += `<p style="line-height: 1.6; font-size: var(--text-sm);"><strong>Storage:</strong> ${reqs.storage}</p>`;
    
    return html || '<p style="color: var(--color-text-muted);">Requirements unavailable</p>';
}

function populatePlatforms(platforms) {
    const container = document.getElementById('platform-badges');
    if (!container) return;

    if (!platforms || platforms.length === 0) {
        container.innerHTML = '<p style="color: var(--color-text-muted);">Platform info unavailable</p>';
        return;
    }

    const platformIcons = {
        'PC': '💻',
        'PlayStation': '🎮',
        'Xbox': '🎯',
        'Switch': '🔀',
        'Mobile': '📱'
    };

    container.innerHTML = platforms.map(platform => {
        const icon = platformIcons[platform] || '🎮';
        return `<span class="badge badge-primary" style="font-size: var(--text-base); padding: var(--space-2) var(--space-4);">${icon} ${platform}</span>`;
    }).join('');
}

function populatePriceComparison(prices) {
    const tbody = document.getElementById('price-comparison-body');
    if (!tbody) return;

    if (!prices || prices.length === 0) {
        tbody.innerHTML = '<tr><td colspan="4" style="text-align: center; padding: var(--space-4); color: var(--color-text-muted);">No price data available</td></tr>';
        return;
    }

    // Sort by price (lowest first)
    const sortedPrices = [...prices].sort((a, b) => a.price_inr - b.price_inr);
    const lowestPrice = sortedPrices[0];

    tbody.innerHTML = sortedPrices.map((price, index) => {
        const isLowest = price.price_inr === lowestPrice.price_inr;
        const discountBadge = price.discount_percent > 0 
            ? `<span class="badge badge-success" style="font-size: var(--text-sm);">-${price.discount_percent}%</span>`
            : '<span style="color: var(--color-text-muted); font-size: var(--text-sm);">No discount</span>';
        
        const storeIcon = getStoreIcon(price.store?.slug || 'store');
        const rowStyle = isLowest ? 'background: var(--color-surface-highlight);' : '';
        
        return `
            <tr style="border-bottom: 1px solid var(--color-border); ${rowStyle}">
                <td style="padding: var(--space-3);">
                    <div style="display: flex; align-items: center; gap: var(--space-2);">
                        ${storeIcon}
                        <span style="font-weight: 500;">${price.store?.name || 'Store'}</span>
                        ${isLowest ? '<span class="badge badge-accent" style="font-size: var(--text-xs);">Best Price</span>' : ''}
                    </div>
                </td>
                <td style="text-align: right; padding: var(--space-3);">
                    <span style="font-family: var(--font-display); font-weight: 700; font-size: var(--text-lg);">${formatINR(price.price_inr)}</span>
                </td>
                <td style="text-align: right; padding: var(--space-3);">${discountBadge}</td>
                <td style="text-align: center; padding: var(--space-3);">
                    <a href="${price.store?.website_url || '#'}" target="_blank" rel="noopener noreferrer" class="btn btn-sm btn-primary">
                        Buy Now
                    </a>
                </td>
            </tr>
        `;
    }).join('');
}

function getStoreIcon(slug) {
    const icons = {
        'steam': '⚡',
        'epic': '🎮',
        'greenmangaming': '🌿',
        'fanatical': '🔥',
        'humble': '📦',
        'indian': '🇮🇳'
    };
    return `<span style="font-size: 1.2em;">${icons[slug] || '🛒'}</span>`;
}

function openImageModal(imageUrl) {
    const modal = document.createElement('div');
    modal.style.cssText = 'position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.9); display: flex; align-items: center; justify-content: center; z-index: 9999; cursor: pointer;';
    modal.onclick = () => modal.remove();
    
    const img = document.createElement('img');
    img.src = imageUrl;
    img.style.cssText = 'max-width: 90%; max-height: 90%; object-fit: contain;';
    
    modal.appendChild(img);
    document.body.appendChild(modal);
}
