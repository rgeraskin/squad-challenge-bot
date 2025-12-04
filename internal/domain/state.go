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

	// Template-based challenge creation states (User)
	StateSelectTemplateOrScratch         = "select_template_or_scratch"
	StateSelectTemplate                  = "select_template"
	StateViewingTemplate                 = "viewing_template"
	StateAwaitingTemplateChallengeName   = "awaiting_template_challenge_name"
	StateAwaitingTemplateCreatorName     = "awaiting_template_creator_name"
	StateAwaitingTemplateCreatorEmoji    = "awaiting_template_creator_emoji"
	StateAwaitingTemplateCreatorSyncTime = "awaiting_template_creator_sync_time"

	// Template admin/editing states (Super Admin)
	StateAwaitingNewTemplateName        = "awaiting_new_template_name"
	StateAwaitingNewTemplateDescription = "awaiting_new_template_description"
	StateAwaitingNewTemplateDailyLimit  = "awaiting_new_template_daily_limit"
	StateAwaitingTplTaskTitle           = "awaiting_tpl_task_title"
	StateAwaitingTplTaskDescription     = "awaiting_tpl_task_description"
	StateAwaitingTplTaskImage           = "awaiting_tpl_task_image"
	StateAwaitingTplEditTitle       = "awaiting_tpl_edit_title"
	StateAwaitingTplEditDescription = "awaiting_tpl_edit_description"
	StateAwaitingTplEditImage       = "awaiting_tpl_edit_image"
)
