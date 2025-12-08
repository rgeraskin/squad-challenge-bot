# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
# Install dependencies
go mod download

# Run the bot locally (requires .env with TELEGRAM_BOT_TOKEN)
go run ./cmd/bot

# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/service/... -v

# Run a single test
go test ./internal/service/... -run TestChallengeService_Create -v

# Build
go build ./cmd/bot

# Docker
docker build -t squad-challenge-bot .
docker-compose up -d
```

## Environment Configuration

Required `.env` variables:
- `TELEGRAM_BOT_TOKEN` - Bot token from @BotFather
- `DATABASE_PATH` - Path to SQLite database (default: `./data/bot.db`)
- `LOG_LEVEL` - Logging level (default: `info`)
- `HEALTH_PORT` - Health check endpoint port (default: `8080`)
- `SUPER_ADMIN_ID` - (Optional) Telegram user ID for initial super admin

## Architecture

Telegram bot built with Clean Architecture patterns using `gopkg.in/telebot.v3`.

**Layer Flow**: `handlers → services → repository → SQLite`

### Key Layers

- **handlers/** - Telegram message/callback handlers. Entry point for all user interactions.
- **service/** - Business logic layer. All domain rules (max 50 tasks, max 50 participants, emoji uniqueness per challenge) are enforced here. Defines sentinel errors like `ErrChallengeNotFound`, `ErrNotAdmin`.
- **repository/** - Data access interfaces in `interfaces.go`, SQLite implementation in `sqlite/`. Migrations are embedded via `embed.FS`.
- **domain/** - Entity definitions with no dependencies. Business logic limits (MaxParticipants, MaxTasksPerChallenge, MaxChallengesPerUser) are centralized in `domain/limits.go` for easy configuration.

### State Machine

The bot uses a conversation state machine (`domain/state.go`) to track multi-step flows like challenge creation and joining. States are persisted in SQLite per user. Key patterns:

- `StateService.SetStateWithData()` - Store temp data during multi-step flows
- `StateService.GetTempData()` - Retrieve accumulated data
- `StateService.ResetKeepChallenge()` - Clear state but preserve current challenge context

### Handler Structure

`Handler` struct aggregates all services. Handlers dispatch based on:
- `/start` command (with optional deep link payload for joining challenges)
- Callback data (parsed as `action|param1|param2`, telebot prefixes with `\f`)
- Text messages (routed by current state in `text.go`)

Admin-protected actions are checked via `adminActions` map in `callbacks.go`.

### Adding Callbacks

1. Add button in `keyboards/inline.go` with unique action name
2. Register action in `callbacks.go` switch statement
3. If admin-only, add to `adminActions` map

### Database Migrations

Migrations in `repository/sqlite/migrations/` are embedded and run in order on startup. Add new migrations with incremental numbering (e.g., `003_feature.sql`). Update the migrations list in `db.go`.

### Testing

- **Flow tests** (`testutil/flow_test.go`) - End-to-end user scenarios
- **MockContext** (`testutil/mock_context.go`) - Implements `tele.Context` for handler testing
- Tests use temp SQLite databases that are cleaned up automatically

## Key Constraints

**Business Logic Limits** (configurable in `internal/domain/limits.go`):
- Max 10 challenges per user (`MaxChallengesPerUser`)
- Max 50 participants per challenge (`MaxParticipants`)
- Max 50 tasks per challenge (`MaxTasksPerChallenge`)
- Task descriptions: max 1200 characters (`MaxTaskDescriptionLength`)

**Other Constraints**:
- Daily task limit: 1-50 tasks per day (0 = unlimited)
- Emojis must be unique within a challenge
- Only challenge creator (admin) can modify tasks/settings
- Super admins can view and modify any challenge in the system

## Key Features

- **Challenge Creation**: Create challenges with custom name, description, and settings
- **Task Management**: Add, edit, delete, and reorder tasks with unique emojis
- **Daily Limits**: Limit tasks completed per day (1-50, enforced per user's local time)
- **Sequential Mode**: `HideFutureTasks` hides tasks until previous ones are completed
- **Time Zone Sync**: Each participant stores `TimeOffsetMinutes` for accurate daily limit resets
- **Team Progress**: Leaderboard visualization sorted by completion percentage
- **Deep Links**: Share challenges via `t.me/botusername?start=CHALLENGE_ID`
- **Notifications**: Users receive notifications when teammates complete tasks or finish challenges
- **Super Admin System**: System-wide admins can:
  - View all challenges in the system
  - Observer mode for any challenge
  - Modify settings for any challenge
  - Grant/revoke super admin privileges
- **Templates System**: Super admins can create reusable challenge templates:
  - Create templates from existing challenges
  - Edit template settings and tasks
  - Users can create new challenges from templates

## Bot Commands

- `/start` - Main entry point. Shows main menu or joins challenge via deep link parameter

## Health Checks

The bot exposes HTTP health check endpoints (port configurable via `HEALTH_PORT`):
- `GET /health` - Liveness probe
- `GET /ready` - Readiness probe

## Message Formatting

- Use `tele.ModeHTML` when sending messages with HTML tags (`<b>`, `<i>`, etc.)
- Bot messages use friendly, casual tone with emojis
- Admin indicator uses 👑 emoji in team progress views
