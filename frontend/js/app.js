document.addEventListener('DOMContentLoaded', () => {
    console.log("DropsAndGrinds App Loaded");
    
    // Check backend health
    fetch('/health')
        .then(res => res.text())
        .then(data => console.log("Backend Status:", data))
        .catch(err => console.error("Backend not reachable:", err));
        
    renderMockDeals();
});

// Mock deals for design purposes
function renderMockDeals() {
    const container = document.getElementById('deals-container');
    
    for (let i = 0; i < 8; i++) {
        const card = document.createElement('div');
        card.className = 'deal-card';
        card.innerHTML = `
            <div class="deal-cover"></div>
            <div class="deal-info">
                <div class="deal-title">Game Title ${i + 1}</div>
                <div class="deal-price-row">
                    <span class="discount">-40%</span>
                    <span class="price">₹1,499</span>
                </div>
            </div>
        `;
        container.appendChild(card);
    }
}
