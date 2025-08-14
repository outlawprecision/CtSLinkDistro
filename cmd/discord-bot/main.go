package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"

	"flavaflav/internal/database"
	"flavaflav/internal/models"
	"flavaflav/internal/services"
)

type Bot struct {
	session             *discordgo.Session
	memberService       *services.MemberService
	distributionService *services.DistributionService
	config              *models.Config
}

func main() {
	// Load configuration
	config := models.DefaultConfig()

	// Override with environment variables
	if token := os.Getenv("DISCORD_BOT_TOKEN"); token != "" {
		config.Discord.BotToken = token
	}
	if guildID := os.Getenv("DISCORD_GUILD_ID"); guildID != "" {
		config.Discord.GuildID = guildID
	}
	if channelID := os.Getenv("DISCORD_CHANNEL_ID"); channelID != "" {
		config.Discord.ChannelID = channelID
	}
	if region := os.Getenv("AWS_REGION"); region != "" {
		config.AWS.Region = region
	}
	if table := os.Getenv("DYNAMODB_TABLE"); table != "" {
		config.AWS.DynamoDBTable = table
	}

	if config.Discord.BotToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN environment variable is required")
	}

	// Get inventory table name
	inventoryTableName := os.Getenv("DYNAMODB_INVENTORY_TABLE")
	if inventoryTableName == "" {
		inventoryTableName = "flavaflav-inventory-dev" // Default fallback
	}

	// Initialize database
	db, err := database.NewDynamoDBService(config.AWS.Region, config.AWS.DynamoDBTable, inventoryTableName)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize services
	memberService := services.NewMemberService(db, config)
	distributionService := services.NewDistributionService(db, memberService, config)

	// Initialize distribution lists
	ctx := context.Background()
	err = distributionService.InitializeDistributionLists(ctx)
	if err != nil {
		log.Printf("Warning: Failed to initialize distribution lists: %v", err)
	}

	// Create Discord session
	session, err := discordgo.New("Bot " + config.Discord.BotToken)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}

	bot := &Bot{
		session:             session,
		memberService:       memberService,
		distributionService: distributionService,
		config:              config,
	}

	// Register event handlers
	session.AddHandler(bot.ready)
	session.AddHandler(bot.messageCreate)
	session.AddHandler(bot.interactionCreate)

	// Set intents
	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	// Open connection
	err = session.Open()
	if err != nil {
		log.Fatalf("Failed to open Discord session: %v", err)
	}
	defer session.Close()

	// Register slash commands
	err = bot.registerCommands()
	if err != nil {
		log.Printf("Failed to register commands: %v", err)
	}

	log.Println("FlavaFlav Discord bot is running. Press CTRL+C to exit.")

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down Discord bot...")
}

func (b *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("Bot is ready! Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
}

func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Handle legacy text commands (optional)
	if strings.HasPrefix(m.Content, "!flava") {
		s.ChannelMessageSend(m.ChannelID, "Please use slash commands instead! Try `/help` to see available commands.")
	}
}

func (b *Bot) interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name == "" {
		return
	}

	ctx := context.Background()

	switch i.ApplicationCommandData().Name {
	case "add-member":
		b.handleAddMember(ctx, s, i)
	case "check-status":
		b.handleCheckStatus(ctx, s, i)
	case "current-lists":
		b.handleCurrentLists(ctx, s, i)
	case "mark-participation":
		b.handleMarkParticipation(ctx, s, i)
	case "spin-wheel":
		b.handleSpinWheel(ctx, s, i)
	case "help":
		b.handleHelp(ctx, s, i)
	}
}

func (b *Bot) registerCommands() error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "add-member",
			Description: "Add a new guild member",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Discord user to add",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "join-date",
					Description: "Guild join date (YYYY-MM-DD)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "characters",
					Description: "Character names (comma separated)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "role",
					Description: "Member role",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "User", Value: "user"},
						{Name: "Officer", Value: "officer"},
						{Name: "Admin", Value: "admin"},
					},
				},
			},
		},
		{
			Name:        "check-status",
			Description: "Check member eligibility status",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Discord user to check",
					Required:    false,
				},
			},
		},
		{
			Name:        "current-lists",
			Description: "Show current distribution lists status",
		},
		{
			Name:        "mark-participation",
			Description: "Mark weekly boss participation",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Discord user",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "participated",
					Description: "Did they participate?",
					Required:    true,
				},
			},
		},
		{
			Name:        "spin-wheel",
			Description: "Spin the wheel for link distribution (Admin only)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "Link type",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Silver", Value: "silver"},
						{Name: "Gold", Value: "gold"},
					},
				},
			},
		},
		{
			Name:        "help",
			Description: "Show available commands",
		},
	}

	for _, command := range commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, b.config.Discord.GuildID, command)
		if err != nil {
			return fmt.Errorf("failed to create command %s: %v", command.Name, err)
		}
	}

	return nil
}

