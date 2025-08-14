package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"flavaflav/internal/db"
	"flavaflav/internal/models"

	"github.com/bwmarrin/discordgo"
)

var (
	dbClient *db.DynamoDBClient
	guildID  string
)

func init() {
	// Get configuration from environment variables
	tableName := os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		tableName = "flavaflav-dev"
	}

	guildID = os.Getenv("DISCORD_GUILD_ID")
	if guildID == "" {
		log.Fatal("DISCORD_GUILD_ID environment variable is required")
	}

	// Initialize DynamoDB client
	var err error
	dbClient, err = db.NewDynamoDBClient(tableName)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB client: %v", err)
	}
}

func main() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_BOT_TOKEN environment variable is required")
	}

	// Create Discord session
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Register slash commands
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		registerCommands(s)
	})

	// Register interaction handler
	dg.AddHandler(interactionHandler)

	// Open connection
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}

	// Wait for interrupt signal
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Clean up
	dg.Close()
}

// Discord slash commands
var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "my-status",
		Description: "Check your rank, eligibility, and distribution history",
	},
	{
		Name:        "inventory",
		Description: "View current mastery link inventory",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "quality",
				Description: "Filter by quality (bronze, silver, gold)",
				Required:    false,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Bronze", Value: "bronze"},
					{Name: "Silver", Value: "silver"},
					{Name: "Gold", Value: "gold"},
				},
			},
		},
	},
	{
		Name:        "check-rank",
		Description: "Check a member's rank and eligibility",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "member",
				Description: "The member to check",
				Required:    true,
			},
		},
	},
	{
		Name:        "add-member",
		Description: "Add a new guild member (Maester only)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "member",
				Description: "The Discord user to add",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "join_date",
				Description: "Guild join date (YYYY-MM-DD)",
				Required:    true,
			},
		},
	},
	{
		Name:        "promote-officer",
		Description: "Promote a member to Maester (Maester only)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "member",
				Description: "The member to promote",
				Required:    true,
			},
		},
	},
	{
		Name:        "add-inventory",
		Description: "Add mastery links to inventory (Maester only)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "link_type",
				Description: "Type of mastery link",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "quality",
				Description: "Quality of the link",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Bronze", Value: "bronze"},
					{Name: "Silver", Value: "silver"},
					{Name: "Gold", Value: "gold"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "count",
				Description: "Number of links to add",
				Required:    true,
			},
		},
	},
	{
		Name:        "pick-winner",
		Description: "Pick a random winner from eligible members (Maester only)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "quality",
				Description: "Quality of links to distribute",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Silver", Value: "silver"},
					{Name: "Gold", Value: "gold"},
				},
			},
		},
	},
}

func registerCommands(s *discordgo.Session) {
	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
		if err != nil {
			log.Printf("Cannot create '%v' command: %v", cmd.Name, err)
		}
	}
}

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name == "" {
		return
	}

	ctx := context.Background()

	switch i.ApplicationCommandData().Name {
	case "my-status":
		handleMyStatus(ctx, s, i)
	case "inventory":
		handleInventory(ctx, s, i)
	case "check-rank":
		handleCheckRank(ctx, s, i)
	case "add-member":
		handleAddMember(ctx, s, i)
	case "promote-officer":
		handlePromoteOfficer(ctx, s, i)
	case "add-inventory":
		handleAddInventory(ctx, s, i)
	case "pick-winner":
		handlePickWinner(ctx, s, i)
	}
}

