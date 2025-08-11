package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"flavaflav/internal/models"
)

// MemoryDatabase implements an in-memory database for local development
type MemoryDatabase struct {
	members           map[string]*models.Member
	linkHistory       []*models.LinkHistory
	distributionLists map[string]*models.DistributionList
	mutex             sync.RWMutex
}

// NewMemoryDatabase creates a new in-memory database
func NewMemoryDatabase() *MemoryDatabase {
	db := &MemoryDatabase{
		members:           make(map[string]*models.Member),
		linkHistory:       []*models.LinkHistory{},
		distributionLists: make(map[string]*models.DistributionList),
	}

	// Initialize with some sample data
	db.initializeSampleData()

	return db
}

// Member operations

func (db *MemoryDatabase) CreateMember(ctx context.Context, member *models.Member) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, exists := db.members[member.DiscordID]; exists {
		return fmt.Errorf("member with Discord ID %s already exists", member.DiscordID)
	}

	db.members[member.DiscordID] = member
	return nil
}

func (db *MemoryDatabase) GetMember(ctx context.Context, discordID string) (*models.Member, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	member, exists := db.members[discordID]
	if !exists {
		return nil, fmt.Errorf("member not found")
	}

	return member, nil
}

func (db *MemoryDatabase) UpdateMember(ctx context.Context, member *models.Member) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, exists := db.members[member.DiscordID]; !exists {
		return fmt.Errorf("member not found")
	}

	db.members[member.DiscordID] = member
	return nil
}

func (db *MemoryDatabase) DeleteMember(ctx context.Context, discordID string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, exists := db.members[discordID]; !exists {
		return fmt.Errorf("member not found")
	}

	delete(db.members, discordID)
	return nil
}

func (db *MemoryDatabase) GetAllMembers(ctx context.Context) ([]*models.Member, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	members := make([]*models.Member, 0, len(db.members))
	for _, member := range db.members {
		members = append(members, member)
	}

	return members, nil
}

// Link History operations

func (db *MemoryDatabase) CreateLinkHistory(ctx context.Context, linkHistory *models.LinkHistory) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.linkHistory = append(db.linkHistory, linkHistory)
	return nil
}

func (db *MemoryDatabase) GetLinkHistoryByMember(ctx context.Context, discordID string) ([]*models.LinkHistory, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	var history []*models.LinkHistory
	for _, link := range db.linkHistory {
		if link.DiscordID == discordID {
			history = append(history, link)
		}
	}

	return history, nil
}

// Distribution List operations

func (db *MemoryDatabase) CreateDistributionList(ctx context.Context, distributionList *models.DistributionList) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.distributionLists[distributionList.ListType] = distributionList
	return nil
}

func (db *MemoryDatabase) GetDistributionList(ctx context.Context, listType string) (*models.DistributionList, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	list, exists := db.distributionLists[listType]
	if !exists {
		return nil, fmt.Errorf("distribution list not found")
	}

	return list, nil
}

func (db *MemoryDatabase) UpdateDistributionList(ctx context.Context, distributionList *models.DistributionList) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.distributionLists[distributionList.ListType] = distributionList
	return nil
}

