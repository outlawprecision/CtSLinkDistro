# FlavaFlav AWS Deployment Guide

This guide walks you through deploying the FlavaFlav application to AWS using CloudFormation.

## Prerequisites

### Required Tools
- AWS CLI v2 installed and configured
- Go 1.21+ installed
- `jq` command-line JSON processor
- `make` utility

### AWS Requirements
- AWS Account with appropriate permissions
- IAM permissions for:
  - CloudFormation (full access)
  - Lambda (full access)
  - DynamoDB (full access)
  - S3 (full access)
  - CloudFront (full access)
  - API Gateway (full access)
  - IAM (role creation)
  - CloudWatch (logs and alarms)

## Quick Start

### 1. Configure AWS CLI
```bash
aws configure
# Enter your AWS Access Key ID, Secret Access Key, Region, and Output format
```

### 2. Deploy to Development Environment
```bash
# Deploy everything
make deploy-dev

# Or use the script directly
./scripts/deploy.sh -e dev
```

### 3. Upload Static Files
After deployment, upload the web interface files:
```bash
# Get the S3 bucket name from the deployment output
make upload-static BUCKET=your-s3-bucket-name
```

## Detailed Deployment Steps

### Step 1: Build the Application
```bash
# Build all components
make build

# Build Lambda function specifically
make lambda-build
```

### Step 2: Configure Parameters
Edit the parameters file for your environment:
```bash
# For development
vim cloudformation/parameters-dev.json

# For production
cp cloudformation/parameters-dev.json cloudformation/parameters-prod.json
vim cloudformation/parameters-prod.json
```

Example parameters file:
```json
[
    {
        "ParameterKey": "Environment",
        "ParameterValue": "prod"
    },
    {
        "ParameterKey": "DomainName",
        "ParameterValue": "flavaflav.yourdomain.com"
    },
    {
        "ParameterKey": "CertificateArn",
        "ParameterValue": "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
    }
]
```

### Step 3: Deploy Infrastructure
```bash
# Deploy to development
make deploy-dev

# Deploy to staging
make deploy-staging

# Deploy to production
make deploy-prod

# Deploy to custom region
./scripts/deploy.sh -e prod -r us-west-2
```

### Step 4: Upload Static Files
```bash
# Get bucket name from CloudFormation outputs
aws cloudformation describe-stacks \
    --stack-name flavaflav-dev \
    --query "Stacks[0].Outputs[?OutputKey=='S3BucketName'].OutputValue" \
    --output text

# Upload files
make upload-static BUCKET=flavaflav-static-dev-123456789012
```

## Environment-Specific Deployments

### Development Environment
- Stack Name: `flavaflav-dev`
- DynamoDB Table: `flavaflav-dev`
- Lambda Function: `flavaflav-api-dev`
- No custom domain required

```bash
make deploy-dev
```

### Staging Environment
- Stack Name: `flavaflav-staging`
- DynamoDB Table: `flavaflav-staging`
- Lambda Function: `flavaflav-api-staging`
- Optional custom domain

```bash
make deploy-staging
```

### Production Environment
- Stack Name: `flavaflav-prod`
- DynamoDB Table: `flavaflav-prod`
- Lambda Function: `flavaflav-api-prod`
- Custom domain recommended

```bash
make deploy-prod
```

## Custom Domain Setup

### 1. Request SSL Certificate
```bash
# Request certificate in us-east-1 (required for CloudFront)
aws acm request-certificate \
    --domain-name flavaflav.yourdomain.com \
    --validation-method DNS \
    --region us-east-1
```

### 2. Validate Certificate
Follow the DNS validation process in the AWS Console or CLI.

### 3. Update Parameters
Add the certificate ARN to your parameters file:
```json
{
    "ParameterKey": "CertificateArn",
    "ParameterValue": "arn:aws:acm:us-east-1:123456789012:certificate/..."
}
```

### 4. Update DNS
After deployment, create a CNAME record pointing to the CloudFront distribution.

## Updating the Application

