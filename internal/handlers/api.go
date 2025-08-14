package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"flavaflav/internal/db"
	"flavaflav/internal/models"
)

// APIHandlers contains all HTTP handlers for the API
type APIHandlers struct {
	db *db.DynamoDBClient
}

// NewAPIHandlers creates a new API handlers instance
func NewAPIHandlers(database *db.DynamoDBClient) *APIHandlers {
	return &APIHandlers{
		db: database,
	}
}

// Response structures
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type CreateMemberRequest struct {
	DiscordID string    `json:"discord_id"`
	Username  string    `json:"username"`
	JoinDate  time.Time `json:"join_date"`
}

type AddInventoryRequest struct {
	LinkType string `json:"link_type"`
	Quality  string `json:"quality"`
	Count    int    `json:"count"`
}

type CreateDistributionListRequest struct {
	ListName string `json:"list_name"`
	Quality  string `json:"quality"`
}

type DistributeRequest struct {
	ListID string `json:"list_id"`
	LinkID string `json:"link_id"`
}

// Member endpoints

// GetMembers returns all members
func (h *APIHandlers) GetMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	members, err := h.db.GetAllMembers(r.Context())
	if err != nil {
		h.sendErrorResponse(w, "Failed to get members", http.StatusInternalServerError)
		return
	}

	// Update rank and eligibility for all members
	for _, member := range members {
		member.UpdateRankAndEligibility()
	}

	h.sendSuccessResponse(w, members)
}

// GetMember returns a specific member
func (h *APIHandlers) GetMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	discordID := r.URL.Query().Get("discord_id")
	if discordID == "" {
		h.sendErrorResponse(w, "discord_id parameter is required", http.StatusBadRequest)
		return
	}

	member, err := h.db.GetMember(r.Context(), discordID)
	if err != nil {
		h.sendErrorResponse(w, "Member not found", http.StatusNotFound)
		return
	}

	// Update rank and eligibility
	member.UpdateRankAndEligibility()

	h.sendSuccessResponse(w, member)
}

// CreateMember creates a new member (Maester only)
func (h *APIHandlers) CreateMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add authentication check for Maester role

	var req CreateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.DiscordID == "" || req.Username == "" {
		h.sendErrorResponse(w, "discord_id and username are required", http.StatusBadRequest)
		return
	}

	member := models.NewMember(req.DiscordID, req.Username, req.JoinDate, "web-admin")

	err := h.db.CreateMember(r.Context(), member)
	if err != nil {
		h.sendErrorResponse(w, "Failed to create member", http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, member)
}

// PromoteMember promotes a member to officer (Maester only)
func (h *APIHandlers) PromoteMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add authentication check for Maester role

	discordID := r.URL.Query().Get("discord_id")
	if discordID == "" {
		h.sendErrorResponse(w, "discord_id parameter is required", http.StatusBadRequest)
		return
	}

	member, err := h.db.GetMember(r.Context(), discordID)
	if err != nil {
		h.sendErrorResponse(w, "Member not found", http.StatusNotFound)
		return
	}

	member.PromoteToOfficer()

	err = h.db.UpdateMember(r.Context(), member)
	if err != nil {
		h.sendErrorResponse(w, "Failed to promote member", http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, member)
}

// Inventory endpoints

// GetInventory returns all available inventory links
func (h *APIHandlers) GetInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	quality := r.URL.Query().Get("quality")

	var links []*models.InventoryLink
	var err error

	if quality != "" {
		links, err = h.db.GetAvailableInventoryLinksByQuality(r.Context(), quality)
	} else {
		links, err = h.db.GetAvailableInventoryLinks(r.Context())
	}

	if err != nil {
		h.sendErrorResponse(w, "Failed to get inventory", http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, links)
}

// GetInventorySummary returns inventory counts by type and quality
func (h *APIHandlers) GetInventorySummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	links, err := h.db.GetAvailableInventoryLinks(r.Context())
	if err != nil {
		h.sendErrorResponse(w, "Failed to get inventory", http.StatusInternalServerError)
		return
	}

	// Group by link type and quality
	summary := make(map[string]map[string]int)
	for _, link := range links {
		if summary[link.LinkType] == nil {
			summary[link.LinkType] = make(map[string]int)
		}
		summary[link.LinkType][link.Quality]++
	}

	h.sendSuccessResponse(w, summary)
}

