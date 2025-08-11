package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"flavaflav/internal/database"
	"flavaflav/internal/models"
)

// DistributionService handles link distribution logic
type DistributionService struct {
	db            database.DatabaseService
	memberService *MemberService
	config        *models.Config
}

// NewDistributionService creates a new distribution service
func NewDistributionService(db database.DatabaseService, memberService *MemberService, config *models.Config) *DistributionService {
	return &DistributionService{
		db:            db,
		memberService: memberService,
		config:        config,
	}
}

// InitializeDistributionLists creates initial distribution lists if they don't exist
func (s *DistributionService) InitializeDistributionLists(ctx context.Context) error {
	// Initialize silver list
	_, err := s.db.GetDistributionList(ctx, models.ListTypeSilver)
	if err != nil {
		silverList := models.NewDistributionList(models.ListTypeSilver, s.config.Rules.MaxAbsenceCount)
		err = s.db.CreateDistributionList(ctx, silverList)
		if err != nil {
			return fmt.Errorf("failed to create silver distribution list: %v", err)
		}
	}

	// Initialize gold list
	_, err = s.db.GetDistributionList(ctx, models.ListTypeGold)
	if err != nil {
		goldList := models.NewDistributionList(models.ListTypeGold, s.config.Rules.MaxAbsenceCount)
		err = s.db.CreateDistributionList(ctx, goldList)
		if err != nil {
			return fmt.Errorf("failed to create gold distribution list: %v", err)
		}
	}

	return nil
}

// UpdateDistributionLists updates the distribution lists based on current member eligibility
func (s *DistributionService) UpdateDistributionLists(ctx context.Context) error {
	// Update silver list
	err := s.updateDistributionList(ctx, models.ListTypeSilver)
	if err != nil {
		return fmt.Errorf("failed to update silver list: %v", err)
	}

	// Update gold list
	err = s.updateDistributionList(ctx, models.ListTypeGold)
	if err != nil {
		return fmt.Errorf("failed to update gold list: %v", err)
	}

	return nil
}

// updateDistributionList updates a specific distribution list
func (s *DistributionService) updateDistributionList(ctx context.Context, listType string) error {
	distributionList, err := s.db.GetDistributionList(ctx, listType)
	if err != nil {
		return fmt.Errorf("failed to get distribution list: %v", err)
	}

	// Get currently eligible members
	eligibleMembers, err := s.memberService.GetActiveEligibleMembers(ctx, listType)
	if err != nil {
		return fmt.Errorf("failed to get eligible members: %v", err)
	}

	// Clear current eligible list and rebuild
	distributionList.EligibleMembers = []string{}

	// Add eligible members who haven't completed this cycle
	for _, member := range eligibleMembers {
		if !contains(distributionList.CompletedMembers, member.DiscordID) &&
			!contains(distributionList.InactiveMembers, member.DiscordID) {
			distributionList.AddEligibleMember(member.DiscordID)
		}
	}

	// Check for members who should be marked inactive due to absence count
	allMembers, err := s.memberService.GetAllMembers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all members: %v", err)
	}

	for _, member := range allMembers {
		if !member.IsActive(s.config.Rules.MaxAbsenceCount) &&
			contains(distributionList.EligibleMembers, member.DiscordID) {
			distributionList.MarkMemberInactive(member.DiscordID)
		}
	}

	return s.db.UpdateDistributionList(ctx, distributionList)
}

// SelectRandomWinner selects a random winner from the eligible list
func (s *DistributionService) SelectRandomWinner(ctx context.Context, listType string) (*WinnerResult, error) {
	distributionList, err := s.db.GetDistributionList(ctx, listType)
	if err != nil {
		return nil, fmt.Errorf("failed to get distribution list: %v", err)
	}

	// Check compensation queue first
	if len(distributionList.CompensationQueue) > 0 {
		winnerID := distributionList.CompensationQueue[0]
		distributionList.RemoveFromCompensationQueue(winnerID)

		member, err := s.memberService.GetMember(ctx, winnerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get winner member: %v", err)
		}

		// Create link history record
		linkHistory := models.NewLinkHistory(
			member.DiscordID,
			member.DiscordUsername,
			listType,
			time.Now(),
			true, // This is compensation
			"Compensation for missed omni events",
		)

		err = s.db.CreateLinkHistory(ctx, linkHistory)
		if err != nil {
			return nil, fmt.Errorf("failed to create link history: %v", err)
		}

		// Update member compensation status
		member.CompensationOwed = false
		err = s.memberService.UpdateMember(ctx, member)
		if err != nil {
			return nil, fmt.Errorf("failed to update member: %v", err)
		}

		// Update distribution list
		err = s.db.UpdateDistributionList(ctx, distributionList)
		if err != nil {
			return nil, fmt.Errorf("failed to update distribution list: %v", err)
		}

		return &WinnerResult{
			Winner:         member,
			LinkHistory:    linkHistory,
			IsCompensation: true,
			ListStatus:     s.getListStatus(distributionList),
		}, nil
	}

	// No compensation queue, select from eligible members
	if len(distributionList.EligibleMembers) == 0 {
		return nil, fmt.Errorf("no eligible members available for %s links", listType)
	}

	// Select random winner
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(distributionList.EligibleMembers))
	winnerID := distributionList.EligibleMembers[randomIndex]

	member, err := s.memberService.GetMember(ctx, winnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get winner member: %v", err)
	}

	// Create link history record
	linkHistory := models.NewLinkHistory(
		member.DiscordID,
		member.DiscordUsername,
		listType,
		time.Now(),
		false, // Not compensation
		"",
	)

	err = s.db.CreateLinkHistory(ctx, linkHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to create link history: %v", err)
	}

	// Mark member as completed
	distributionList.MarkMemberCompleted(winnerID)

	// Check if list is complete and should reset
	if distributionList.IsListComplete() {
		distributionList.ResetList()
	}

	// Update distribution list
	err = s.db.UpdateDistributionList(ctx, distributionList)
	if err != nil {
		return nil, fmt.Errorf("failed to update distribution list: %v", err)
	}

	return &WinnerResult{
		Winner:         member,
		LinkHistory:    linkHistory,
		IsCompensation: false,
		ListStatus:     s.getListStatus(distributionList),
	}, nil
}

