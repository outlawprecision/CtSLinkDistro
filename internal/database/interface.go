package database

import (
	"context"
	"flavaflav/internal/models"
)

// DatabaseService defines the interface for database operations
type DatabaseService interface {
	// Member operations
	CreateMember(ctx context.Context, member *models.Member) error
	GetMember(ctx context.Context, discordID string) (*models.Member, error)
	UpdateMember(ctx context.Context, member *models.Member) error
	DeleteMember(ctx context.Context, discordID string) error
	GetAllMembers(ctx context.Context) ([]*models.Member, error)

	// Link History operations
	CreateLinkHistory(ctx context.Context, linkHistory *models.LinkHistory) error
	GetLinkHistoryByMember(ctx context.Context, discordID string) ([]*models.LinkHistory, error)

	// Distribution List operations
	CreateDistributionList(ctx context.Context, distributionList *models.DistributionList) error
	GetDistributionList(ctx context.Context, listType string) (*models.DistributionList, error)
	UpdateDistributionList(ctx context.Context, distributionList *models.DistributionList) error
}