// AddInventory adds new inventory links (Maester only)
func (h *APIHandlers) AddInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add authentication check for Maester role

	var req AddInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.LinkType == "" || req.Quality == "" || req.Count <= 0 {
		h.sendErrorResponse(w, "link_type, quality, and count (>0) are required", http.StatusBadRequest)
		return
	}

	// Get bonus and category for this link type
	bonus := models.GetLinkBonus(req.LinkType, req.Quality)
	category := models.GetLinkCategory(req.LinkType)

	var createdLinks []*models.InventoryLink
	for i := 0; i < req.Count; i++ {
		link := models.NewInventoryLink(req.LinkType, req.Quality, category, bonus, "web-admin")
		err := h.db.CreateInventoryLink(r.Context(), link)
		if err != nil {
			h.sendErrorResponse(w, fmt.Sprintf("Failed to create inventory link %d", i+1), http.StatusInternalServerError)
			return
		}
		createdLinks = append(createdLinks, link)
	}

	h.sendSuccessResponse(w, map[string]interface{}{
		"message": fmt.Sprintf("Added %d %s %s links", req.Count, req.Quality, req.LinkType),
		"links":   createdLinks,
	})
}

// Distribution endpoints

// GetEligibleMembers returns members eligible for a specific quality
func (h *APIHandlers) GetEligibleMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	quality := r.URL.Query().Get("quality")
	if quality != "silver" && quality != "gold" {
		h.sendErrorResponse(w, "quality parameter must be 'silver' or 'gold'", http.StatusBadRequest)
		return
	}

	members, err := h.db.GetAllMembers(r.Context())
	if err != nil {
		h.sendErrorResponse(w, "Failed to get members", http.StatusInternalServerError)
		return
	}

	var eligibleMembers []*models.Member
	for _, member := range members {
		member.UpdateRankAndEligibility()
		if (quality == "silver" && member.SilverEligible) || (quality == "gold" && member.GoldEligible) {
			eligibleMembers = append(eligibleMembers, member)
		}
	}

	h.sendSuccessResponse(w, eligibleMembers)
}

// CreateDistributionList creates a new distribution list (Maester only)
func (h *APIHandlers) CreateDistributionList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add authentication check for Maester role

	var req CreateDistributionListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ListName == "" || (req.Quality != "silver" && req.Quality != "gold") {
		h.sendErrorResponse(w, "list_name and quality (silver/gold) are required", http.StatusBadRequest)
		return
	}

	// Get eligible members
	members, err := h.db.GetAllMembers(r.Context())
	if err != nil {
		h.sendErrorResponse(w, "Failed to get members", http.StatusInternalServerError)
		return
	}

	var eligibleMemberIDs []string
	for _, member := range members {
		member.UpdateRankAndEligibility()
		if (req.Quality == "silver" && member.SilverEligible) || (req.Quality == "gold" && member.GoldEligible) {
			eligibleMemberIDs = append(eligibleMemberIDs, member.DiscordID)
		}
	}

	list := models.NewDistributionList(req.ListName, req.Quality, eligibleMemberIDs, "web-admin")

	err = h.db.CreateDistributionList(r.Context(), list)
	if err != nil {
		h.sendErrorResponse(w, "Failed to create distribution list", http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, list)
}

// GetDistributionLists returns all active distribution lists
func (h *APIHandlers) GetDistributionLists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lists, err := h.db.GetActiveDistributionLists(r.Context())
	if err != nil {
		h.sendErrorResponse(w, "Failed to get distribution lists", http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, lists)
}

// PickWinner randomly selects a winner from a distribution list (Maester only)
func (h *APIHandlers) PickWinner(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add authentication check for Maester role

	listID := r.URL.Query().Get("list_id")
	if listID == "" {
		h.sendErrorResponse(w, "list_id parameter is required", http.StatusBadRequest)
		return
	}

	list, err := h.db.GetDistributionList(r.Context(), listID)
	if err != nil {
		h.sendErrorResponse(w, "Distribution list not found", http.StatusNotFound)
		return
	}

	if len(list.EligibleMembers) == 0 {
		h.sendErrorResponse(w, "No eligible members in list", http.StatusBadRequest)
		return
	}

	// Pick random member
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(list.EligibleMembers))
	winnerID := list.EligibleMembers[randomIndex]

	// Get winner details
	winner, err := h.db.GetMember(r.Context(), winnerID)
	if err != nil {
		h.sendErrorResponse(w, "Winner member not found", http.StatusNotFound)
		return
	}

	h.sendSuccessResponse(w, map[string]interface{}{
		"winner":       winner,
		"list":         list,
		"winner_index": randomIndex,
	})
}

