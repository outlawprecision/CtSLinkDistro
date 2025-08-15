#!/bin/bash

DISTRIBUTION_ID="E188SG8Z03BNLO"

echo "Getting current CloudFront configuration..."
aws cloudfront get-distribution-config --id $DISTRIBUTION_ID > /tmp/dist-config.json

# Extract just the DistributionConfig and ETag
ETAG=$(jq -r '.ETag' /tmp/dist-config.json)
jq '.DistributionConfig' /tmp/dist-config.json > /tmp/config.json

# Update the ApiGatewayOrigin path from "/dev" to ""
echo "Updating API Gateway origin path..."
jq '.Origins.Items |= map(if .Id == "ApiGatewayOrigin" then .OriginPath = "" else . end)' /tmp/config.json > /tmp/updated-config.json

echo "Updating CloudFront distribution..."
aws cloudfront update-distribution \
  --id $DISTRIBUTION_ID \
  --distribution-config file:///tmp/updated-config.json \
  --if-match $ETAG \
  --output json > /tmp/update-result.json

if [ $? -eq 0 ]; then
  echo "CloudFront distribution updated successfully!"
  echo "Creating invalidation..."
  aws cloudfront create-invalidation \
    --distribution-id $DISTRIBUTION_ID \
    --paths "/*" \
    --query 'Invalidation.Id' \
    --output text
  echo "Update complete! Changes may take 5-10 minutes to propagate."
else
  echo "Failed to update distribution"
  cat /tmp/update-result.json
fi
