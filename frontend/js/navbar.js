// Shared Navbar Component
// Usage: Include this script in your HTML and call initNavbar() in DOMContentLoaded

const navbarTemplate = `
<nav class="nav">
    <div class="container flex items-center justify-between" style="height: 64px;">
        <a href="index.html" class="flex items-center gap-2" style="text-decoration: none;">
            <div style="width: 36px; height: 36px; background: var(--gradient-primary); border-radius: var(--radius-md); display: flex; align-items: center; justify-content: center; font-size: 1.2rem;"><i data-lucide="gamepad-2"></i></div>
            <span style="font-family: var(--font-display); font-size: var(--text-xl); font-weight: 700; color: var(--color-text-primary);">DropsAndGrinds</span>
        </a>
        
        <div class="flex items-center gap-1" id="nav-links">
            <a href="index.html" class="nav-link" data-page="index">
                <i data-lucide="badge-percent"></i>
                Deals
            </a>
            <a href="search.html" class="nav-link" data-page="search">
                <i data-lucide="search"></i>
                Search
            </a>
            <a href="library.html" class="nav-link" data-page="library">
                <i data-lucide="library"></i>
                Library
            </a>
            <a href="wishlist.html" class="nav-link" data-page="wishlist">
                <i data-lucide="heart"></i>
                Wishlist
            </a>
            <a href="savings.html" class="nav-link" data-page="savings">
                <i data-lucide="badge-dollar-sign"></i>
                Savings
            </a>
            <a href="bundle.html" class="nav-link" data-page="bundle">
                <i data-lucide="package"></i>
                Bundle
            </a>
            <a href="buy-timing.html" class="nav-link" data-page="buy-timing">
                <i data-lucide="clock"></i>
                Timing
            </a>
            <a href="about.html" class="nav-link" data-page="about">
                <i data-lucide="info"></i>
                About
            </a>
        </div>
        
        <div class="flex items-center gap-3">
            <button id="theme-toggle" class="btn btn-ghost btn-sm" title="Toggle theme">
                <i data-lucide="moon" class="theme-icon-moon"></i>
                <i data-lucide="sun" class="theme-icon-sun" style="display: none;"></i>
            </button>
            <button class="btn btn-primary btn-sm" id="auth-btn" onclick="window.location.href='login.html'">Sign In</button>
        </div>
    </div>
</nav>
`;

function initNavbar() {
    // Find existing nav and replace with template, or insert at body start
    const existingNav = document.querySelector('.nav');
    if (existingNav) {
        existingNav.outerHTML = navbarTemplate;
    } else {
        // Insert as first child of body
        document.body.insertAdjacentHTML('afterbegin', navbarTemplate);
    }
    
    // Set active page based on current URL
    const currentPage = window.location.pathname.split('/').pop().replace('.html', '') || 'index';
    const activeLink = document.querySelector(`[data-page="${currentPage}"]`);
    if (activeLink) {
        activeLink.classList.add('active');
    }
    
    // Initialize auth button
    initNavbarAuth();
    
    // Initialize theme toggle
    initNavbarTheme();
}

function initNavbarAuth() {
    const btn = document.getElementById('auth-btn');
    if (!btn) return;
    
    const token = sessionStorage.getItem('dropsandgrinds_access_token');
    if (token) {
        btn.textContent = 'Logout';
        btn.onclick = () => {
            sessionStorage.removeItem('dropsandgrinds_access_token');
            sessionStorage.removeItem('dropsandgrinds_refresh_token');
            sessionStorage.removeItem('dropsandgrinds_user_id');
            sessionStorage.removeItem('dropsandgrinds_is_authenticated');
            window.location.href = 'index.html';
        };
    }
}

function initNavbarTheme() {
    const themeToggle = document.getElementById('theme-toggle');
    if (!themeToggle) return;
    
    const moonIcon = themeToggle.querySelector('.theme-icon-moon');
    const sunIcon = themeToggle.querySelector('.theme-icon-sun');
    
    // Check for saved theme preference
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
    
    if (savedTheme === 'light') {
        if (moonIcon) moonIcon.style.display = 'none';
        if (sunIcon) sunIcon.style.display = 'block';
    }
    
    themeToggle.addEventListener('click', () => {
        const currentTheme = document.documentElement.getAttribute('data-theme');
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        
        document.documentElement.setAttribute('data-theme', newTheme);
        localStorage.setItem('theme', newTheme);
        
        if (moonIcon) moonIcon.style.display = newTheme === 'dark' ? 'block' : 'none';
        if (sunIcon) sunIcon.style.display = newTheme === 'light' ? 'block' : 'none';
    });
}
