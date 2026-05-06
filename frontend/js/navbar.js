// Shared Navbar Component
// Usage: Include this script in your HTML and call initNavbar() in DOMContentLoaded

async function initNavbar() {
    const currentPage = window.location.pathname.split('/').pop().replace('.html', '') || 'index';
    const existingNav = document.querySelector('.nav');
    
    if (!existingNav) return;
    
    // Determine which navbar to use
    let navbarFile = 'includes/navbar.html';
    if (currentPage === 'search') {
        navbarFile = 'includes/navbar-search.html';
    }
    
    try {
        const response = await fetch(navbarFile);
        const navbarHTML = await response.text();
        existingNav.outerHTML = navbarHTML;
        
        // Set active page based on current URL
        const activeLink = document.querySelector(`[data-page="${currentPage}"]`);
        if (activeLink) {
            activeLink.classList.add('active');
        }
        
        // Initialize auth button
        initNavbarAuth();
        
        // Initialize theme toggle
        initNavbarTheme();
        
        // Initialize mobile menu toggle
        initMobileMenuToggle();
        
    } catch (error) {
        console.error('Failed to load navbar:', error);
    }
}

function initMobileMenuToggle() {
    const toggle = document.getElementById('mobile-menu-toggle');
    const navLinks = document.getElementById('nav-links');
    
    if (!toggle || !navLinks) return;
    
    toggle.addEventListener('click', () => {
        toggle.classList.toggle('active');
        navLinks.classList.toggle('active');
        document.body.classList.toggle('menu-open');
    });
    
    // Close menu when clicking a link
    navLinks.querySelectorAll('a').forEach(link => {
        link.addEventListener('click', () => {
            toggle.classList.remove('active');
            navLinks.classList.remove('active');
            document.body.classList.remove('menu-open');
        });
    });
    
    // Close menu when clicking outside
    document.addEventListener('click', (e) => {
        if (!toggle.contains(e.target) && !navLinks.contains(e.target)) {
            toggle.classList.remove('active');
            navLinks.classList.remove('active');
            document.body.classList.remove('menu-open');
        }
    });
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
    
    // Check for saved theme preference
    const savedTheme = localStorage.getItem('theme') || 'light';
    document.documentElement.setAttribute('data-theme', savedTheme);
    
    themeToggle.addEventListener('click', () => {
        const currentTheme = document.documentElement.getAttribute('data-theme');
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        
        document.documentElement.setAttribute('data-theme', newTheme);
        localStorage.setItem('theme', newTheme);
    });
}
