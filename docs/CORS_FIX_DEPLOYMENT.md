# CORS Fix Deployment Instructions

## Overview
This document outlines the changes made to fix CORS issues in the FlavaFlav application when deployed to AWS.

## Changes Made

### 1. API Handler CORS Improvements (`internal/handlers/api.go`)
- Enhanced CORS middleware to properly handle preflight requests
- Changed OPTIONS response from `StatusOK (200)` to `StatusNoContent (204)` for better compliance
- Added `X-Requested-With` to allowed headers
- Added `Access-Control-Max-Age` header for caching preflight responses

### 2. CloudFormation Template Updates (`cloudformation/flavaflav-infrastructure.yaml`)
- Added explicit OPTIONS method handlers for both root and proxy resources in API Gateway
- Configured MOCK integration for OPTIONS requests to handle CORS preflight
- Added proper CORS response headers at the API Gateway level
- Updated CloudFront cache behavior to forward necessary headers for API requests:
  - Authorization
  - Content-Type
  - X-Requested-With
  - Accept
  - Origin
  - Referer
- Added OPTIONS to cached methods for better performance

## Deployment Steps

### 1. Build and Deploy Lambda Function

```bash
# Build the Lambda function
make build-lambda

# Deploy the Lambda package to S3 (if using S3 for deployment)
aws s3 cp bin/lambda.zip s3://sherwood-artifacts/flavaflav/lambda/flavaflav-lambda-${ENVIRONMENT}.zip
```

### 2. Update CloudFormation Stack

```bash
# For development environment
aws cloudformation update-stack \
  --stack-name flavaflav-dev \
  --template-body file://cloudformation/flavaflav-infrastructure.yaml \
  --parameters file://cloudformation/parameters-dev.json \
  --capabilities CAPABILITY_NAMED_IAM

# For production environment
aws cloudformation update-stack \
  --stack-name flavaflav-prod \
  --template-body file://cloudformation/flavaflav-infrastructure.yaml \
  --parameters file://cloudformation/parameters-prod.json \
  --capabilities CAPABILITY_NAMED_IAM
```

### 3. Wait for Stack Update to Complete

```bash
# Monitor stack update progress
aws cloudformation wait stack-update-complete --stack-name flavaflav-${ENVIRONMENT}

# Or watch the progress
aws cloudformation describe-stack-events \
  --stack-name flavaflav-${ENVIRONMENT} \
  --query 'StackEvents[0:10]' \
  --output table
```

### 4. Deploy Frontend Files to S3

```bash
# Sync frontend files to S3
aws s3 sync web/static/ s3://flavaflav-static-${ENVIRONMENT}-${AWS_ACCOUNT_ID}/ \
  --delete \
  --cache-control "public, max-age=3600"
```

### 5. Invalidate CloudFront Cache

```bash
# Get CloudFront distribution ID
DISTRIBUTION_ID=$(aws cloudformation describe-stacks \
  --stack-name flavaflav-${ENVIRONMENT} \
  --query 'Stacks[0].Outputs[?OutputKey==`CloudFrontDistributionId`].OutputValue' \
  --output text)

# Create invalidation
aws cloudfront create-invalidation \
  --distribution-id ${DISTRIBUTION_ID} \
  --paths "/*"
```

## Testing CORS

After deployment, test CORS functionality:

### 1. Test Preflight Request
```bash
curl -X OPTIONS \
  https://your-cloudfront-domain.cloudfront.net/api/members \
  -H "Origin: https://your-cloudfront-domain.cloudfront.net" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -v
```

Expected response headers:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization, X-Requested-With`

### 2. Test Actual API Request
```bash
curl -X GET \
  https://your-cloudfront-domain.cloudfront.net/api/members \
  -H "Origin: https://your-cloudfront-domain.cloudfront.net" \
  -v
```

### 3. Browser Console Test
Open the deployed application in your browser and check the console for any CORS errors. The application should now be able to make API calls without 403 errors.

## Troubleshooting

### If CORS errors persist:

1. **Check CloudWatch Logs**
   ```bash
   aws logs tail /aws/lambda/flavaflav-api-${ENVIRONMENT} --follow
   ```

2. **Verify API Gateway Deployment**
   - Ensure the API Gateway deployment includes the new OPTIONS methods
   - Check that the deployment stage matches your environment

3. **CloudFront Cache**
   - Ensure CloudFront invalidation completed successfully
   - Wait 5-10 minutes for global propagation

4. **Test Direct API Gateway URL**
   - Try accessing the API directly via API Gateway URL (bypassing CloudFront)
   - This helps isolate whether the issue is with API Gateway or CloudFront

## Rollback Instructions

If issues occur after deployment:

```bash
# Rollback CloudFormation stack to previous version
aws cloudformation cancel-update-stack --stack-name flavaflav-${ENVIRONMENT}

# Or update with previous template
aws cloudformation update-stack \
  --stack-name flavaflav-${ENVIRONMENT} \
  --template-body file://backup/cloudformation/flavaflav-infrastructure.yaml \
  --parameters file://cloudformation/parameters-${ENVIRONMENT}.json \
  --capabilities CAPABILITY_NAMED_IAM
```

## Additional Notes

- The CORS configuration allows all origins (`*`) for simplicity. In production, consider restricting to specific domains.
- The `Access-Control-Max-Age` is set to 3600 seconds (1 hour) to reduce preflight requests.
- CloudFront is configured to forward necessary headers for API requests while maintaining caching for static content.
