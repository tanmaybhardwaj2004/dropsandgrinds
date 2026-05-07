// Auth State Management
// JWT Tokens must be stored in memory per project requirements, not localStorage
let authState = {
    accessToken: null,
    refreshToken: null,
    userId: null,
    isAuthenticated: false
};

// Expose state for cross-script reads in same page context.
window.authState = authState;

document.addEventListener('DOMContentLoaded', () => {
    hydrateAuthState();
    showPendingLoginMessage();
    scheduleTokenRefresh();

    const loginForm = document.getElementById('login-form');
    const registerForm = document.getElementById('register-form');

    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }

    if (registerForm) {
        registerForm.addEventListener('submit', handleRegister);
    }
});

function showError(message) {
    const errorDiv = document.getElementById('auth-error');
    if (errorDiv) {
        const messageEl = document.getElementById('error-message');
        if (messageEl) {
            messageEl.textContent = message;
        } else {
            errorDiv.textContent = message;
        }
        errorDiv.classList.remove('hidden');
    }
}

function hideError() {
    const errorDiv = document.getElementById('auth-error');
    if (errorDiv) {
        errorDiv.classList.add('hidden');
    }
}

function showPendingLoginMessage() {
    const message = sessionStorage.getItem('dropsandgrinds_login_message');
    if (!message) return;
    sessionStorage.removeItem('dropsandgrinds_login_message');
    showError(message);
}

async function handleLogin(e) {
    e.preventDefault();
    hideError();
    
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const btn = document.getElementById('login-btn');

    btn.disabled = true;
    btn.textContent = "Logging in...";

    try {
        const response = await fetch('/api/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Failed to login');
        }

        // Store tokens in memory
        setAuthTokens(data.access_token, data.refresh_token, data.user_id);
        
        // Redirect to dashboard/home
        window.location.href = 'index.html';
    } catch (error) {
        showError(error.message);
    } finally {
        btn.disabled = false;
        btn.textContent = "Log In";
    }
}

async function handleRegister(e) {
    e.preventDefault();
    hideError();

    const username = document.getElementById('username').value;
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const steamId = document.getElementById('steamId').value;
    const consentAlerts = document.getElementById('consentAlerts').checked;
    const consentAnalytics = document.getElementById('consentAnalytics').checked;
    const btn = document.getElementById('register-btn');

    btn.disabled = true;
    btn.textContent = "Creating Account...";

    try {
        const response = await fetch('/api/auth/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                username,
                email,
                password,
                steam_id: steamId || undefined,
                consent_alerts: consentAlerts,
                consent_analytics: consentAnalytics
            })
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Failed to register');
        }

        // Store tokens in memory
        setAuthTokens(data.access_token, data.refresh_token, data.user_id);
        
        // Redirect to dashboard/home
        window.location.href = 'index.html';
    } catch (error) {
        showError(error.message);
    } finally {
        btn.disabled = false;
        btn.textContent = "Create Account";
    }
}

// Utility to set tokens in memory space
function setAuthTokens(access, refresh, userId) {
    authState.accessToken = access;
    authState.refreshToken = refresh;
    authState.userId = userId;
    authState.isAuthenticated = true;
    window.authState = authState;

    // Keep tokens for tab/session navigation between static HTML pages.
    sessionStorage.setItem('dropsandgrinds_access_token', access);
    sessionStorage.setItem('dropsandgrinds_refresh_token', refresh);
    sessionStorage.setItem('dropsandgrinds_user_id', String(userId));
    sessionStorage.setItem('dropsandgrinds_is_authenticated', 'true');

    console.log("Tokens stored securely in memory.");
}

function hydrateAuthState() {
    const access = sessionStorage.getItem('dropsandgrinds_access_token');
    const refresh = sessionStorage.getItem('dropsandgrinds_refresh_token');
    const userID = sessionStorage.getItem('dropsandgrinds_user_id');
    const isAuthenticated = sessionStorage.getItem('dropsandgrinds_is_authenticated') === 'true';

    if (!access || !refresh || !userID || !isAuthenticated) {
        return;
    }

    authState.accessToken = access;
    authState.refreshToken = refresh;
    authState.userId = Number(userID);
    authState.isAuthenticated = true;
    window.authState = authState;
}