func handleMyStatus(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID

	member, err := dbClient.GetMember(ctx, userID)
	if err != nil {
		respondError(s, i, "You are not registered as a guild member. Contact a Maester to be added.")
		return
	}

	member.UpdateRankAndEligibility()

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Status for %s", member.Username),
		Color: getRankColor(member.Rank),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Rank", Value: member.Rank, Inline: true},
			{Name: "Days in Guild", Value: strconv.Itoa(member.DaysInGuild), Inline: true},
			{Name: "Silver Eligible", Value: boolToEmoji(member.SilverEligible), Inline: true},
			{Name: "Gold Eligible", Value: boolToEmoji(member.GoldEligible), Inline: true},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Get distribution history
	distributions, err := dbClient.GetDistributionsByMember(ctx, userID)
	if err == nil && len(distributions) > 0 {
		historyText := fmt.Sprintf("Total distributions: %d\n", len(distributions))
		if len(distributions) <= 5 {
			for _, dist := range distributions {
				historyText += fmt.Sprintf("• %s (%s)\n", dist.GetDisplayName(), dist.DistributedAt.Format("Jan 2"))
			}
		} else {
			for i := 0; i < 3; i++ {
				dist := distributions[i]
				historyText += fmt.Sprintf("• %s (%s)\n", dist.GetDisplayName(), dist.DistributedAt.Format("Jan 2"))
			}
			historyText += fmt.Sprintf("... and %d more", len(distributions)-3)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Recent Distributions",
			Value: historyText,
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func handleInventory(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	var quality string
	if len(options) > 0 {
		quality = options[0].StringValue()
	}

	var links []*models.InventoryLink
	var err error

	if quality != "" {
		links, err = dbClient.GetAvailableInventoryLinksByQuality(ctx, quality)
	} else {
		links, err = dbClient.GetAvailableInventoryLinks(ctx)
	}

	if err != nil {
		respondError(s, i, "Failed to get inventory")
		return
	}

	// Group by link type and quality
	summary := make(map[string]map[string]int)
	for _, link := range links {
		if summary[link.LinkType] == nil {
			summary[link.LinkType] = make(map[string]int)
		}
		summary[link.LinkType][link.Quality]++
	}

	embed := &discordgo.MessageEmbed{
		Title: "Mastery Link Inventory",
		Color: 0x00ff00,
	}

	if quality != "" {
		embed.Title += fmt.Sprintf(" (%s)", strings.Title(quality))
	}

	if len(summary) == 0 {
		embed.Description = "No links available in inventory"
	} else {
		var fields []*discordgo.MessageEmbedField
		for linkType, qualities := range summary {
			var value string
			for qual, count := range qualities {
				emoji := getQualityEmoji(qual)
				value += fmt.Sprintf("%s %s: %d\n", emoji, strings.Title(qual), count)
			}
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   linkType,
				Value:  value,
				Inline: true,
			})
		}
		embed.Fields = fields
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func handleCheckRank(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	targetUser := i.ApplicationCommandData().Options[0].UserValue(s)

	member, err := dbClient.GetMember(ctx, targetUser.ID)
	if err != nil {
		respondError(s, i, fmt.Sprintf("%s is not registered as a guild member.", targetUser.Username))
		return
	}

	member.UpdateRankAndEligibility()

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Status for %s", member.Username),
		Color: getRankColor(member.Rank),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Rank", Value: member.Rank, Inline: true},
			{Name: "Days in Guild", Value: strconv.Itoa(member.DaysInGuild), Inline: true},
			{Name: "Silver Eligible", Value: boolToEmoji(member.SilverEligible), Inline: true},
			{Name: "Gold Eligible", Value: boolToEmoji(member.GoldEligible), Inline: true},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func handleAddMember(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isMaester(s, i.Member) {
		respondError(s, i, "Only Maesters can add members.")
		return
	}

	options := i.ApplicationCommandData().Options
	targetUser := options[0].UserValue(s)
	joinDateStr := options[1].StringValue()

	joinDate, err := time.Parse("2006-01-02", joinDateStr)
	if err != nil {
		respondError(s, i, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	member := models.NewMember(targetUser.ID, targetUser.Username, joinDate, i.Member.User.ID)

	err = dbClient.CreateMember(ctx, member)
	if err != nil {
		respondError(s, i, "Failed to add member")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "Member Added",
		Color: 0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Member", Value: targetUser.Username, Inline: true},
			{Name: "Rank", Value: member.Rank, Inline: true},
			{Name: "Join Date", Value: joinDate.Format("Jan 2, 2006"), Inline: true},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func handlePromoteOfficer(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isMaester(s, i.Member) {
		respondError(s, i, "Only Maesters can promote members.")
		return
	}

	targetUser := i.ApplicationCommandData().Options[0].UserValue(s)

	member, err := dbClient.GetMember(ctx, targetUser.ID)
	if err != nil {
		respondError(s, i, "Member not found")
		return
	}

	member.PromoteToOfficer()

	err = dbClient.UpdateMember(ctx, member)
	if err != nil {
		respondError(s, i, "Failed to promote member")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Member Promoted",
		Color:       0x800080,
		Description: fmt.Sprintf("%s has been promoted to %s!", member.Username, member.Rank),
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func handleAddInventory(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isMaester(s, i.Member) {
		respondError(s, i, "Only Maesters can add inventory.")
		return
	}

	options := i.ApplicationCommandData().Options
	linkType := options[0].StringValue()
	quality := options[1].StringValue()
	count := int(options[2].IntValue())

	if count <= 0 {
		respondError(s, i, "Count must be greater than 0")
		return
	}

	bonus := models.GetLinkBonus(linkType, quality)
	category := models.GetLinkCategory(linkType)

	var createdCount int
	for j := 0; j < count; j++ {
		link := models.NewInventoryLink(linkType, quality, category, bonus, i.Member.User.ID)
		err := dbClient.CreateInventoryLink(ctx, link)
		if err != nil {
			break
		}
		createdCount++
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Inventory Updated",
		Color:       0x00ff00,
		Description: fmt.Sprintf("Added %d %s %s links to inventory", createdCount, quality, linkType),
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func handlePickWinner(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isMaester(s, i.Member) {
		respondError(s, i, "Only Maesters can pick winners.")
		return
	}

	quality := i.ApplicationCommandData().Options[0].StringValue()

	// Get eligible members
	members, err := dbClient.GetAllMembers(ctx)
	if err != nil {
		respondError(s, i, "Failed to get members")
		return
	}

	var eligibleMembers []*models.Member
	for _, member := range members {
		member.UpdateRankAndEligibility()
		if (quality == "silver" && member.SilverEligible) || (quality == "gold" && member.GoldEligible) {
			eligibleMembers = append(eligibleMembers, member)
		}
	}

	if len(eligibleMembers) == 0 {
		respondError(s, i, fmt.Sprintf("No members eligible for %s links", quality))
		return
	}

	// Pick random winner
	rand.Seed(time.Now().UnixNano())
	winner := eligibleMembers[rand.Intn(len(eligibleMembers))]

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎉 %s Link Winner!", strings.Title(quality)),
		Color:       getQualityColor(quality),
		Description: fmt.Sprintf("**%s** has been selected for a %s link!", winner.Username, quality),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Winner Rank", Value: winner.Rank, Inline: true},
			{Name: "Days in Guild", Value: strconv.Itoa(winner.DaysInGuild), Inline: true},
			{Name: "Total Eligible", Value: strconv.Itoa(len(eligibleMembers)), Inline: true},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// Helper functions

func isMaester(s *discordgo.Session, member *discordgo.Member) bool {
	// Check if user has a role named "Maester" or is admin
	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			continue
		}
		if role.Name == "Maester" || role.Permissions&discordgo.PermissionAdministrator != 0 {
			return true
		}
	}
	return false
}

func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "❌ " + message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func boolToEmoji(b bool) string {
	if b {
		return "✅"
	}
	return "❌"
}

func getRankColor(rank string) int {
	switch rank {
	case models.RankBookWorm:
		return 0x8B4513 // Brown
	case models.RankScholar:
		return 0xC0C0C0 // Silver
	case models.RankSage:
		return 0xFFD700 // Gold
	case models.RankMaester:
		return 0x800080 // Purple
	default:
		return 0x666666 // Gray
	}
}

func getQualityEmoji(quality string) string {
	switch quality {
	case "bronze":
		return "🥉"
	case "silver":
		return "🥈"
	case "gold":
		return "🥇"
	default:
		return "⚪"
	}
}

func getQualityColor(quality string) int {
	switch quality {
	case "bronze":
		return 0xCD7F32
	case "silver":
		return 0xC0C0C0
	case "gold":
		return 0xFFD700
	default:
		return 0x666666
	}
}
