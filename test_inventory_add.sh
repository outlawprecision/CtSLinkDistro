#!/bin/bash

echo "Testing inventory add endpoint..."
echo ""

# Set timeout to 5 seconds to avoid hanging
response=$(timeout 5 curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST \
  https://xl6a8tnacj.execute-api.us-east-1.amazonaws.com/dev/api/inventory/add \
  -H "Content-Type: application/json" \
  -d '{
    "link_type": "Melee Damage",
    "quality": "gold",
    "count": 1
  }' 2>&1)

if [ $? -eq 124 ]; then
    echo "❌ Request timed out after 5 seconds"
    exit 1
fi

# Extract HTTP status code
http_status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_STATUS:/d')

echo "Response Status: $http_status"
echo "Response Body: $body"
echo ""

if [ "$http_status" = "200" ] || [ "$http_status" = "201" ]; then
    echo "✅ Success! Inventory item added successfully"
    echo "The 500 error has been fixed!"
else
    echo "❌ Failed with status $http_status"
    if [ "$http_status" = "500" ]; then
        echo "Still getting 500 error. Check CloudWatch logs for details."
    fi
fi
