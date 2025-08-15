package models

import (
	"fmt"
	"strings"
	"time"
)

// Quality constants for mastery links
const (
	QualityBronze = "bronze"
	QualitySilver = "silver"
	QualityGold   = "gold"
)

// InventoryLink represents a single mastery link in inventory
type InventoryLink struct {
	LinkID      string    `json:"link_id" dynamodbav:"link_id"`           // unique ID for this specific link
	LinkType    string    `json:"link_type" dynamodbav:"link_type"`       // e.g., "Melee Damage"
	Quality     string    `json:"quality" dynamodbav:"quality"`           // bronze, silver, gold
	Category    string    `json:"category" dynamodbav:"category"`         // e.g., "Melee Type Links"
	Bonus       string    `json:"bonus" dynamodbav:"bonus"`               // e.g., "3.75%"
	IsAvailable string    `json:"is_available" dynamodbav:"is_available"` // "true" if not distributed yet, "false" otherwise
	AddedBy     string    `json:"added_by" dynamodbav:"added_by"`
	AddedDate   time.Time `json:"added_date" dynamodbav:"added_date"`
	Notes       string    `json:"notes" dynamodbav:"notes"` // optional notes about this specific link
}

// NewInventoryLink creates a new individual mastery link
func NewInventoryLink(linkType, quality, category, bonus, addedBy string) *InventoryLink {
	return &InventoryLink{
		LinkID:      generateLinkID(linkType, quality),
		LinkType:    linkType,
		Quality:     quality,
		Category:    category,
		Bonus:       bonus,
		IsAvailable: "true",
		AddedBy:     addedBy,
		AddedDate:   time.Now(),
		Notes:       "",
	}
}

// MarkDistributed marks this link as distributed (no longer available)
func (l *InventoryLink) MarkDistributed() {
	l.IsAvailable = "false"
}

// MarkAvailable marks this link as available again
func (l *InventoryLink) MarkAvailable() {
	l.IsAvailable = "true"
}

// GetDisplayName returns a formatted display name for the link
func (l *InventoryLink) GetDisplayName() string {
	return fmt.Sprintf("%s %s (%s)", l.Quality, l.LinkType, l.Bonus)
}

// generateLinkID creates a unique ID for a link (sanitized for DynamoDB)
func generateLinkID(linkType, quality string) string {
	// Sanitize linkType by replacing spaces and special characters with underscores
	sanitizedLinkType := strings.ReplaceAll(linkType, " ", "_")
	sanitizedLinkType = strings.ReplaceAll(sanitizedLinkType, "/", "_")
	sanitizedLinkType = strings.ReplaceAll(sanitizedLinkType, "-", "_")
	sanitizedLinkType = strings.ReplaceAll(sanitizedLinkType, "%", "pct")

	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%s_%d", quality, sanitizedLinkType, timestamp)
}

// GetLinkBonus returns the standard bonus for a link type and quality
func GetLinkBonus(linkType, quality string) string {
	// Use the comprehensive link type data from link_types.go
	return GetLinkTypeBonus(linkType, quality)
}

// GetLinkCategory returns the category for a link type
func GetLinkCategory(linkType string) string {
	// Simplified category mapping - can be expanded if needed
	// For now, just return a generic category since categories aren't important
	return "Mastery Links"
}
