package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"flavaflav/internal/services"
)

// WebHandlers contains all HTTP handlers for the web application
type WebHandlers struct {
	memberService       *services.MemberService
	distributionService *services.DistributionService
	inventoryService    *services.InventoryService
}

// NewWebHandlers creates a new web handlers instance
func NewWebHandlers(memberService *services.MemberService, distributionService *services.DistributionService, inventoryService *services.InventoryService) *WebHandlers {
	return &WebHandlers{
		memberService:       memberService,
		distributionService: distributionService,
		inventoryService:    inventoryService,
	}
}

// API Response structures
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type CreateMemberRequest struct {
	DiscordID       string    `json:"discord_id"`
	DiscordUsername string    `json:"discord_username"`
	CharacterNames  []string  `json:"character_names"`
	GuildJoinDate   time.Time `json:"guild_join_date"`
	Role            string    `json:"role"`
}

type UpdateParticipationRequest struct {
	DiscordID    string    `json:"discord_id"`
	Participated bool      `json:"participated"`
	OmniDate     time.Time `json:"omni_date,omitempty"`
}

type UpdateInventoryRequest struct {
	LinkType string `json:"link_type"`
	Quality  string `json:"quality"`
	NewCount int    `json:"new_count"`
	Reason   string `json:"reason"`
}

type BulkInventoryUpdateRequest struct {
	Updates []services.InventoryUpdate `json:"updates"`
	Reason  string                     `json:"reason"`
}

// Member endpoints

// GetMembers returns all members
func (h *WebHandlers) GetMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	members, err := h.memberService.GetAllMembers(r.Context())
	if err != nil {
		h.sendErrorResponse(w, "Failed to get members", http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, members)
}

// GetMember returns a specific member
func (h *WebHandlers) GetMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	discordID := r.URL.Query().Get("discord_id")
	if discordID == "" {
		h.sendErrorResponse(w, "discord_id parameter is required", http.StatusBadRequest)
		return
	}

	member, err := h.memberService.GetMember(r.Context(), discordID)
	if err != nil {
		h.sendErrorResponse(w, "Member not found", http.StatusNotFound)
		return
	}

	h.sendSuccessResponse(w, member)
}

// CreateMember creates a new member
func (h *WebHandlers) CreateMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	member, err := h.memberService.CreateMember(
		r.Context(),
		req.DiscordID,
		req.DiscordUsername,
		req.CharacterNames,
		req.GuildJoinDate,
		req.Role,
	)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendSuccessResponse(w, member)
}

// GetMemberStatus returns detailed member status
func (h *WebHandlers) GetMemberStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	discordID := r.URL.Query().Get("discord_id")
	if discordID == "" {
		h.sendErrorResponse(w, "discord_id parameter is required", http.StatusBadRequest)
		return
	}

	status, err := h.memberService.GetMemberStatus(r.Context(), discordID)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	h.sendSuccessResponse(w, status)
}

// UpdateWeeklyParticipation updates weekly boss participation
func (h *WebHandlers) UpdateWeeklyParticipation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateParticipationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.memberService.MarkWeeklyBossParticipation(r.Context(), req.DiscordID, req.Participated)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendSuccessResponse(w, map[string]string{"message": "Participation updated successfully"})
}

// UpdateOmniParticipation updates omni boss participation
func (h *WebHandlers) UpdateOmniParticipation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateParticipationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	omniDate := req.OmniDate
	if omniDate.IsZero() {
		omniDate = time.Now()
	}

	err := h.memberService.MarkOmniParticipation(r.Context(), req.DiscordID, req.Participated, omniDate)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendSuccessResponse(w, map[string]string{"message": "Omni participation updated successfully"})
}

// Distribution endpoints

// GetDistributionStatus returns status of both distribution lists
func (h *WebHandlers) GetDistributionStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status, err := h.distributionService.GetAllListStatuses(r.Context())
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, status)
}

