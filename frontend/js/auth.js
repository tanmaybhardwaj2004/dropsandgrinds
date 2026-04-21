// Auth State Management
// JWT Tokens must be stored in memory per project requirements, not localStorage
let authState = {
    accessToken: null,
    refreshToken: null,
    userId: null,
    isAuthenticated: false
};

document.addEventListener('DOMContentLoaded', () => {
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
        errorDiv.textContent = message;
        errorDiv.classList.remove('hidden');
    }
}

function hideError() {
    const errorDiv = document.getElementById('auth-error');
    if (errorDiv) {
        errorDiv.classList.add('hidden');
    }
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
    console.log("Tokens stored securely in memory.");
}

// Auto-refresh mechanism (if an active session exists in memory)
async function refreshAuthToken() {
    if (!authState.refreshToken) return;
    
    try {
        const response = await fetch('/api/auth/refresh', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: authState.refreshToken })
        });
        
        if (response.ok) {
            const data = await response.json();
            authState.accessToken = data.access_token;
            authState.refreshToken = data.refresh_token; 
        } else {
            handleLogout();
        }
    } catch {
        handleLogout();
    }
}

// Central logout function
function handleLogout() {
    authState = {
        accessToken: null,
        refreshToken: null,
        userId: null,
        isAuthenticated: false
    };
    window.location.href = 'login.html';
}
