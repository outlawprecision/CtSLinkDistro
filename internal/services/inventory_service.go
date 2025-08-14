package services

import (
	"context"
	"fmt"
	"time"

	"flavaflav/internal/database"
	"flavaflav/internal/models"
)

// InventoryService handles link inventory management
type InventoryService struct {
	db     *database.DynamoDBService
	config *models.Config
}

// NewInventoryService creates a new inventory service
func NewInventoryService(db *database.DynamoDBService, config *models.Config) *InventoryService {
	return &InventoryService{
		db:     db,
		config: config,
	}
}

// InitializeInventory creates all link types in the database if they don't exist
func (s *InventoryService) InitializeInventory(ctx context.Context) error {
	categories := models.GetLinkCategories()

	for _, category := range categories {
		for _, linkType := range category.LinkTypes {
			// Check if item already exists
			_, err := s.db.GetLinkInventoryItem(ctx, linkType)
			if err != nil {
				// Item doesn't exist, create it
				item := models.NewLinkInventoryItem(linkType, category.CategoryName)
				err = s.db.CreateLinkInventoryItem(ctx, item)
				if err != nil {
					return fmt.Errorf("failed to create inventory item for %s: %v", linkType, err)
				}
			}
		}
	}

	return nil
}

// GetAllInventoryItems retrieves all link inventory items
func (s *InventoryService) GetAllInventoryItems(ctx context.Context) ([]*models.LinkInventoryItem, error) {
	return s.db.GetAllLinkInventoryItems(ctx)
}

// GetInventoryItemsByCategory retrieves inventory items by category
func (s *InventoryService) GetInventoryItemsByCategory(ctx context.Context, category string) ([]*models.LinkInventoryItem, error) {
	return s.db.GetLinkInventoryItemsByCategory(ctx, category)
}

// GetInventoryItem retrieves a specific inventory item
func (s *InventoryService) GetInventoryItem(ctx context.Context, linkType string) (*models.LinkInventoryItem, error) {
	return s.db.GetLinkInventoryItem(ctx, linkType)
}

// UpdateInventoryCount updates the count for a specific link type and quality
func (s *InventoryService) UpdateInventoryCount(ctx context.Context, linkType, quality string, newCount int, reason, updatedBy string) error {
	// Get the current item
	item, err := s.db.GetLinkInventoryItem(ctx, linkType)
	if err != nil {
		return fmt.Errorf("failed to get inventory item: %v", err)
	}

	// Update the count and create transaction
	transaction := item.UpdateCount(quality, newCount, reason, updatedBy)
	if transaction == nil {
		return fmt.Errorf("invalid quality specified: %s", quality)
	}

	// Save the updated item
	err = s.db.UpdateLinkInventoryItem(ctx, item)
	if err != nil {
		return fmt.Errorf("failed to update inventory item: %v", err)
	}

	// Save the transaction
	err = s.db.CreateLinkInventoryTransaction(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %v", err)
	}

	return nil
}

// AdjustInventoryCount adjusts the count by a delta amount (positive or negative)
func (s *InventoryService) AdjustInventoryCount(ctx context.Context, linkType, quality string, delta int, reason, updatedBy string) error {
	// Get the current item
	item, err := s.db.GetLinkInventoryItem(ctx, linkType)
	if err != nil {
		return fmt.Errorf("failed to get inventory item: %v", err)
	}

	var currentCount int
	switch quality {
	case "bronze":
		currentCount = item.BronzeCount
	case "silver":
		currentCount = item.SilverCount
	case "gold":
		currentCount = item.GoldCount
	default:
		return fmt.Errorf("invalid quality specified: %s", quality)
	}

	newCount := currentCount + delta
	if newCount < 0 {
		newCount = 0 // Don't allow negative counts
	}

	return s.UpdateInventoryCount(ctx, linkType, quality, newCount, reason, updatedBy)
}

