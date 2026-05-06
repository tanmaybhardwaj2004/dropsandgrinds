document.addEventListener('DOMContentLoaded', () => {
    initAuthButton();
    initThemeToggle();
    checkAllStatus();
});

const apis = [
    { name: 'Steam', slug: 'steam', endpoint: '/api/health/steam' },
    { name: 'Epic Games', slug: 'epic', endpoint: '/api/health/epic' },
    { name: 'GOG', slug: 'gog', endpoint: '/api/health/gog' },
    { name: 'Xbox', slug: 'xbox', endpoint: '/api/health/xbox' },
    { name: 'PlayStation', slug: 'playstation', endpoint: '/api/health/playstation' },
    { name: 'Nintendo', slug: 'nintendo', endpoint: '/api/health/nintendo' },
    { name: 'GreenManGaming', slug: 'greenmangaming', endpoint: '/api/health/greenmangaming' },
    { name: 'Fanatical', slug: 'fanatical', endpoint: '/api/health/fanatical' },
    { name: 'Humble Bundle', slug: 'humble', endpoint: '/api/health/humble' },
    { name: 'Instant Gaming', slug: 'instantgaming', endpoint: '/api/health/instantgaming' },
    { name: 'Gamivo', slug: 'gamivo', endpoint: '/api/health/gamivo' },
    { name: 'Eneba', slug: 'eneba', endpoint: '/api/health/eneba' },
    { name: 'Database', slug: 'database', endpoint: '/api/health/database' },
    { name: 'Redis', slug: 'redis', endpoint: '/api/health/redis' },
    { name: 'Meilisearch', slug: 'meilisearch', endpoint: '/api/health/meilisearch' }
];

async function checkAllStatus() {
    const grid = document.getElementById('api-status-grid');
    const overallIndicator = document.getElementById('overall-indicator');
    const overallText = document.getElementById('overall-text');
    const lastUpdated = document.getElementById('last-updated');
    
    if (!grid) return;
    
    grid.innerHTML = '';
    overallIndicator.style.background = 'var(--color-warning)';
    overallText.textContent = 'Checking...';
    
    // Show checking status for all APIs first
    apis.forEach((api, index) => {
        createAPICard(api, 'checking');
    });
    
    // Then perform actual checks
    const results = await Promise.allSettled(apis.map(api => checkAPIStatus(api)));
    
    // Clear grid and show actual results
    grid.innerHTML = '';
    
    let healthyCount = 0;
    let degradedCount = 0;
    let downCount = 0;
    
    results.forEach((result, index) => {
        if (result.status === 'fulfilled') {
            const status = result.value;
            if (status === 'healthy') healthyCount++;
            else if (status === 'degraded') degradedCount++;
            else downCount++;
            
            createAPICard(apis[index], status);
        } else {
            downCount++;
            createAPICard(apis[index], 'down');
        }
    });
    
    // Update overall status
    if (downCount === 0 && degradedCount === 0) {
        overallIndicator.style.background = 'var(--color-success)';
        overallText.textContent = 'All Systems Operational';
    } else if (downCount === 0) {
        overallIndicator.style.background = 'var(--color-warning)';
        overallText.textContent = 'Some Systems Degraded';
    } else {
        overallIndicator.style.background = 'var(--color-error)';
        overallText.textContent = 'Some Systems Down';
    }
    
    lastUpdated.textContent = new Date().toLocaleString();
}

async function checkAPIStatus(api) {
    try {
        const response = await fetch(api.endpoint);
        if (response.ok) {
            const data = await response.json();
            if (data.status === 'healthy') {
                return 'healthy';
            } else if (data.status === 'degraded') {
                return 'degraded';
            }
        }
        return 'down';
    } catch (error) {
        console.error(`Failed to check ${api.name}:`, error);
        return 'down';
    }
}

function createAPICard(api, status) {
    const grid = document.getElementById('api-status-grid');
    if (!grid) return;
    
    const card = document.createElement('div');
    card.className = 'card';
    card.style.cssText = 'padding: var(--space-4);';
    
    const statusColors = {
        healthy: 'var(--color-success)',
        degraded: 'var(--color-warning)',
        down: 'var(--color-error)',
        checking: 'var(--color-warning)'
    };
    
    const statusText = {
        healthy: 'Operational',
        degraded: 'Degraded',
        down: 'Down',
        checking: 'Checking...'
    };
    
    const statusBadges = {
        healthy: '<span style="display: inline-flex; align-items: center; gap: var(--space-1);"><span style="width: 8px; height: 8px; background: var(--color-success); border-radius: 50%;"></span><span style="color: var(--color-success); font-weight: 600;">Operational</span></span>',
        degraded: '<span style="display: inline-flex; align-items: center; gap: var(--space-1);"><span style="width: 8px; height: 8px; background: var(--color-warning); border-radius: 50%;"></span><span style="color: var(--color-warning); font-weight: 600;">Degraded</span></span>',
        down: '<span style="display: inline-flex; align-items: center; gap: var(--space-1);"><span style="width: 8px; height: 8px; background: var(--color-error); border-radius: 50%;"></span><span style="color: var(--color-error); font-weight: 600;">Down</span></span>',
        checking: '<span style="display: inline-flex; align-items: center; gap: var(--space-1);"><span style="width: 8px; height: 8px; background: var(--color-warning); border-radius: 50%; animation: pulse 1s infinite;"></span><span style="color: var(--color-warning); font-weight: 600;">Checking...</span></span>'
    };
    
    card.innerHTML = `
        <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--space-2);">
            <div style="display: flex; align-items: center; gap: var(--space-3);">
                <div style="width: 12px; height: 12px; border-radius: 50%; background: ${statusColors[status]};"></div>
                <span style="font-weight: 600;">${api.name}</span>
            </div>
            <span class="badge" style="background: ${statusColors[status]}; color: white; padding: var(--space-1) var(--space-2); border-radius: var(--radius-full); font-size: var(--text-sm); font-weight: 500;">${statusBadges[status]}</span>
        </div>
        <div style="font-size: var(--text-sm); color: var(--color-text-muted);">
            Endpoint: <code style="background: var(--color-surface-secondary); padding: var(--space-1); border-radius: var(--radius-sm);">${api.endpoint}</code>
        </div>
    `;
    
    grid.appendChild(card);
}

function initThemeToggle() {
    const toggle = document.getElementById('theme-toggle');
    if (!toggle) return;
    
    const savedTheme = localStorage.getItem('theme') || 'light';
    document.documentElement.setAttribute('data-theme', savedTheme);
    updateThemeIcons(savedTheme);
    
    toggle.addEventListener('click', () => {
        const currentTheme = document.documentElement.getAttribute('data-theme');
        const newTheme = currentTheme === 'light' ? 'dark' : 'light';
        document.documentElement.setAttribute('data-theme', newTheme);
        localStorage.setItem('theme', newTheme);
        updateThemeIcons(newTheme);
    });
}

function updateThemeIcons(theme) {
    const moonIcon = document.querySelector('.theme-icon-moon');
    const sunIcon = document.querySelector('.theme-icon-sun');
    
    if (theme === 'dark') {
        moonIcon.style.display = 'none';
        sunIcon.style.display = 'block';
    } else {
        moonIcon.style.display = 'block';
        sunIcon.style.display = 'none';
    }
}
