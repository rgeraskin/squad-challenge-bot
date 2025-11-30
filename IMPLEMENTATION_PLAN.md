# SquadChallengeBot - Implementation Plan

## Overview

A Telegram bot that allows users to create challenges with multiple tasks and invite others to participate as a team. Users track their progress, see teammates' status, and celebrate together upon completion.

---

## Tech Stack

- **Language:** Go
- **Bot Framework:** [telebot](https://github.com/tucnak/telebot) v3
- **Database:** SQLite with [go-sqlite3](https://github.com/mattn/go-sqlite3)
- **SQL Toolkit:** [sqlx](https://github.com/jmoiron/sqlx) (struct scanning, named params)
- **Deployment:** Docker on VPS
- **Architecture:** Clean Architecture (handlers → services → repository)

---

## Data Models

### Challenge
```go
type Challenge struct {
    ID           string    // Random 8-char alphanumeric (e.g., "A3X9K2M1")
    Name         string    // Challenge title
    CreatorID    int64     // Telegram user ID of admin
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### Task
```go
type Task struct {
    ID            int64
    ChallengeID   string
    OrderNum      int      // 1-based task number
    Title         string
    Description   string   // Optional
    ImageFileID   string   // Optional - Telegram file_id for image
    CreatedAt     time.Time
}
```

### Participant
```go
type Participant struct {
    ID              int64
    ChallengeID     string
    TelegramID      int64
    DisplayName     string
    Emoji           string   // Single emoji chosen by user
    NotifyEnabled   bool     // Default: true
    JoinedAt        time.Time
}
```

### TaskCompletion
```go
type TaskCompletion struct {
    ID            int64
    TaskID        int64
    ParticipantID int64
    CompletedAt   time.Time
}
```

### UserState (for conversation flow)
```go
type UserState struct {
    TelegramID     int64
    State          string    // e.g., "idle", "awaiting_challenge_id", "awaiting_name", etc.
    TempData       string    // JSON blob for intermediate data
    CurrentChallenge string  // Active challenge ID (if any)
    UpdatedAt      time.Time
}
```

---

## Database Schema

```sql
CREATE TABLE challenges (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    creator_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    challenge_id TEXT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    order_num INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    image_file_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(challenge_id, order_num)
);

CREATE TABLE participants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    challenge_id TEXT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    telegram_id INTEGER NOT NULL,
    display_name TEXT NOT NULL,
    emoji TEXT NOT NULL,
    notify_enabled BOOLEAN DEFAULT 1,
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(challenge_id, telegram_id)
);

CREATE TABLE task_completions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    participant_id INTEGER NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    completed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(task_id, participant_id)
);

CREATE TABLE user_states (
    telegram_id INTEGER PRIMARY KEY,
    state TEXT NOT NULL DEFAULT 'idle',
    temp_data TEXT,
    current_challenge TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tasks_challenge ON tasks(challenge_id);
CREATE INDEX idx_participants_challenge ON participants(challenge_id);
CREATE INDEX idx_participants_telegram ON participants(telegram_id);
CREATE INDEX idx_completions_task ON task_completions(task_id);
CREATE INDEX idx_completions_participant ON task_completions(participant_id);
```

---

## User Flows

### Flow 1: Bot Start (Entry Point)

```
User sends /start
         │
         ▼
┌──────────────────────────────────────────────────┐
│  Welcome to SquadChallengeBot!                   │
│                                                  │
│  Your challenges:                                │
│  [🏆 30-Day Fitness (3/10 tasks)]               │
│  [🏆 Reading Challenge (5/5 ✅)]                │
│  [🏆 Morning Routine (0/7 tasks)]               │
│                                                  │
│  [🎯 Create Challenge]                           │
│  [🚀 Join Challenge]                             │
└──────────────────────────────────────────────────┘

Notes:
- Shows all challenges user is participating in
- Each challenge button shows: name + progress (completed/total)
- ✅ indicates fully completed challenges
- Clicking a challenge → goes to Main Challenge View for that challenge
- If user has no challenges, only show Create/Join buttons
```

---

### Flow 2: Create Challenge (Admin)

```
[Create Challenge] clicked
         │
         ▼
"Enter challenge name:"
[❌ Cancel]
         │
         ▼
User enters: "30-Day Fitness"
         │
         ▼
"Enter your display name:"
[❌ Cancel]
         │
         ▼
User enters: "John"
         │
         ▼
"Choose your emoji or send your own:"
[❌ Cancel]
         │
         ▼
User sends: "💪"
         │
         ▼
┌─────────────────────────────────────────┐
│  ✅ Challenge "30-Day Fitness" created! │
│                                         │
│  You are the admin of this challenge.   │
│  Now add tasks to your challenge.       │
│                                         │
└─────────────────────────────────────────┘
         │
         ▼
(Automatically goes to Admin View - Flow 12)

Note: [❌ Cancel] at any step → returns to Bot Start, clears temp data
```

---

### Flow 3: Add Task (Admin)

```
[Add Task] clicked
         │
         ▼
"Enter task title:"
[❌ Cancel]
         │
         ▼
User enters: "100 Push-ups"
         │
         ▼
"Send an image for this task (or click Skip):"
[⏭ Skip]  [❌ Cancel]
         │
         ▼
User sends image OR clicks Skip
         │
         ▼
"Enter task description (or click Skip):"
[⏭ Skip]  [❌ Cancel]
         │
         ▼
User enters description OR clicks Skip
         │
         ▼
┌─────────────────────────────────────────┐
│  ✅ Task #1 "100 Push-ups" added!       │
│                                         │
│  [➕ Add Another Task]                  │
│  [✅ Done Adding Tasks]                 │
└─────────────────────────────────────────┘

Buttons:
- "Add Another Task" → restart Add Task flow
- "Done Adding Tasks" → returns to Admin View (Flow 12)

Note: [❌ Cancel] at any step → returns to Admin View, discards partial task
```

---

### Flow 4: Join Challenge

```
[Join Challenge] clicked
         │
         ▼
"Enter the Challenge ID:"
[❌ Cancel]
         │
         ▼
User enters: "A3X9K2M1"
         │
         ▼
(Validate: exists? not full? not already member?)
         │
         ▼
"Challenge: 30-Day Fitness (3 tasks, 2 members)"
"Enter your display name:"
[❌ Cancel]
         │
         ▼
User enters: "Sarah"
         │
         ▼
"Choose your emoji or send your own:"
(Show suggested emojis not taken by other participants)
[❌ Cancel]
         │
         ▼
User sends: "🔥"

Note: [❌ Cancel] at any step → returns to Bot Start, clears temp data
         │
         ▼
┌─────────────────────────────────────────┐
│                                         │
│         🎯 CHALLENGE ACCEPTED! 🎯        │
│                                         │
│  Welcome to "30-Day Fitness", Sarah!    │
│                                         │
│  [🚀 Start Challenge]                   │
└─────────────────────────────────────────┘
         │
         ▼
[🚀 Start Challenge] clicked → Main Challenge View (Flow 5)
         │
         ▼
(Notify all participants: "🔥 Sarah joined the challenge!")
```

---

### Flow 5: Main Challenge View (Task List)

**Regular User Layout:**
```
┌──────────────────────────────────────────────────┐
│  🏆 30-Day Fitness                               │
│  Progress: 2/5 tasks • 3 members                 │
├──────────────────────────────────────────────────┤
│                                                  │
│  ✅ 1. Morning Stretch        💪🔥              │
│  ✅ 2. 50 Squats              💪🔥⭐            │
│  ⬜ 3. 100 Push-ups           💪⭐    ← YOU     │
│  ⬜ 4. 5K Run                                    │
│  ⬜ 5. Plank Challenge                           │
│  ⬜ 6. Cool Down                                 │
│  ⬜ 7. Rest Day                                  │
│                                                  │
├──────────────────────────────────────────────────┤
│  [✅ Complete #3]                                │  ← Row 1 (shows current task number)
│  [👥 Team Progress]  [🔗 Share ID]              │  ← Row 2
│  [⚙️ Settings]  [🚪 Exit]                       │  ← Row 3
└──────────────────────────────────────────────────┘
```

**Admin Layout:**
```
┌──────────────────────────────────────────────────┐
│  🏆 30-Day Fitness                               │
│  Progress: 2/5 tasks • 3 members                 │
├──────────────────────────────────────────────────┤
│                                                  │
│  (same task list as above)                       │
│                                                  │
├──────────────────────────────────────────────────┤
│  [✅ Complete #3]                                │  ← Row 1 (shows current task number)
│  [👥 Team Progress]  [🔗 Share ID]              │  ← Row 2
│  [🔧 Admin]  [⚙️ Settings]  [🚪 Exit]           │  ← Row 3
└──────────────────────────────────────────────────┘
```

Legend:
- ✅ = Completed task (by current user)
- ⬜ = Not completed task
- 💪🔥⭐ = Emojis of users currently on that task
- "← YOU" indicator shows current user's position

Buttons:
- "Complete #N" - marks task N as completed, button text updates to next incomplete task
- "Team Progress" - shows team progress view
- "Share ID" - shows challenge ID and deep link for sharing (available to all users)
- "Admin" - (admin only) goes to Admin View
- "Settings" - goes to settings
- "Exit" - returns to Bot Start (Entry Point) screen

**"Current Task" Logic:**
- Current task = first uncompleted task after the last completed task (in order)
- If user completed tasks 1, 2, 5 → current task is 3 (first gap)
- If user completed all tasks → no current task (hide "Complete current" button)
- Each user's emoji appears on THEIR current task (shows team progress visually)

**Task List Display Logic:**
- Show 2 previous tasks + current + 5 next tasks (max 8 visible)
- If user is on task 1, show tasks 1-7
- If user is on task 10 of 15, show tasks 8-15
- Each task button shows: `[status_emoji] [number]. [title] [participant_emojis]`
- Tasks are clickable → opens Task Detail View
- Emoji overflow: show max 4 emojis, then `+N` (e.g., `💪🔥⭐🎯 +3`)

**Empty Challenge (0 tasks):**
```
┌──────────────────────────────────────────────────┐
│  🏆 30-Day Fitness                               │
│  Progress: 0/0 tasks • 3 members                 │
├──────────────────────────────────────────────────┤
│                                                  │
│  📭 No tasks yet                                 │
│  Waiting for admin to add tasks...              │
│                                                  │
├──────────────────────────────────────────────────┤
│  [👥 Team Progress]  [🔗 Share ID]              │
│  [⚙️ Settings]  [🚪 Exit]                       │
└──────────────────────────────────────────────────┘
```
Note: "Complete current" button is hidden when no tasks exist

**All Tasks Completed:**
- When user completes the final task → immediately show Celebration (Flow 8)
- If user later uncompletes a task → return to Main View, re-enable "Complete current"

**Share ID Flow:**
```
[🔗 Share ID] clicked
         │
         ▼
┌──────────────────────────────────────────────────┐
│  📋 Share Challenge                              │
│                                                  │
│  Challenge ID: A3X9K2M1                          │
│                                                  │
│  Or share this link:                             │
│  t.me/SquadChallengeBot?start=A3X9K2M1           │
│                                                  │
│  [📋 Copy ID]  [🔗 Copy Link]                   │
│  [⬅️ Back]                                       │
└──────────────────────────────────────────────────┘
         │
         ▼ (Copy ID clicked)
"A3X9K2M1" (copyable text message)
         │
         ▼ (Copy Link clicked)
"t.me/SquadChallengeBot?start=A3X9K2M1" (copyable text message)
```

**Deep Link Handling:**
When user opens bot via deep link `t.me/SquadChallengeBot?start=A3X9K2M1`:
1. Extract challenge ID from start parameter
2. Check if challenge exists
3. Check if user is already a member → go to Main Challenge View
4. Otherwise → skip "Enter Challenge ID" step, go directly to name input

---

### Flow 6: Task Detail View

```
User clicks on "3. 100 Push-ups"
         │
         ▼
┌──────────────────────────────────────────────────┐
│  Task #3: 100 Push-ups                           │
├──────────────────────────────────────────────────┤
│                                                  │
│  [📷 Task Image Here - if exists]               │
│                                                  │
├──────────────────────────────────────────────────┤
│  Your status: ⬜ Not completed                   │
├──────────────────────────────────────────────────┤
│  Description:                                    │
│  Complete 100 push-ups throughout the day.       │
│  You can split them into sets of 10-20.          │
├──────────────────────────────────────────────────┤
│  Completed by:                                   │
│  💪 John • 🔥 Sarah                              │
│                                                  │
│  Not yet: ⭐ Mike                                │
├──────────────────────────────────────────────────┤
│  [✅ Complete]  [⬅️ Back]                        │  ← if NOT completed
│  [↩️ Uncomplete]  [⬅️ Back]                      │  ← if completed
└──────────────────────────────────────────────────┘

Buttons (conditional):
- If task NOT completed by user → show [✅ Complete] button
- If task IS completed by user → show [↩️ Uncomplete] button
- [⬅️ Back] → returns to Main Challenge View
```

---

### Flow 7: Team Progress View

```
[Team Progress] clicked
         │
         ▼
┌──────────────────────────────────────────────────┐
│  👥 Team Progress - 30-Day Fitness               │
├──────────────────────────────────────────────────┤
│                                                  │
│  💪 John (Admin)     ████████░░ 80% (4/5)       │
│  🔥 Sarah            ██████░░░░ 60% (3/5)       │
│  ⭐ Mike             ████░░░░░░ 40% (2/5)       │
│                                                  │
├──────────────────────────────────────────────────┤
│  [⬅️ Back]                                       │
└──────────────────────────────────────────────────┘

Note: Participants sorted by completion % descending (highest first)
```

---

### Flow 8: Challenge Completion (Celebration)

```
(When user completes final task)
         │
         ▼
┌──────────────────────────────────────────────────┐
│                                                  │
│  🎉🎊🏆 CONGRATULATIONS! 🏆🎊🎉                  │
│                                                  │
│  You completed "30-Day Fitness"!                 │
│                                                  │
│  ⏱ Time taken: 28 days                          │
│  📊 Tasks completed: 30/30                       │
│                                                  │
│  Team Status:                                    │
│  💪 John - ✅ Completed                          │
│  🔥 Sarah - 🔄 28/30 tasks                       │
│  ⭐ Mike - 🔄 25/30 tasks                        │
│                                                  │
├──────────────────────────────────────────────────┤
│  [👥 View Team]  [🏠 Main Menu]                  │
└──────────────────────────────────────────────────┘

(Notify team: "🎉 💪 John completed the challenge!")
```

---

### Flow 9: Edit Tasks (Admin)

```
[📋 Edit Tasks] clicked (from Admin View)
         │
         ▼
┌──────────────────────────────────────────────────┐
│  📋 Edit Tasks - 30-Day Fitness                  │
├──────────────────────────────────────────────────┤
│                                                  │
│  [1. Morning Stretch ✏️]                         │
│  [2. 50 Squats ✏️]                               │
│  [3. 100 Push-ups ✏️]                            │
│  ...                                             │
│                                                  │
├──────────────────────────────────────────────────┤
│  [🔀 Reorder Tasks]  [⬅️ Back]                   │  ← single row
└──────────────────────────────────────────────────┘
         │
         ▼ (click on task)
┌──────────────────────────────────────────────────┐
│  Edit Task #1: Morning Stretch                   │
├──────────────────────────────────────────────────┤
│  [📝 Edit Title]  [📷 Change Image]              │  ← Row 1
│  [📄 Edit Description]                           │  ← Row 2
│  [🗑 Delete Task]  [⬅️ Back]                     │  ← Row 3
└──────────────────────────────────────────────────┘

[🗑 Delete Task] clicked:
         │
         ▼
┌──────────────────────────────────────────────────┐
│  ⚠️ Delete task "Morning Stretch"?               │
│                                                  │
│  This will remove completion data for all users. │
│                                                  │
│  [✅ Yes, delete]  [❌ Cancel]                   │
└──────────────────────────────────────────────────┘
         │
         ▼ (if confirmed)
(Delete task, renumber remaining tasks)
(Return to Edit Tasks list)
```

---

### Flow 10: Settings

```
[Settings] clicked
         │
         ▼
┌──────────────────────────────────────────────────┐
│  ⚙️ Settings                                     │
├──────────────────────────────────────────────────┤
│                                                  │
│  Current challenge: 30-Day Fitness               │
│  Your emoji: 💪                                  │
│  Your name: John                                 │
│                                                  │
├──────────────────────────────────────────────────┤
│  [🔔 Notifications: ON]                          │  ← Row 1 (toggle button)
│  [✏️ Change Name]  [😀 Change Emoji]             │  ← Row 2
│  [🚫 Leave Challenge]  [⬅️ Back]                 │  ← Row 3
└──────────────────────────────────────────────────┘

Notifications toggle:
- Shows current state: [🔔 Notifications: ON] or [🔕 Notifications: OFF]
- Tapping toggles between ON/OFF
- Updates immediately, no confirmation needed
```

[🚫 Leave Challenge] clicked:
         │
         ▼
┌──────────────────────────────────────────────────┐
│  ⚠️ Are you sure you want to leave               │
│  "30-Day Fitness"?                               │
│                                                  │
│  Your progress will be deleted.                  │
│                                                  │
│  [✅ Yes, leave]  [❌ Cancel]                    │
└──────────────────────────────────────────────────┘
         │
         ▼ (if confirmed)
(Remove participant from challenge)
(Return to Bot Start)

Note: Admin cannot leave their own challenge (button hidden for admin)
```

---

### Flow 11: Reorder Tasks (Admin)

```
[🔀 Reorder Tasks] clicked (from Edit Tasks view)
         │
         ▼
┌──────────────────────────────────────────────────┐
│  🔀 Reorder Tasks - 30-Day Fitness               │
│                                                  │
│  Select a task to move:                          │
├──────────────────────────────────────────────────┤
│                                                  │
│  [1. Morning Stretch]                            │
│  [2. 50 Squats]                                  │
│  [3. 100 Push-ups]                               │
│  [4. 5K Run]                                     │
│  [5. Plank Challenge]                            │
│                                                  │
├──────────────────────────────────────────────────┤
│  [⬅️ Back]                                       │
└──────────────────────────────────────────────────┘
         │
         ▼ (user clicks on "3. 100 Push-ups")
┌──────────────────────────────────────────────────┐
│  🔀 Moving: "100 Push-ups"                       │
│                                                  │
│  Select new position:                            │
├──────────────────────────────────────────────────┤
│                                                  │
│  [⬆️ Move to position 1]                         │
│  [⬆️ Move to position 2]                         │
│  [   Current position: 3]  (disabled/grayed)     │
│  [⬇️ Move to position 4]                         │
│  [⬇️ Move to position 5]                         │
│                                                  │
├──────────────────────────────────────────────────┤
│  [❌ Cancel]                                     │
└──────────────────────────────────────────────────┘
         │
         ▼ (user clicks "Move to position 1")
┌──────────────────────────────────────────────────┐
│  ✅ Task moved!                                  │
│                                                  │
│  New order:                                      │
│  1. 100 Push-ups      ← moved here               │
│  2. Morning Stretch                              │
│  3. 50 Squats                                    │
│  4. 5K Run                                       │
│  5. Plank Challenge                              │
│                                                  │
├──────────────────────────────────────────────────┤
│  [🔀 Move Another]  [⬅️ Done]                    │
└──────────────────────────────────────────────────┘

Reorder Logic:
1. Admin selects task to move
2. Admin selects target position
3. All tasks between old and new position shift accordingly
4. Order numbers are recalculated
5. Show confirmation with new order
```

---

### Flow 12: Admin View

```
[🔧 Admin] clicked (from Main Challenge View)
OR automatically after creating a challenge
         │
         ▼
┌──────────────────────────────────────────────────┐
│  🔧 Admin Panel - 30-Day Fitness                 │
├──────────────────────────────────────────────────┤
│                                                  │
│  Challenge ID: A3X9K2M1                          │
│  Participants: 3/10                              │
│  Tasks: 5                                        │
│                                                  │
├──────────────────────────────────────────────────┤
│  [➕ Add Task]  [📋 Edit Tasks]                  │  ← Row 1
│  [✏️ Edit Challenge Name]                        │  ← Row 2
│  [🗑 Delete Challenge]  [🏠 Main Menu]           │  ← Row 3
└──────────────────────────────────────────────────┘

Buttons:
- "Add Task" → goes to Add Task flow (Flow 3)
- "Edit Tasks" → goes to Edit Tasks list (Flow 9)
- "Edit Challenge Name" → prompts for new name
- "Delete Challenge" → shows delete confirmation
- "Main Menu" → returns to Main Challenge View (Flow 5)
```

**Edit Challenge Name flow:**
```
[✏️ Edit Challenge Name] clicked
         │
         ▼
"Enter new challenge name:"
[❌ Cancel]
         │
         ▼
User enters: "60-Day Fitness"
         │
         ▼
┌──────────────────────────────────────────────────┐
│  ✅ Challenge renamed to "60-Day Fitness"        │
│                                                  │
│  [⬅️ Back to Admin]                              │
└──────────────────────────────────────────────────┘
```

**Delete Challenge flow:**
```
[🗑 Delete Challenge] clicked
         │
         ▼
┌──────────────────────────────────────────────────┐
│  ⚠️ DELETE CHALLENGE?                            │
│                                                  │
│  "30-Day Fitness" will be permanently deleted.   │
│                                                  │
│  This will remove:                               │
│  • All 5 tasks                                   │
│  • All 3 participants                            │
│  • All progress data                             │
│                                                  │
│  This action cannot be undone!                   │
│                                                  │
│  [🗑 Yes, delete everything]  [❌ Cancel]        │
└──────────────────────────────────────────────────┘
         │
         ▼ (if confirmed)
(Delete challenge and all related data)
(Notify all participants: "❌ Challenge '30-Day Fitness' has been deleted by admin")
(Return to Bot Start)
```

---

## Project Structure

```
challenge-accepted-bot/
├── cmd/
│   └── bot/
│       └── main.go              # Entry point
├── internal/
│   ├── bot/
│   │   ├── bot.go               # Bot initialization
│   │   ├── handlers/
│   │   │   ├── start.go         # /start command, challenge list
│   │   │   ├── start_test.go    # Start handler tests
│   │   │   ├── challenge.go     # Create/join challenge handlers
│   │   │   ├── challenge_test.go
│   │   │   ├── task.go          # Task add/edit/delete handlers
│   │   │   ├── task_test.go
│   │   │   ├── reorder.go       # Task reordering handlers
│   │   │   ├── admin.go         # Admin panel handlers
│   │   │   ├── admin_test.go
│   │   │   ├── progress.go      # Complete/uncomplete handlers
│   │   │   ├── progress_test.go
│   │   │   ├── settings.go      # Settings handlers
│   │   │   ├── settings_test.go
│   │   │   └── callbacks.go     # Callback query router
│   │   ├── keyboards/
│   │   │   ├── inline.go        # Inline keyboard builders
│   │   │   └── reply.go         # Reply keyboard builders
│   │   ├── views/
│   │   │   ├── tasklist.go      # Task list view builder
│   │   │   ├── tasklist_test.go
│   │   │   ├── taskdetail.go    # Task detail view builder
│   │   │   ├── progress.go      # Team progress view
│   │   │   └── celebration.go   # Completion celebration view
│   │   └── middleware/
│   │       ├── state.go         # User state middleware
│   │       └── admin.go         # Admin authorization middleware
│   ├── domain/
│   │   ├── challenge.go         # Challenge entity
│   │   ├── task.go              # Task entity
│   │   ├── participant.go       # Participant entity
│   │   └── completion.go        # TaskCompletion entity
│   ├── repository/
│   │   ├── sqlite/
│   │   │   ├── challenge.go     # Challenge repository
│   │   │   ├── challenge_test.go
│   │   │   ├── task.go          # Task repository
│   │   │   ├── task_test.go
│   │   │   ├── participant.go   # Participant repository
│   │   │   ├── participant_test.go
│   │   │   ├── completion.go    # Completion repository
│   │   │   ├── completion_test.go
│   │   │   ├── state.go         # User state repository
│   │   │   └── migrations.go    # DB migrations
│   │   └── interfaces.go        # Repository interfaces
│   ├── service/
│   │   ├── challenge.go         # Challenge business logic
│   │   ├── challenge_test.go
│   │   ├── task.go              # Task business logic
│   │   ├── task_test.go
│   │   ├── participant.go       # Participant business logic
│   │   ├── participant_test.go
│   │   ├── notification.go      # Notification service
│   │   ├── notification_test.go
│   │   ├── state.go             # State machine service
│   │   └── state_test.go
│   ├── testutil/
│   │   ├── mock_context.go      # Fake tele.Context for testing
│   │   ├── mock_repository.go   # In-memory repository mock
│   │   ├── mock_notifier.go     # Mock notification service
│   │   ├── flow_runner.go       # Flow test helper
│   │   └── fixtures.go          # Test data fixtures
│   ├── util/
│   │   ├── id.go                # ID generation (8-char random)
│   │   ├── id_test.go
│   │   ├── emoji.go             # Emoji helpers
│   │   └── emoji_test.go
│   └── config/
│       └── config.go            # Configuration
├── migrations/
│   └── 001_initial.sql          # Initial schema
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

---

## Implementation Details

### Challenge ID Generation
Generate unique 8-character alphanumeric IDs with collision retry logic:

```go
func generateUniqueID(repo ChallengeRepository) (string, error) {
    const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    const maxRetries = 10

    for i := 0; i < maxRetries; i++ {
        id := make([]byte, 8)
        for j := range id {
            id[j] = charset[rand.Intn(len(charset))]
        }

        // Check if ID already exists
        if !repo.ChallengeExists(string(id)) {
            return string(id), nil
        }
    }

    return "", errors.New("failed to generate unique ID after max retries")
}
```

### State Conflict Handling
When a callback button is pressed while user is in an input state (e.g., `awaiting_task_title`):
1. **Reset state to `idle`** - Clear any temp data
2. **Process the callback normally** - Handle the button action
3. **No error message needed** - User intentionally clicked a button

```go
// In callback handler middleware:
func handleCallback(c tele.Context) error {
    state := repo.GetUserState(c.Sender().ID)

    // If user was in an input state, reset it
    if state.State != "idle" {
        repo.ResetUserState(c.Sender().ID)
    }

    // Process callback normally
    return next(c)
}
```

---

## Security & Authorization

### Admin Authorization
Every admin action must verify that the requesting user is the challenge creator:

```go
// Before executing any admin action:
func isAdmin(challengeID string, telegramUserID int64) bool {
    challenge := repo.GetChallenge(challengeID)
    return challenge.CreatorID == telegramUserID
}
```

### Protected Admin Actions
The following callbacks/actions require admin verification:
- `admin_panel` - Access admin view
- `add_task` - Add new task
- `edit_tasks` - View edit tasks list
- `edit_task:*` - Edit specific task
- `delete_task:*` - Delete task
- `reorder_tasks` - Enter reorder mode
- `reorder_select:*` - Select task to move
- `reorder_move:*:*` - Move task to position
- `edit_challenge_name` - Change challenge name
- `delete_challenge` - Show delete challenge confirmation
- `confirm_delete_challenge` - Delete entire challenge

### Implementation Requirements
1. **Callback handler middleware**: Check admin status before processing admin callbacks
2. **Hide admin UI**: Don't show "Admin" button to non-admin users (already in Flow 5)
3. **Server-side validation**: Even if a user somehow sends an admin callback, reject it
4. **Fail securely**: On authorization failure, show generic error (don't reveal admin exists)

### Error Response for Unauthorized Access
```
"⚠️ You don't have permission to perform this action."
```

---

## Error Handling

| Scenario | Response |
|----------|----------|
| Invalid challenge ID | "❌ Challenge not found. Check the ID and try again." |
| Challenge full (10/10) | "❌ This challenge is full (10/10 participants)." |
| Already a member | "ℹ️ You're already participating in this challenge." |
| Emoji already taken | "❌ This emoji is already taken. Choose another:" |
| Empty task title | "❌ Task title cannot be empty." |
| Invalid emoji | "❌ Please send a single emoji." |
| Network/DB error | "⚠️ Something went wrong. Please try again." |
| Challenge has no tasks | "📭 No tasks yet. Waiting for admin to add tasks..." |
| User not in challenge | "❌ You're not a participant of this challenge." |
| Max tasks reached | "❌ Challenge has reached maximum of 50 tasks." |
| Max challenges reached | "❌ You've reached the maximum of 10 active challenges." |
| Image too large | "❌ Image is too large. Maximum size is 5MB." |

---

## Input Validation

| Field | Rules |
|-------|-------|
| Challenge name | 1-50 characters |
| Display name | 1-30 characters |
| Emoji | Single emoji only (validated with regex) |
| Task title | 1-100 characters |
| Task description | 0-500 characters (optional) |
| Challenge ID | Exactly 8 alphanumeric characters |
| Task image | Telegram photo only (file_id stored), max 5MB |
| Tasks per challenge | Maximum 50 tasks |
| Challenges per user | Maximum 10 active challenges |

---

## Testing Strategy

### Architecture for Testability
Structure code to separate Telegram from business logic:
```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Handlers      │ ──▶ │   Services      │ ──▶ │  Repository     │
│  (Telegram)     │     │ (Business Logic)│     │   (Database)    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
   thin layer              testable               mockable
```

### Test Types

**1. Unit Tests (Services Layer)**
Test business logic without Telegram dependencies:
```go
func TestCreateChallenge(t *testing.T) {
    repo := mocks.NewMockRepository()
    svc := service.NewChallengeService(repo)

    challenge, err := svc.CreateChallenge("30-Day Fitness", 12345)

    assert.NoError(t, err)
    assert.Len(t, challenge.ID, 8)
    assert.Equal(t, "30-Day Fitness", challenge.Name)
}
```

**2. Repository Tests (SQLite)**
Test database operations with test DB:
```go
func TestRepository_CascadeDelete(t *testing.T) {
    db := setupTestDB(t)
    db.CreateChallenge("ABC123", "Test", 1)
    db.CreateTask("ABC123", 1, "Task 1")

    db.DeleteChallenge("ABC123")

    tasks, _ := db.GetTasks("ABC123")
    assert.Empty(t, tasks)
}
```

**3. Flow Tests (Integration)**
Test complete user flows with mock Telegram context:
```go
func TestFlow_CreateChallenge(t *testing.T) {
    db := setupTestDB(t)
    handler := NewHandler(db)
    ctx := &MockContext{sender: &tele.User{ID: 12345}}

    // Step 1: Click Create Challenge
    ctx.callback = &tele.Callback{Data: "create_challenge"}
    handler.OnCallback(ctx)
    assert.Contains(t, ctx.LastMessage(), "Enter challenge name")

    // Step 2: Enter name
    ctx.text = "30-Day Fitness"
    handler.OnText(ctx)
    assert.Contains(t, ctx.LastMessage(), "Enter your display name")

    // ... continue flow
}
```

**4. State Transition Tests**
Test state machine transitions:
```go
func TestStateTransitions(t *testing.T) {
    tests := []struct {
        name         string
        initialState string
        input        string
        wantState    string
    }{
        {"create flow - name", "awaiting_challenge_name", "My Challenge", "awaiting_creator_name"},
        {"cancel resets state", "awaiting_task_title", "CANCEL", "idle"},
    }
    // ...
}
```

### Test Structure
```
internal/
├── bot/handlers/
│   ├── challenge.go
│   └── challenge_test.go      # Handler tests with mock context
├── service/
│   ├── challenge.go
│   └── challenge_test.go      # Unit tests (no Telegram)
├── repository/sqlite/
│   ├── challenge.go
│   └── challenge_test.go      # DB tests
└── testutil/
    ├── mock_context.go        # Fake tele.Context
    ├── mock_repository.go     # In-memory repository
    └── flow_runner.go         # Flow test helper
```

### Required Test Coverage
| Component | Min Coverage |
|-----------|--------------|
| Services | 80% |
| Repository | 70% |
| Handlers | 60% |
| State transitions | 100% |

---

## Implementation Phases

### Phase 1: Foundation (Core Setup)
- [ ] Initialize Go module
- [ ] Set up project structure
- [ ] Configure SQLite database
- [ ] Implement database migrations
- [ ] Create domain entities
- [ ] Implement repository layer
- [ ] Set up basic bot with /start command
- [ ] Implement user state management
- [ ] Implement admin authorization middleware
- [ ] Set up test infrastructure (mocks, helpers)

### Phase 2: Challenge Creation (Admin Flow)
- [ ] Create challenge flow (name input)
- [ ] Admin registration (name + emoji)
- [ ] Admin View panel
- [ ] Add task flow (title → image → description)
- [ ] Edit tasks list for admin
- [ ] Edit task flow (title, image, description)
- [ ] Delete task flow (with confirmation)
- [ ] Delete challenge flow (with confirmation)
- [ ] Reorder tasks flow
- [ ] Edit challenge name flow
- [ ] **Tests:** Challenge service unit tests
- [ ] **Tests:** Task service unit tests
- [ ] **Tests:** Create challenge flow test
- [ ] **Tests:** Add/edit/delete task flow tests

### Phase 3: Join Challenge (Participant Flow)
- [ ] Join challenge by ID
- [ ] Deep link support (`t.me/bot?start=ID`)
- [ ] Validate: exists, not full, not duplicate
- [ ] Participant registration (name + emoji selection)
- [ ] "Challenge Accepted!" welcome message
- [ ] Notify existing participants of new joiner
- [ ] **Tests:** Join challenge service tests
- [ ] **Tests:** Validation edge cases (full, duplicate, invalid ID)
- [ ] **Tests:** Join flow integration test

### Phase 4: Main Challenge View
- [ ] Build task list view with pagination (2 prev + 5 next)
- [ ] Show completion status per task
- [ ] Show participant emojis on tasks
- [ ] Highlight current user position
- [ ] "Complete #N" button
- [ ] "Share ID" button with deep link
- [ ] "Admin" button (for admin only)
- [ ] "Exit" button (returns to start screen)
- [ ] Start screen with user's challenge list
- [ ] **Tests:** Task list pagination logic
- [ ] **Tests:** Current task calculation
- [ ] **Tests:** Emoji overflow (+N) display

### Phase 5: Task Detail View
- [ ] Task detail view with image
- [ ] Show completion status
- [ ] Show description
- [ ] List who completed / not completed
- [ ] Conditional Complete/Uncomplete button
- [ ] Back button
- [ ] **Tests:** Complete/uncomplete service tests
- [ ] **Tests:** Task detail view rendering

### Phase 6: Team Progress, Settings & Notifications
- [ ] Team progress view (sorted by %)
- [ ] Progress bars per participant
- [ ] Notification on task completion
- [ ] Notification on user join
- [ ] Notification on challenge deletion
- [ ] Settings view
- [ ] Settings: toggle notifications
- [ ] Settings: change name
- [ ] Settings: change emoji
- [ ] Settings: leave challenge (with confirmation)
- [ ] **Tests:** Notification service tests
- [ ] **Tests:** Settings update tests
- [ ] **Tests:** Leave challenge flow test

### Phase 7: Celebration & Polish
- [ ] Challenge completion detection
- [ ] Celebration screen
- [ ] Summary statistics
- [ ] Notify team of completion
- [ ] Error handling improvements
- [ ] Edge cases (empty tasks, 0 participants, etc.)
- [ ] **Tests:** All state transition tests (100% coverage)
- [ ] **Tests:** Error handling tests
- [ ] **Tests:** Edge case tests

### Phase 8: Deployment
- [ ] Dockerfile
- [ ] docker-compose.yml
- [ ] Environment configuration
- [ ] Health checks
- [ ] Logging
- [ ] Deploy to VPS
- [ ] **Tests:** Run full test suite in CI
- [ ] **Tests:** Verify coverage thresholds met

---

## Callback Data Format

To handle inline button presses, use structured callback data:

```
Format: action:param1:param2

Examples:

# Start screen
- "create_challenge"
- "join_challenge"
- "open_challenge:ABC123"   # Open specific challenge from list

# Main view
- "task_detail:5"           # View task ID 5
- "complete_current"        # Complete current task (main view)
- "exit_challenge"          # Return to start screen
- "team_progress"
- "share_id"                # Show challenge ID view
- "copy_id"                 # Send copyable ID text
- "copy_link"               # Send copyable deep link
- "settings"
- "admin_panel"             # Admin: go to admin view

# Join flow (after deep link or successful join)
- "start_challenge:ABC123"  # Start challenge → go to Main Challenge View

# Task detail view
- "complete_task:5"         # Complete specific task (detail view)
- "uncomplete_task:5"       # Uncomplete specific task (detail view)

# Admin panel
- "add_task"
- "edit_tasks"              # Admin: edit tasks list
- "edit_challenge_name"     # Admin: edit challenge name
- "edit_task:5"             # Admin: edit specific task
- "delete_task:5"           # Admin: show delete confirmation
- "confirm_delete_task:5"   # Admin: confirm delete task
- "cancel_delete_task"      # Admin: cancel delete task
- "delete_challenge"        # Admin: show delete challenge confirmation
- "confirm_delete_challenge"# Admin: confirm delete entire challenge
- "cancel_delete_challenge" # Admin: cancel delete challenge
- "reorder_tasks"           # Admin: enter reorder mode
- "reorder_select:5"        # Admin: select task to move
- "reorder_move:5:2"        # Admin: move task 5 to position 2
- "reorder_cancel"          # Admin: cancel reorder

# Settings
- "toggle_notifications"
- "change_name"
- "change_emoji"
- "leave_challenge"         # Show leave confirmation
- "confirm_leave"           # Confirm leaving
- "cancel_leave"            # Cancel leaving

# Navigation
- "back_to_main"
- "back_to_admin"
- "back_to_tasks"
- "back_to_share"           # Back to share ID view
- "skip"
- "cancel"                  # Cancel current multi-step flow
```

---

## State Machine

User states for conversation flow:

```
idle                        → Default state

# Challenge creation
awaiting_challenge_name     → Creating challenge: waiting for name
awaiting_creator_name       → Creating challenge: waiting for display name
awaiting_creator_emoji      → Creating challenge: waiting for emoji

# Task management
awaiting_task_title         → Adding task: waiting for title
awaiting_task_image         → Adding task: waiting for image or skip
awaiting_task_description   → Adding task: waiting for description or skip
awaiting_edit_title         → Editing task: waiting for new title
awaiting_edit_description   → Editing task: waiting for new description
awaiting_edit_image         → Editing task: waiting for new image
reorder_select_task         → Reordering: waiting to select task to move
reorder_select_position     → Reordering: waiting to select target position

# Joining challenge
awaiting_challenge_id       → Joining: waiting for challenge ID
awaiting_participant_name   → Joining: waiting for display name
awaiting_participant_emoji  → Joining: waiting for emoji

# Admin
awaiting_new_challenge_name → Admin: waiting for new challenge name

# User settings
awaiting_new_name           → Settings: waiting for new display name
awaiting_new_emoji          → Settings: waiting for new emoji
```

**State Transitions:**
- Any state → `idle`: User clicks [❌ Cancel] or any callback button
- When transitioning to `idle`: Clear `temp_data` field
- All non-idle states have implicit "cancel" transition back to appropriate view

---

## Notification Messages

| Event | Message |
|-------|---------|
| User joins | `🎉 {emoji} {name} joined the challenge!` |
| Task completed | `✅ {emoji} {name} completed "{task_title}"!` |
| Challenge completed | `🏆 {emoji} {name} finished the challenge!` |
| Challenge deleted | `❌ Challenge "{challenge_name}" has been deleted by admin.` |

---

## Configuration (.env)

```env
TELEGRAM_BOT_TOKEN=your_bot_token_here
DATABASE_PATH=./data/bot.db
LOG_LEVEL=info
```

---

## Docker Setup

### Dockerfile
```dockerfile
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o bot ./cmd/bot

FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/bot .

VOLUME ["/app/data"]

CMD ["./bot"]
```

### docker-compose.yml
```yaml
version: '3.8'

services:
  bot:
    build: .
    restart: unless-stopped
    env_file:
      - .env
    volumes:
      - bot_data:/app/data

volumes:
  bot_data:
```

---

## Estimated Complexity

| Component | Files | Complexity |
|-----------|-------|------------|
| Database/Repository | 6-8 | Medium |
| Domain entities | 4-5 | Low |
| Handlers | 6-8 | High |
| Views/Keyboards | 5-6 | Medium |
| Services | 5-6 | Medium |
| State machine | 1-2 | Medium |
| **Total** | ~35 files | |

---

## Next Steps

1. **Confirm this plan** - Review flows and suggest changes
2. **Set up project** - Initialize Go module, create structure
3. **Start Phase 1** - Foundation and database
4. **Iterate** - Build phase by phase with testing

---

## Open Questions / Future Features

- [ ] Should there be a way to "pause" a challenge?
- [ ] Archive completed challenges?
- [ ] Challenge templates (predefined task sets)?
- [ ] Photo proof required for task completion?
- [ ] Leaderboard/ranking by completion speed?
- [ ] Due dates for individual tasks?
