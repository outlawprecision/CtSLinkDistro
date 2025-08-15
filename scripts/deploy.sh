#!/bin/bash

# FlavaFlav AWS Deployment Script
set -e

# Configuration
STACK_NAME="flavaflav"
REGION="us-east-1"
ENVIRONMENT=""
TEMPLATE_FILE="cloudformation/flavaflav-infrastructure.yaml"
PARAMETERS_FILE=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            PARAMETERS_FILE="cloudformation/parameters-${ENVIRONMENT}.json"
            shift 2
            ;;
        -r|--region)
            REGION="$2"
            shift 2
            ;;
        -s|--stack-name)
            STACK_NAME="$2"
            shift 2
            ;;
        --delete)
            DELETE_STACK=true
            shift
            ;;
        --update-lambda)
            UPDATE_LAMBDA_ONLY=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -e, --environment ENV    Environment (dev, staging, prod) [REQUIRED]"
            echo "  -r, --region REGION      AWS region [default: us-east-1]"
            echo "  -s, --stack-name NAME    CloudFormation stack name [default: flavaflav]"
            echo "      --delete             Delete the stack instead of creating/updating"
            echo "      --update-lambda      Update Lambda function code only"
            echo "  -h, --help               Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0 -e dev                Deploy dev environment"
            echo "  $0 -e prod -r us-west-2  Deploy prod environment in us-west-2"
            echo "  $0 -e dev --delete       Delete the dev stack"
            echo "  $0 -e dev --update-lambda Update Lambda function code only"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Validate required parameters
if [[ -z "$ENVIRONMENT" ]]; then
    print_error "Environment is required. Use -e or --environment to specify dev, staging, or prod"
    echo ""
    echo "Examples:"
    echo "  $0 -e dev     Deploy to development"
    echo "  $0 -e prod    Deploy to production"
    exit 1
fi

# Validate environment value
if [[ "$ENVIRONMENT" != "dev" && "$ENVIRONMENT" != "staging" && "$ENVIRONMENT" != "prod" ]]; then
    print_error "Invalid environment: $ENVIRONMENT. Must be dev, staging, or prod"
    exit 1
fi

# Set parameters file based on environment
PARAMETERS_FILE="cloudformation/parameters-${ENVIRONMENT}.json"

# Update stack name with environment
FULL_STACK_NAME="${STACK_NAME}-${ENVIRONMENT}"

print_info "Starting FlavaFlav deployment..."
print_info "Environment: ${ENVIRONMENT}"
print_info "Region: ${REGION}"
print_info "Stack Name: ${FULL_STACK_NAME}"

# Check if AWS CLI is installed
if ! command -v aws &> /dev/null; then
    print_error "AWS CLI is not installed. Please install it first."
    exit 1
fi

# Check if parameters file exists
if [[ ! -f "$PARAMETERS_FILE" ]]; then
    print_error "Parameters file not found: $PARAMETERS_FILE"
    print_info "Creating default parameters file..."
    cat > "$PARAMETERS_FILE" << EOF
[
    {
        "ParameterKey": "Environment",
        "ParameterValue": "${ENVIRONMENT}"
    },
    {
        "ParameterKey": "DomainName",
        "ParameterValue": ""
    },
    {
        "ParameterKey": "CertificateArn",
        "ParameterValue": ""
    }
]
EOF
    print_success "Created parameters file: $PARAMETERS_FILE"
    print_warning "Please edit the parameters file if you need custom domain or certificate"
fi

# Delete stack if requested
if [[ "$DELETE_STACK" == "true" ]]; then
    print_warning "Deleting stack: $FULL_STACK_NAME"
    aws cloudformation delete-stack \
        --stack-name "$FULL_STACK_NAME" \
        --region "$REGION"
    
    print_info "Waiting for stack deletion to complete..."
    aws cloudformation wait stack-delete-complete \
        --stack-name "$FULL_STACK_NAME" \
        --region "$REGION"
    
    print_success "Stack deleted successfully!"
    exit 0
fi

