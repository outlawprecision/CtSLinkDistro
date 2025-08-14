package models

import (
	"fmt"
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
	IsAvailable bool      `json:"is_available" dynamodbav:"is_available"` // true if not distributed yet
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
		IsAvailable: true,
		AddedBy:     addedBy,
		AddedDate:   time.Now(),
		Notes:       "",
	}
}

// MarkDistributed marks this link as distributed (no longer available)
func (l *InventoryLink) MarkDistributed() {
	l.IsAvailable = false
}

// MarkAvailable marks this link as available again
func (l *InventoryLink) MarkAvailable() {
	l.IsAvailable = true
}

// GetDisplayName returns a formatted display name for the link
func (l *InventoryLink) GetDisplayName() string {
	return fmt.Sprintf("%s %s (%s)", l.Quality, l.LinkType, l.Bonus)
}

// generateLinkID creates a unique ID for a link
func generateLinkID(linkType, quality string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%s_%d", quality, linkType, timestamp)
}

// Common mastery link types with their standard bonuses
var MasteryLinkBonuses = map[string]map[string]string{
	"Melee Damage": {
		QualityBronze: "3.00%",
		QualitySilver: "3.75%",
		QualityGold:   "4.50%",
	},
	"Melee Accuracy": {
		QualityBronze: "1.75%",
		QualitySilver: "2.19%",
		QualityGold:   "2.62%",
	},
	"Spell Damage": {
		QualityBronze: "3.00%",
		QualitySilver: "3.75%",
		QualityGold:   "4.50%",
	},
	"Damage to Undead Creatures": {
		QualityBronze: "2.50%",
		QualitySilver: "3.13%",
		QualityGold:   "3.75%",
	},
	"Inferno Damage": {
		QualityBronze: "2.00%",
		QualitySilver: "2.50%",
		QualityGold:   "3.00%",
	},
	"Follower Damage": {
		QualityBronze: "2.00%",
		QualitySilver: "2.50%",
		QualityGold:   "3.00%",
	},
	"Gold/Doubloon Drop Increase": {
		QualityBronze: "3.00%",
		QualitySilver: "3.75%",
		QualityGold:   "4.50%",
	},
}

// GetLinkBonus returns the standard bonus for a link type and quality
func GetLinkBonus(linkType, quality string) string {
	if linkBonuses, exists := MasteryLinkBonuses[linkType]; exists {
		if bonus, qualityExists := linkBonuses[quality]; qualityExists {
			return bonus
		}
	}
	return "TBD" // To be determined for custom links
}

// GetLinkCategory returns the category for a link type
func GetLinkCategory(linkType string) string {
	categoryMap := map[string]string{
		"Melee Damage":                "Melee Type Links",
		"Melee Accuracy":              "Melee Type Links",
		"Spell Damage":                "Spell Type Links",
		"Damage to Undead Creatures":  "Monster Slayer Links",
		"Inferno Damage":              "Dungeon Slayer Links",
		"Follower Damage":             "Follower Type Links",
		"Gold/Doubloon Drop Increase": "Other Links",
	}

	if category, exists := categoryMap[linkType]; exists {
		return category
	}
	return "Other Links"
}
