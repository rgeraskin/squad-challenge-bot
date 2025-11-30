# Implementation Plan: Daily Task Limit & Time Sync

## Overview
Add ability to limit tasks completed per day per user, with time synchronization to handle different timezones.

---

## 1. Database Migration

**File**: `internal/repository/sqlite/migrations/003_daily_limit.sql`

```sql
ALTER TABLE challenges ADD COLUMN daily_task_limit INTEGER DEFAULT 0;
ALTER TABLE participants ADD COLUMN time_offset_minutes INTEGER DEFAULT 0;
```

- `daily_task_limit`: 0 = unlimited, >0 = max completions per day
- `time_offset_minutes`: offset from server time in minutes (can be negative)

**Update**: `internal/repository/sqlite/db.go` - add migration to list

---

## 2. Domain Changes

**File**: `internal/domain/challenge.go`
- Add `DailyTaskLimit int` field

**File**: `internal/domain/participant.go`
- Add `TimeOffsetMinutes int` field

---

## 3. New States

**File**: `internal/domain/state.go`

Add:
- `StateAwaitingDailyLimit` - challenge creation flow
- `StateAwaitingCreatorSyncTime` - creator time sync during challenge creation
- `StateAwaitingNewDailyLimit` - admin editing daily limit
- `StateAwaitingSyncTime` - user joining + settings re-sync

---

## 4. Repository Changes

**File**: `internal/repository/sqlite/challenge.go`
- Update `Create()` to include `daily_task_limit`
- Update `GetByID()` to read `daily_task_limit`
- Add `UpdateDailyLimit(challengeID string, limit int) error`

**File**: `internal/repository/sqlite/participant.go`
- Update `Create()` to include `time_offset_minutes`
- Update queries to read `time_offset_minutes`
- Add `UpdateTimeOffset(participantID int64, offsetMinutes int) error`

**File**: `internal/repository/sqlite/completion.go`
- Add `CountCompletionsInRange(participantID int64, from, to time.Time) (int, error)`

---

## 5. Service Changes

**File**: `internal/service/challenge.go`
- Update `Create()` signature to accept `dailyLimit int`
- Add `UpdateDailyLimit(challengeID string, limit int) error`

**File**: `internal/service/participant.go`
- Update `Join()` signature to accept `timeOffsetMinutes int`
- Add `UpdateTimeOffset(participantID int64, offsetMinutes int) error`

**File**: `internal/service/completion.go`
- Add `GetCompletionsToday(participant *domain.Participant) (int, error)`
- Add `CanCompleteTask(participant *domain.Participant, challenge *domain.Challenge) (allowed bool, completed int, limit int, timeToReset time.Duration, err error)`
- Add helper `getUserDayBoundaries(offsetMinutes int) (start, end time.Time)`

---

## 6. Challenge Creation Flow Changes

**File**: `internal/bot/handlers/challenge.go`

### Current flow:
1. `StateAwaitingChallengeName`
2. `StateAwaitingChallengeDescription` (skippable)
3. `StateAwaitingCreatorName`
4. `StateAwaitingCreatorEmoji` → creates challenge

### New flow:
1. `StateAwaitingChallengeName`
2. `StateAwaitingChallengeDescription` (skippable)
3. `StateAwaitingCreatorName`
4. `StateAwaitingCreatorEmoji` → transitions to daily limit
5. `StateAwaitingDailyLimit` (skippable) → transitions to time sync
6. `StateAwaitingCreatorSyncTime` (skippable) → creates challenge

### Changes:
- Modify `processCreatorEmoji()`: don't create challenge, store emoji in temp data, transition to `StateAwaitingDailyLimit`
- Add `processDailyLimit()`: validate input (1-50), store in temp data, transition to `StateAwaitingCreatorSyncTime`
- Add `skipDailyLimit()`: store limit=0 in temp data, transition to `StateAwaitingCreatorSyncTime`
- Add `processCreatorSyncTime()`: validate HH:MM, calculate offset, create challenge with all temp data
- Add `skipCreatorSyncTime()`: create challenge with offset=0

### UI for daily limit prompt:
```
⏱ Daily Task Limit

How many tasks can each participant complete per day?

Enter a number (1-50) or tap Skip for unlimited.
```

