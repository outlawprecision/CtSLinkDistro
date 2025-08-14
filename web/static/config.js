// Configuration for different environments
const CONFIG = {
    development: {
        API_BASE: 'https://kiyajsbp09.execute-api.us-east-1.amazonaws.com/dev/api'
    },
    production: {
        API_BASE: '/api' // Use CloudFront routing in production
    }
};

// Determine environment based on hostname
function getEnvironment() {
    const hostname = window.location.hostname;

    // If running on CloudFront distribution, assume production
    if (hostname.includes('cloudfront.net')) {
        return 'production';
    }

    // If running on localhost or development domain, use development
    if (hostname === 'localhost' || hostname.includes('dev')) {
        return 'development';
    }

    // Default to development for now (since CloudFront API routing has issues)
    return 'development';
}

// Get the current configuration
const CURRENT_ENV = getEnvironment();
const API_BASE = CONFIG[CURRENT_ENV].API_BASE;

// Export for use in other scripts
window.FlavaFlavConfig = {
    API_BASE: API_BASE,
    ENVIRONMENT: CURRENT_ENV
};