// GetInventoryTransactions retrieves transaction history for a link type
func (s *InventoryService) GetInventoryTransactions(ctx context.Context, linkType string) ([]*models.LinkInventoryTransaction, error) {
	return s.db.GetLinkInventoryTransactions(ctx, linkType)
}

// GetInventorySummary returns a summary of the entire inventory
func (s *InventoryService) GetInventorySummary(ctx context.Context) (*InventorySummary, error) {
	items, err := s.db.GetAllLinkInventoryItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory items: %v", err)
	}

	summary := &InventorySummary{
		Categories: make(map[string]*CategorySummary),
		TotalItems: len(items),
	}

	for _, item := range items {
		// Update category summary
		if summary.Categories[item.Category] == nil {
			summary.Categories[item.Category] = &CategorySummary{
				CategoryName: item.Category,
				LinkTypes:    0,
				TotalBronze:  0,
				TotalSilver:  0,
				TotalGold:    0,
			}
		}

		catSummary := summary.Categories[item.Category]
		catSummary.LinkTypes++
		catSummary.TotalBronze += item.BronzeCount
		catSummary.TotalSilver += item.SilverCount
		catSummary.TotalGold += item.GoldCount

		// Update overall totals
		summary.TotalBronze += item.BronzeCount
		summary.TotalSilver += item.SilverCount
		summary.TotalGold += item.GoldCount
	}

	summary.TotalLinks = summary.TotalBronze + summary.TotalSilver + summary.TotalGold
	summary.LastUpdated = time.Now()

	return summary, nil
}

// GetLowStockItems returns items with counts below specified thresholds
func (s *InventoryService) GetLowStockItems(ctx context.Context, bronzeThreshold, silverThreshold, goldThreshold int) ([]*models.LinkInventoryItem, error) {
	items, err := s.db.GetAllLinkInventoryItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory items: %v", err)
	}

	var lowStockItems []*models.LinkInventoryItem
	for _, item := range items {
		if item.BronzeCount <= bronzeThreshold ||
			item.SilverCount <= silverThreshold ||
			item.GoldCount <= goldThreshold {
			lowStockItems = append(lowStockItems, item)
		}
	}

	return lowStockItems, nil
}

// BulkUpdateInventory updates multiple inventory items at once
func (s *InventoryService) BulkUpdateInventory(ctx context.Context, updates []InventoryUpdate, reason, updatedBy string) error {
	for _, update := range updates {
		err := s.UpdateInventoryCount(ctx, update.LinkType, update.Quality, update.NewCount, reason, updatedBy)
		if err != nil {
			return fmt.Errorf("failed to update %s %s: %v", update.Quality, update.LinkType, err)
		}
	}

	return nil
}

// InventoryUpdate represents a single inventory update operation
type InventoryUpdate struct {
	LinkType string `json:"link_type"`
	Quality  string `json:"quality"`
	NewCount int    `json:"new_count"`
}

// InventorySummary represents a summary of the entire inventory
type InventorySummary struct {
	Categories  map[string]*CategorySummary `json:"categories"`
	TotalItems  int                         `json:"total_items"`
	TotalLinks  int                         `json:"total_links"`
	TotalBronze int                         `json:"total_bronze"`
	TotalSilver int                         `json:"total_silver"`
	TotalGold   int                         `json:"total_gold"`
	LastUpdated time.Time                   `json:"last_updated"`
}

// CategorySummary represents a summary for a specific category
type CategorySummary struct {
	CategoryName string `json:"category_name"`
	LinkTypes    int    `json:"link_types"`
	TotalBronze  int    `json:"total_bronze"`
	TotalSilver  int    `json:"total_silver"`
	TotalGold    int    `json:"total_gold"`
}

// GetCategoryList returns a list of all categories with their totals
func (summary *InventorySummary) GetCategoryList() []*CategorySummary {
	var categories []*CategorySummary
	for _, category := range summary.Categories {
		categories = append(categories, category)
	}
	return categories
}
