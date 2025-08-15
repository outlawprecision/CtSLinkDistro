#!/bin/bash

echo "Testing FlavaFlav API endpoints..."
echo "=================================="

# CloudFront URL
CF_URL="https://djo8q7hdb90li.cloudfront.net"

# Direct API Gateway URL for comparison
API_GW_URL="https://efw22hfwjh.execute-api.us-east-1.amazonaws.com/dev"

echo ""
echo "1. Testing CloudFront -> API Gateway routing:"
echo "----------------------------------------------"

echo "Testing /api/health through CloudFront:"
curl -s -o /dev/null -w "Status: %{http_code}\n" "$CF_URL/api/health"

echo ""
echo "Testing /api/members through CloudFront:"
curl -s -o /dev/null -w "Status: %{http_code}\n" "$CF_URL/api/members"

echo ""
echo "Testing /api/inventory through CloudFront:"
curl -s -o /dev/null -w "Status: %{http_code}\n" "$CF_URL/api/inventory"

echo ""
echo "2. Testing Direct API Gateway:"
echo "-------------------------------"

echo "Testing /api/health directly:"
curl -s -o /dev/null -w "Status: %{http_code}\n" "$API_GW_URL/api/health"

echo ""
echo "3. Getting actual response from health endpoint:"
echo "-------------------------------------------------"
curl -s "$CF_URL/api/health" | jq '.' 2>/dev/null || curl -s "$CF_URL/api/health"

echo ""
echo "=================================="
echo "If you see 200 status codes above, the API is working!"
echo ""
echo "IMPORTANT: Clear your browser cache before testing the UI:"
echo "1. Use an incognito/private window, OR"
echo "2. Press Ctrl+Shift+Delete and clear cached files"
echo ""
echo "Then visit: $CF_URL"
