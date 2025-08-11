package models

import (
	"time"
)

// LinkHistory represents a record of link distribution
type LinkHistory struct {
	LinkID         string    `json:"link_id" dynamodbav:"link_id"`
	DiscordID      string    `json:"discord_id" dynamodbav:"discord_id"`
	DiscordUsername string   `json:"discord_username" dynamodbav:"discord_username"`
	LinkType       string    `json:"link_type" dynamodbav:"link_type"` // silver, gold
	DateReceived   time.Time `json:"date_received" dynamodbav:"date_received"`
	OmniEventDate  time.Time `json:"omni_event_date" dynamodbav:"omni_event_date"`
	IsCompensation bool      `json:"is_compensation" dynamodbav:"is_compensation"`
	Notes          string    `json:"notes,omitempty" dynamodbav:"notes,omitempty"`
	Timestamp      time.Time `json:"timestamp" dynamodbav:"timestamp"`
}

// LinkType constants
const (
	LinkTypeSilver = "silver"
	LinkTypeGold   = "gold"
)

// NewLinkHistory creates a new link history record
func NewLinkHistory(discordID, discordUsername, linkType string, omniEventDate time.Time, isCompensation bool, notes string) *LinkHistory {
	now := time.Now()
	return &LinkHistory{
		LinkID:          generateLinkID(),
		DiscordID:       discordID,
		DiscordUsername: discordUsername,
		LinkType:        linkType,
		DateReceived:    now,
		OmniEventDate:   omniEventDate,
		IsCompensation:  isCompensation,
		Notes:           notes,
		Timestamp:       now,
	}
}

// generateLinkID generates a unique ID for the link record
func generateLinkID() string {
	// This would typically use UUID or similar
	return time.Now().Format("20060102150405") + "-" + time.Now().Format("000")
}
