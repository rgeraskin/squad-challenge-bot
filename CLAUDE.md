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

## Architecture

Telegram bot built with Clean Architecture patterns using `gopkg.in/telebot.v3`.

**Layer Flow**: `handlers → services → repository → SQLite`

### Key Layers

- **handlers/** - Telegram message/callback handlers. Entry point for all user interactions.
- **service/** - Business logic layer. All domain rules (max 50 tasks, max 10 participants, emoji uniqueness per challenge) are enforced here. Defines sentinel errors like `ErrChallengeNotFound`, `ErrNotAdmin`.
- **repository/** - Data access interfaces in `interfaces.go`, SQLite implementation in `sqlite/`. Migrations are embedded via `embed.FS`.
- **domain/** - Entity definitions with no dependencies.

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

- Max 10 challenges per user
- Max 10 participants per challenge
- Max 50 tasks per challenge
- Daily task limit: 0-50 (0 = unlimited)
- Emojis must be unique within a challenge
- Only challenge creator (admin) can modify tasks/settings

## Key Features

- **Daily Limits**: Challenges can limit tasks completed per day (enforced per user's local time)
- **Sequential Mode**: `HideFutureTasks` hides tasks until previous ones are completed
- **Time Zone Sync**: Each participant stores `TimeOffsetMinutes` for accurate daily limit resets

## Message Formatting

- Use `tele.ModeHTML` when sending messages with HTML tags (`<b>`, `<i>`, etc.)
- Bot messages use friendly, casual tone with emojis
- Admin indicator uses 👑 emoji in team progress views
