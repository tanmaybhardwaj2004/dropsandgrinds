// Theme Toggle functionality
(function() {
    const STORAGE_KEY = 'theme-preference';
    const THEME_ATTR = 'data-theme';
    
    // Get stored theme or detect from system preference
    function getStoredTheme() {
        return localStorage.getItem(STORAGE_KEY);
    }
    
    function setStoredTheme(theme) {
        localStorage.setItem(STORAGE_KEY, theme);
    }
    
    function getSystemTheme() {
        return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }
    
    function getEffectiveTheme() {
        const stored = getStoredTheme();
        if (stored) return stored;
        // Dark is the product default (India-first gaming aesthetic). We only respect
        // system preference after the user explicitly opts in by toggling.
        return 'dark';
    }
    
    // Apply theme to document
    function applyTheme(theme) {
        const root = document.documentElement;
        
        if (theme === 'light') {
            root.setAttribute(THEME_ATTR, 'light');
        } else {
            root.removeAttribute(THEME_ATTR);
        }
        
        // Update button icons (SVG)
        const toggleBtn = document.getElementById('theme-toggle');
        if (toggleBtn) {
            const moonIcon = toggleBtn.querySelector('.theme-icon-moon');
            const sunIcon = toggleBtn.querySelector('.theme-icon-sun');
            if (moonIcon && sunIcon) {
                moonIcon.style.display = theme === 'light' ? 'block' : 'none';
                sunIcon.style.display = theme === 'light' ? 'none' : 'block';
            }
            toggleBtn.title = theme === 'light' ? 'Switch to dark mode' : 'Switch to light mode';
        }
        
        // Update theme-color meta tag
        const themeColorMeta = document.querySelector('meta[name="theme-color"]');
        if (themeColorMeta) {
            themeColorMeta.content = theme === 'light' ? '#2563eb' : '#58a6ff';
        }
    }
    
    // Toggle between light and dark
    function toggleTheme() {
        const current = getEffectiveTheme();
        const next = current === 'light' ? 'dark' : 'light';
        setStoredTheme(next);
        applyTheme(next);
    }
    
    // Initialize theme
    function initTheme() {
        const theme = getEffectiveTheme();
        applyTheme(theme);
        
        // Set up toggle button
        const toggleBtn = document.getElementById('theme-toggle');
        if (toggleBtn) {
            toggleBtn.addEventListener('click', toggleTheme);
        }
        
        // Listen for system preference changes
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
            // Only auto-switch if user hasn't set a preference
            if (!getStoredTheme()) {
                applyTheme(e.matches ? 'dark' : 'light');
            }
        });
    }
    
    // Run on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initTheme);
    } else {
        initTheme();
    }
})();
