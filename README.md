# FlavaFlav - Simplified Guild Link Distribution System

A clean, focused web application and Discord bot for managing UO Outlands guild mastery link distribution. Built with Go, AWS Lambda, DynamoDB, and modern web technologies.

## 🎯 Core Features

### Guild Rank System
- **Book Worm** - New members (<30 days) - No link eligibility
- **Scholar** - Veterans (30+ days) - Silver link eligible  
- **Sage** - Elders (90+ days) - Gold link eligible
- **Maester** - Officers - Admin access + all links

### Web Application
- **Dashboard** - Real-time stats and inventory overview
- **Member Management** - View members with automatic rank calculation
- **Inventory Tracking** - Individual mastery link management
- **Picker Wheel** - Visual spinning wheel for fair winner selection
- **Distribution History** - Complete audit trail

### Discord Bot
- **Member Commands** - Check status, view inventory, see history
- **Officer Commands** - Add members, manage inventory, pick winners
- **Automatic Permissions** - Role-based access control

## 🏗️ Architecture

### Technology Stack
- **Backend**: Go with clean, simple architecture
- **Database**: AWS DynamoDB (single table design)
- **API**: AWS Lambda with API Gateway
- **Frontend**: Vanilla HTML/CSS/JavaScript
- **Discord**: DiscordGo with slash commands
- **Deployment**: CloudFormation infrastructure as code

### Simplified Structure
```
FlavaFlav/
├── cmd/
│   ├── lambda/main.go      # Web API Lambda function
│   └── discord-bot/main.go # Discord bot application
├── internal/
│   ├── models/             # Simple data models
│   ├── handlers/           # API handlers
│   └── db/                 # DynamoDB operations
├── web/static/             # Frontend files
├── cloudformation/         # AWS infrastructure
└── scripts/               # Deployment scripts
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- AWS Account with CLI configured
- Discord Bot Token (optional)

### 1. Clone and Build
```bash
git clone https://github.com/outlawprecision/CtSLinkDistro.git
cd CtSLinkDistro
go mod tidy
```

### 2. Test Locally
```bash
# Build Lambda function
go build ./cmd/lambda

# Build Discord bot
go build ./cmd/discord-bot
```

### 3. Deploy to AWS
```bash
# Deploy infrastructure
make deploy-dev

# Upload static files
make upload-static BUCKET=your-s3-bucket-name
```

### 4. Configure Environment
Set environment variables:
```bash
# Required
DYNAMODB_TABLE=flavaflav-dev
AWS_REGION=us-east-1

# Discord Bot (optional)
DISCORD_BOT_TOKEN=your_bot_token
DISCORD_GUILD_ID=your_guild_id
```

## 📱 Discord Commands

### Everyone Can Use
- `/my-status` - Check your rank, eligibility, and history
- `/inventory [quality]` - View current mastery link inventory
- `/check-rank @member` - Check any member's rank and eligibility

### Maesters Only
- `/add-member @user YYYY-MM-DD` - Add new guild member
- `/promote-officer @member` - Promote member to Maester
- `/add-inventory "Link Type" quality count` - Add mastery links
- `/pick-winner quality` - Random winner selection with announcement

## 🎮 Web Interface

### Member Dashboard
- View your current rank and eligibility status
- Check days in guild and distribution history
- Browse current inventory

### Maester Admin Panel
- Manage all guild members
- Add and track mastery link inventory
- Use visual picker wheel for distributions
- View complete distribution history
- Bulk operations and reporting

## 🔧 API Endpoints

### Members
- `GET /api/members` - List all members
- `GET /api/member?discord_id=<id>` - Get specific member
- `POST /api/member/create` - Add new member (Maester only)
- `POST /api/member/promote?discord_id=<id>` - Promote to officer

### Inventory
- `GET /api/inventory` - List available links
- `GET /api/inventory/summary` - Inventory counts by type/quality
- `POST /api/inventory/add` - Add new links (Maester only)

### Distribution
- `GET /api/distribution/eligible?quality=<silver|gold>` - Get eligible members
- `POST /api/distribution/pick-winner` - Random winner selection
- `GET /api/distribution/history` - Distribution history

## 🎯 Business Rules

### Automatic Rank Calculation
- Ranks update automatically based on guild join date
- Officers manually promoted to Maester rank
- Eligibility calculated in real-time

### Link Distribution
- Silver links: 30+ days in guild
- Gold links: 90+ days in guild
- Random selection ensures fairness
- Complete audit trail maintained

### Security
- Role-based permissions (view vs admin)
- Discord role integration for bot commands
- All admin actions require Maester privileges

## 🛠️ Development

### Build Commands
```bash
# Build Lambda function
go build -o bootstrap cmd/lambda/main.go

# Build Discord bot
go build -o discord-bot cmd/discord-bot/main.go

# Test compilation
go build ./...
```

### Local Development
```bash
# Run with local DynamoDB
export DYNAMODB_TABLE=flavaflav-local
go run cmd/lambda/main.go

# Run Discord bot
export DISCORD_BOT_TOKEN=your_token
export DISCORD_GUILD_ID=your_guild
go run cmd/discord-bot/main.go
```

## 📊 Data Models

### Member
```go
type Member struct {
    DiscordID      string    // Discord user ID
    Username       string    // Display name
    JoinDate       time.Time // Guild join date
    Rank           string    // Auto-calculated rank
    IsOfficer      bool      // Maester status
    SilverEligible bool      // Auto-calculated
    GoldEligible   bool      // Auto-calculated
    DaysInGuild    int       // Auto-calculated
}
```

### Inventory Link
```go
type InventoryLink struct {
    LinkID      string    // Unique identifier
    LinkType    string    // e.g., "Melee Damage"
    Quality     string    // bronze, silver, gold
    Category    string    // Link category
    Bonus       string    // e.g., "3.75%"
    IsAvailable bool      // Distribution status
}
```

### Distribution
```go
type Distribution struct {
    DistributionID string    // Unique identifier
    MemberID       string    // Who received it
    LinkID         string    // Which link
    LinkType       string    // Link details
    Quality        string    // Link quality
    Method         string    // "web" or "discord"
    DistributedAt  time.Time // When
}
```

## 🚀 Deployment

### AWS Infrastructure
- **Lambda Function** - Serverless API
- **DynamoDB Table** - Single table for all data
- **API Gateway** - HTTP endpoints
- **S3 + CloudFront** - Static file hosting
- **IAM Roles** - Least-privilege access

### Make Targets
```bash
make build          # Build all components
make deploy-dev     # Deploy to development
make deploy-prod    # Deploy to production
make clean          # Clean build artifacts
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details.

## 🎉 What's New

This is a **complete rewrite** focusing on:
- ✅ Simplified architecture (removed over-engineering)
- ✅ Clean, maintainable code
- ✅ Essential features only
- ✅ Better user experience
- ✅ Reliable deployment process
- ✅ Clear documentation

The old complex system has been archived and this new version provides all the core functionality you need without the complexity that caused problems.
