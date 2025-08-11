package models

import (
	"time"
)

// Member represents a guild member
type Member struct {
	DiscordID              string    `json:"discord_id" dynamodbav:"discord_id"`
	DiscordUsername        string    `json:"discord_username" dynamodbav:"discord_username"`
	CharacterNames         []string  `json:"character_names" dynamodbav:"character_names"`
	GuildJoinDate          time.Time `json:"guild_join_date" dynamodbav:"guild_join_date"`
	Role                   string    `json:"role" dynamodbav:"role"` // admin, officer, user
	WeeklyBossParticipation bool     `json:"weekly_boss_participation" dynamodbav:"weekly_boss_participation"`
	OmniParticipationDates []time.Time `json:"omni_participation_dates" dynamodbav:"omni_participation_dates"`
	OmniAbsenceCount       int       `json:"omni_absence_count" dynamodbav:"omni_absence_count"`
	CompensationOwed       bool      `json:"compensation_owed" dynamodbav:"compensation_owed"`
	LastOmniParticipation  *time.Time `json:"last_omni_participation,omitempty" dynamodbav:"last_omni_participation,omitempty"`
	CreatedAt              time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" dynamodbav:"updated_at"`
}

// MemberRole constants
const (
	RoleAdmin   = "admin"
	RoleOfficer = "officer"
	RoleUser    = "user"
)

// IsEligibleForSilver checks if member is eligible for silver links
func (m *Member) IsEligibleForSilver() bool {
	// Must be in guild for at least 30 days and have weekly boss participation
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	return m.GuildJoinDate.Before(thirtyDaysAgo) && m.WeeklyBossParticipation
}

// IsEligibleForGold checks if member is eligible for gold links
func (m *Member) IsEligibleForGold() bool {
	// Must be in guild for at least 90 days and have weekly boss participation
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
	return m.GuildJoinDate.Before(ninetyDaysAgo) && m.WeeklyBossParticipation
}

// IsActive checks if member is considered active (not exceeding absence threshold)
func (m *Member) IsActive(maxAbsences int) bool {
	return m.OmniAbsenceCount < maxAbsences
}

// GetDaysInGuild returns the number of days the member has been in the guild
func (m *Member) GetDaysInGuild() int {
	return int(time.Since(m.GuildJoinDate).Hours() / 24)
}
