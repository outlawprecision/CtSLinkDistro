#!/bin/bash

# Get S3 bucket name
BUCKET_NAME=$(aws cloudformation describe-stacks \
  --stack-name flavaflav-dev \
  --region us-east-1 \
  --query "Stacks[0].Outputs[?OutputKey=='S3BucketName'].OutputValue" \
  --output text)

echo "S3 Bucket: $BUCKET_NAME"

# Upload the updated frontend files
echo "Uploading updated files to S3..."
aws s3 sync web/static/ s3://$BUCKET_NAME/ --region us-east-1 --delete

# Get CloudFront distribution ID
DISTRIBUTION_ID=$(aws cloudfront list-distributions \
  --query "DistributionList.Items[?Comment=='FlavaFlav CDN - dev'].Id" \
  --output text)

echo "CloudFront Distribution ID: $DISTRIBUTION_ID"

# Create CloudFront invalidation
echo "Creating CloudFront invalidation..."
INVALIDATION_ID=$(aws cloudfront create-invalidation \
  --distribution-id $DISTRIBUTION_ID \
  --paths "/*" \
  --query "Invalidation.Id" \
  --output text)

echo "Invalidation created with ID: $INVALIDATION_ID"

# Wait for invalidation to complete
echo "Waiting for invalidation to complete (this may take a few minutes)..."
aws cloudfront wait invalidation-completed \
  --distribution-id $DISTRIBUTION_ID \
  --id $INVALIDATION_ID

echo "CloudFront cache cleared successfully!"
echo ""
echo "IMPORTANT: Clear your browser cache!"
echo "1. Press Ctrl+Shift+Delete (or Cmd+Shift+Delete on Mac)"
echo "2. Select 'Cached images and files'"
echo "3. Clear the cache"
echo "4. Or use an incognito/private window"
echo ""
echo "Then visit: https://djo8q7hdb90li.cloudfront.net/"
