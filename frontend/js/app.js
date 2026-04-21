document.addEventListener('DOMContentLoaded', () => {
    console.log("DropsAndGrinds App Loaded");
    
    // Check backend health
    fetch('/health')
        .then(res => res.text())
        .then(data => console.log("Backend Status:", data))
        .catch(err => console.error("Backend not reachable:", err));
        
    initApp();
});

// Hardcoded Mock Dataset (Normally fetched from /api/games)
const mockDeals = [
    { id: 1, title: "Cyberpunk 2077", cover: "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1091500/header.jpg", store: "Steam", price: 1499, original: 2999, discount: 50, score: 86, isGSTAdded: true },
    { id: 2, title: "Elden Ring", cover: "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1245620/header.jpg", store: "Steam", price: 2399, original: 3999, discount: 40, score: 96, isGSTAdded: true },
    { id: 3, title: "Alan Wake", cover: "https://cdn2.unrealengine.com/egs-alanwakeRemastered-remedyentertainment-s2-1200x1600-b6f4e150f584.jpg", store: "Epic Games", price: 450, original: 1500, discount: 70, score: 83, isGSTAdded: false },
    { id: 4, title: "The Witcher 3", cover: "https://images.gog-statics.com/1445585698466185bb212ae17d45e5df5a36371c10787a740703e2c340d12e8b_glx_logo_284x400.png", store: "GOG", price: 299, original: 999, discount: 70, score: 93, isGSTAdded: false },
    { id: 5, title: "Red Dead Redemption 2", cover: "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1174180/header.jpg", store: "Steam", price: 999, original: 3199, discount: 69, score: 97, isGSTAdded: true },
    { id: 6, title: "Hades", cover: "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1145360/header.jpg", store: "Epic Games", price: 549, original: 1099, discount: 50, score: 93, isGSTAdded: false },
    { id: 7, title: "Stardew Valley", cover: "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/413150/header.jpg", store: "Steam", price: 384, original: 479, discount: 20, score: 89, isGSTAdded: true },
    { id: 8, title: "Control", cover: "https://images.gog-statics.com/97adbbdcdab1889814c8d5c4142f2edab2838bed820d885b546bed1d5a711422_glx_logo_284x400.png", store: "GOG", price: 899, original: 2999, discount: 70, score: 85, isGSTAdded: false },
];

function initApp() {
    renderDeals(mockDeals);

    // Attach Event Listeners to Filters
    const filters = ['store-steam', 'store-epic', 'store-gog'];
    filters.forEach(id => {
        document.getElementById(id).addEventListener('change', updateFilters);
    });

    const priceSlider = document.getElementById('price-slider');
    const priceDisplay = document.getElementById('price-display');
    priceSlider.addEventListener('input', (e) => {
        priceDisplay.textContent = `₹${e.target.value}`;
        updateFilters();
    });
    
    const searchInput = document.getElementById('search-input');
    searchInput.addEventListener('input', updateFilters);
}

function updateFilters() {
    const steamChecked = document.getElementById('store-steam').checked;
    const epicChecked = document.getElementById('store-epic').checked;
    const gogChecked = document.getElementById('store-gog').checked;
    const maxPrice = parseInt(document.getElementById('price-slider').value);
    const searchTerm = document.getElementById('search-input').value.toLowerCase();

    const filtered = mockDeals.filter(deal => {
        // Store filter
        if (deal.store === "Steam" && !steamChecked) return false;
        if (deal.store === "Epic Games" && !epicChecked) return false;
        if (deal.store === "GOG" && !gogChecked) return false;

        // Price filter
        if (deal.price > maxPrice) return false;
        
        // Search Filter
        if (searchTerm && !deal.title.toLowerCase().includes(searchTerm)) return false;

        return true;
    });

    renderDeals(filtered);
}

function renderDeals(dealsArray) {
    const container = document.getElementById('deals-container');
    container.innerHTML = ''; // clear grid

    if(dealsArray.length === 0) {
        container.innerHTML = '<p style="color: var(--text-muted); grid-column: 1/-1;">No deals found matching criteria.</p>';
        return;
    }

    dealsArray.forEach(deal => {
        const card = document.createElement('div');
        card.className = 'deal-card';
        
        card.innerHTML = `
            <img src="${deal.cover}" class="deal-cover" alt="${deal.title} cover">
            <div class="deal-info">
                <div class="meta-row">
                    <span>${deal.store} ${deal.isGSTAdded ? '(Inc. GST)' : ''}</span>
                    <span class="score-badge">★ ${deal.score}</span>
                </div>
                <div class="deal-title">${deal.title}</div>
                <div class="deal-price-row">
                    <span class="discount">-${deal.discount}%</span>
                    <div style="text-align: right;">
                        <span style="text-decoration: line-through; color: var(--text-muted); font-size: 0.8rem; display: block;">₹${deal.original}</span>
                        <span class="price">₹${deal.price}</span>
                    </div>
                </div>
            </div>
        `;
        container.appendChild(card);
    });
}
