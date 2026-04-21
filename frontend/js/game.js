document.addEventListener('DOMContentLoaded', () => {
    console.log("Game Tracking Metrics UI Loaded");
    
    initWishlistButton();
});

// Implementation of LocalStorage Wishlist Temporary Fix (Phase 4 - User Approved)
function initWishlistButton() {
    const btn = document.getElementById('wishlist-btn');
    if(!btn) return;

    const gameId = "cyberpunk-2077"; // Hardcoded for this mockup
    let wishlist = [];
    
    // Parse existing wishlist
    try {
        const stored = localStorage.getItem('dropsandgrinds_wishlist');
        if (stored) wishlist = JSON.parse(stored);
    } catch {
        wishlist = [];
    }

    // Initialize UI state
    const isWishlisted = wishlist.includes(gameId);
    if(isWishlisted) {
        btn.classList.add('active');
        btn.innerHTML = '♥ Wishlisted';
    }

    // Attach interaction logic
    btn.addEventListener('click', () => {
        // Since Dev A hasn't built POST /api/tracking yet, we manage state strictly in local memory/cache.
        const currentList = JSON.parse(localStorage.getItem('dropsandgrinds_wishlist') || '[]');
        const idx = currentList.indexOf(gameId);
        
        if(idx > -1) {
            // Remove
            currentList.splice(idx, 1);
            btn.classList.remove('active');
            btn.innerHTML = '♡ Add to Wishlist';
        } else {
            // Add
            currentList.push(gameId);
            btn.classList.add('active');
            btn.innerHTML = '♥ Wishlisted';
        }
        
        localStorage.setItem('dropsandgrinds_wishlist', JSON.stringify(currentList));
        console.log("Local Wishlist Updated:", currentList);
    });
}
