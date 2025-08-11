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
	if port := os.Getenv("PORT"); port != "" {
		config.Web.Port = port
	}
	if host := os.Getenv("HOST"); host != "" {
		config.Web.Host = host
	}

	log.Println("🚀 Starting FlavaFlav in LOCAL DEVELOPMENT mode")
	log.Println("📊 Using in-memory database with sample data")

	// Initialize in-memory database with sample data
	memoryDB := database.NewMemoryDatabase()

	// Initialize services
	memberService := services.NewMemberService(memoryDB, config)
	distributionService := services.NewDistributionService(memoryDB, memberService, config)

	// Initialize distribution lists
	ctx := context.Background()
	err := distributionService.InitializeDistributionLists(ctx)
	if err != nil {
		log.Printf("Warning: Failed to initialize distribution lists: %v", err)
	}

	// Initialize handlers
	webHandlers := handlers.NewWebHandlers(memberService, distributionService)

	// Setup routes
	mux := webHandlers.SetupRoutes()

	// Start server
	addr := fmt.Sprintf("%s:%s", config.Web.Host, config.Web.Port)
	log.Printf("🌐 FlavaFlav web server running on http://%s", addr)
	log.Printf("💡 Sample data includes 5 members with different eligibility statuses")
	log.Printf("🎯 Try the picker wheel with both silver and gold links!")
	log.Printf("📋 Check the Members tab to see all sample data")
	log.Printf("🔄 Use Ctrl+C to stop the server")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
