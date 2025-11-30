# Implementation Plan: Hide Future Tasks

## Overview
Add a "Hide Future Tasks" setting that hides task names for tasks beyond the participant's current task. Each participant sees tasks based on their own progress. Hidden tasks show a spoiler placeholder and cannot be completed until previous tasks are done.

---

## 1. Database Migration

**File**: `internal/repository/sqlite/migrations/004_hide_future_tasks.sql`

```sql
ALTER TABLE challenges ADD COLUMN hide_future_tasks INTEGER DEFAULT 0;
```

- `hide_future_tasks`: 0 = show all task names, 1 = hide tasks after current

**Update**: `internal/repository/sqlite/db.go` - add migration to list

---

## 2. Domain Changes

**File**: `internal/domain/challenge.go`
- Add `HideFutureTasks bool` field

---

## 3. New State

**File**: `internal/domain/state.go`

Add:
- `StateAwaitingHideFutureTasks` - challenge creation flow (after daily limit)
- `StateAwaitingNewHideFutureTasks` - admin toggling setting (if needed, or just use callback)

---

## 4. Repository Changes

**File**: `internal/repository/sqlite/challenge.go`
- Update `Create()` to include `hide_future_tasks`
- Update `GetByID()` to read `hide_future_tasks`
- Add `UpdateHideFutureTasks(challengeID string, hide bool) error`

**File**: `internal/repository/interfaces.go`
- Update `ChallengeRepository` interface with new method

---

## 5. Service Changes

**File**: `internal/service/challenge.go`
- Update `Create()` signature to accept `hideFutureTasks bool`
- Add `UpdateHideFutureTasks(challengeID string, hide bool) error`
- Add `ToggleHideFutureTasks(challengeID string) (newValue bool, err error)` - toggle and return new state

---

## 6. Challenge Creation Flow Changes

**File**: `internal/bot/handlers/challenge.go`

### Current flow:
1. `StateAwaitingChallengeName`
2. `StateAwaitingChallengeDescription` (skippable)
3. `StateAwaitingCreatorName`
4. `StateAwaitingCreatorEmoji`
5. `StateAwaitingDailyLimit` (skippable)
6. `StateAwaitingCreatorSyncTime` (skippable) → creates challenge

### New flow:
1. `StateAwaitingChallengeName`
2. `StateAwaitingChallengeDescription` (skippable)
3. `StateAwaitingCreatorName`
4. `StateAwaitingCreatorEmoji`
5. `StateAwaitingDailyLimit` (skippable)
6. **`StateAwaitingHideFutureTasks`** (Yes/No buttons) ← NEW
7. `StateAwaitingCreatorSyncTime` (skippable) → creates challenge

### Changes:
- Modify `processDailyLimit()` and `skipDailyLimit()`: transition to `StateAwaitingHideFutureTasks` instead of `StateAwaitingCreatorSyncTime`
- Add `processHideFutureTasks(hide bool)`: store in temp data, transition to `StateAwaitingCreatorSyncTime`
- Update `finishChallengeCreation()`: pass `hideFutureTasks` to service

### UI for hide future tasks prompt:
```
👁 Hide Future Tasks

Do you want to hide task names until participants reach them?

When enabled:
• Participants only see names of completed tasks and their current task
• Future tasks show "🔒 Complete previous tasks to unlock"
• Each participant sees based on their own progress
```

**Keyboard**: `HideFutureTasksChoice()` - two buttons in a row:
- "✅ Yes, hide" → callback `hide_future_yes`
- "❌ No, show all" → callback `hide_future_no`

---

## 7. Admin Panel Changes

**File**: `internal/bot/keyboards/inline.go`

### Current layout:
```
Row 1: "➕ Add Task" | "📋 Edit Tasks"
Row 2: "✏️ Name" | "📝 Description"
Row 3: "⏱ Daily Limit: X"
Row 4: "🗑 Delete Challenge" | "🏠 Main Menu"
```

### New layout:
```
Row 1: "➕ Add Task" | "📋 Edit Tasks"
Row 2: "✏️ Name" | "📝 Description"
Row 3: "⏱ Daily Limit: X" | "👁 Tasks: Sequential" (or "👁 Tasks: All Visible")
Row 4: "🗑 Delete Challenge" | "🏠 Main Menu"
```

