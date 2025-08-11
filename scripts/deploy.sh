#!/bin/bash

# FlavaFlav AWS Deployment Script
set -e

# Configuration
STACK_NAME="flavaflav"
REGION="us-east-1"
ENVIRONMENT="dev"
TEMPLATE_FILE="cloudformation/flavaflav-infrastructure.yaml"
PARAMETERS_FILE="cloudformation/parameters-dev.json"

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
            echo "  -e, --environment ENV    Environment (dev, staging, prod) [default: dev]"
            echo "  -r, --region REGION      AWS region [default: us-east-1]"
            echo "  -s, --stack-name NAME    CloudFormation stack name [default: flavaflav]"
            echo "      --delete             Delete the stack instead of creating/updating"
            echo "      --update-lambda      Update Lambda function code only"
            echo "  -h, --help               Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                       Deploy dev environment"
            echo "  $0 -e prod -r us-west-2 Deploy prod environment in us-west-2"
            echo "  $0 --delete              Delete the stack"
            echo "  $0 --update-lambda       Update Lambda function code only"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

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
    make lambda-build
    
    if [[ ! -f "bootstrap" ]]; then
        print_error "Lambda binary 'bootstrap' not found. Run 'make lambda-build' first."
        exit 1
    fi
    
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
    
    # Create deployment package
    if command -v zip &> /dev/null; then
        zip -q lambda-deployment.zip bootstrap
        DEPLOYMENT_PACKAGE="lambda-deployment.zip"
    else
        tar -czf lambda-deployment.tar.gz bootstrap
        DEPLOYMENT_PACKAGE="lambda-deployment.tar.gz"
        print_warning "Using tar.gz instead of zip (zip not available)"
    fi
    
    # Update function code
    aws lambda update-function-code \
        --function-name "$LAMBDA_FUNCTION_NAME" \
        --zip-file "fileb://$DEPLOYMENT_PACKAGE" \
        --region "$REGION"
    
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

# Validate template
print_info "Validating CloudFormation template..."
aws cloudformation validate-template \
    --template-body "file://$TEMPLATE_FILE" \
    --region "$REGION" > /dev/null

print_success "Template validation passed!"

# Deploy stack
print_info "Deploying CloudFormation stack..."
aws cloudformation "$OPERATION" \
    --stack-name "$FULL_STACK_NAME" \
    --template-body "file://$TEMPLATE_FILE" \
    --parameters "file://$PARAMETERS_FILE" \
    --capabilities CAPABILITY_NAMED_IAM \
    --region "$REGION" \
    --tags Key=Environment,Value="$ENVIRONMENT" Key=Application,Value=FlavaFlav

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

# Build and deploy Lambda function
print_info "Building and deploying Lambda function..."
make lambda-build

if [[ ! -f "bootstrap" ]]; then
    print_error "Lambda binary 'bootstrap' not found."
    exit 1
fi

# Get Lambda function name from outputs
LAMBDA_FUNCTION_NAME=$(echo "$OUTPUTS" | jq -r '.[] | select(.OutputKey=="LambdaFunctionName") | .OutputValue')

if [[ -n "$LAMBDA_FUNCTION_NAME" && "$LAMBDA_FUNCTION_NAME" != "null" ]]; then
    print_info "Updating Lambda function code..."
    
    # Create deployment package
    if command -v zip &> /dev/null; then
        zip -q lambda-deployment.zip bootstrap
        DEPLOYMENT_PACKAGE="lambda-deployment.zip"
    else
        tar -czf lambda-deployment.tar.gz bootstrap
        DEPLOYMENT_PACKAGE="lambda-deployment.tar.gz"
        print_warning "Using tar.gz instead of zip (zip not available)"
    fi
    
    # Update function code
    aws lambda update-function-code \
        --function-name "$LAMBDA_FUNCTION_NAME" \
        --zip-file "fileb://$DEPLOYMENT_PACKAGE" \
        --region "$REGION" > /dev/null
    
    print_success "Lambda function code updated!"
else
    print_warning "Could not find Lambda function name in stack outputs"
fi

# Get important URLs
API_URL=$(echo "$OUTPUTS" | jq -r '.[] | select(.OutputKey=="ApiGatewayUrl") | .OutputValue')
CLOUDFRONT_URL=$(echo "$OUTPUTS" | jq -r '.[] | select(.OutputKey=="CloudFrontUrl") | .OutputValue')
S3_BUCKET=$(echo "$OUTPUTS" | jq -r '.[] | select(.OutputKey=="S3BucketName") | .OutputValue')

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
print_info "Next steps:"
echo "  1. Upload static files to S3 bucket: $S3_BUCKET"
echo "  2. Test the API endpoints"
echo "  3. Configure your Discord bot (if needed)"
echo ""
print_info "To upload static files:"
echo "  aws s3 sync web/static/ s3://$S3_BUCKET/ --region $REGION"