**Keyboard**: Skip button

---

## 7. Join Challenge Flow Changes

**File**: `internal/bot/handlers/join.go`

### Current flow:
1. `StateAwaitingChallengeID` (or deep link)
2. `StateAwaitingParticipantName`
3. `StateAwaitingParticipantEmoji` → joins challenge

### New flow:
1. `StateAwaitingChallengeID` (or deep link)
2. `StateAwaitingParticipantName`
3. `StateAwaitingParticipantEmoji` → transitions to time sync
4. `StateAwaitingSyncTime` (skippable) → joins challenge

### Changes:
- Modify `processParticipantEmoji()`: don't join, store emoji in temp data, transition to `StateAwaitingSyncTime`
- Add `processSyncTime()`: validate HH:MM format, calculate offset, join challenge
- Add `skipSyncTime()`: join with offset=0

### UI for time sync prompt:
```
🕐 Synchronize Your Clock

Please enter your current time in 24-hour format (HH:MM).
Example: 14:30 or 09:15

This helps us track your daily progress correctly.

Current server time: HH:MM
```

**Keyboard**: Skip button with note "(will use server time)"

---

## 8. Admin Panel Changes

**File**: `internal/bot/keyboards/inline.go`

### Current layout:
```
Row 1: "➕ Add Task" | "📋 Edit Tasks"
Row 2: "✏️ Name" | "📝 Description"
Row 3: "🗑 Delete Challenge" | "🏠 Main Menu"
```

### New layout:
```
Row 1: "➕ Add Task" | "📋 Edit Tasks"
Row 2: "✏️ Name" | "📝 Description"
Row 3: "⏱ Daily Limit: X" or "⏱ Daily Limit: ∞"
Row 4: "🗑 Delete Challenge" | "🏠 Main Menu"
```

**File**: `internal/bot/handlers/admin.go`
- Update `showAdminPanel()` to show daily limit in info text
- Add `handleEditDailyLimit()`: show prompt, set `StateAwaitingNewDailyLimit`
- Add `processNewDailyLimit()`: validate, update, show admin panel

### UI for editing:
```
⏱ Edit Daily Limit

Current limit: X tasks/day (or "unlimited")

Enter a new limit (1-50) or 0 for unlimited.
```

**Keyboard**: Cancel button

---

## 9. User Settings Changes

**File**: `internal/bot/keyboards/inline.go`

### Current layout:
```
Row 1: "🔔 Notifications: ON/OFF"
Row 2: "✏️ Change Name" | "😀 Change Emoji"
Row 3: "🔗 Share the Challenge"
Row 4: "🚫 Leave" | "⬅️ Back" (non-admin) or just "⬅️ Back" (admin)
```

### New layout:
```
Row 1: "🔔 Notifications: ON/OFF"
Row 2: "✏️ Change Name" | "😀 Change Emoji"
Row 3: "🕐 Sync Time" | "🔗 Share the Challenge"
Row 4: "🚫 Leave" | "⬅️ Back" (non-admin) or just "⬅️ Back" (admin)
```

**File**: `internal/bot/handlers/settings.go`
- Update `showSettings()` to display user's synced time: `🕐 Your time: HH:MM`
- Add `handleSyncTime()`: show prompt, set `StateAwaitingSyncTime`

---

## 10. Task Completion Enforcement

**File**: `internal/bot/handlers/progress.go`

### Modify `handleCompleteTask()` and `handleCompleteCurrent()`:

Before completing:
1. Get challenge and check if `DailyTaskLimit > 0`
2. If limited, call `CanCompleteTask()`
3. If not allowed, show limit reached message and return
4. If allowed, proceed with completion

### Limit reached message:
```
⏱ Daily Limit Reached

You've completed X/X tasks today.

New day starts in: HH:MM:SS

Come back tomorrow to continue!
```

**Keyboard**: "⬅️ Back" button

### After successful completion (if daily limit set):
Show feedback: `✅ Task completed! (X/Y today, resets in HH:MM)`

---

## 11. Main View Changes (Optional Enhancement)

**File**: `internal/bot/handlers/main.go` or wherever main view is rendered