function initAuthButton() {
    const token = getAccessToken();
    const authBtn = document.getElementById('auth-btn');
    const loginBtn = document.getElementById('login-btn');
    const logoutBtn = document.getElementById('logout-btn');
    const userMenu = document.getElementById('user-menu');

    if (authBtn) {
        authBtn.textContent = token ? 'Logout' : 'Sign In';
        authBtn.onclick = token ? handleLogout : () => {
            window.location.href = 'login.html';
        };
    }

    if (loginBtn) {
        loginBtn.style.display = token ? 'none' : 'inline-flex';
        loginBtn.onclick = (event) => {
            event.preventDefault();
            window.location.href = 'login.html';
        };
    }

    if (userMenu) {
        userMenu.style.display = token ? 'flex' : 'none';
    }

    if (logoutBtn) {
        logoutBtn.onclick = handleLogout;
    }
}

function getAccessToken() {
    if (window.authState?.accessToken) {
        return window.authState.accessToken;
    }
    return sessionStorage.getItem('dropsandgrinds_access_token');
}

function decodeJwtPayload(token) {
    try {
        const parts = token.split('.');
        if (parts.length !== 3) return null;
        const base64Url = parts[1];
        const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
        const padded = base64.padEnd(Math.ceil(base64.length / 4) * 4, '=');
        const json = atob(padded);
        return JSON.parse(json);
    } catch {
        return null;
    }
}

let refreshTimeoutId = null;
function scheduleTokenRefresh() {
    if (refreshTimeoutId) {
        clearTimeout(refreshTimeoutId);
        refreshTimeoutId = null;
    }
    const access = authState.accessToken || sessionStorage.getItem('dropsandgrinds_access_token');
    const refresh = authState.refreshToken || sessionStorage.getItem('dropsandgrinds_refresh_token');
    if (!access || !refresh) return;

    const payload = decodeJwtPayload(access);
    const expSeconds = payload?.exp;
    if (!expSeconds) return;

    const nowMs = Date.now();
    const expMs = expSeconds * 1000;
    const refreshAtMs = Math.max(nowMs + 5_000, expMs - 60_000); // 60s before expiry (min 5s)
    const delayMs = refreshAtMs - nowMs;

    refreshTimeoutId = setTimeout(async () => {
        await refreshAuthToken();
        scheduleTokenRefresh();
    }, delayMs);
}

// Auto-refresh mechanism (if an active session exists in memory)
async function refreshAuthToken() {
    const refreshToken = authState.refreshToken || sessionStorage.getItem('dropsandgrinds_refresh_token');
    if (!refreshToken) return;
    
    try {
        const response = await fetch('/api/auth/refresh', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refreshToken })
        });
        
        if (response.ok) {
            const data = await response.json();
            setAuthTokens(data.access_token, data.refresh_token, data.user_id || authState.userId || 0);
        } else {
            await handleLogout();
        }
    } catch {
        await handleLogout();
    }
}

// Central logout function
async function handleLogout() {
    const refreshToken = authState.refreshToken || sessionStorage.getItem('dropsandgrinds_refresh_token');
    if (refreshToken) {
        // Best effort: invalidate refresh token server-side.
        try {
            await fetch('/api/auth/logout', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ refresh_token: refreshToken }),
            });
        } catch {
            // ignore network errors during logout
        }
    }

    authState = {
        accessToken: null,
        refreshToken: null,
        userId: null,
        isAuthenticated: false
    };
    window.authState = authState;

    sessionStorage.removeItem('dropsandgrinds_access_token');
    sessionStorage.removeItem('dropsandgrinds_refresh_token');
    sessionStorage.removeItem('dropsandgrinds_user_id');
    sessionStorage.removeItem('dropsandgrinds_is_authenticated');

    window.location.href = 'login.html';
}

function initAuthButton() {
    const btn = document.getElementById('auth-btn');
    if (!btn) return;

    const token = getAccessToken();
    if (!token) {
        btn.textContent = 'Login';
        btn.onclick = () => {
            window.location.href = 'login.html';
        };
        return;
    }

    btn.textContent = 'Logout';
    btn.onclick = () => {
        sessionStorage.removeItem('dropsandgrinds_access_token');
        sessionStorage.removeItem('dropsandgrinds_refresh_token');
        sessionStorage.removeItem('dropsandgrinds_user_id');
        sessionStorage.removeItem('dropsandgrinds_is_authenticated');
        window.location.href = 'login.html';
    };
}

function getAccessToken() {
    if (window.authState?.accessToken) {
        return window.authState.accessToken;
    }
    return sessionStorage.getItem('dropsandgrinds_access_token');
}
