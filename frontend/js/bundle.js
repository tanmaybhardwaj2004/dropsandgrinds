document.addEventListener('DOMContentLoaded', () => {
    console.log('Bundle Breaker page loaded');
    initAuthButton();
    initBundleForm();
});

function initBundleForm() {
    const form = document.getElementById('bundle-form');
    if (!form) return;

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const url = document.getElementById('bundle-url').value;
        const bundlePrice = parseFloat(document.getElementById('bundle-price').value);

        if (!url || !bundlePrice) {
            alert('Please fill in all fields');
            return;
        }

        await analyzeBundle(url, bundlePrice);
    });
}

async function analyzeBundle(url, bundlePrice) {
    const loadingState = document.getElementById('loading-state');
    const errorState = document.getElementById('error-state');
    const resultsContainer = document.getElementById('results-container');

    // Show loading state
    if (loadingState) loadingState.style.display = 'block';
    if (errorState) errorState.style.display = 'none';
    if (resultsContainer) resultsContainer.style.display = 'none';

    try {
        const response = await fetch('/api/bundles/analyze', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                url: url,
                bundle_price_inr: bundlePrice
            })
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Failed to analyze bundle');
        }

        // Hide loading state
        if (loadingState) loadingState.style.display = 'none';

        // Display results
        displayResults(data);

    } catch (error) {
        console.error('Failed to analyze bundle:', error);
        if (loadingState) loadingState.style.display = 'none';
        if (errorState) {
            errorState.style.display = 'block';
            document.getElementById('error-message').textContent = error.message;
        }
    }
}

function displayResults(data) {
    const resultsContainer = document.getElementById('results-container');
    const verdictBanner = document.getElementById('verdict-banner');
    
    // Set verdict banner
    const verdictClass = data.verdict === 'buy_bundle' ? 'buy_bundle' : 
                        data.verdict === 'buy_separately' ? 'buy_separately' : 'mixed';
    
    const verdictText = data.verdict === 'buy_bundle' ? 'BUY THE BUNDLE' :
                       data.verdict === 'buy_separately' ? 'BUY SEPARATELY' : 'MIXED DECISION';
    
    verdictBanner.className = `verdict-banner ${verdictClass}`;
    verdictBanner.textContent = verdictText;

    // Set summary values
    document.getElementById('bundle-name').textContent = data.bundle_name || 'Unknown Bundle';
    document.getElementById('bundle-price-display').textContent = `₹${data.bundle_price_inr.toFixed(2)}`;
    document.getElementById('individual-sum').textContent = `₹${data.individual_sum_inr.toFixed(2)}`;
    
    const savingsEl = document.getElementById('savings');
    savingsEl.textContent = `₹${data.savings_inr.toFixed(2)}`;
    savingsEl.style.color = data.savings_inr >= 0 ? '#238636' : '#da3633';

    // Populate games table
    const tableBody = document.getElementById('games-table-body');
    tableBody.innerHTML = data.games.map(game => {
        const statusClass = game.current_price_inr > 0 ? 
                           game.bundle_share_inr < game.current_price_inr ? 'status-good' : 'status-bad' : 
                           'status-unknown';
        
        const statusText = game.current_price_inr > 0 ?
                          game.bundle_share_inr < game.current_price_inr ? 'Good deal' : 'Overpriced' :
                          'Price unknown';

        return `
            <tr>
                <td>${game.title}</td>
                <td>${game.current_price_inr > 0 ? `₹${game.current_price_inr.toFixed(2)}` : 'N/A'}</td>
                <td>₹${game.bundle_share_inr.toFixed(2)}</td>
                <td class="${statusClass}">${statusText}</td>
            </tr>
        `;
    }).join('');

    // Show results
    resultsContainer.style.display = 'block';
}
