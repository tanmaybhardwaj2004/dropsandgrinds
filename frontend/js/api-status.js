document.addEventListener('DOMContentLoaded', async () => {
    await initNavbar();
    checkAllStatus();
});

const apis = [
    { name: 'Steam', slug: 'steam', endpoint: '/health/stores/steam' },
    { name: 'Epic Games', slug: 'epic', endpoint: '/health/stores/epic' },
    { name: 'GOG', slug: 'gog', endpoint: '/health/stores/gog' },
    { name: 'Xbox', slug: 'xbox', endpoint: '/health/stores/xbox' },
    { name: 'PlayStation', slug: 'playstation', endpoint: '/health/stores/playstation' },
    { name: 'Nintendo', slug: 'nintendo', endpoint: '/health/stores/nintendo' },
    { name: 'GreenManGaming', slug: 'greenmangaming', endpoint: '/health/stores/greenmangaming' },
    { name: 'Fanatical', slug: 'fanatical', endpoint: '/health/stores/fanatical' },
    { name: 'Humble Bundle', slug: 'humble', endpoint: '/health/stores/humble' },
    { name: 'Instant Gaming', slug: 'instantgaming', endpoint: '/health/stores/instantgaming' },
    { name: 'Gamivo', slug: 'gamivo', endpoint: '/health/stores/gamivo' },
    { name: 'Eneba', slug: 'eneba', endpoint: '/health/stores/eneba' },
    { name: 'Database', slug: 'database', endpoint: '/health/deps' },
    { name: 'Redis', slug: 'redis', endpoint: '/health/deps' },
    { name: 'Meilisearch', slug: 'meilisearch', endpoint: '/health/deps' }
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
            
            // Handle different response formats
            let status = 'down';
            
            // Store health endpoints return StoreHealthStatus with status field
            if (data.status) {
                if (data.status === 'up') status = 'healthy';
                else if (data.status === 'degraded') status = 'degraded';
                else status = 'down';
            }
            // Dependency health endpoint returns different format
            else if (api.endpoint === '/health/deps') {
                // Check if database, redis, etc. are up
                if (data.database === 'up' || data.redis === 'up') status = 'healthy';
                else if (data.database === 'down' || data.redis === 'down') status = 'down';
                else status = 'degraded';
            }
            
            return status;
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