// ForceCompleteList forces completion of a distribution list
func (s *DistributionService) ForceCompleteList(ctx context.Context, listType string, reason string) error {
	distributionList, err := s.db.GetDistributionList(ctx, listType)
	if err != nil {
		return fmt.Errorf("failed to get distribution list: %v", err)
	}

	if !distributionList.CanForceComplete() {
		return fmt.Errorf("list cannot be force completed - no inactive members")
	}

	// Move remaining eligible members to inactive (they'll be added to compensation queue)
	for _, memberID := range distributionList.EligibleMembers {
		distributionList.MarkMemberInactive(memberID)
	}

	// Reset the list
	distributionList.ResetList()

	return s.db.UpdateDistributionList(ctx, distributionList)
}

// GetDistributionListStatus returns the current status of a distribution list
func (s *DistributionService) GetDistributionListStatus(ctx context.Context, listType string) (*ListStatus, error) {
	distributionList, err := s.db.GetDistributionList(ctx, listType)
	if err != nil {
		return nil, fmt.Errorf("failed to get distribution list: %v", err)
	}

	return s.getListStatus(distributionList), nil
}

// getListStatus creates a ListStatus from a DistributionList
func (s *DistributionService) getListStatus(distributionList *models.DistributionList) *ListStatus {
	return &ListStatus{
		ListType:             distributionList.ListType,
		EligibleCount:        len(distributionList.EligibleMembers),
		CompletedCount:       len(distributionList.CompletedMembers),
		InactiveCount:        len(distributionList.InactiveMembers),
		CompensationCount:    len(distributionList.CompensationQueue),
		CompletionPercentage: distributionList.GetCompletionPercentage(),
		CurrentCycleStart:    distributionList.CurrentCycleStart,
		LastResetDate:        distributionList.LastResetDate,
		CanForceComplete:     distributionList.CanForceComplete(),
		IsComplete:           distributionList.IsListComplete(),
	}
}

// GetAllListStatuses returns status for both silver and gold lists
func (s *DistributionService) GetAllListStatuses(ctx context.Context) (*AllListStatuses, error) {
	silverStatus, err := s.GetDistributionListStatus(ctx, models.ListTypeSilver)
	if err != nil {
		return nil, fmt.Errorf("failed to get silver list status: %v", err)
	}

	goldStatus, err := s.GetDistributionListStatus(ctx, models.ListTypeGold)
	if err != nil {
		return nil, fmt.Errorf("failed to get gold list status: %v", err)
	}

	return &AllListStatuses{
		Silver: silverStatus,
		Gold:   goldStatus,
	}, nil
}

// WinnerResult represents the result of a winner selection
type WinnerResult struct {
	Winner         *models.Member
	LinkHistory    *models.LinkHistory
	IsCompensation bool
	ListStatus     *ListStatus
}

// ListStatus represents the current status of a distribution list
type ListStatus struct {
	ListType             string    `json:"list_type"`
	EligibleCount        int       `json:"eligible_count"`
	CompletedCount       int       `json:"completed_count"`
	InactiveCount        int       `json:"inactive_count"`
	CompensationCount    int       `json:"compensation_count"`
	CompletionPercentage float64   `json:"completion_percentage"`
	CurrentCycleStart    time.Time `json:"current_cycle_start"`
	LastResetDate        time.Time `json:"last_reset_date"`
	CanForceComplete     bool      `json:"can_force_complete"`
	IsComplete           bool      `json:"is_complete"`
}

// AllListStatuses represents status for both distribution lists
type AllListStatuses struct {
	Silver *ListStatus `json:"silver"`
	Gold   *ListStatus `json:"gold"`
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
