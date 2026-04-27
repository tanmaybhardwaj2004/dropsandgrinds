document.addEventListener('DOMContentLoaded', () => {
    console.log('Buy Timing page loaded');
    initAuthButton();
    initBuyTiming();
    loadSalesCalendar();
});

function initBuyTiming() {
    const checkBtn = document.getElementById('check-timing-btn');
    const gameInput = document.getElementById('game-id');
    
    if (!checkBtn || !gameInput) return;

    checkBtn.addEventListener('click', () => {
        const gameID = parseInt(gameInput.value);
        if (!gameID || gameID <= 0) {
            alert('Please enter a valid game ID');
            return;
        }
        checkBuyTiming(gameID);
    });

    // Allow Enter key to trigger check
    gameInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            checkBtn.click();
        }
    });
}

async function checkBuyTiming(gameID) {
    const loadingState = document.getElementById('loading-state');
    const recommendationCard = document.getElementById('recommendation-card');

    // Show loading state
    if (loadingState) loadingState.style.display = 'block';
    if (recommendationCard) recommendationCard.style.display = 'none';

    try {
        const response = await fetch(`/api/games/${gameID}/buy-timing`);
        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Failed to get buy timing');
        }

        // Hide loading state
        if (loadingState) loadingState.style.display = 'none';

        // Display recommendation
        displayRecommendation(data);

    } catch (error) {
        console.error('Failed to get buy timing:', error);
        if (loadingState) loadingState.style.display = 'none';
        alert('Failed to get buy timing. Please try again.');
    }
}

function displayRecommendation(data) {
    const recommendationCard = document.getElementById('recommendation-card');
    const recommendationBanner = document.getElementById('recommendation-banner');
    
    // Set recommendation banner
    const recClass = data.recommendation === 'buy_now' ? 'buy_now' : 
                     data.recommendation === 'wait_soon' ? 'wait_soon' : 'wait_next';
    
    const recText = data.recommendation === 'buy_now' ? 'BUY NOW! 🎉' :
                    data.recommendation === 'wait_soon' ? 'WAIT - SALE SOON ⏳' : 'WAIT FOR NEXT SALE 📅';
    
    recommendationBanner.className = `recommendation-banner ${recClass}`;
    recommendationBanner.textContent = recText;

    // Set details
    document.getElementById('recommendation-text').textContent = data.recommendation.replace('_', ' ').toUpperCase();
    document.getElementById('recommendation-reason').textContent = data.reason;

    // Show days until sale if available
    const daysUntilItem = document.getElementById('days-until-item');
    if (data.days_until_sale !== undefined && data.days_until_sale !== null) {
        daysUntilItem.style.display = 'block';
        document.getElementById('days-until').textContent = data.days_until_sale;
    } else {
        daysUntilItem.style.display = 'none';
    }

    // Show recommendation card
    recommendationCard.style.display = 'block';
}

async function loadSalesCalendar() {
    try {
        const response = await fetch('/api/sales/calendar');
        const sales = await response.json();

        if (!response.ok) {
            throw new Error('Failed to load sales calendar');
        }

        displaySalesCalendar(sales);

    } catch (error) {
        console.error('Failed to load sales calendar:', error);
    }
}

function displaySalesCalendar(sales) {
    const calendarContainer = document.getElementById('sales-calendar');
    if (!calendarContainer) return;

    const now = new Date();

    calendarContainer.innerHTML = sales.map(sale => {
        const startDate = new Date(sale.start_date);
        const endDate = new Date(sale.end_date);
        const isActive = now >= startDate && now <= endDate;

        const formatDate = (date) => date.toLocaleDateString('en-US', { 
            month: 'short', 
            day: 'numeric',
            year: 'numeric'
        });

        return `
            <div class="sale-card ${isActive ? 'active' : ''}">
                <div class="sale-name">${sale.name}</div>
                <div class="sale-platform">${sale.platform}</div>
                <div class="sale-dates">${formatDate(startDate)} - ${formatDate(endDate)}</div>
            </div>
        `;
    }).join('');
}