If challenge has daily limit > 0, add to the view text:
```
📅 Today: X/Y completed (Resets in HH:MM)
```

Only show this line if daily limit is set (not unlimited).

---

## 12. Text Handler Routing

**File**: `internal/bot/handlers/text.go`

Add cases for new states:
- `StateAwaitingDailyLimit` → `processDailyLimit()`
- `StateAwaitingCreatorSyncTime` → `processCreatorSyncTime()`
- `StateAwaitingNewDailyLimit` → `processNewDailyLimit()`
- `StateAwaitingSyncTime` → `processSyncTime()`

---

## 13. Callback Handler Updates

**File**: `internal/bot/handlers/callbacks.go`

Add new actions:
- `edit_daily_limit` → `handleEditDailyLimit()`
- `sync_time` → `handleSyncTime()`
- `skip_daily_limit` → `skipDailyLimit()`
- `skip_creator_sync_time` → `skipCreatorSyncTime()`
- `skip_sync_time` → `skipSyncTime()`

---

## 14. Validation Functions

**File**: `internal/util/validation.go` (or new file)

- `ParseTime(input string) (hours, minutes int, err error)` - validates HH:MM format
- `CalculateTimeOffset(userHours, userMinutes int) int` - returns offset in minutes
- `ValidateDailyLimit(input string) (int, error)` - validates 1-50 or 0

---

## 15. Time Calculation Helpers

**File**: `internal/service/completion.go` or `internal/util/time.go`

```go
// GetUserLocalTime returns current time adjusted by user's offset
func GetUserLocalTime(offsetMinutes int) time.Time

// GetUserDayStart returns start of user's current day (midnight)
func GetUserDayStart(offsetMinutes int) time.Time

// GetUserDayEnd returns end of user's current day
func GetUserDayEnd(offsetMinutes int) time.Time

// TimeUntilUserMidnight returns duration until user's next day
func TimeUntilUserMidnight(offsetMinutes int) time.Duration

// FormatDuration formats duration as "HH:MM:SS" or "HH:MM"
func FormatDuration(d time.Duration) string
```

---

## File Change Summary

| File | Changes |
|------|---------|
| `migrations/003_daily_limit.sql` | NEW - add columns |
| `repository/sqlite/db.go` | Add migration to list |
| `domain/challenge.go` | Add DailyTaskLimit field |
| `domain/participant.go` | Add TimeOffsetMinutes field |
| `domain/state.go` | Add 4 new states |
| `repository/sqlite/challenge.go` | Update Create, GetByID, add UpdateDailyLimit |
| `repository/sqlite/participant.go` | Update Create, queries, add UpdateTimeOffset |
| `repository/sqlite/completion.go` | Add CountCompletionsInRange |
| `repository/interfaces.go` | Update interfaces |
| `service/challenge.go` | Update Create, add UpdateDailyLimit |
| `service/participant.go` | Update Join, add UpdateTimeOffset |
| `service/completion.go` | Add daily limit checking functions |
| `handlers/challenge.go` | Modify emoji handler, add daily limit + creator sync handlers |
| `handlers/join.go` | Modify emoji handler, add sync time handlers |
| `handlers/admin.go` | Update panel, add daily limit editing |
| `handlers/settings.go` | Update view, add sync time button handler |
| `handlers/progress.go` | Add daily limit enforcement |
| `handlers/main.go` | Add daily progress to main view |
| `handlers/text.go` | Add state routing |
| `handlers/callbacks.go` | Add new actions |
| `keyboards/inline.go` | Update AdminPanel, Settings, add Skip keyboards |
| `util/validation.go` | Add time/limit validation (or existing util file) |

---

## Testing Plan

1. **Unit tests** for new service functions:
   - `GetCompletionsToday()`
   - `CanCompleteTask()`
   - Time calculation helpers

2. **Flow tests** for:
   - Challenge creation with daily limit
   - Challenge creation skipping daily limit
   - Join with time sync
   - Join skipping time sync
   - Admin editing daily limit
   - User re-syncing time
   - Task completion within limit
   - Task completion blocked by limit

3. **Edge cases**:
   - Day boundary transitions
   - Negative time offsets
   - Limit = 1 (edge case)
   - Changing limit mid-challenge
