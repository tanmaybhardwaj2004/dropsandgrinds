document.addEventListener('DOMContentLoaded', () => {
    console.log('Savings page loaded');
    initAuthButton();
    checkAuthAndLoadSavings();
    initPurchaseForm();
});

function checkAuthAndLoadSavings() {
    const token = getAccessToken();
    if (!token) {
        // Redirect to login if not authenticated
        window.location.href = 'login.html';
        return;
    }
    loadSavings();
}

function initPurchaseForm() {
    const form = document.getElementById('purchase-form');
    if (!form) return;

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const gameTitle = document.getElementById('game-title').value;
        const paidPrice = parseInt(document.getElementById('paid-price').value);
        const originalPrice = parseInt(document.getElementById('original-price').value);
        const gameID = document.getElementById('game-id').value;

        if (!gameTitle || !paidPrice || !originalPrice) {
            alert('Please fill in all required fields');
            return;
        }

        if (paidPrice > originalPrice) {
            alert('Paid price cannot be greater than original price');
            return;
        }

        try {
            const token = getAccessToken();
            const response = await fetch('/api/savings/purchase', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({
                    game_id: gameID ? parseInt(gameID) : 0,
                    game_title: gameTitle,
                    paid_price_inr: paidPrice,
                    original_price_inr: originalPrice
                })
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || 'Failed to log purchase');
            }

            // Reset form and reload savings
            form.reset();
            await loadSavings();
            alert('Purchase logged successfully!');

        } catch (error) {
            console.error('Failed to log purchase:', error);
            alert('Failed to log purchase. Please try again.');
        }
    });
}

async function loadSavings() {
    try {
        const token = getAccessToken();
        const response = await fetch('/api/savings', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Failed to load savings');
        }

        // Update total savings
        const totalSavingsEl = document.getElementById('total-savings');
        if (totalSavingsEl) {
            totalSavingsEl.textContent = `₹${data.total_savings}`;
        }

        // Update equivalent games message
        const equivalentGamesEl = document.getElementById('equivalent-games');
        if (equivalentGamesEl) {
            equivalentGamesEl.textContent = data.equivalent_games;
        }

        // Render monthly breakdown
        renderMonthlyBreakdown(data.monthly_breakdown);

        // Load purchase history
        loadPurchaseHistory();

    } catch (error) {
        console.error('Failed to load savings:', error);
    }
}

function renderMonthlyBreakdown(monthlyData) {
    const chartContainer = document.getElementById('monthly-chart');
    if (!chartContainer) return;

    if (!monthlyData || monthlyData.length === 0) {
        chartContainer.innerHTML = '<p style="color: var(--text-muted);">No savings data yet. Log your first purchase!</p>';
        return;
    }

    // Find max value for scaling
    const maxValue = Math.max(...monthlyData.map(m => m.total_savings));

    chartContainer.innerHTML = monthlyData.map(month => {
        const percentage = maxValue > 0 ? (month.total_savings / maxValue) * 100 : 0;
        return `
            <div class="monthly-bar">
                <div class="monthly-label">${formatMonth(month.month)}</div>
                <div class="monthly-bar-container">
                    <div class="monthly-bar-fill" style="width: ${percentage}%">
                        <span class="monthly-value">₹${month.total_savings}</span>
                    </div>
                </div>
            </div>
        `;
    }).join('');
}

function formatMonth(monthStr) {
    // monthStr is in format "YYYY-MM"
    const [year, month] = monthStr.split('-');
    const date = new Date(year, month - 1);
    return date.toLocaleDateString('en-US', { month: 'short', year: '2-digit' });
}

async function loadPurchaseHistory() {
    try {
        const token = getAccessToken();
        const response = await fetch('/api/savings/history?limit=20&offset=0', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Failed to load purchase history');
        }

        renderPurchaseHistory(data.purchases || []);

    } catch (error) {
        console.error('Failed to load purchase history:', error);
    }
}

function renderPurchaseHistory(purchases) {
    const listContainer = document.getElementById('purchase-history-list');
    if (!listContainer) return;

    if (purchases.length === 0) {
        listContainer.innerHTML = '<div class="empty-purchases">No purchases logged yet.</div>';
        return;
    }

    listContainer.innerHTML = purchases.map(purchase => `
        <div class="purchase-item">
            <div class="purchase-info">
                <div class="purchase-title">${purchase.game_title}</div>
                <div class="purchase-date">${formatDate(purchase.purchased_at)}</div>
            </div>
            <div class="purchase-savings">
                <div class="saved-amount">₹${purchase.saved_amount_inr}</div>
                <div class="price-breakdown">Paid ₹${purchase.paid_price_inr} / Was ₹${purchase.original_price_inr}</div>
            </div>
        </div>
    `).join('');
}

function formatDate(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { 
        year: 'numeric', 
        month: 'short', 
        day: 'numeric' 
    });
}