### Update `AdminPanel()` function:
- Accept additional parameter: `hideFutureTasks bool`
- Create button text based on state:
  - If `hideFutureTasks`: `"👁 Tasks: Sequential"`
  - If not: `"👁 Tasks: All Visible"`
- Button callback: `toggle_hide_future`
- Place in same row as daily limit button (Row 3)

**File**: `internal/bot/handlers/admin.go`
- Update `showAdminPanel()` to pass `challenge.HideFutureTasks` to keyboard
- Add `handleToggleHideFutureTasks()`: toggle setting, refresh admin panel

---

## 8. Task List View Changes

**File**: `internal/bot/views/tasklist.go`

### Update `TaskListData` struct:
```go
type TaskListData struct {
    // ... existing fields ...
    HideFutureTasks bool  // NEW: challenge setting
    CurrentTaskNum  int   // Already exists: participant's current task number
}
```

### Update `RenderTaskList()` function:

When rendering each task in the window:
1. Determine if task should be hidden:
   - `isHidden := data.HideFutureTasks && task.OrderNum > data.CurrentTaskNum`
2. If hidden, render as:
   ```
   ⬜ Task X: <tg-spoiler>🔒 Complete previous tasks to unlock</tg-spoiler>
   ```
3. If not hidden, render normally with task title

### Example output (participant on task 3, setting enabled):
```
📋 Challenge Name

✅ Task 1: Morning stretch
✅ Task 2: Drink water
⬜ Task 3: Read 10 pages        ← current task (visible)
⬜ Task 4: <tg-spoiler>🔒 Complete previous tasks to unlock</tg-spoiler>
⬜ Task 5: <tg-spoiler>🔒 Complete previous tasks to unlock</tg-spoiler>

Progress: 2/10 tasks
```

---

## 9. Task Detail View Changes

**File**: `internal/bot/views/taskdetail.go`

### Update `TaskDetailData` struct:
```go
type TaskDetailData struct {
    // ... existing fields ...
    IsHidden bool  // NEW: whether this task is hidden for the viewer
}
```

### Update `RenderTaskDetail()` function:

If `data.IsHidden`:
```
🔒 Task X: Hidden

This task is not yet unlocked.

Complete your previous tasks first to reveal this task's details.

Your current task: Task Y
```

If not hidden, render normally.

**File**: `internal/bot/handlers/task.go`

### Update `handleViewTask()`:
1. Get participant's current task number
2. Get challenge's `HideFutureTasks` setting
3. Calculate if task is hidden: `isHidden := challenge.HideFutureTasks && task.OrderNum > currentTaskNum`
4. Pass `IsHidden` to view data
5. Show back button regardless

---

## 10. Task Completion Changes

**File**: `internal/bot/handlers/progress.go`

### Update `handleCompleteTask()` and `handleCompleteCurrent()`:

Before completing, add check:
```go
// Check if task is hidden (cannot complete hidden tasks)
if challenge.HideFutureTasks {
    currentTaskNum, _ := h.completion.GetCurrentTaskNum(participant.ID, challengeID)
    if task.OrderNum > currentTaskNum {
        // Task is hidden, cannot complete
        return c.Send("🔒 This task is locked.\n\nComplete your previous tasks first.")
    }
}
```

This ensures participants cannot complete tasks out of order when the setting is enabled.

---

## 11. Callback Handler Updates

**File**: `internal/bot/handlers/callbacks.go`

Add new actions:
- `hide_future_yes` → `processHideFutureTasks(true)` (creation flow)
- `hide_future_no` → `processHideFutureTasks(false)` (creation flow)
- `toggle_hide_future` → `handleToggleHideFutureTasks()` (admin panel)

Add `toggle_hide_future` to `adminActions` map.

---

## 12. Text Handler Routing

**File**: `internal/bot/handlers/text.go`

No changes needed - this feature uses callback buttons, not text input.

---

## 13. New Keyboard Function

**File**: `internal/bot/keyboards/inline.go`

