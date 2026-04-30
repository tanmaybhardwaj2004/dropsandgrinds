// Frontend Logger - Logs to localStorage and console
(function() {
    const STORAGE_KEY = 'frontend_logs';
    const MAX_LOGS = 1000;
    const LOG_LEVELS = {
        DEBUG: 0,
        INFO: 1,
        WARN: 2,
        ERROR: 3
    };
    
    let currentLevel = LOG_LEVELS.INFO;
    let logs = [];
    
    // Load logs from localStorage
    function loadLogs() {
        try {
            const stored = localStorage.getItem(STORAGE_KEY);
            if (stored) {
                logs = JSON.parse(stored);
            }
        } catch (e) {
            console.error('Failed to load logs from localStorage:', e);
        }
    }
    
    // Save logs to localStorage
    function saveLogs() {
        try {
            localStorage.setItem(STORAGE_KEY, JSON.stringify(logs));
        } catch (e) {
            console.error('Failed to save logs to localStorage:', e);
        }
    }
    
    // Add a log entry
    function addLog(level, message, details = {}) {
        const entry = {
            timestamp: new Date().toISOString(),
            level: level,
            message: message,
            details: details,
            url: window.location.href,
            userAgent: navigator.userAgent
        };
        
        logs.push(entry);
        
        // Trim logs if too many
        if (logs.length > MAX_LOGS) {
            logs = logs.slice(-MAX_LOGS);
        }
        
        saveLogs();
        
        // Also log to console
        const consoleMethod = level === 'ERROR' ? 'error' : level === 'WARN' ? 'warn' : 'log';
        console[consoleMethod](`[${level}] ${message}`, details);
    }
    
    // Public API
    window.FrontendLogger = {
        init: function(config = {}) {
            loadLogs();
            
            if (config.level) {
                currentLevel = LOG_LEVELS[config.level] || LOG_LEVELS.INFO;
            }
            
            // Log frontend startup
            this.info('Frontend initialized', {
                version: config.version || '1.0',
                url: window.location.href,
                timestamp: new Date().toISOString()
            });
            
            // Log page load
            window.addEventListener('load', () => {
                this.info('Page loaded', {
                    url: window.location.href,
                    loadTime: performance.now().toFixed(2) + 'ms'
                });
            });
            
            // Log page unload
            window.addEventListener('beforeunload', () => {
                this.info('Page unloading', {
                    url: window.location.href,
                    sessionDuration: ((Date.now() - window.performance.timing.navigationStart) / 1000).toFixed(2) + 's'
                });
            });
            
            // Log errors
            window.addEventListener('error', (event) => {
                this.error('JavaScript error', {
                    message: event.message,
                    filename: event.filename,
                    lineno: event.lineno,
                    colno: event.colno,
                    stack: event.error?.stack
                });
            });
            
            // Log unhandled promise rejections
            window.addEventListener('unhandledrejection', (event) => {
                this.error('Unhandled promise rejection', {
                    reason: event.reason,
                    promise: event.promise
                });
            });
        },
        
        debug: function(message, details) {
            if (currentLevel <= LOG_LEVELS.DEBUG) {
                addLog('DEBUG', message, details);
            }
        },
        
        info: function(message, details) {
            if (currentLevel <= LOG_LEVELS.INFO) {
                addLog('INFO', message, details);
            }
        },
        
        warn: function(message, details) {
            if (currentLevel <= LOG_LEVELS.WARN) {
                addLog('WARN', message, details);
            }
        },
        
        error: function(message, details) {
            if (currentLevel <= LOG_LEVELS.ERROR) {
                addLog('ERROR', message, details);
            }
        },
        
        // Get all logs
        getLogs: function() {
            return logs;
        },
        
        // Clear logs
        clearLogs: function() {
            logs = [];
            saveLogs();
            this.info('Logs cleared');
        },
        
        // Export logs as JSON
        exportLogs: function() {
            const dataStr = JSON.stringify(logs, null, 2);
            const dataBlob = new Blob([dataStr], { type: 'application/json' });
            const url = URL.createObjectURL(dataBlob);
            const link = document.createElement('a');
            link.href = url;
            link.download = `frontend_logs_${new Date().toISOString().replace(/[:.]/g, '-')}.json`;
            link.click();
            URL.revokeObjectURL(url);
            this.info('Logs exported');
        },
        
        // Set log level
        setLevel: function(level) {
            currentLevel = LOG_LEVELS[level] || LOG_LEVELS.INFO;
            this.info('Log level changed', { level });
        }
    };
    
    // Auto-initialize
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => window.FrontendLogger.init());
    } else {
        window.FrontendLogger.init();
    }
})();
