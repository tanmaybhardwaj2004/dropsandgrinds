document.addEventListener('DOMContentLoaded', () => {
    console.log('Game Tracking Metrics UI Loaded');
    initGamePage();
});

async function initGamePage() {
    const gameID = getGameIDFromURL();
    if (!gameID) {
        renderGameError('Invalid game id in URL');
        return;
    }

    initAuthButton();
    await Promise.all([loadGameDetails(gameID), loadBuyAdvice(gameID)]);
    initWishlistButton(gameID);
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
        document.getElementById('game-cover').src = game.cover_url || '';
        document.getElementById('game-cover').alt = `${game.title} cover`;

        document.getElementById('main-price').textContent = formatINR(game.price_inr);
        document.getElementById('original-price').textContent = formatINR(game.original_inr);
        document.getElementById('discount-badge').textContent = `-${game.discount_percent || 0}%`;
        document.getElementById('store-badge').textContent = game.platform || 'Store';

        const arbitrage = document.getElementById('arbitrage-cost');
        arbitrage.textContent = `${formatINR(game.price_inr)}.00`;

        const verdict = document.getElementById('arbitrage-verdict');
        if (game.is_all_time_low) {
            verdict.textContent = 'VERDICT: All-time low detected. Excellent buy window.';
            verdict.classList.add('green');
        } else {
            verdict.textContent = `VERDICT: Current lowest seen: ${formatINR(game.lowest_price_inr)}.`;
        }
    } catch (error) {
        renderGameError(error.message);
    }
}

async function loadBuyAdvice(gameID) {
    try {
        const response = await fetch(`/api/games/${gameID}/buy-advice`);
        const advice = await response.json();
        if (!response.ok) {
            throw new Error(advice.error || 'Failed to load buy advice');
        }

        const verdict = document.getElementById('timeline-verdict');
        const recommendation = (advice.recommendation || 'unknown').toUpperCase();
        verdict.textContent = `${recommendation}: ${advice.reason || 'No recommendation available.'}`;

        const activeCost = document.querySelector('.current-node-cost');
        if (activeCost) {
            activeCost.textContent = formatINR(advice.current_price_inr);
        }

        const futureNode = document.querySelector('.time-node.future .node-cost');
        if (futureNode) {
            futureNode.textContent = `Predicted: ${formatINR(advice.lowest_price_inr)}`;
        }

        const confidence = document.querySelector('.time-node.future .confidence');
        if (confidence) {
            confidence.textContent = `${advice.confidence_percent || 0}% Confidence`;
        }
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