### Update Lambda Code Only
```bash
# Quick update for code changes
make update-lambda
```

### Full Infrastructure Update
```bash
# Update entire stack
make deploy-dev
```

### Update Static Files Only
```bash
make upload-static BUCKET=your-bucket-name
```

## Monitoring and Troubleshooting

### View CloudFormation Events
```bash
aws cloudformation describe-stack-events \
    --stack-name flavaflav-dev \
    --query "StackEvents[0:10].[Timestamp,ResourceStatus,ResourceType,LogicalResourceId]" \
    --output table
```

### View Lambda Logs
```bash
aws logs tail /aws/lambda/flavaflav-api-dev --follow
```

### Test API Endpoints
```bash
# Get API Gateway URL
API_URL=$(aws cloudformation describe-stacks \
    --stack-name flavaflav-dev \
    --query "Stacks[0].Outputs[?OutputKey=='ApiGatewayUrl'].OutputValue" \
    --output text)

# Test health endpoint
curl $API_URL/api/health
```

### Common Issues

#### 1. Lambda Function Too Large
If the Lambda deployment package exceeds 50MB:
- Use Lambda Layers for dependencies
- Optimize binary size with build flags
- Consider using container images

#### 2. DynamoDB Throttling
If you see throttling errors:
- Check CloudWatch alarms
- Consider switching to provisioned capacity
- Implement exponential backoff in application

#### 3. CloudFront Cache Issues
If static files aren't updating:
```bash
# Create invalidation
aws cloudfront create-invalidation \
    --distribution-id YOUR_DISTRIBUTION_ID \
    --paths "/*"
```

## Cleanup

### Delete Stack
```bash
# Delete development stack
make delete-stack

# Delete specific environment
./scripts/deploy.sh -e prod --delete
```

### Manual Cleanup
Some resources may need manual cleanup:
- S3 bucket contents
- CloudWatch log groups (if retention is set to never expire)
- Route 53 records (if using custom domain)

## Security Considerations

### IAM Permissions
The deployment creates minimal IAM roles with least-privilege access:
- Lambda execution role with DynamoDB access only
- No public access to DynamoDB tables
- S3 bucket with public read access for static files only

### Network Security
- API Gateway with HTTPS only
- CloudFront with HTTPS redirect
- No direct access to Lambda functions

### Data Protection
- DynamoDB encryption at rest (default)
- Point-in-time recovery enabled
- CloudWatch logs with 14-day retention

## Cost Optimization

### Development Environment
- Use DynamoDB on-demand pricing
- Lambda with minimal memory allocation
- CloudFront with basic caching

### Production Environment
- Consider DynamoDB provisioned capacity for predictable workloads
- Optimize Lambda memory based on performance testing
- Use CloudFront with longer cache TTLs

## Backup and Recovery

### DynamoDB Backup
Point-in-time recovery is enabled by default. For additional protection:
```bash
# Create on-demand backup
aws dynamodb create-backup \
    --table-name flavaflav-prod \
    --backup-name flavaflav-prod-backup-$(date +%Y%m%d)
```

### Application Code Backup
- Source code is in Git repository
- Lambda function code can be downloaded from AWS Console
- CloudFormation templates provide infrastructure as code

## Performance Tuning

### Lambda Optimization
- Monitor execution duration and memory usage
- Adjust memory allocation based on CloudWatch metrics
- Consider provisioned concurrency for consistent performance

### DynamoDB Optimization
- Monitor read/write capacity utilization
- Use appropriate partition keys for even distribution
- Consider Global Secondary Indexes for query patterns

### CloudFront Optimization
- Configure appropriate cache behaviors
- Use compression for static assets
- Monitor cache hit ratio

## Support and Maintenance

### Regular Tasks
- Monitor CloudWatch alarms
- Review AWS costs monthly
- Update dependencies and security patches
- Test disaster recovery procedures

### Scaling Considerations
- Lambda automatically scales with demand
- DynamoDB on-demand scales automatically
- CloudFront handles global distribution
- Consider API Gateway throttling limits for high traffic
