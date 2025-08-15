#!/bin/bash

DISTRIBUTION_ID="E188SG8Z03BNLO"

echo "Getting current CloudFront configuration..."
aws cloudfront get-distribution-config --id $DISTRIBUTION_ID > /tmp/dist-config.json

# Extract just the DistributionConfig and ETag
ETAG=$(jq -r '.ETag' /tmp/dist-config.json)
jq '.DistributionConfig' /tmp/dist-config.json > /tmp/config.json

# Update the ApiGatewayOrigin path back to "/dev"
echo "Setting API Gateway origin path to /dev..."
jq '.Origins.Items |= map(if .Id == "ApiGatewayOrigin" then .OriginPath = "/dev" else . end)' /tmp/config.json > /tmp/updated-config.json

echo "Updating CloudFront distribution..."
aws cloudfront update-distribution \
  --id $DISTRIBUTION_ID \
  --distribution-config file:///tmp/updated-config.json \
  --if-match $ETAG \
  --output json > /tmp/update-result.json

if [ $? -eq 0 ]; then
  echo "CloudFront distribution updated successfully!"
  echo "Distribution will use /dev stage for API Gateway"
else
  echo "Failed to update distribution"
  cat /tmp/update-result.json
fi