// Initialize sample data for demonstration
func (db *MemoryDatabase) initializeSampleData() {
	now := time.Now()

	// Sample members
	members := []*models.Member{
		{
			DiscordID:               "123456789012345678",
			DiscordUsername:         "GuildMaster",
			CharacterNames:          []string{"Aragorn", "Legolas"},
			GuildJoinDate:           now.AddDate(0, 0, -120), // 120 days ago
			Role:                    models.RoleAdmin,
			WeeklyBossParticipation: true,
			OmniParticipationDates:  []time.Time{now.AddDate(0, 0, -7), now.AddDate(0, 0, -14)},
			OmniAbsenceCount:        0,
			CompensationOwed:        false,
			LastOmniParticipation:   &[]time.Time{now.AddDate(0, 0, -7)}[0],
			CreatedAt:               now.AddDate(0, 0, -120),
			UpdatedAt:               now,
		},
		{
			DiscordID:               "234567890123456789",
			DiscordUsername:         "VeteranPlayer",
			CharacterNames:          []string{"Gimli", "Boromir", "Faramir"},
			GuildJoinDate:           now.AddDate(0, 0, -95), // 95 days ago
			Role:                    models.RoleOfficer,
			WeeklyBossParticipation: true,
			OmniParticipationDates:  []time.Time{now.AddDate(0, 0, -7)},
			OmniAbsenceCount:        0,
			CompensationOwed:        false,
			LastOmniParticipation:   &[]time.Time{now.AddDate(0, 0, -7)}[0],
			CreatedAt:               now.AddDate(0, 0, -95),
			UpdatedAt:               now,
		},
		{
			DiscordID:               "345678901234567890",
			DiscordUsername:         "NewMember",
			CharacterNames:          []string{"Frodo"},
			GuildJoinDate:           now.AddDate(0, 0, -45), // 45 days ago
			Role:                    models.RoleUser,
			WeeklyBossParticipation: true,
			OmniParticipationDates:  []time.Time{now.AddDate(0, 0, -7)},
			OmniAbsenceCount:        0,
			CompensationOwed:        false,
			LastOmniParticipation:   &[]time.Time{now.AddDate(0, 0, -7)}[0],
			CreatedAt:               now.AddDate(0, 0, -45),
			UpdatedAt:               now,
		},
		{
			DiscordID:               "456789012345678901",
			DiscordUsername:         "RecentJoiner",
			CharacterNames:          []string{"Sam", "Merry"},
			GuildJoinDate:           now.AddDate(0, 0, -15), // 15 days ago
			Role:                    models.RoleUser,
			WeeklyBossParticipation: false,
			OmniParticipationDates:  []time.Time{},
			OmniAbsenceCount:        2,
			CompensationOwed:        false,
			CreatedAt:               now.AddDate(0, 0, -15),
			UpdatedAt:               now,
		},
		{
			DiscordID:               "567890123456789012",
			DiscordUsername:         "InactiveMember",
			CharacterNames:          []string{"Pippin"},
			GuildJoinDate:           now.AddDate(0, 0, -60), // 60 days ago
			Role:                    models.RoleUser,
			WeeklyBossParticipation: true,
			OmniParticipationDates:  []time.Time{},
			OmniAbsenceCount:        4, // Exceeds threshold
			CompensationOwed:        true,
			CreatedAt:               now.AddDate(0, 0, -60),
			UpdatedAt:               now,
		},
	}

	for _, member := range members {
		db.members[member.DiscordID] = member
	}

	// Sample link history
	linkHistory := []*models.LinkHistory{
		{
			LinkID:          "link-001",
			DiscordID:       "123456789012345678",
			DiscordUsername: "GuildMaster",
			LinkType:        models.LinkTypeGold,
			DateReceived:    now.AddDate(0, 0, -30),
			OmniEventDate:   now.AddDate(0, 0, -30),
			IsCompensation:  false,
			Notes:           "",
			Timestamp:       now.AddDate(0, 0, -30),
		},
		{
			LinkID:          "link-002",
			DiscordID:       "234567890123456789",
			DiscordUsername: "VeteranPlayer",
			LinkType:        models.LinkTypeSilver,
			DateReceived:    now.AddDate(0, 0, -20),
			OmniEventDate:   now.AddDate(0, 0, -20),
			IsCompensation:  false,
			Notes:           "",
			Timestamp:       now.AddDate(0, 0, -20),
		},
		{
			LinkID:          "link-003",
			DiscordID:       "567890123456789012",
			DiscordUsername: "InactiveMember",
			LinkType:        models.LinkTypeSilver,
			DateReceived:    now.AddDate(0, 0, -10),
			OmniEventDate:   now.AddDate(0, 0, -10),
			IsCompensation:  true,
			Notes:           "Compensation for missed omni events",
			Timestamp:       now.AddDate(0, 0, -10),
		},
	}

	db.linkHistory = linkHistory

	// Initialize distribution lists
	silverList := models.NewDistributionList(models.ListTypeSilver, 3)
	silverList.EligibleMembers = []string{"234567890123456789", "345678901234567890"} // VeteranPlayer, NewMember
	silverList.CompletedMembers = []string{"123456789012345678"}                      // GuildMaster
	silverList.InactiveMembers = []string{"567890123456789012"}                       // InactiveMember
	silverList.CompensationQueue = []string{}

	goldList := models.NewDistributionList(models.ListTypeGold, 3)
	goldList.EligibleMembers = []string{"234567890123456789"}  // VeteranPlayer (95+ days)
	goldList.CompletedMembers = []string{"123456789012345678"} // GuildMaster
	goldList.InactiveMembers = []string{}
	goldList.CompensationQueue = []string{}

	db.distributionLists[models.ListTypeSilver] = silverList
	db.distributionLists[models.ListTypeGold] = goldList
}
