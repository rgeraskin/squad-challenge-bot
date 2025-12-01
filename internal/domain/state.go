package domain

import "time"

// UserState tracks the conversation state for a user
type UserState struct {
	TelegramID       int64     `db:"telegram_id"`
	State            string    `db:"state"`
	TempData         string    `db:"temp_data"`
	CurrentChallenge string    `db:"current_challenge"`
	UpdatedAt        time.Time `db:"updated_at"`
}

// User states for conversation flow
const (
	StateIdle = "idle"

	// Challenge creation
	StateAwaitingChallengeName        = "awaiting_challenge_name"
	StateAwaitingChallengeDescription = "awaiting_challenge_description"
	StateAwaitingCreatorName          = "awaiting_creator_name"
	StateAwaitingCreatorEmoji         = "awaiting_creator_emoji"
	StateAwaitingDailyLimit           = "awaiting_daily_limit"
	StateAwaitingHideFutureTasks      = "awaiting_hide_future_tasks"
	StateAwaitingCreatorSyncTime      = "awaiting_creator_sync_time"

	// Task management
	StateAwaitingTaskTitle       = "awaiting_task_title"
	StateAwaitingTaskImage       = "awaiting_task_image"
	StateAwaitingTaskDescription = "awaiting_task_description"
	StateAwaitingEditTitle       = "awaiting_edit_title"
	StateAwaitingEditDescription = "awaiting_edit_description"
	StateAwaitingEditImage       = "awaiting_edit_image"
	StateReorderSelectTask       = "reorder_select_task"
	StateReorderSelectPosition   = "reorder_select_position"

	// Joining challenge
	StateAwaitingChallengeID      = "awaiting_challenge_id"
	StateAwaitingParticipantName  = "awaiting_participant_name"
	StateAwaitingParticipantEmoji = "awaiting_participant_emoji"
	StateAwaitingSyncTime         = "awaiting_sync_time"

	// Admin
	StateAwaitingNewChallengeName        = "awaiting_new_challenge_name"
	StateAwaitingNewChallengeDescription = "awaiting_new_challenge_description"
	StateAwaitingNewDailyLimit           = "awaiting_new_daily_limit"

	// User settings
	StateAwaitingNewName  = "awaiting_new_name"
	StateAwaitingNewEmoji = "awaiting_new_emoji"

	// Super Admin
	StateAwaitingSuperAdminID = "awaiting_super_admin_id"
)
