// FlavaFlav Configuration
const CONFIG = {
    // API Configuration
    API_BASE_URL: window.location.origin + '/api',

    // Environment detection
    ENVIRONMENT: 'dev', // Will be overridden by deployment

    // UI Configuration
    UI: {
        REFRESH_INTERVAL: 30000, // 30 seconds
        ANIMATION_DURATION: 300,
        MAX_RETRIES: 3
    }
};

// Auto-detect environment based on hostname
if (window.location.hostname.includes('cloudfront') || window.location.hostname.includes('amazonaws')) {
    CONFIG.ENVIRONMENT = 'prod';
} else if (window.location.hostname.includes('staging')) {
    CONFIG.ENVIRONMENT = 'staging';
}

// Export for use in other scripts
window.FLAVAFLAV_CONFIG = CONFIG;
