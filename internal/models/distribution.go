package models

import (
	"time"
)

// Distribution represents a link distributed to a member
type Distribution struct {
	DistributionID string    `json:"distribution_id" dynamodbav:"distribution_id"` // unique ID for this distribution
	MemberID       string    `json:"member_id" dynamodbav:"member_id"`             // Discord ID of member who received link
	MemberUsername string    `json:"member_username" dynamodbav:"member_username"` // Discord username for display
	LinkID         string    `json:"link_id" dynamodbav:"link_id"`                 // ID of the specific link distributed
	LinkType       string    `json:"link_type" dynamodbav:"link_type"`             // e.g., "Melee Damage"
	Quality        string    `json:"quality" dynamodbav:"quality"`                 // bronze, silver, gold
	Bonus          string    `json:"bonus" dynamodbav:"bonus"`                     // e.g., "3.75%"
	Method         string    `json:"method" dynamodbav:"method"`                   // "web" or "discord"
	DistributedBy  string    `json:"distributed_by" dynamodbav:"distributed_by"`   // who gave the link
	DistributedAt  time.Time `json:"distributed_at" dynamodbav:"distributed_at"`
	Notes          string    `json:"notes" dynamodbav:"notes"` // optional notes about this distribution
}

// NewDistribution creates a new distribution record
func NewDistribution(memberID, memberUsername, linkID, linkType, quality, bonus, method, distributedBy string) *Distribution {
	return &Distribution{
		DistributionID: generateDistributionID(),
		MemberID:       memberID,
		MemberUsername: memberUsername,
		LinkID:         linkID,
		LinkType:       linkType,
		Quality:        quality,
		Bonus:          bonus,
		Method:         method,
		DistributedBy:  distributedBy,
		DistributedAt:  time.Now(),
		Notes:          "",
	}
}

// GetDisplayName returns a formatted display name for the distribution
func (d *Distribution) GetDisplayName() string {
	return d.Quality + " " + d.LinkType + " (" + d.Bonus + ")"
}

// generateDistributionID creates a unique ID for a distribution
func generateDistributionID() string {
	return "dist_" + time.Now().Format("20060102150405") + "_" + string(rune(time.Now().UnixNano()%1000))
}

// DistributionList represents a list of eligible members for distribution
type DistributionList struct {
	ListID          string    `json:"list_id" dynamodbav:"list_id"`
	ListName        string    `json:"list_name" dynamodbav:"list_name"`               // e.g., "Silver Links - January 2024"
	Quality         string    `json:"quality" dynamodbav:"quality"`                   // silver or gold
	EligibleMembers []string  `json:"eligible_members" dynamodbav:"eligible_members"` // Discord IDs
	CreatedBy       string    `json:"created_by" dynamodbav:"created_by"`
	CreatedAt       time.Time `json:"created_at" dynamodbav:"created_at"`
	IsActive        bool      `json:"is_active" dynamodbav:"is_active"`
}

// NewDistributionList creates a new distribution list
func NewDistributionList(listName, quality string, eligibleMembers []string, createdBy string) *DistributionList {
	return &DistributionList{
		ListID:          generateListID(),
		ListName:        listName,
		Quality:         quality,
		EligibleMembers: eligibleMembers,
		CreatedBy:       createdBy,
		CreatedAt:       time.Now(),
		IsActive:        true,
	}
}

// RemoveMember removes a member from the eligible list
func (dl *DistributionList) RemoveMember(memberID string) {
	for i, id := range dl.EligibleMembers {
		if id == memberID {
			dl.EligibleMembers = append(dl.EligibleMembers[:i], dl.EligibleMembers[i+1:]...)
			break
		}
	}
}

// HasMember checks if a member is in the eligible list
func (dl *DistributionList) HasMember(memberID string) bool {
	for _, id := range dl.EligibleMembers {
		if id == memberID {
			return true
		}
	}
	return false
}

// GetMemberCount returns the number of eligible members
func (dl *DistributionList) GetMemberCount() int {
	return len(dl.EligibleMembers)
}

// generateListID creates a unique ID for a distribution list
func generateListID() string {
	return "list_" + time.Now().Format("20060102150405")
}
