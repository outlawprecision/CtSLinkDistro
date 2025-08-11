package services

import (
	"context"
	"fmt"
	"time"

	"flavaflav/internal/database"
	"flavaflav/internal/models"
)

// MemberService handles member-related business logic
type MemberService struct {
	db     *database.DynamoDBService
	config *models.Config
}

// NewMemberService creates a new member service
func NewMemberService(db *database.DynamoDBService, config *models.Config) *MemberService {
	return &MemberService{
		db:     db,
		config: config,
	}
}

// CreateMember creates a new guild member
func (s *MemberService) CreateMember(ctx context.Context, discordID, discordUsername string, characterNames []string, guildJoinDate time.Time, role string) (*models.Member, error) {
	// Check if member already exists
	existingMember, err := s.db.GetMember(ctx, discordID)
	if err == nil && existingMember != nil {
		return nil, fmt.Errorf("member with Discord ID %s already exists", discordID)
	}

	now := time.Now()
	member := &models.Member{
		DiscordID:               discordID,
		DiscordUsername:         discordUsername,
		CharacterNames:          characterNames,
		GuildJoinDate:           guildJoinDate,
		Role:                    role,
		WeeklyBossParticipation: false,
		OmniParticipationDates:  []time.Time{},
		OmniAbsenceCount:        0,
		CompensationOwed:        false,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	err = s.db.CreateMember(ctx, member)
	if err != nil {
		return nil, fmt.Errorf("failed to create member: %v", err)
	}

	return member, nil
}

// GetMember retrieves a member by Discord ID
func (s *MemberService) GetMember(ctx context.Context, discordID string) (*models.Member, error) {
	return s.db.GetMember(ctx, discordID)
}

// GetAllMembers retrieves all members
func (s *MemberService) GetAllMembers(ctx context.Context) ([]*models.Member, error) {
	return s.db.GetAllMembers(ctx)
}

// UpdateMember updates an existing member
func (s *MemberService) UpdateMember(ctx context.Context, member *models.Member) error {
	member.UpdatedAt = time.Now()
	return s.db.UpdateMember(ctx, member)
}

// DeleteMember removes a member
func (s *MemberService) DeleteMember(ctx context.Context, discordID string) error {
	return s.db.DeleteMember(ctx, discordID)
}

// MarkWeeklyBossParticipation marks a member as having participated in weekly boss
func (s *MemberService) MarkWeeklyBossParticipation(ctx context.Context, discordID string, participated bool) error {
	member, err := s.db.GetMember(ctx, discordID)
	if err != nil {
		return fmt.Errorf("member not found: %v", err)
	}

	member.WeeklyBossParticipation = participated
	member.UpdatedAt = time.Now()

	return s.db.UpdateMember(ctx, member)
}

// MarkOmniParticipation marks a member as having participated in omni boss
func (s *MemberService) MarkOmniParticipation(ctx context.Context, discordID string, participated bool, omniDate time.Time) error {
	member, err := s.db.GetMember(ctx, discordID)
	if err != nil {
		return fmt.Errorf("member not found: %v", err)
	}

	if participated {
		// Add participation date and reset absence count
		member.OmniParticipationDates = append(member.OmniParticipationDates, omniDate)
		member.LastOmniParticipation = &omniDate
		member.OmniAbsenceCount = 0
	} else {
		// Increment absence count
		member.OmniAbsenceCount++
	}

	member.UpdatedAt = time.Now()
	return s.db.UpdateMember(ctx, member)
}

// GetEligibleMembers returns members eligible for silver or gold links
func (s *MemberService) GetEligibleMembers(ctx context.Context, linkType string) ([]*models.Member, error) {
	allMembers, err := s.db.GetAllMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %v", err)
	}

	var eligibleMembers []*models.Member
	for _, member := range allMembers {
		if linkType == models.LinkTypeSilver && member.IsEligibleForSilver() {
			eligibleMembers = append(eligibleMembers, member)
		} else if linkType == models.LinkTypeGold && member.IsEligibleForGold() {
			eligibleMembers = append(eligibleMembers, member)
		}
	}

	return eligibleMembers, nil
}

