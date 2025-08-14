package models

import (
	"time"
)

// DistributionList represents the current distribution lists for silver and gold links
type DistributionList struct {
	DiscordID         string    `json:"discord_id" dynamodbav:"discord_id"` // Use list_type as discord_id for storage
	ListType          string    `json:"list_type" dynamodbav:"list_type"`   // silver, gold
	EligibleMembers   []string  `json:"eligible_members" dynamodbav:"eligible_members"`
	InactiveMembers   []string  `json:"inactive_members" dynamodbav:"inactive_members"`
	CompensationQueue []string  `json:"compensation_queue" dynamodbav:"compensation_queue"`
	CurrentCycleStart time.Time `json:"current_cycle_start" dynamodbav:"current_cycle_start"`
	LastResetDate     time.Time `json:"last_reset_date" dynamodbav:"last_reset_date"`
	CompletedMembers  []string  `json:"completed_members" dynamodbav:"completed_members"`
	MaxAbsenceCount   int       `json:"max_absence_count" dynamodbav:"max_absence_count"`
	UpdatedAt         time.Time `json:"updated_at" dynamodbav:"updated_at"`
}

// DistributionListType constants
const (
	ListTypeSilver = "silver"
	ListTypeGold   = "gold"
)

// NewDistributionList creates a new distribution list
func NewDistributionList(listType string, maxAbsenceCount int) *DistributionList {
	now := time.Now()
	return &DistributionList{
		DiscordID:         "list_" + listType, // Use a unique identifier for the discord_id field
		ListType:          listType,
		EligibleMembers:   []string{},
		InactiveMembers:   []string{},
		CompensationQueue: []string{},
		CurrentCycleStart: now,
		LastResetDate:     now,
		CompletedMembers:  []string{},
		MaxAbsenceCount:   maxAbsenceCount,
		UpdatedAt:         now,
	}
}

// AddEligibleMember adds a member to the eligible list if not already present
func (dl *DistributionList) AddEligibleMember(discordID string) {
	if !contains(dl.EligibleMembers, discordID) {
		dl.EligibleMembers = append(dl.EligibleMembers, discordID)
		dl.UpdatedAt = time.Now()
	}
}

// RemoveEligibleMember removes a member from the eligible list
func (dl *DistributionList) RemoveEligibleMember(discordID string) {
	dl.EligibleMembers = removeFromSlice(dl.EligibleMembers, discordID)
	dl.UpdatedAt = time.Now()
}

// MarkMemberCompleted moves a member from eligible to completed
func (dl *DistributionList) MarkMemberCompleted(discordID string) {
	dl.RemoveEligibleMember(discordID)
	if !contains(dl.CompletedMembers, discordID) {
		dl.CompletedMembers = append(dl.CompletedMembers, discordID)
	}
	dl.UpdatedAt = time.Now()
}

// MarkMemberInactive moves a member from eligible to inactive
func (dl *DistributionList) MarkMemberInactive(discordID string) {
	dl.RemoveEligibleMember(discordID)
	if !contains(dl.InactiveMembers, discordID) {
		dl.InactiveMembers = append(dl.InactiveMembers, discordID)
	}
	dl.UpdatedAt = time.Now()
}

// AddToCompensationQueue adds a member to the compensation queue
func (dl *DistributionList) AddToCompensationQueue(discordID string) {
	if !contains(dl.CompensationQueue, discordID) {
		dl.CompensationQueue = append(dl.CompensationQueue, discordID)
		dl.UpdatedAt = time.Now()
	}
}

// RemoveFromCompensationQueue removes a member from the compensation queue
func (dl *DistributionList) RemoveFromCompensationQueue(discordID string) {
	dl.CompensationQueue = removeFromSlice(dl.CompensationQueue, discordID)
	dl.UpdatedAt = time.Now()
}

// IsListComplete checks if all eligible members have received links
func (dl *DistributionList) IsListComplete() bool {
	return len(dl.EligibleMembers) == 0
}

// CanForceComplete checks if list can be force completed (has inactive members)
func (dl *DistributionList) CanForceComplete() bool {
	return len(dl.InactiveMembers) > 0
}

// ResetList resets the distribution list for a new cycle
func (dl *DistributionList) ResetList() {
	// Move inactive members to compensation queue
	for _, memberID := range dl.InactiveMembers {
		dl.AddToCompensationQueue(memberID)
	}

	// Reset lists
	dl.EligibleMembers = []string{}
	dl.InactiveMembers = []string{}
	dl.CompletedMembers = []string{}
	dl.CurrentCycleStart = time.Now()
	dl.LastResetDate = time.Now()
	dl.UpdatedAt = time.Now()
}

// GetActiveEligibleCount returns the count of active eligible members
func (dl *DistributionList) GetActiveEligibleCount() int {
	return len(dl.EligibleMembers)
}

// GetCompletionPercentage returns the percentage of members who have completed the cycle
func (dl *DistributionList) GetCompletionPercentage() float64 {
	total := len(dl.EligibleMembers) + len(dl.CompletedMembers) + len(dl.InactiveMembers)
	if total == 0 {
		return 100.0
	}
	return float64(len(dl.CompletedMembers)) / float64(total) * 100.0
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func removeFromSlice(slice []string, item string) []string {
	result := []string{}
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