// DistributeLink distributes a specific link to a member (Maester only)
func (h *APIHandlers) DistributeLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add authentication check for Maester role

	var req DistributeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	memberID := r.URL.Query().Get("member_id")
	if memberID == "" || req.LinkID == "" {
		h.sendErrorResponse(w, "member_id and link_id are required", http.StatusBadRequest)
		return
	}

	// Get member and link
	member, err := h.db.GetMember(r.Context(), memberID)
	if err != nil {
		h.sendErrorResponse(w, "Member not found", http.StatusNotFound)
		return
	}

	link, err := h.db.GetInventoryLink(r.Context(), req.LinkID)
	if err != nil {
		h.sendErrorResponse(w, "Link not found", http.StatusNotFound)
		return
	}

	if !link.IsAvailable {
		h.sendErrorResponse(w, "Link is not available", http.StatusBadRequest)
		return
	}

	// Mark link as distributed
	link.MarkDistributed()
	err = h.db.UpdateInventoryLink(r.Context(), link)
	if err != nil {
		h.sendErrorResponse(w, "Failed to update link", http.StatusInternalServerError)
		return
	}

	// Create distribution record
	distribution := models.NewDistribution(
		member.DiscordID,
		member.Username,
		link.LinkID,
		link.LinkType,
		link.Quality,
		link.Bonus,
		"web",
		"web-admin",
	)

	err = h.db.CreateDistribution(r.Context(), distribution)
	if err != nil {
		h.sendErrorResponse(w, "Failed to create distribution record", http.StatusInternalServerError)
		return
	}

	// Remove member from distribution list if provided
	if req.ListID != "" {
		list, err := h.db.GetDistributionList(r.Context(), req.ListID)
		if err == nil {
			list.RemoveMember(memberID)
			h.db.UpdateDistributionList(r.Context(), list)
		}
	}

	h.sendSuccessResponse(w, map[string]interface{}{
		"distribution": distribution,
		"member":       member,
		"link":         link,
	})
}

// GetMemberHistory returns distribution history for a member
func (h *APIHandlers) GetMemberHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	memberID := r.URL.Query().Get("member_id")
	if memberID == "" {
		h.sendErrorResponse(w, "member_id parameter is required", http.StatusBadRequest)
		return
	}

	distributions, err := h.db.GetDistributionsByMember(r.Context(), memberID)
	if err != nil {
		h.sendErrorResponse(w, "Failed to get member history", http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, distributions)
}

// GetAllHistory returns all distribution history (Maester only)
func (h *APIHandlers) GetAllHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add authentication check for Maester role

	distributions, err := h.db.GetAllDistributions(r.Context())
	if err != nil {
		h.sendErrorResponse(w, "Failed to get distribution history", http.StatusInternalServerError)
		return
	}

	h.sendSuccessResponse(w, distributions)
}

// Health check endpoint
func (h *APIHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.sendSuccessResponse(w, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
	})
}

// Helper methods

func (h *APIHandlers) sendSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

func (h *APIHandlers) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   message,
	})
}

// CORS middleware
func (h *APIHandlers) EnableCORS(next http.HandlerFunc) http.HandlerFunc {
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

// SetupRoutes sets up all API routes
func (h *APIHandlers) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Member endpoints
	mux.HandleFunc("/api/members", h.EnableCORS(h.GetMembers))
	mux.HandleFunc("/api/member", h.EnableCORS(h.GetMember))
	mux.HandleFunc("/api/member/create", h.EnableCORS(h.CreateMember))
	mux.HandleFunc("/api/member/promote", h.EnableCORS(h.PromoteMember))
	mux.HandleFunc("/api/member/history", h.EnableCORS(h.GetMemberHistory))

	// Inventory endpoints
	mux.HandleFunc("/api/inventory", h.EnableCORS(h.GetInventory))
	mux.HandleFunc("/api/inventory/summary", h.EnableCORS(h.GetInventorySummary))
	mux.HandleFunc("/api/inventory/add", h.EnableCORS(h.AddInventory))

	// Distribution endpoints
	mux.HandleFunc("/api/distribution/eligible", h.EnableCORS(h.GetEligibleMembers))
	mux.HandleFunc("/api/distribution/lists", h.EnableCORS(h.GetDistributionLists))
	mux.HandleFunc("/api/distribution/create-list", h.EnableCORS(h.CreateDistributionList))
	mux.HandleFunc("/api/distribution/pick-winner", h.EnableCORS(h.PickWinner))
	mux.HandleFunc("/api/distribution/distribute", h.EnableCORS(h.DistributeLink))
	mux.HandleFunc("/api/distribution/history", h.EnableCORS(h.GetAllHistory))

	// Health check
	mux.HandleFunc("/api/health", h.EnableCORS(h.HealthCheck))

	// Serve static files
	mux.Handle("/", http.FileServer(http.Dir("web/static/")))

	return mux
}