// SpinWheel performs random winner selection
func (h *WebHandlers) SpinWheel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	linkType := r.URL.Query().Get("type")
	if linkType != "silver" && linkType != "gold" {
		h.sendErrorResponse(w, "type parameter must be 'silver' or 'gold'", http.StatusBadRequest)
		return
	}

	// Update distribution lists first
	err := h.distributionService.UpdateDistributionLists(r.Context())
	if err != nil {
		h.sendErrorResponse(w, "Failed to update distribution lists", http.StatusInternalServerError)
		return
	}

	result, err := h.distributionService.SelectRandomWinner(r.Context(), linkType)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendSuccessResponse(w, result)
}

// ForceCompleteList forces completion of a distribution list
func (h *WebHandlers) ForceCompleteList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	linkType := r.URL.Query().Get("type")
	if linkType != "silver" && linkType != "gold" {
		h.sendErrorResponse(w, "type parameter must be 'silver' or 'gold'", http.StatusBadRequest)
		return
	}

	reason := r.URL.Query().Get("reason")
	if reason == "" {
		reason = "Manually force completed"
	}

	err := h.distributionService.ForceCompleteList(r.Context(), linkType, reason)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendSuccessResponse(w, map[string]string{"message": "List force completed successfully"})
}

// GetEligibleMembers returns eligible members for a specific link type
func (h *WebHandlers) GetEligibleMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	linkType := r.URL.Query().Get("type")
	if linkType != "silver" && linkType != "gold" {
		h.sendErrorResponse(w, "type parameter must be 'silver' or 'gold'", http.StatusBadRequest)
		return
	}

	activeOnly := r.URL.Query().Get("active_only") == "true"

	var members interface{}
	var err error

	if activeOnly {
		members, err = h.memberService.GetActiveEligibleMembers(r.Context(), linkType)
	} else {
		members, err = h.memberService.GetEligibleMembers(r.Context(), linkType)
	}

	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, members)
}

// Utility endpoints

// ResetWeeklyParticipation resets weekly participation for all members
func (h *WebHandlers) ResetWeeklyParticipation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := h.memberService.ResetWeeklyParticipation(r.Context())
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, map[string]string{"message": "Weekly participation reset successfully"})
}

// UpdateDistributionLists manually updates distribution lists
func (h *WebHandlers) UpdateDistributionLists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := h.distributionService.UpdateDistributionLists(r.Context())
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, map[string]string{"message": "Distribution lists updated successfully"})
}

// Inventory endpoints

// GetInventory returns all inventory items
func (h *WebHandlers) GetInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	category := r.URL.Query().Get("category")

	var items interface{}
	var err error

	if category != "" {
		items, err = h.inventoryService.GetInventoryItemsByCategory(r.Context(), category)
	} else {
		items, err = h.inventoryService.GetAllInventoryItems(r.Context())
	}

	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, items)
}

// GetInventoryItem returns a specific inventory item
func (h *WebHandlers) GetInventoryItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	linkType := r.URL.Query().Get("link_type")
	if linkType == "" {
		h.sendErrorResponse(w, "link_type parameter is required", http.StatusBadRequest)
		return
	}

	item, err := h.inventoryService.GetInventoryItem(r.Context(), linkType)
	if err != nil {
		h.sendErrorResponse(w, "Inventory item not found", http.StatusNotFound)
		return
	}

	h.sendSuccessResponse(w, item)
}

// GetInventorySummary returns inventory summary with totals by category
func (h *WebHandlers) GetInventorySummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	summary, err := h.inventoryService.GetInventorySummary(r.Context())
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, summary)
}

// UpdateInventoryItem updates inventory count for a specific item
func (h *WebHandlers) UpdateInventoryItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.LinkType == "" || req.Quality == "" {
		h.sendErrorResponse(w, "link_type and quality are required", http.StatusBadRequest)
		return
	}

	if req.Quality != "bronze" && req.Quality != "silver" && req.Quality != "gold" {
		h.sendErrorResponse(w, "quality must be 'bronze', 'silver', or 'gold'", http.StatusBadRequest)
		return
	}

	updatedBy := "web-user" // In a real app, this would come from authentication
	err := h.inventoryService.UpdateInventoryCount(r.Context(), req.LinkType, req.Quality, req.NewCount, req.Reason, updatedBy)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendSuccessResponse(w, map[string]string{"message": "Inventory updated successfully"})
}

