package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"flavaflav/internal/database"
	"flavaflav/internal/handlers"
	"flavaflav/internal/models"
	"flavaflav/internal/services"
)

type LambdaHandler struct {
	webHandlers *handlers.WebHandlers
}

func main() {
	// Load configuration
	config := models.DefaultConfig()

	// Override with environment variables
	if region := os.Getenv("AWS_REGION"); region != "" {
		config.AWS.Region = region
	}
	if table := os.Getenv("DYNAMODB_TABLE"); table != "" {
		config.AWS.DynamoDBTable = table
	}

	// Get inventory table name
	inventoryTableName := os.Getenv("DYNAMODB_INVENTORY_TABLE")
	if inventoryTableName == "" {
		inventoryTableName = "flavaflav-inventory-dev" // Default fallback
	}

	// Initialize database
	db, err := database.NewDynamoDBService(config.AWS.Region, config.AWS.DynamoDBTable, inventoryTableName)
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

	handler := &LambdaHandler{
		webHandlers: webHandlers,
	}

	lambda.Start(handler.HandleRequest)
}

func (h *LambdaHandler) HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Convert Lambda request to HTTP request
	httpReq, err := h.convertToHTTPRequest(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"error": "Failed to convert request: %v"}`, err),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	// Create a response recorder
	recorder := &ResponseRecorder{
		statusCode: 200,
		headers:    make(map[string]string),
		body:       "",
	}

	// Route the request
	h.routeRequest(recorder, httpReq)

	// Convert response back to Lambda format
	return events.APIGatewayProxyResponse{
		StatusCode: recorder.statusCode,
		Headers:    recorder.headers,
		Body:       recorder.body,
	}, nil
}

func (h *LambdaHandler) convertToHTTPRequest(request events.APIGatewayProxyRequest) (*http.Request, error) {
	// Create HTTP request from Lambda event
	httpReq, err := http.NewRequest(request.HTTPMethod, request.Path, nil)
	if err != nil {
		return nil, err
	}

	// Add query parameters
	q := httpReq.URL.Query()
	for key, value := range request.QueryStringParameters {
		q.Add(key, value)
	}
	httpReq.URL.RawQuery = q.Encode()

	// Add headers
	for key, value := range request.Headers {
		httpReq.Header.Add(key, value)
	}

	// Add body if present
	if request.Body != "" {
		httpReq.Header.Set("Content-Length", fmt.Sprintf("%d", len(request.Body)))
	}

	return httpReq, nil
}

func (h *LambdaHandler) routeRequest(w http.ResponseWriter, r *http.Request) {
	// Simple routing based on path
	switch {
	case r.URL.Path == "/api/health":
		h.webHandlers.HealthCheck(w, r)
	case r.URL.Path == "/api/members":
		h.webHandlers.GetMembers(w, r)
	case r.URL.Path == "/api/member":
		h.webHandlers.GetMember(w, r)
	case r.URL.Path == "/api/member/create":
		h.webHandlers.CreateMember(w, r)
	case r.URL.Path == "/api/member/status":
		h.webHandlers.GetMemberStatus(w, r)
	case r.URL.Path == "/api/member/weekly-participation":
		h.webHandlers.UpdateWeeklyParticipation(w, r)
	case r.URL.Path == "/api/member/omni-participation":
		h.webHandlers.UpdateOmniParticipation(w, r)
	case r.URL.Path == "/api/distribution/status":
		h.webHandlers.GetDistributionStatus(w, r)
	case r.URL.Path == "/api/distribution/spin":
		h.webHandlers.SpinWheel(w, r)
	case r.URL.Path == "/api/distribution/force-complete":
		h.webHandlers.ForceCompleteList(w, r)
	case r.URL.Path == "/api/distribution/eligible":
		h.webHandlers.GetEligibleMembers(w, r)
	case r.URL.Path == "/api/inventory":
		h.webHandlers.GetInventory(w, r)
	case r.URL.Path == "/api/inventory/item":
		h.webHandlers.GetInventoryItem(w, r)
	case r.URL.Path == "/api/inventory/summary":
		h.webHandlers.GetInventorySummary(w, r)
	case r.URL.Path == "/api/inventory/update":
		h.webHandlers.UpdateInventoryItem(w, r)
	case r.URL.Path == "/api/inventory/bulk-update":
		h.webHandlers.BulkUpdateInventory(w, r)
	case r.URL.Path == "/api/inventory/transactions":
		h.webHandlers.GetInventoryTransactions(w, r)
	case r.URL.Path == "/api/inventory/initialize":
		h.webHandlers.InitializeInventory(w, r)
	case r.URL.Path == "/api/utility/reset-weekly":
		h.webHandlers.ResetWeeklyParticipation(w, r)
	case r.URL.Path == "/api/utility/update-lists":
		h.webHandlers.UpdateDistributionLists(w, r)
	default:
		// Serve static files or return 404
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			h.serveStaticFile(w, r, "web/static/index.html")
		} else if r.URL.Path == "/styles.css" {
			h.serveStaticFile(w, r, "web/static/styles.css")
		} else if r.URL.Path == "/app.js" {
			h.serveStaticFile(w, r, "web/static/app.js")
		} else {
			http.NotFound(w, r)
		}
	}
}

func (h *LambdaHandler) serveStaticFile(w http.ResponseWriter, r *http.Request, filePath string) {
	// In a real Lambda deployment, you'd want to embed static files
	// or serve them from S3. For now, return a simple response.
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>FlavaFlav - Lambda Deployment</title>
</head>
<body>
    <h1>FlavaFlav API</h1>
    <p>This is the Lambda deployment of FlavaFlav. The API endpoints are available at /api/*</p>
    <p>For the full web interface, please deploy the static files to S3 and CloudFront.</p>
</body>
</html>
	`))
}

// ResponseRecorder implements http.ResponseWriter for Lambda
type ResponseRecorder struct {
	statusCode int
	headers    map[string]string
	body       string
}

func (r *ResponseRecorder) Header() http.Header {
	h := make(http.Header)
	for k, v := range r.headers {
		h.Set(k, v)
	}
	return h
}

func (r *ResponseRecorder) Write(data []byte) (int, error) {
	r.body += string(data)
	return len(data), nil
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}
