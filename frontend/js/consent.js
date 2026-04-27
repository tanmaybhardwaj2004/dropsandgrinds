// GDPR Consent Management
const CONSENT_ANALYTICS_KEY = 'consent_analytics';
const CONSENT_ALERTS_KEY = 'consent_alerts';

// Check if consent has been given
function hasAnalyticsConsent() {
    return localStorage.getItem(CONSENT_ANALYTICS_KEY) === 'true';
}

function hasAlertsConsent() {
    return localStorage.getItem(CONSENT_ALERTS_KEY) === 'true';
}

// Set consent
function setAnalyticsConsent(consent) {
    localStorage.setItem(CONSENT_ANALYTICS_KEY, consent);
}

function setAlertsConsent(consent) {
    localStorage.setItem(CONSENT_ALERTS_KEY, consent);
}

// Show consent modal if not yet decided
function showConsentModalIfNeeded() {
    if (!hasAnalyticsConsent() || !hasAlertsConsent()) {
        showConsentModal();
    }
}

function showConsentModal() {
    // Create modal if it doesn't exist
    let modal = document.getElementById('consent-modal');
    if (!modal) {
        modal = document.createElement('div');
        modal.id = 'consent-modal';
        modal.className = 'consent-modal';
        modal.innerHTML = `
            <div class="consent-modal-content">
                <h2>Privacy Consent</h2>
                <p>We value your privacy. Please choose your preferences:</p>
                
                <div class="consent-option">
                    <label>
                        <input type="checkbox" id="consent-analytics" ${hasAnalyticsConsent() ? 'checked' : ''}>
                        <span>Allow analytics tracking (helps us improve the service)</span>
                    </label>
                </div>
                
                <div class="consent-option">
                    <label>
                        <input type="checkbox" id="consent-alerts" ${hasAlertsConsent() ? 'checked' : ''}>
                        <span>Allow email alerts for price drops</span>
                    </label>
                </div>
                
                <div class="consent-actions">
                    <button class="btn btn-primary" onclick="saveConsent()">Save Preferences</button>
                    <button class="btn btn-secondary" onclick="hideConsentModal()">Later</button>
                </div>
            </div>
        `;
        document.body.appendChild(modal);
    }
    
    modal.style.display = 'flex';
}

function hideConsentModal() {
    const modal = document.getElementById('consent-modal');
    if (modal) {
        modal.style.display = 'none';
    }
}

function saveConsent() {
    const analyticsConsent = document.getElementById('consent-analytics').checked;
    const alertsConsent = document.getElementById('consent-alerts').checked;
    
    setAnalyticsConsent(analyticsConsent);
    setAlertsConsent(alertsConsent);
    
    // Send consent to backend if authenticated
    const token = getAccessToken();
    if (token) {
        fetch('/api/me/consent', {
            method: 'PATCH',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                consent_analytics: analyticsConsent,
                consent_alerts: alertsConsent
            })
        }).catch(err => console.error('Failed to save consent to server:', err));
    }
    
    hideConsentModal();
}

// Initialize consent on page load
document.addEventListener('DOMContentLoaded', () => {
    // Delay showing modal slightly to not interrupt user
    setTimeout(() => {
        showConsentModalIfNeeded();
    }, 2000);
});
