/**
 * UI Components - Reusable loading, error, and empty states
 * DropsAndGrinds Frontend
 */

// Show loading spinner in a container
function showLoading(containerId, message = 'Loading...') {
    const container = document.getElementById(containerId);
    if (!container) return;
    
    container.innerHTML = `
        <div class="loading-state">
            <div class="spinner"></div>
            <p class="loading-text">${message}</p>
        </div>
    `;
}

// Show error state with retry button
function showError(containerId, message, retryCallback = null) {
    const container = document.getElementById(containerId);
    if (!container) return;
    
    const retryButton = retryCallback 
        ? `<button class="btn btn-primary btn-sm" onclick="${retryCallback}">Retry</button>` 
        : '';
    
    container.innerHTML = `
        <div class="error-state">
            <div class="error-icon">⚠️</div>
            <p class="error-text">${message}</p>
            ${retryButton}
        </div>
    `;
}

// Show empty state
function showEmpty(containerId, message, icon = '📭') {
    const container = document.getElementById(containerId);
    if (!container) return;
    
    container.innerHTML = `
        <div class="empty-state">
            <div class="empty-icon">${icon}</div>
            <p class="empty-text">${message}</p>
        </div>
    `;
}

// Toast notification system
function showToast(message, type = 'info', duration = 3000) {
    // Remove existing toast container if any
    let container = document.getElementById('toast-container');
    if (!container) {
        container = document.createElement('div');
        container.id = 'toast-container';
        container.style.cssText = `
            position: fixed;
            bottom: 20px;
            right: 20px;
            z-index: 10000;
            display: flex;
            flex-direction: column;
            gap: 10px;
        `;
        document.body.appendChild(container);
    }
    
    const toast = document.createElement('div');
    const colors = {
        success: '#238636',
        error: '#da3633',
        warning: '#d29922',
        info: '#58a6ff'
    };
    
    toast.style.cssText = `
        background: ${colors[type] || colors.info};
        color: white;
        padding: 12px 20px;
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.3);
        font-size: 14px;
        min-width: 250px;
        max-width: 400px;
        animation: slideIn 0.3s ease;
        display: flex;
        align-items: center;
        gap: 10px;
    `;
    
    const icons = {
        success: '✓',
        error: '✕',
        warning: '⚠',
        info: 'ℹ'
    };
    
    toast.innerHTML = `<span>${icons[type] || icons.info}</span> ${message}`;
    container.appendChild(toast);
    
    // Auto remove
    setTimeout(() => {
        toast.style.animation = 'slideOut 0.3s ease';
        setTimeout(() => toast.remove(), 300);
    }, duration);
}

// Global fetch wrapper with error handling
async function safeFetch(url, options = {}) {
    try {
        const response = await fetch(url, options);
        
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
        }
        
        return response;
    } catch (error) {
        if (error.message.includes('Failed to fetch') || error.message.includes('NetworkError')) {
            throw new Error('Network error. Please check your connection.');
        }
        throw error;
    }
}

// Retry wrapper for async functions
async function withRetry(fn, maxRetries = 3, delay = 1000) {
    for (let i = 0; i < maxRetries; i++) {
        try {
            return await fn();
        } catch (error) {
            if (i === maxRetries - 1) throw error;
            await new Promise(resolve => setTimeout(resolve, delay * (i + 1)));
        }
    }
}

// Add animation styles
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from { transform: translateX(100%); opacity: 0; }
        to { transform: translateX(0); opacity: 1; }
    }
    @keyframes slideOut {
        from { transform: translateX(0); opacity: 1; }
        to { transform: translateX(100%); opacity: 0; }
    }
    
    .loading-state, .error-state, .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: 40px 20px;
        text-align: center;
        color: var(--color-text-secondary);
    }
    
    .loading-state .spinner {
        width: 40px;
        height: 40px;
        border: 3px solid var(--color-border);
        border-top-color: var(--color-primary);
        border-radius: 50%;
        animation: spin 1s linear infinite;
        margin-bottom: 16px;
    }
    
    .loading-text {
        font-size: 14px;
        color: var(--color-text-secondary);
    }
    
    .error-icon {
        font-size: 32px;
        margin-bottom: 12px;
    }
    
    .error-text {
        font-size: 14px;
        color: var(--color-text-secondary);
        margin-bottom: 16px;
    }
    
    .empty-icon {
        font-size: 48px;
        margin-bottom: 16px;
        opacity: 0.5;
    }
    
    .empty-text {
        font-size: 14px;
        color: var(--color-text-secondary);
    }
    
    @keyframes spin {
        to { transform: rotate(360deg); }
    }
`;
document.head.appendChild(style);

// Export for module systems if available
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { showLoading, showError, showEmpty, showToast, safeFetch, withRetry };
}
