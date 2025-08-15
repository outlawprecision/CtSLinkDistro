# FlavaFlav CORS Fix Deployment Guide

## Problem Summary
The frontend was getting 403 errors when trying to access the API because:
1. CloudFront was correctly forwarding `/api/*` requests to API Gateway
2. API Gateway was adding the stage prefix (`/dev`) to the path
3. The Lambda function wasn't recognizing paths like `/dev/api/members`

## Solution Applied

### 1. Frontend Changes (`web/static/app.js`)
- Updated to use the configuration from `config.js` properly
- Changed from hardcoded `window.location.origin` to `window.FLAVAFLAV_CONFIG.API_BASE_URL`
- This ensures all API calls go through the same CloudFront distribution at `/api/*`

### 2. Backend Changes (`internal/handlers/api.go`)
- Modified the route setup to handle API Gateway stage prefixes
- Now registers all routes with multiple patterns: `/api/*`, `/dev/api/*`, `/staging/api/*`, `/prod/api/*`
- This allows the Lambda to handle requests regardless of the stage prefix

## Deployment Steps

### Step 1: Build and Deploy Lambda Function
```bash
# Build the Lambda function with the updated handler
make lambda-build

# Deploy the Lambda update (this will build and upload)
./scripts/deploy.sh -e dev --update-lambda
```

### Step 2: Upload Updated Frontend Files
```bash
# Get the S3 bucket name from your stack outputs
BUCKET_NAME=$(aws cloudformation describe-stacks \
  --stack-name flavaflav-dev \
  --region us-east-1 \
  --query "Stacks[0].Outputs[?OutputKey=='S3BucketName'].OutputValue" \
  --output text)

# Upload the updated frontend files
aws s3 cp web/static/app.js s3://$BUCKET_NAME/app.js --region us-east-1
aws s3 cp web/static/config.js s3://$BUCKET_NAME/config.js --region us-east-1

# Or sync all static files
aws s3 sync web/static/ s3://$BUCKET_NAME/ --region us-east-1
```

### Step 3: Clear CloudFront Cache
```bash
# Get the CloudFront distribution ID
DISTRIBUTION_ID=$(aws cloudformation describe-stacks \
  --stack-name flavaflav-dev \
  --region us-east-1 \
  --query "Stacks[0].Outputs[?OutputKey=='CloudFrontDistributionId'].OutputValue" \
  --output text)

# If the above doesn't work (no output for distribution ID), list distributions
aws cloudfront list-distributions --query "DistributionList.Items[?Comment=='FlavaFlav CDN - dev'].Id" --output text

# Create an invalidation to clear the cache
aws cloudfront create-invalidation \
  --distribution-id $DISTRIBUTION_ID \
  --paths "/*"
```

### Step 4: Test the Application
1. Open your browser to: https://djo8q7hdb90li.cloudfront.net/
2. Open the browser's Developer Tools (F12)
3. Go to the Network tab
4. Try loading different tabs (Members, Inventory, etc.)
5. Verify that API calls to `/api/*` return 200 status codes

## How It Works Now

1. **Frontend** at `https://djo8q7hdb90li.cloudfront.net/` makes requests to `/api/*`
2. **CloudFront** receives requests at `https://djo8q7hdb90li.cloudfront.net/api/*`
3. **CloudFront** forwards these to API Gateway at `https://efw22hfwjh.execute-api.us-east-1.amazonaws.com/dev/api/*`
4. **API Gateway** passes the request to Lambda with the full path including `/dev`
5. **Lambda** now recognizes both `/api/*` and `/dev/api/*` patterns and handles them correctly
6. Response flows back through the same path

## Future Custom Domain Setup

When you add a custom domain (e.g., `flavaflav.yourdomain.com`):
1. The same CloudFront distribution will serve both static content and API
2. No code changes needed - the frontend uses relative paths (`/api/*`)
3. Everything will work seamlessly under your custom domain

## Troubleshooting

### If you still see 403 errors:
1. **Check CloudFront invalidation status**: 
   ```bash
   aws cloudfront list-invalidations --distribution-id $DISTRIBUTION_ID
   ```
2. **Hard refresh your browser**: Ctrl+Shift+R (or Cmd+Shift+R on Mac)
3. **Check Lambda logs**:
   ```bash
   aws logs tail /aws/lambda/flavaflav-api-dev --follow
   ```

### If API calls show old CloudFront URL:
1. Your browser has cached the old JavaScript files
2. Clear browser cache or use incognito/private mode
3. Verify the S3 upload was successful:
   ```bash
   aws s3 ls s3://$BUCKET_NAME/ --region us-east-1
   ```

## Notes
- The solution maintains flexibility for custom domains
- No hardcoded URLs in the frontend
- Works with any API Gateway stage (dev, staging, prod)
- CORS headers are properly set in the Lambda function
