# SquadChallengeBot

A Telegram bot for creating and managing team challenges with task tracking and progress visualization.

## Features

- **Challenge Management**: Create challenges with up to 50 tasks and 10 participants
- **Task Tracking**: Complete tasks in any order, track progress with visual progress bars
- **Daily Limits**: Set a daily task limit (1-50 tasks/day) to pace your challenge
- **Sequential Mode**: Hide future tasks until previous ones are completed
- **Time Zone Sync**: Sync your local time for accurate daily limit resets
- **Team Progress**: View team leaderboard sorted by completion percentage
- **Deep Links**: Share challenges via `t.me/bot?start=CHALLENGE_ID`
- **Admin Controls**: Rename challenges, reorder/edit/delete tasks, configure limits
- **Super Admin**: System-wide admin can view all challenges, modify settings, and grant super admin to others
- **Templates**: Super admins can create reusable templates from existing challenges for quick challenge creation
- **Notifications**: Get notified when teammates complete tasks or finish challenges

## Requirements

- Go 1.21+
- SQLite3
- Telegram Bot Token (from [@BotFather](https://t.me/BotFather))

## Installation

```bash
# Clone the repository
git clone https://github.com/rgeraskin/squad-challenge-bot.git
cd squad-challenge-bot

# Install dependencies
go mod download

# Copy environment file and configure
cp .env.example .env
# Edit .env with your bot token
```

## Configuration

Create a `.env` file with:

```env
TELEGRAM_BOT_TOKEN=your_bot_token_here
DATABASE_PATH=./data/bot.db
LOG_LEVEL=info
HEALTH_PORT=8080
SUPER_ADMIN_ID=123456789  # Optional: Your Telegram user ID for super admin access
```

## Running

### Local

```bash
go run ./cmd/bot
```

### Docker

```bash
docker build -t squad-challenge-bot .
docker run -d \
  -e TELEGRAM_BOT_TOKEN=your_token \
  -v $(pwd)/data:/app/data \
  squad-challenge-bot
```

### Docker Compose

```bash
docker-compose up -d
```

## Development

### Project Structure

```
.
├── cmd/bot/              # Application entry point
├── internal/
│   ├── bot/              # Bot setup and routing
│   │   ├── handlers/     # Message and callback handlers
│   │   ├── keyboards/    # Inline keyboard builders
│   │   └── views/        # Message formatters
│   ├── config/           # Configuration loading
│   ├── domain/           # Domain entities
│   ├── logger/           # Structured logging
│   ├── repository/       # Data access layer
│   │   └── sqlite/       # SQLite implementation
│   ├── service/          # Business logic
│   └── util/             # Utilities (ID generation, emoji validation)
└── testutil/             # Test helpers
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/service/... -v
```

### Test Coverage

| Package | Coverage |
|---------|----------|
| service | 70% |
| repository/sqlite | 58% |
| bot/views | 57% |
| util | 96% |

## Bot Commands

- `/start` - Show main menu or join via deep link
- `/help` - Show help message

## User Flows

1. **Create Challenge**: Enter name → Description (optional) → Display name → Pick emoji → Daily limit → Sequential mode → Sync time
2. **Join Challenge**: Enter challenge ID → Display name → Pick emoji → Sync time
3. **Complete Tasks**: Tap task → Mark complete/incomplete (respects daily limits and sequential mode)
4. **View Progress**: See team progress with visual progress bars
5. **Admin Settings**: Edit name/description, set daily limits, toggle sequential mode, manage tasks

## Super Admin

Super admins have system-wide privileges:

- **View All Challenges**: See every challenge in the system (separated into "Your Challenges" and "Other Challenges")
- **Observer Mode**: View any challenge's tasks and progress without being a participant
- **Modify Settings**: Change daily limits and sequential mode for any challenge
- **Grant/Revoke**: Grant super admin privileges to other users by their Telegram ID
- **Templates**: Create, edit, and delete reusable challenge templates

To become the initial super admin, set `SUPER_ADMIN_ID` in your `.env` file to your Telegram user ID. You can find your ID in the bot's Settings menu.

## Health Checks

The bot exposes health check endpoints:

- `GET /health` - Liveness probe
- `GET /ready` - Readiness probe

## License

MIT
