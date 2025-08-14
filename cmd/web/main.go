package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"flavaflav/internal/database"
	"flavaflav/internal/handlers"
	"flavaflav/internal/models"
	"flavaflav/internal/services"
)

func main() {
	// Load configuration
	config := models.DefaultConfig()

	// Override with environment variables if available
	if region := os.Getenv("AWS_REGION"); region != "" {
		config.AWS.Region = region
	}
	if table := os.Getenv("DYNAMODB_TABLE"); table != "" {
		config.AWS.DynamoDBTable = table
	}
	if port := os.Getenv("PORT"); port != "" {
		config.Web.Port = port
	}

	// Initialize database
	db, err := database.NewDynamoDBService(config.AWS.Region, config.AWS.DynamoDBTable)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize services
	memberService := services.NewMemberService(db, config)
	distributionService := services.NewDistributionService(db, memberService, config)
	inventoryService := services.NewInventoryService(db, config)

	// Initialize distribution lists
	ctx := context.Background()
	err = distributionService.InitializeDistributionLists(ctx)
	if err != nil {
		log.Printf("Warning: Failed to initialize distribution lists: %v", err)
	}

	// Initialize inventory
	err = inventoryService.InitializeInventory(ctx)
	if err != nil {
		log.Printf("Warning: Failed to initialize inventory: %v", err)
	}

	// Initialize handlers
	webHandlers := handlers.NewWebHandlers(memberService, distributionService, inventoryService)

	// Setup routes
	mux := webHandlers.SetupRoutes()

	// Start server
	addr := fmt.Sprintf("%s:%s", config.Web.Host, config.Web.Port)
	log.Printf("Starting FlavaFlav web server on %s", addr)
	log.Printf("Health check available at: http://%s/api/health", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
