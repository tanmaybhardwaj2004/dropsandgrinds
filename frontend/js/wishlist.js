document.addEventListener('DOMContentLoaded', () => {
    console.log('Wishlist page loaded');
    initAuthButton();
    checkAuthAndLoadWishlist();
});

// Transform external image URLs to use local proxy (bypasses hotlink protection)
function getProxiedImageUrl(originalUrl) {
    if (!originalUrl) return '';
    
    if (originalUrl.includes('steamstatic.com')) {
        return originalUrl.replace('https://shared.cloudflare.steamstatic.com/', '/img/steam/');
    }
    if (originalUrl.includes('gog-statics.com')) {
        return originalUrl.replace('https://images.gog-statics.com/', '/img/gog/');
    }
    if (originalUrl.includes('unrealengine.com')) {
        return originalUrl.replace('https://cdn2.unrealengine.com/', '/img/epic/');
    }
    
    return originalUrl;
}

function checkAuthAndLoadWishlist() {
    const token = getAccessToken();
    if (!token) {
        // Redirect to login if not authenticated
        window.location.href = 'login.html';
        return;
    }
    loadWishlist();
}

async function loadWishlist() {
    const container = document.getElementById('wishlist-container');
    const loadingState = document.getElementById('loading-state');
    const emptyState = document.getElementById('empty-state');
    const errorState = document.getElementById('error-state');

    // Show loading state
    if (container) container.style.display = 'none';
    if (loadingState) loadingState.style.display = 'block';
    if (emptyState) emptyState.style.display = 'none';
    if (errorState) errorState.style.display = 'none';

    try {
        const token = getAccessToken();
        const response = await fetch('/api/wishlist?limit=50&offset=0', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Failed to load wishlist');
        }

        // Hide loading state
        if (loadingState) loadingState.style.display = 'none';

        if (data.items && data.items.length > 0) {
            if (container) {
                container.style.display = 'grid';
                container.innerHTML = data.items.map(item => renderWishlistItem(item)).join('');
            }
        } else {
            if (emptyState) emptyState.style.display = 'block';
        }

    } catch (error) {
        console.error('Failed to load wishlist:', error);
        if (loadingState) loadingState.style.display = 'none';
        if (errorState) errorState.style.display = 'block';
    }
}

function renderWishlistItem(item) {
    const alertClass = item.triggered ? 'triggered' : 'waiting';
    const alertText = item.triggered ? 'Price Alert Triggered!' : 'Waiting for price drop';

    return `
        <div class="wishlist-card" data-id="${item.id}">
            <img src="${getProxiedImageUrl(item.cover_url)}" class="wishlist-cover" alt="${item.title} cover" onerror="this.src='data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22 width=%22200%22 height=%22150%22%3E%3Crect fill=%22%23333%22 width=%22200%22 height=%22150%22/%3E%3Ctext fill=%22%23666%22 x=%2250%25%22 y=%2250%25%22 text-anchor=%22middle%22%3ENo Image%3C/text%3E%3C/svg%3E'">
            <div class="wishlist-info">
                <div class="wishlist-title">${item.title}</div>
                <div class="wishlist-platform">${item.platform}</div>
                <div class="alert-status ${alertClass}">${alertText}</div>
                <div class="wishlist-price-row">
                    <div>
                        <div class="current-price">₹${item.current_price_inr}</div>
                        <div class="target-price">Target: ₹${item.target_price_inr}</div>
                    </div>
                </div>
                <div class="threshold-input-group">
                    <input type="number" 
                           class="threshold-input" 
                           placeholder="Update target price" 
                           min="0" 
                           value="${item.target_price_inr}"
                           onchange="updateThreshold(${item.id}, this.value)">
                </div>
                <div class="wishlist-actions">
                    <a href="game.html?id=${item.game_id}" class="btn btn-primary">View Deal</a>
                    <button class="btn btn-remove" onclick="removeFromWishlist(${item.id})">Remove</button>
                </div>
            </div>
        </div>
    `;
}

async function updateThreshold(wishlistId, newThreshold) {
    try {
        const token = getAccessToken();
        const response = await fetch(`/api/wishlist/${wishlistId}`, {
            method: 'PATCH',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                target_price_inr: parseInt(newThreshold)
            })
        });

        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to update threshold');
        }

        // Reload wishlist to show updated data
        loadWishlist();

    } catch (error) {
        console.error('Failed to update threshold:', error);
        alert('Failed to update threshold. Please try again.');
    }
}

async function removeFromWishlist(wishlistId) {
    if (!confirm('Are you sure you want to remove this game from your wishlist?')) {
        return;
    }

    try {
        const token = getAccessToken();
        const response = await fetch(`/api/wishlist/${wishlistId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to remove from wishlist');
        }

        // Reload wishlist
        loadWishlist();

    } catch (error) {
        console.error('Failed to remove from wishlist:', error);
        alert('Failed to remove from wishlist. Please try again.');
    }
}