# Update Lambda function code only
if [[ "$UPDATE_LAMBDA_ONLY" == "true" ]]; then
    print_info "Building Lambda function..."
    make build-lambda
    
    if [[ ! -f "bootstrap" ]]; then
        print_error "Lambda binary 'bootstrap' not found. Run 'make lambda-build' first."
        exit 1
    fi
    
    # Create and upload new package
    LAMBDA_ZIP_NAME="flavaflav-lambda-${ENVIRONMENT}.zip"
    print_info "Creating Lambda deployment package: $LAMBDA_ZIP_NAME"
    zip -q "$LAMBDA_ZIP_NAME" bootstrap
    
    # Upload to S3
    S3_BUCKET="sherwood-artifacts"
    S3_KEY="flavaflav/lambda/${LAMBDA_ZIP_NAME}"
    print_info "Uploading Lambda package to s3://${S3_BUCKET}/${S3_KEY}"
    aws s3 cp "$LAMBDA_ZIP_NAME" "s3://${S3_BUCKET}/${S3_KEY}" --region "$REGION"
    
    # Get Lambda function name from stack outputs
    LAMBDA_FUNCTION_NAME=$(aws cloudformation describe-stacks \
        --stack-name "$FULL_STACK_NAME" \
        --region "$REGION" \
        --query "Stacks[0].Outputs[?OutputKey=='LambdaFunctionName'].OutputValue" \
        --output text 2>/dev/null || echo "")
    
    if [[ -z "$LAMBDA_FUNCTION_NAME" ]]; then
        print_error "Could not find Lambda function name. Make sure the stack is deployed first."
        exit 1
    fi
    
    print_info "Updating Lambda function: $LAMBDA_FUNCTION_NAME"
    
    # Update function code from S3
    aws lambda update-function-code \
        --function-name "$LAMBDA_FUNCTION_NAME" \
        --s3-bucket "$S3_BUCKET" \
        --s3-key "$S3_KEY" \
        --region "$REGION"
    
    # Clean up local zip
    rm -f "$LAMBDA_ZIP_NAME"
    
    print_success "Lambda function updated successfully!"
    exit 0
fi

# Check if stack exists
STACK_EXISTS=$(aws cloudformation describe-stacks \
    --stack-name "$FULL_STACK_NAME" \
    --region "$REGION" \
    --query "Stacks[0].StackStatus" \
    --output text 2>/dev/null || echo "DOES_NOT_EXIST")

if [[ "$STACK_EXISTS" == "DOES_NOT_EXIST" ]]; then
    print_info "Creating new stack..."
    OPERATION="create-stack"
    WAIT_CONDITION="stack-create-complete"
else
    print_info "Updating existing stack..."
    OPERATION="update-stack"
    WAIT_CONDITION="stack-update-complete"
fi

# Build Lambda function first
print_info "Building Lambda function..."
make build-lambda

if [[ ! -f "bootstrap" ]]; then
    print_error "Lambda binary 'bootstrap' not found. Build failed."
    exit 1
fi

# Create Lambda deployment package
LAMBDA_ZIP_NAME="flavaflav-lambda-${ENVIRONMENT}.zip"
print_info "Creating Lambda deployment package: $LAMBDA_ZIP_NAME"
zip -q "$LAMBDA_ZIP_NAME" bootstrap

# Upload to S3
S3_BUCKET="sherwood-artifacts"
S3_KEY="flavaflav/lambda/${LAMBDA_ZIP_NAME}"
print_info "Uploading Lambda package to s3://${S3_BUCKET}/${S3_KEY}"
aws s3 cp "$LAMBDA_ZIP_NAME" "s3://${S3_BUCKET}/${S3_KEY}" --region "$REGION"

print_success "Lambda package uploaded successfully!"

# Clean up local zip file
rm -f "$LAMBDA_ZIP_NAME"

# Validate template
print_info "Validating CloudFormation template..."
aws cloudformation validate-template \
    --template-body "file://$TEMPLATE_FILE" \
    --region "$REGION" > /dev/null

print_success "Template validation passed!"

# Add Lambda S3 location to parameters
TEMP_PARAMS_FILE="/tmp/params-with-lambda-$$.json"
if [[ -f "$PARAMETERS_FILE" ]]; then
    # Add LambdaCodeBucket and LambdaCodeKey to existing parameters
    jq '. + [{"ParameterKey": "LambdaCodeBucket", "ParameterValue": "'$S3_BUCKET'"}, {"ParameterKey": "LambdaCodeKey", "ParameterValue": "'$S3_KEY'"}]' "$PARAMETERS_FILE" > "$TEMP_PARAMS_FILE"