```go
// HideFutureTasksChoice returns keyboard for hide future tasks selection
func HideFutureTasksChoice() *tele.ReplyMarkup {
    menu := &tele.ReplyMarkup{}

    yesBtn := menu.Data("✅ Yes, hide", "hide_future_yes")
    noBtn := menu.Data("❌ No, show all", "hide_future_no")

    menu.Inline(
        menu.Row(yesBtn, noBtn),
    )
    return menu
}
```

---

## 14. Handler Updates for Creation Flow

**File**: `internal/bot/handlers/challenge.go`

### Add `askHideFutureTasks()`:
```go
func (h *Handler) askHideFutureTasks(c tele.Context) error {
    h.state.SetState(c.Sender().ID, domain.StateAwaitingHideFutureTasks)

    text := `👁 Hide Future Tasks

Do you want to hide task names until participants reach them?

When enabled:
• Participants only see names of completed tasks and their current task
• Future tasks show "🔒 Complete previous tasks to unlock"
• Each participant sees based on their own progress`

    return c.Send(text, keyboards.HideFutureTasksChoice())
}
```

### Add `processHideFutureTasks(hide bool)`:
```go
func (h *Handler) processHideFutureTasks(c tele.Context, hide bool) error {
    userID := c.Sender().ID

    // Store in temp data
    tempData, _ := h.state.GetTempData(userID)
    if tempData == nil {
        tempData = make(map[string]string)
    }
    if hide {
        tempData["hide_future_tasks"] = "1"
    } else {
        tempData["hide_future_tasks"] = "0"
    }
    h.state.SetStateWithData(userID, domain.StateAwaitingCreatorSyncTime, tempData)

    // Proceed to time sync
    return h.askCreatorSyncTime(c)
}
```

### Modify `processDailyLimit()`:
Change transition from `StateAwaitingCreatorSyncTime` to calling `askHideFutureTasks()`

### Modify `skipDailyLimit()`:
Change transition from `StateAwaitingCreatorSyncTime` to calling `askHideFutureTasks()`

### Modify `finishChallengeCreation()`:
```go
// Parse hide future tasks setting
hideFutureTasks := tempData["hide_future_tasks"] == "1"

// Update Create call
challenge, err := h.challenge.Create(name, description, creatorID, dailyLimit, hideFutureTasks)
```

---

## File Change Summary

| File | Changes |
|------|---------|
| `migrations/004_hide_future_tasks.sql` | NEW - add column |
| `repository/sqlite/db.go` | Add migration to list |
| `domain/challenge.go` | Add HideFutureTasks field |
| `domain/state.go` | Add StateAwaitingHideFutureTasks |
| `repository/sqlite/challenge.go` | Update Create, GetByID, add UpdateHideFutureTasks |
| `repository/interfaces.go` | Update ChallengeRepository interface |
| `service/challenge.go` | Update Create signature, add UpdateHideFutureTasks, ToggleHideFutureTasks |
| `handlers/challenge.go` | Add askHideFutureTasks, processHideFutureTasks, modify flow |
| `handlers/admin.go` | Update showAdminPanel, add handleToggleHideFutureTasks |
| `handlers/task.go` | Update handleViewTask for hidden task check |
| `handlers/progress.go` | Add hidden task completion check |
| `handlers/callbacks.go` | Add hide_future_yes, hide_future_no, toggle_hide_future actions |
| `keyboards/inline.go` | Update AdminPanel, add HideFutureTasksChoice |
| `views/tasklist.go` | Update TaskListData, RenderTaskList for hidden tasks |
| `views/taskdetail.go` | Update TaskDetailData, RenderTaskDetail for hidden tasks |

---

## Testing Plan

1. **Unit tests** for new service functions:
   - `UpdateHideFutureTasks()`
   - `ToggleHideFutureTasks()`

2. **Flow tests** for:
   - Challenge creation with hide future tasks enabled
   - Challenge creation with hide future tasks disabled
   - Admin toggling the setting
   - Task list rendering with setting enabled (various progress states)
   - Task list rendering with setting disabled
   - Viewing hidden task details
   - Attempting to complete hidden task (should be blocked)
   - Completing current task reveals next task

3. **Edge cases**:
   - Participant has completed all tasks (nothing hidden)
   - Participant on first task (all future tasks hidden)
   - Admin in edit tasks view (should see all tasks)
   - Toggle setting mid-challenge (existing participants see updated view)
   - Only one task in challenge (nothing to hide)
