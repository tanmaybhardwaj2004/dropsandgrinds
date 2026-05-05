document.addEventListener('DOMContentLoaded', () => {
    console.log('Wishlist page loaded');
    initAuthButton();
    checkAuthAndLoadWishlist();
});

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
                container.style.display = 'block';
                container.innerHTML = renderWishlistTable(data.items);
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

function renderWishlistTable(items) {
    return `
        <div class="card" style="overflow:auto;">
            <table class="data-table">
                <thead>
                    <tr>
                        <th>Game</th>
                        <th>Platform</th>
                        <th>Current</th>
                        <th>Target</th>
                        <th>Status</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>${items.map(renderWishlistRow).join('')}</tbody>
            </table>
        </div>
    `;
}

function renderWishlistRow(item) {
    const alertText = item.triggered ? 'Triggered' : 'Watching';
    return `
        <tr data-id="${item.id}">
            <td><a href="game.html?id=${item.game_id}">${item.title}</a></td>
            <td>${item.platform}</td>
            <td>₹${item.current_price_inr}</td>
            <td><input type="number" class="threshold-input" min="1" value="${item.target_price_inr}" onchange="updateThreshold(${item.id}, this.value)"></td>
            <td><span class="status-pill ${item.triggered ? 'success' : 'neutral'}">${alertText}</span></td>
            <td>
                <a href="game.html?id=${item.game_id}" class="btn btn-secondary">View</a>
                <button class="btn btn-remove" onclick="removeFromWishlist(${item.id})">Remove</button>
            </td>
        </tr>
    `;
}

async function updateThreshold(wishlistId, newThreshold) {
    try {
        const token = getAccessToken();
        const response = await fetch(`/api/wishlist/${wishlistId}/threshold`, {
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