// BulkUpdateInventory updates multiple inventory items at once
func (h *WebHandlers) BulkUpdateInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BulkInventoryUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Updates) == 0 {
		h.sendErrorResponse(w, "updates array cannot be empty", http.StatusBadRequest)
		return
	}

	updatedBy := "web-user" // In a real app, this would come from authentication
	err := h.inventoryService.BulkUpdateInventory(r.Context(), req.Updates, req.Reason, updatedBy)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendSuccessResponse(w, map[string]string{"message": "Bulk inventory update completed successfully"})
}

// GetInventoryTransactions returns transaction history for a link type
func (h *WebHandlers) GetInventoryTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	linkType := r.URL.Query().Get("link_type")
	if linkType == "" {
		h.sendErrorResponse(w, "link_type parameter is required", http.StatusBadRequest)
		return
	}

	transactions, err := h.inventoryService.GetInventoryTransactions(r.Context(), linkType)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, transactions)
}

// InitializeInventory initializes all link types in the database
func (h *WebHandlers) InitializeInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := h.inventoryService.InitializeInventory(r.Context())
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, map[string]string{"message": "Inventory initialized successfully"})
}

// Health check endpoint
func (h *WebHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.sendSuccessResponse(w, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
	})
}

// Helper methods

func (h *WebHandlers) sendSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

func (h *WebHandlers) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   message,
	})
}

// CORS middleware
func (h *WebHandlers) EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// SetupRoutes sets up all HTTP routes
func (h *WebHandlers) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Member endpoints
	mux.HandleFunc("/api/members", h.EnableCORS(h.GetMembers))
	mux.HandleFunc("/api/member", h.EnableCORS(h.GetMember))
	mux.HandleFunc("/api/member/create", h.EnableCORS(h.CreateMember))
	mux.HandleFunc("/api/member/status", h.EnableCORS(h.GetMemberStatus))
	mux.HandleFunc("/api/member/weekly-participation", h.EnableCORS(h.UpdateWeeklyParticipation))
	mux.HandleFunc("/api/member/omni-participation", h.EnableCORS(h.UpdateOmniParticipation))

	// Distribution endpoints
	mux.HandleFunc("/api/distribution/status", h.EnableCORS(h.GetDistributionStatus))
	mux.HandleFunc("/api/distribution/spin", h.EnableCORS(h.SpinWheel))
	mux.HandleFunc("/api/distribution/force-complete", h.EnableCORS(h.ForceCompleteList))
	mux.HandleFunc("/api/distribution/eligible", h.EnableCORS(h.GetEligibleMembers))

	// Inventory endpoints
	mux.HandleFunc("/api/inventory", h.EnableCORS(h.GetInventory))
	mux.HandleFunc("/api/inventory/item", h.EnableCORS(h.GetInventoryItem))
	mux.HandleFunc("/api/inventory/summary", h.EnableCORS(h.GetInventorySummary))
	mux.HandleFunc("/api/inventory/update", h.EnableCORS(h.UpdateInventoryItem))
	mux.HandleFunc("/api/inventory/bulk-update", h.EnableCORS(h.BulkUpdateInventory))
	mux.HandleFunc("/api/inventory/transactions", h.EnableCORS(h.GetInventoryTransactions))
	mux.HandleFunc("/api/inventory/initialize", h.EnableCORS(h.InitializeInventory))

	// Utility endpoints
	mux.HandleFunc("/api/utility/reset-weekly", h.EnableCORS(h.ResetWeeklyParticipation))
	mux.HandleFunc("/api/utility/update-lists", h.EnableCORS(h.UpdateDistributionLists))

	// Health check
	mux.HandleFunc("/api/health", h.EnableCORS(h.HealthCheck))

	// Serve static files
	mux.Handle("/", http.FileServer(http.Dir("web/static/")))

	return mux
}