else
    # Create parameters with Lambda location
    echo '[
        {
            "ParameterKey": "Environment",
            "ParameterValue": "'$ENVIRONMENT'"
        },
        {
            "ParameterKey": "LambdaCodeBucket",
            "ParameterValue": "'$S3_BUCKET'"
        },
        {
            "ParameterKey": "LambdaCodeKey",
            "ParameterValue": "'$S3_KEY'"
        }
    ]' > "$TEMP_PARAMS_FILE"
fi

# Deploy stack
print_info "Deploying CloudFormation stack..."
aws cloudformation "$OPERATION" \
    --stack-name "$FULL_STACK_NAME" \
    --template-body "file://$TEMPLATE_FILE" \
    --parameters "file://$TEMP_PARAMS_FILE" \
    --capabilities CAPABILITY_NAMED_IAM \
    --region "$REGION" \
    --tags Key=Environment,Value="$ENVIRONMENT" Key=Application,Value=FlavaFlav

# Clean up temp file
rm -f "$TEMP_PARAMS_FILE"

# Wait for completion
print_info "Waiting for stack operation to complete..."
aws cloudformation wait "$WAIT_CONDITION" \
    --stack-name "$FULL_STACK_NAME" \
    --region "$REGION"

# Get stack outputs
print_success "Stack deployment completed!"
print_info "Getting stack outputs..."

OUTPUTS=$(aws cloudformation describe-stacks \
    --stack-name "$FULL_STACK_NAME" \
    --region "$REGION" \
    --query "Stacks[0].Outputs")

echo "$OUTPUTS" | jq -r '.[] | "\(.OutputKey): \(.OutputValue)"'

# Lambda is already deployed with correct code from S3
print_success "Lambda function deployed with code from S3!"

# Get important URLs
API_URL=$(echo "$OUTPUTS" | jq -r '.[] | select(.OutputKey=="ApiGatewayUrl") | .OutputValue')
CLOUDFRONT_URL=$(echo "$OUTPUTS" | jq -r '.[] | select(.OutputKey=="CloudFrontUrl") | .OutputValue')
S3_BUCKET=$(echo "$OUTPUTS" | jq -r '.[] | select(.OutputKey=="S3BucketName") | .OutputValue')

# Update config.js with the actual API Gateway URL
if [[ -n "$API_URL" && "$API_URL" != "null" ]]; then
    print_info "Updating config.js with API Gateway URL..."
    # Create a temporary config.js with the actual API URL
    sed "s|__API_GATEWAY_URL__|${API_URL}|g" web/static/config.js > /tmp/config.js.tmp
    cp /tmp/config.js.tmp web/static/config.js
    rm -f /tmp/config.js.tmp
    print_success "config.js updated with API Gateway URL"
fi

# Upload static files to S3
if [[ -n "$S3_BUCKET" && "$S3_BUCKET" != "null" ]]; then
    print_info "Uploading static files to S3..."
    aws s3 sync web/static/ s3://$S3_BUCKET/ --region $REGION --delete
    print_success "Static files uploaded to S3"
    
    # Invalidate CloudFront cache if distribution exists
    if [[ -n "$CLOUDFRONT_URL" && "$CLOUDFRONT_URL" != "null" ]]; then
        DISTRIBUTION_ID=$(aws cloudfront list-distributions --query "DistributionList.Items[?Comment=='FlavaFlav CDN - ${ENVIRONMENT}'].Id" --output text --region $REGION)
        if [[ -n "$DISTRIBUTION_ID" ]]; then
            print_info "Creating CloudFront invalidation..."
            aws cloudfront create-invalidation --distribution-id $DISTRIBUTION_ID --paths "/*" --region $REGION > /dev/null
            print_success "CloudFront cache invalidated"
        fi
    fi
fi

print_success "Deployment completed successfully!"
echo ""
print_info "Important URLs:"
if [[ -n "$API_URL" && "$API_URL" != "null" ]]; then
    echo "  API Gateway: $API_URL"
fi
if [[ -n "$CLOUDFRONT_URL" && "$CLOUDFRONT_URL" != "null" ]]; then
    echo "  CloudFront:  $CLOUDFRONT_URL"
fi

echo ""
print_info "Your application is ready!"
echo "  Access your app at: $CLOUDFRONT_URL"
echo ""
print_info "The four DynamoDB tables have been created:"
echo "  - flavaflav-members-${ENVIRONMENT}"
echo "  - flavaflav-inventory-${ENVIRONMENT}"
echo "  - flavaflav-distributions-${ENVIRONMENT}"
echo "  - flavaflav-lists-${ENVIRONMENT}"
