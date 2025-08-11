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
}

// NewWebHandlers creates a new web handlers instance
func NewWebHandlers(memberService *services.MemberService, distributionService *services.DistributionService) *WebHandlers {
	return &WebHandlers{
		memberService:       memberService,
		distributionService: distributionService,
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

	// Utility endpoints
	mux.HandleFunc("/api/utility/reset-weekly", h.EnableCORS(h.ResetWeeklyParticipation))
	mux.HandleFunc("/api/utility/update-lists", h.EnableCORS(h.UpdateDistributionLists))

	// Health check
	mux.HandleFunc("/api/health", h.EnableCORS(h.HealthCheck))

	// Serve static files
	mux.Handle("/", http.FileServer(http.Dir("web/static/")))

	return mux
}
