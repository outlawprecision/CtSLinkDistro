// FlavaFlav Configuration
const CONFIG = {
    // API Configuration - Direct API Gateway URL
    API_BASE_URL: 'https://xl6a8tnacj.execute-api.us-east-1.amazonaws.com/dev/api',

    // Environment detection
    ENVIRONMENT: 'dev',

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
window.FLAVAFLAV_CONFIG.API_BASE_URL = CONFIG.API_BASE_URL;