func (b *Bot) handleAddMember(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	user := options[0].UserValue(s)
	joinDateStr := options[1].StringValue()

	var characterNames []string
	var role string = "user"

	if len(options) > 2 && options[2].Name == "characters" {
		characterNames = strings.Split(options[2].StringValue(), ",")
		for i := range characterNames {
			characterNames[i] = strings.TrimSpace(characterNames[i])
		}
	}

	if len(options) > 3 && options[3].Name == "role" {
		role = options[3].StringValue()
	}

	// Parse join date
	joinDate, err := time.Parse("2006-01-02", joinDateStr)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid date format. Please use YYYY-MM-DD format.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Create member
	member, err := b.memberService.CreateMember(ctx, user.ID, user.Username, characterNames, joinDate, role)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Failed to add member: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "Member Added Successfully",
		Color: 0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Discord User", Value: user.Username, Inline: true},
			{Name: "Join Date", Value: joinDate.Format("2006-01-02"), Inline: true},
			{Name: "Role", Value: member.Role, Inline: true},
		},
	}

	if len(characterNames) > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Characters",
			Value:  strings.Join(characterNames, ", "),
			Inline: false,
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func (b *Bot) handleCheckStatus(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	var userID string

	if len(i.ApplicationCommandData().Options) > 0 {
		user := i.ApplicationCommandData().Options[0].UserValue(s)
		userID = user.ID
	} else {
		userID = i.Member.User.ID
	}

	status, err := b.memberService.GetMemberStatus(ctx, userID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Member not found or error occurred.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Status for %s", status.Member.DiscordUsername),
		Color: 0x0099ff,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Days in Guild", Value: strconv.Itoa(status.DaysInGuild), Inline: true},
			{Name: "Silver Eligible", Value: boolToEmoji(status.SilverEligible), Inline: true},
			{Name: "Gold Eligible", Value: boolToEmoji(status.GoldEligible), Inline: true},
			{Name: "Weekly Boss Participation", Value: boolToEmoji(status.Member.WeeklyBossParticipation), Inline: true},
			{Name: "Omni Absences", Value: strconv.Itoa(status.Member.OmniAbsenceCount), Inline: true},
			{Name: "Active Status", Value: boolToEmoji(status.IsActive), Inline: true},
			{Name: "Total Silver Links", Value: strconv.Itoa(status.TotalSilverLinks), Inline: true},
			{Name: "Total Gold Links", Value: strconv.Itoa(status.TotalGoldLinks), Inline: true},
			{Name: "Compensation Links", Value: strconv.Itoa(status.CompensationLinks), Inline: true},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func (b *Bot) handleCurrentLists(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	status, err := b.distributionService.GetAllListStatuses(ctx)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to get list status.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "Current Distribution Lists",
		Color: 0xffd700,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Silver - Eligible", Value: strconv.Itoa(status.Silver.EligibleCount), Inline: true},
			{Name: "Silver - Completed", Value: strconv.Itoa(status.Silver.CompletedCount), Inline: true},
			{Name: "Silver - Compensation", Value: strconv.Itoa(status.Silver.CompensationCount), Inline: true},
			{Name: "Gold - Eligible", Value: strconv.Itoa(status.Gold.EligibleCount), Inline: true},
			{Name: "Gold - Completed", Value: strconv.Itoa(status.Gold.CompletedCount), Inline: true},
			{Name: "Gold - Compensation", Value: strconv.Itoa(status.Gold.CompensationCount), Inline: true},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func (b *Bot) handleMarkParticipation(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	user := options[0].UserValue(s)
	participated := options[1].BoolValue()

	err := b.memberService.MarkWeeklyBossParticipation(ctx, user.ID, participated)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Failed to update participation: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	participationText := "marked as participated"
	if !participated {
		participationText = "marked as not participated"
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s has been %s in weekly boss runs.", user.Username, participationText),
		},
	})
}

func (b *Bot) handleSpinWheel(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check if user has admin permissions (simplified check)
	if !b.hasAdminPermissions(i.Member) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You don't have permission to use this command.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	linkType := i.ApplicationCommandData().Options[0].StringValue()

	// Update lists first
	err := b.distributionService.UpdateDistributionLists(ctx)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to update distribution lists.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	result, err := b.distributionService.SelectRandomWinner(ctx, linkType)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Failed to select winner: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	compensationText := ""
	if result.IsCompensation {
		compensationText = " (Compensation)"
	}

	embed := &discordgo.MessageEmbed{
		Title: "🎉 Link Winner Selected!",
		Color: 0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Winner", Value: result.Winner.DiscordUsername + compensationText, Inline: true},
			{Name: "Link Type", Value: strings.Title(result.LinkHistory.LinkType), Inline: true},
			{Name: "Date", Value: result.LinkHistory.DateReceived.Format("2006-01-02"), Inline: true},
		},
		Timestamp: result.LinkHistory.DateReceived.Format(time.RFC3339),
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func (b *Bot) handleHelp(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "FlavaFlav Bot Commands",
		Description: "Available commands for the UO Outlands guild link distribution system:",
		Color:       0x0099ff,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "/add-member", Value: "Add a new guild member", Inline: false},
			{Name: "/check-status", Value: "Check member eligibility status", Inline: false},
			{Name: "/current-lists", Value: "Show current distribution lists", Inline: false},
			{Name: "/mark-participation", Value: "Mark weekly boss participation", Inline: false},
			{Name: "/spin-wheel", Value: "Spin the wheel for link distribution (Admin only)", Inline: false},
			{Name: "/help", Value: "Show this help message", Inline: false},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *Bot) hasAdminPermissions(member *discordgo.Member) bool {
	// Simplified permission check - in production, you'd want more sophisticated role checking
	for _, roleID := range member.Roles {
		// You would configure these role IDs based on your Discord server
		if roleID == "ADMIN_ROLE_ID" || roleID == "OFFICER_ROLE_ID" {
			return true
		}
	}

	// Check if user has administrator permissions
	permissions, err := b.session.UserChannelPermissions(member.User.ID, b.config.Discord.ChannelID)
	if err == nil && permissions&discordgo.PermissionAdministrator != 0 {
		return true
	}

	return false
}

func boolToEmoji(b bool) string {
	if b {
		return "✅"
	}
	return "❌"
}
