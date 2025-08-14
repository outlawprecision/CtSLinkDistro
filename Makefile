.PHONY: build clean run-bot test help

# Build targets
build: build-lambda build-bot

build-lambda:
	@echo "Building Lambda function..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap cmd/lambda/main.go
	@echo "Lambda binary 'bootstrap' created successfully!"

build-bot:
	@echo "Building Discord bot..."
	go build -o bin/flavaflav-bot cmd/discord-bot/main.go

# Run targets (Note: Lambda functions run in AWS, not locally)
run-bot:
	@echo "Starting Discord bot..."
	go run cmd/discord-bot/main.go

# Development targets

test:
	@echo "Running tests..."
	go test ./...

# Utility targets
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Docker targets
docker-build:
	@echo "Building Docker image..."
	docker build -t flavaflav:latest .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 flavaflav:latest

# AWS Lambda targets
lambda-package: build-lambda
	@echo "Creating Lambda deployment package..."
	@if [ -z "$(ENV)" ]; then \
		echo "Error: ENV variable not set. Usage: make lambda-package ENV=dev"; \
		exit 1; \
	fi
	zip -q flavaflav-lambda-$(ENV).zip bootstrap
	@echo "Lambda package 'flavaflav-lambda-$(ENV).zip' created successfully!"

lambda-upload: lambda-package
	@echo "Uploading Lambda package to S3..."
	@if [ -z "$(ENV)" ]; then \
		echo "Error: ENV variable not set. Usage: make lambda-upload ENV=dev"; \
		exit 1; \
	fi
	aws s3 cp flavaflav-lambda-$(ENV).zip s3://sherwood-artifacts/flavaflav/lambda/
	@echo "Lambda package uploaded to s3://sherwood-artifacts/flavaflav/lambda/flavaflav-lambda-$(ENV).zip"
	@rm -f flavaflav-lambda-$(ENV).zip

# AWS Deployment targets
deploy:
	@echo "Deploying to AWS..."
	./scripts/deploy.sh

deploy-dev:
	@echo "Deploying to dev environment..."
	./scripts/deploy.sh -e dev

deploy-staging:
	@echo "Deploying to staging environment..."
	./scripts/deploy.sh -e staging

deploy-prod:
	@echo "Deploying to production environment..."
	./scripts/deploy.sh -e prod

update-lambda-dev:
	@echo "Updating Lambda function code for dev..."
	./scripts/deploy.sh -e dev --update-lambda

update-lambda-staging:
	@echo "Updating Lambda function code for staging..."
	./scripts/deploy.sh -e staging --update-lambda

update-lambda-prod:
	@echo "Updating Lambda function code for production..."
	./scripts/deploy.sh -e prod --update-lambda

delete-stack:
	@echo "Deleting CloudFormation stack..."
	./scripts/deploy.sh --delete

upload-static:
	@echo "Uploading static files to S3..."
	@if [ -z "$(BUCKET)" ]; then \
		echo "Error: BUCKET variable not set. Usage: make upload-static BUCKET=your-bucket-name"; \
		exit 1; \
	fi
	aws s3 sync web/static/ s3://$(BUCKET)/ --delete

# Help
help:
	@echo "FlavaFlav Serverless Application - Makefile Targets"
	@echo "===================================================="
	@echo ""
	@echo "Build targets:"
	@echo "  build           - Build Lambda function and Discord bot"
	@echo "  build-lambda    - Build Lambda function for AWS deployment"
	@echo "  build-bot       - Build Discord bot"
	@echo ""
	@echo "Run targets:"
	@echo "  run-bot         - Run Discord bot locally"
	@echo "  Note: Lambda functions run in AWS, use 'make deploy-*' to deploy"
	@echo ""
	@echo "Development:"
	@echo "  test            - Run tests"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Install dependencies"
	@echo ""
	@echo "Lambda Deployment:"
	@echo "  lambda-package  - Build and package Lambda (requires ENV=dev/staging/prod)"
	@echo "  lambda-upload   - Build, package and upload to S3 (requires ENV=dev/staging/prod)"
	@echo "  update-lambda-dev     - Update Lambda function code for dev"
	@echo "  update-lambda-staging - Update Lambda function code for staging"
	@echo "  update-lambda-prod    - Update Lambda function code for production"
	@echo ""
	@echo "Full Stack Deployment:"
	@echo "  deploy-dev      - Deploy entire stack to dev environment"
	@echo "  deploy-staging  - Deploy entire stack to staging environment"
	@echo "  deploy-prod     - Deploy entire stack to production environment"
	@echo "  delete-stack    - Delete CloudFormation stack"
	@echo ""
	@echo "Static Files:"
	@echo "  upload-static   - Upload static web files to S3 (requires BUCKET=name)"
	@echo ""
	@echo "  help            - Show this help message"
