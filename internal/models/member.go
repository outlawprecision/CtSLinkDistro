package models

import (
	"time"
)

// Guild rank constants
const (
	RankBookWorm = "Book Worm" // <30 days, no links
	RankScholar  = "Scholar"   // 30+ days, silver links
	RankSage     = "Sage"      // 90+ days, gold links
	RankMaester  = "Maester"   // Officer, admin access
)

// Member represents a guild member with simplified structure
type Member struct {
	DiscordID      string    `json:"discord_id" dynamodbav:"discord_id"`
	Username       string    `json:"username" dynamodbav:"username"`
	JoinDate       time.Time `json:"join_date" dynamodbav:"join_date"`
	Rank           string    `json:"rank" dynamodbav:"rank"`
	IsOfficer      bool      `json:"is_officer" dynamodbav:"is_officer"`
	SilverEligible bool      `json:"silver_eligible" dynamodbav:"silver_eligible"`
	GoldEligible   bool      `json:"gold_eligible" dynamodbav:"gold_eligible"`
	DaysInGuild    int       `json:"days_in_guild" dynamodbav:"days_in_guild"`
	AddedBy        string    `json:"added_by" dynamodbav:"added_by"`
	AddedDate      time.Time `json:"added_date" dynamodbav:"added_date"`
	UpdatedAt      time.Time `json:"updated_at" dynamodbav:"updated_at"`
}

// NewMember creates a new member with calculated rank and eligibility
func NewMember(discordID, username string, joinDate time.Time, addedBy string) *Member {
	member := &Member{
		DiscordID: discordID,
		Username:  username,
		JoinDate:  joinDate,
		IsOfficer: false,
		AddedBy:   addedBy,
		AddedDate: time.Now(),
		UpdatedAt: time.Now(),
	}

	member.UpdateRankAndEligibility()
	return member
}

// UpdateRankAndEligibility calculates and updates rank and eligibility based on join date
func (m *Member) UpdateRankAndEligibility() {
	m.DaysInGuild = int(time.Since(m.JoinDate).Hours() / 24)
	m.UpdatedAt = time.Now()

	// Don't change officer rank automatically
	if m.IsOfficer {
		m.Rank = RankMaester
		m.SilverEligible = true
		m.GoldEligible = true
		return
	}

	// Calculate rank based on days in guild
	if m.DaysInGuild < 30 {
		m.Rank = RankBookWorm
		m.SilverEligible = false
		m.GoldEligible = false
	} else if m.DaysInGuild < 90 {
		m.Rank = RankScholar
		m.SilverEligible = true
		m.GoldEligible = false
	} else {
		m.Rank = RankSage
		m.SilverEligible = true
		m.GoldEligible = true
	}
}

// PromoteToOfficer promotes member to Maester rank
func (m *Member) PromoteToOfficer() {
	m.IsOfficer = true
	m.Rank = RankMaester
	m.SilverEligible = true
	m.GoldEligible = true
	m.UpdatedAt = time.Now()
}

// DemoteFromOfficer removes officer status and recalculates rank
func (m *Member) DemoteFromOfficer() {
	m.IsOfficer = false
	m.UpdateRankAndEligibility()
}

// CanEditSystem returns true if member has admin privileges
func (m *Member) CanEditSystem() bool {
	return m.IsOfficer && m.Rank == RankMaester
}

// GetRankColor returns a color code for the rank (for UI)
func (m *Member) GetRankColor() string {
	switch m.Rank {
	case RankBookWorm:
		return "#8B4513" // Brown
	case RankScholar:
		return "#C0C0C0" // Silver
	case RankSage:
		return "#FFD700" // Gold
	case RankMaester:
		return "#800080" // Purple
	default:
		return "#666666" // Gray
	}
}
