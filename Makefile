.PHONY: build clean run-web run-bot test help

# Build targets
build: build-web build-bot

build-web:
	@echo "Building web application..."
	go build -o bin/flavaflav-web cmd/web/main.go

build-bot:
	@echo "Building Discord bot..."
	go build -o bin/flavaflav-bot cmd/discord-bot/main.go

# Run targets
run-web:
	@echo "Starting web application..."
	go run cmd/web/main.go

run-local:
	@echo "Starting web application in LOCAL DEVELOPMENT mode..."
	go run cmd/web-local/main.go

run-bot:
	@echo "Starting Discord bot..."
	go run cmd/discord-bot/main.go

# Development targets
dev-web:
	@echo "Starting web application in development mode..."
	air -c .air.toml cmd/web/main.go

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
lambda-build:
	@echo "Building for AWS Lambda..."
	GOOS=linux GOARCH=amd64 go build -o bootstrap cmd/lambda/main.go
	@echo "Lambda binary 'bootstrap' created successfully!"
	@echo "To create deployment package, run: zip lambda-deployment.zip bootstrap"
	@echo "Or use: tar -czf lambda-deployment.tar.gz bootstrap"

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

update-lambda:
	@echo "Updating Lambda function code..."
	./scripts/deploy.sh --update-lambda

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
	@echo "Available targets:"
	@echo "  build         - Build both web app and Discord bot"
	@echo "  build-web     - Build web application only"
	@echo "  build-bot     - Build Discord bot only"
	@echo "  run-web       - Run web application"
	@echo "  run-local     - Run web application with sample data (no AWS needed)"
	@echo "  run-bot       - Run Discord bot"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
	@echo "  lambda-build  - Build for AWS Lambda deployment"
	@echo ""
	@echo "AWS Deployment:"
	@echo "  deploy        - Deploy to AWS (dev environment)"
	@echo "  deploy-dev    - Deploy to dev environment"
	@echo "  deploy-staging- Deploy to staging environment"
	@echo "  deploy-prod   - Deploy to production environment"
	@echo "  update-lambda - Update Lambda function code only"
	@echo "  delete-stack  - Delete CloudFormation stack"
	@echo "  upload-static - Upload static files to S3 (requires BUCKET=name)"
	@echo ""
	@echo "  help          - Show this help message"
