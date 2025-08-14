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
	// Get table name from environment variable
	tableName := os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		log.Fatal("DYNAMODB_TABLE environment variable is required")
	}

	// Initialize DynamoDB client
	dbClient, err := db.NewDynamoDBClient(tableName)
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
