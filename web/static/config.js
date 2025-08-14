// Dynamic API configuration - reads from meta tag, no hardcoded URLs
function getApiBase() {
    // First, try to get from meta tag
    const metaTag = document.querySelector('meta[name="api-base-url"]');
    if (metaTag && metaTag.content) {
        return metaTag.content;
    }

    // Fallback: try to detect from current location
    const hostname = window.location.hostname;

    // For localhost development, assume local server
    if (hostname === 'localhost' || hostname === '127.0.0.1') {
        return '/api'; // Assume local development uses relative paths
    }

    // Default fallback
    return '/api';
}

// Determine environment
function getEnvironment() {
    const hostname = window.location.hostname;

    if (hostname === 'localhost' || hostname === '127.0.0.1') {
        return 'local';
    }

    if (hostname.includes('cloudfront.net')) {
        return 'production';
    }

    return 'unknown';
}

// Export configuration
window.FlavaFlavConfig = {
    API_BASE: getApiBase(),
    ENVIRONMENT: getEnvironment(),
    HOSTNAME: window.location.hostname
};

// Debug logging
console.log('FlavaFlav Config:', window.FlavaFlavConfig);
