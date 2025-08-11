# CtSLinkDistro - UO Outlands Guild Link Distribution System

A comprehensive web application and Discord bot for managing link distribution in UO Outlands guilds. Built with Go, AWS services, and modern web technologies.

## Features

### Core Functionality
- **Member Management**: Track guild members with Discord integration
- **Eligibility System**: Automatic eligibility calculation based on guild tenure and participation
- **Distribution Lists**: Separate silver and gold link distribution queues
- **Random Selection**: Visual picker wheel for fair link distribution
- **Absence Tracking**: Automatic inactive member management with compensation system
- **History Tracking**: Complete audit trail of all link distributions

### Web Application
- **Dashboard**: Real-time status of distribution lists
- **Picker Wheel**: Interactive spinning wheel for winner selection
- **Member Directory**: Searchable member list with detailed status
- **Admin Panel**: Complete member and list management

### Discord Bot
- **Slash Commands**: Modern Discord integration with interactive commands
- **Member Management**: Add/remove members directly from Discord
- **Status Checking**: Quick eligibility and participation status
- **Automated Notifications**: Winner announcements and system updates

## Architecture

### Technology Stack
- **Backend**: Go with clean architecture patterns
- **Database**: AWS DynamoDB for scalable NoSQL storage
- **Frontend**: Vanilla JavaScript with modern CSS
- **Discord**: DiscordGo library for bot functionality
- **Cloud**: AWS Lambda serverless deployment

### Project Structure
```
CtSLinkDistro/
├── cmd/
│   ├── web/           # Web application entry point
│   ├── discord-bot/   # Discord bot entry point
│   └── lambda/        # AWS Lambda functions
├── internal/
│   ├── models/        # Data models and business logic
│   ├── services/      # Business logic layer
│   ├── handlers/      # HTTP handlers
│   └── database/      # DynamoDB operations
├── web/
│   └── static/        # Frontend assets (HTML, CSS, JS)
├── cloudformation/    # AWS CloudFormation templates
├── scripts/          # Deployment and utility scripts
├── docs/             # Documentation
└── configs/          # Configuration files
```

## Prerequisites

- Go 1.21 or higher
- AWS Account with appropriate permissions
- AWS CLI configured
- Discord Bot Token (for Discord integration)

## Quick Start

### 1. Clone and Setup
```bash
git clone https://github.com/outlawprecision/CtSLinkDistro.git
cd CtSLinkDistro
go mod tidy
```

### 2. Deploy to AWS
```bash
# Deploy to development environment
make deploy-dev

# Upload static files
make upload-static BUCKET=your-s3-bucket-name
```

### 3. Configure Environment
Copy `.env.example` to `.env` and configure:
```bash
# AWS Configuration
AWS_REGION=us-east-1
DYNAMODB_TABLE=ctslinkdistro-dev
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key

# Discord Configuration (optional)
DISCORD_BOT_TOKEN=your_bot_token
DISCORD_GUILD_ID=your_guild_id
DISCORD_CHANNEL_ID=your_channel_id
```

## Usage

### Web Application
1. Access your deployed CloudFront URL
2. Use the Dashboard to view current distribution status
3. Add members through the Members tab
4. Use the Picker Wheel to select winners
5. Track history and manage lists

### Discord Bot Commands
- `/add-member` - Add a new guild member
- `/check-status` - Check member eligibility status
- `/current-lists` - Show distribution list status
- `/mark-participation` - Mark weekly boss participation
- `/spin-wheel` - Select random winner (Admin only)
- `/help` - Show available commands

## Business Rules

### Eligibility Requirements
- **Silver Links**: 30+ days in guild + weekly boss participation
- **Gold Links**: 90+ days in guild + weekly boss participation

### Distribution System
- Members are added to eligible lists based on criteria
- Random selection ensures fairness
- Completed members wait for list reset
- Inactive members (3+ consecutive absences) are moved to compensation queue

### List Management
- Lists automatically reset when all eligible members have received links
- Force completion available for lists with inactive members
- Compensation system ensures returning members receive priority

## API Endpoints

### Member Management
- `GET /api/members` - Get all members
- `GET /api/member?discord_id=<id>` - Get specific member
- `POST /api/member/create` - Create new member
- `GET /api/member/status?discord_id=<id>` - Get member status

### Distribution
- `GET /api/distribution/status` - Get distribution list status
- `POST /api/distribution/spin?type=<silver|gold>` - Select random winner
- `POST /api/distribution/force-complete?type=<silver|gold>` - Force complete list
- `GET /api/distribution/eligible?type=<silver|gold>` - Get eligible members

### Utilities
- `POST /api/utility/reset-weekly` - Reset weekly participation
- `POST /api/utility/update-lists` - Update distribution lists
- `GET /api/health` - Health check

## AWS Deployment

### CloudFormation Deployment (Recommended)
The application includes complete infrastructure as code:

1. **Deploy to Development**
   ```bash
   make deploy-dev
   ```

2. **Deploy to Production**
   ```bash
   # Configure parameters first
   cp cloudformation/parameters-dev.json cloudformation/parameters-prod.json
   # Edit parameters-prod.json with your domain and certificate
   make deploy-prod
   ```

3. **Upload Static Files**
   ```bash
   make upload-static BUCKET=your-s3-bucket-name
   ```

**What gets deployed:**
- AWS Lambda function for the API
- DynamoDB table with proper indexes
- API Gateway with custom domain support
- S3 bucket for static files
- CloudFront CDN distribution
- CloudWatch monitoring and alarms
- IAM roles with least-privilege access

### Manual Lambda Deployment
For custom deployments:

1. **Build for Lambda**
   ```bash
   make lambda-build
   ```

2. **Deploy to AWS Lambda**
   - Upload the `bootstrap` binary to AWS Lambda
   - Configure environment variables
   - Set up API Gateway for HTTP endpoints

For detailed deployment instructions, see [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md).

## Available Make Targets

```bash
# Development
make build          # Build both applications
make run-web        # Run web application locally (requires AWS)
make run-bot        # Run Discord bot

# AWS Deployment
make deploy-dev     # Deploy to development environment
make deploy-prod    # Deploy to production environment
make update-lambda  # Update Lambda function code only
make delete-stack   # Delete CloudFormation stack

# Utilities
make lambda-build   # Build for AWS Lambda
make clean          # Clean build artifacts
make help           # Show all available targets
```

## Configuration

### Default Settings
- Silver eligibility: 30 days
- Gold eligibility: 90 days
- Maximum absence count: 3
- Web server port: 8080

### Customization
Modify `internal/models/config.go` or use environment variables to customize settings.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions:
- Create an issue in the repository
- Contact the development team
- Check the documentation for common solutions

## Acknowledgments

- UO Outlands community for requirements and feedback
- Discord.js community for bot development guidance
- AWS documentation for serverless architecture patterns
