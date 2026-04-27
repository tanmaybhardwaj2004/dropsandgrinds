document.addEventListener('DOMContentLoaded', () => {
    console.log('Wishlist page loaded');
    initAuthButton();
    checkAuthAndLoadWishlist();
});

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
            <img src="${item.cover_url}" class="wishlist-cover" alt="${item.title} cover">
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
