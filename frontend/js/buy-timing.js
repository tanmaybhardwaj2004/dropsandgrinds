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
        const gameName = gameInput.value.trim();
        if (!gameName) {
            showTimingError('Please enter a game name');
            return;
        }
        findGameAndCheckTiming(gameName);
    });

    // Allow Enter key to trigger check
    gameInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            checkBtn.click();
        }
    });
}

async function findGameAndCheckTiming(gameName) {
    const loadingState = document.getElementById('loading-state');
    const recommendationCard = document.getElementById('recommendation-card');

    if (loadingState) loadingState.style.display = 'block';
    if (recommendationCard) recommendationCard.style.display = 'none';

    try {
        const response = await fetch(`/api/games/search?q=${encodeURIComponent(gameName)}&limit=1&offset=0`);
        const data = await response.json();
        if (!response.ok) {
            throw new Error(data.error || 'Search failed');
        }

        const game = (data.games || [])[0];
        if (!game) {
            showTimingError('Game not found');
            return;
        }

        await checkBuyTiming(game.id, game.title);
    } catch (error) {
        console.error('Failed to search game:', error);
        showTimingError(error.message || 'Failed to search game');
    }
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

        displayRecommendation(data);

    } catch (error) {
        console.error('Failed to get buy timing:', error);
        if (loadingState) loadingState.style.display = 'none';
        showTimingError('Failed to get buy timing. Please try again.');
    }
}

function displayRecommendation(data) {
    const recommendationCard = document.getElementById('recommendation-card');
    const recommendationBanner = document.getElementById('recommendation-banner');
    
    // Set recommendation banner
    const recClass = data.recommendation === 'buy_now' ? 'buy_now' : 
                     data.recommendation === 'wait_soon' ? 'wait_soon' : 'wait_next';
    
    const recText = data.recommendation === 'buy_now' ? 'Buy Now' :
                    data.recommendation === 'wait_soon' ? 'Wait' : 'Wait';
    
    recommendationBanner.className = `recommendation-banner ${recClass}`;
    recommendationBanner.textContent = recText;

    // Set details
    document.getElementById('recommendation-text').textContent = recText;
    document.getElementById('recommendation-reason').textContent = data.reason;

    // Show days until sale if available
    const daysUntilItem = document.getElementById('days-until-item');
    if (data.days_until_sale !== undefined && data.days_until_sale !== null) {
        daysUntilItem.style.display = 'block';
        document.getElementById('days-until').textContent = `${data.days_until_sale} days`;
    } else {
        daysUntilItem.style.display = 'none';
    }

    const sale = data.active_sale || data.next_sale;
    const saleNameItem = document.getElementById('sale-name-item');
    if (saleNameItem && sale) {
        saleNameItem.style.display = 'block';
        document.getElementById('sale-name').textContent = sale.name || 'Sale window';
    } else if (saleNameItem) {
        saleNameItem.style.display = 'none';
    }

    // Show recommendation card
    recommendationCard.style.display = 'block';
}

function showTimingError(message) {
    const loadingState = document.getElementById('loading-state');
    const recommendationCard = document.getElementById('recommendation-card');
    const recommendationBanner = document.getElementById('recommendation-banner');
    if (loadingState) loadingState.style.display = 'none';
    if (!recommendationCard || !recommendationBanner) {
        alert(message);
        return;
    }
    recommendationBanner.className = 'recommendation-banner wait_next';
    recommendationBanner.textContent = message === 'Game not found' ? 'Game not found' : 'Unable to check';
    document.getElementById('recommendation-text').textContent = message;
    document.getElementById('recommendation-reason').textContent = '';
    document.getElementById('days-until-item').style.display = 'none';
    const saleNameItem = document.getElementById('sale-name-item');
    if (saleNameItem) saleNameItem.style.display = 'none';
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
