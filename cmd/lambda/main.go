package main

import (
	"context"
	"log"
	"os"

	"flavaflav/internal/db"
	"flavaflav/internal/handlers"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

var httpLambda *httpadapter.HandlerAdapter

// init initializes the Lambda function
func init() {
	// Get table names from environment variables
	membersTable := os.Getenv("DYNAMODB_MEMBERS_TABLE")
	inventoryTable := os.Getenv("DYNAMODB_INVENTORY_TABLE")
	distributionsTable := os.Getenv("DYNAMODB_DISTRIBUTIONS_TABLE")
	listsTable := os.Getenv("DYNAMODB_LISTS_TABLE")

	// Fallback to legacy single table if new variables not set (for backward compatibility)
	if membersTable == "" {
		membersTable = os.Getenv("DYNAMODB_TABLE")
	}

	// Validate required table names
	if membersTable == "" {
		log.Fatal("DYNAMODB_MEMBERS_TABLE environment variable is required")
	}
	if inventoryTable == "" {
		log.Fatal("DYNAMODB_INVENTORY_TABLE environment variable is required")
	}
	if distributionsTable == "" {
		log.Fatal("DYNAMODB_DISTRIBUTIONS_TABLE environment variable is required")
	}
	if listsTable == "" {
		log.Fatal("DYNAMODB_LISTS_TABLE environment variable is required")
	}

	// Initialize DynamoDB client with four tables
	dbClient, err := db.NewDynamoDBClient(membersTable, inventoryTable, distributionsTable, listsTable)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB client: %v", err)
	}

	// Initialize API handlers
	apiHandlers := handlers.NewAPIHandlers(dbClient)

	// Setup routes
	mux := apiHandlers.SetupRoutes()

	// Create Lambda adapter
	httpLambda = httpadapter.New(mux)
}

// Handler is the Lambda function handler
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return httpLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