// GetActiveEligibleMembers returns members eligible and active (not exceeding absence threshold)
func (s *MemberService) GetActiveEligibleMembers(ctx context.Context, linkType string) ([]*models.Member, error) {
	eligibleMembers, err := s.GetEligibleMembers(ctx, linkType)
	if err != nil {
		return nil, err
	}

	var activeMembers []*models.Member
	for _, member := range eligibleMembers {
		if member.IsActive(s.config.Rules.MaxAbsenceCount) {
			activeMembers = append(activeMembers, member)
		}
	}

	return activeMembers, nil
}

// ResetWeeklyParticipation resets weekly boss participation for all members
func (s *MemberService) ResetWeeklyParticipation(ctx context.Context) error {
	allMembers, err := s.db.GetAllMembers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get members: %v", err)
	}

	for _, member := range allMembers {
		member.WeeklyBossParticipation = false
		member.UpdatedAt = time.Now()

		err = s.db.UpdateMember(ctx, member)
		if err != nil {
			return fmt.Errorf("failed to update member %s: %v", member.DiscordID, err)
		}
	}

	return nil
}

// MarkMemberInactive marks a member as inactive and adds them to compensation queue
func (s *MemberService) MarkMemberInactive(ctx context.Context, discordID string, reason string) error {
	member, err := s.db.GetMember(ctx, discordID)
	if err != nil {
		return fmt.Errorf("member not found: %v", err)
	}

	member.CompensationOwed = true
	member.UpdatedAt = time.Now()

	return s.db.UpdateMember(ctx, member)
}

// GetMemberStatus returns detailed status information for a member
func (s *MemberService) GetMemberStatus(ctx context.Context, discordID string) (*MemberStatus, error) {
	member, err := s.db.GetMember(ctx, discordID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %v", err)
	}

	linkHistory, err := s.db.GetLinkHistoryByMember(ctx, discordID)
	if err != nil {
		linkHistory = []*models.LinkHistory{} // Continue with empty history if error
	}

	status := &MemberStatus{
		Member:            member,
		DaysInGuild:       member.GetDaysInGuild(),
		SilverEligible:    member.IsEligibleForSilver(),
		GoldEligible:      member.IsEligibleForGold(),
		IsActive:          member.IsActive(s.config.Rules.MaxAbsenceCount),
		LinkHistory:       linkHistory,
		LastSilverLink:    getLastLinkByType(linkHistory, models.LinkTypeSilver),
		LastGoldLink:      getLastLinkByType(linkHistory, models.LinkTypeGold),
		TotalSilverLinks:  countLinksByType(linkHistory, models.LinkTypeSilver),
		TotalGoldLinks:    countLinksByType(linkHistory, models.LinkTypeGold),
		CompensationLinks: countCompensationLinks(linkHistory),
	}

	return status, nil
}

// MemberStatus represents detailed member status information
type MemberStatus struct {
	Member            *models.Member
	DaysInGuild       int
	SilverEligible    bool
	GoldEligible      bool
	IsActive          bool
	LinkHistory       []*models.LinkHistory
	LastSilverLink    *models.LinkHistory
	LastGoldLink      *models.LinkHistory
	TotalSilverLinks  int
	TotalGoldLinks    int
	CompensationLinks int
}

// Helper functions
func getLastLinkByType(history []*models.LinkHistory, linkType string) *models.LinkHistory {
	for _, link := range history {
		if link.LinkType == linkType {
			return link
		}
	}
	return nil
}

func countLinksByType(history []*models.LinkHistory, linkType string) int {
	count := 0
	for _, link := range history {
		if link.LinkType == linkType {
			count++
		}
	}
	return count
}

func countCompensationLinks(history []*models.LinkHistory) int {
	count := 0
	for _, link := range history {
		if link.IsCompensation {
			count++
		}
	}
	return count
}
