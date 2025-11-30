package handlers

import (
	"strings"

	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/logger"
	tele "gopkg.in/telebot.v3"
)

// HandleCallback handles inline button callbacks
func (h *Handler) HandleCallback(c tele.Context) error {
	userID := c.Sender().ID
	callback := c.Callback()
	if callback == nil {
		logger.Error("HandleCallback called with nil callback", "user_id", userID)
		return nil
	}
	data := callback.Data
	logger.Info("HandleCallback received", "user_id", userID, "data", data, "unique", callback.Unique)

	// Acknowledge callback to stop loading spinner
	if err := c.Respond(); err != nil {
		logger.Warn("Failed to respond to callback", "user_id", userID, "error", err)
	}

	// Parse callback data: action|param1|param2
	// telebot v3 prefixes callback data with \f (form feed) character
	// and uses | as separator between unique identifier and data
	data = strings.TrimPrefix(data, "\f")
	parts := strings.Split(data, "|")
	action := parts[0]
	logger.Debug("Parsed callback action", "user_id", userID, "action", action, "parts", parts)

	// Get user state (needed for state-dependent callbacks)
	userState, err := h.state.Get(userID)
	if err != nil {
		logger.Warn("Failed to get user state in callback", "user_id", userID, "error", err)
	}
	logger.Debug("User state", "user_id", userID, "state", userState.State, "current_challenge", userState.CurrentChallenge)

	// State-dependent callbacks that should NOT reset state
	stateDependentActions := map[string]bool{
		"select_emoji":           true,
		"skip":                   true,
		"skip_daily_limit":       true,
		"skip_creator_sync_time": true,
		"skip_sync_time":         true,
		"hide_future_yes":        true,
		"hide_future_no":         true,
		"cancel":                 true,
	}

	// Reset state for non-state-dependent callbacks (user clicked a different button)
	if !stateDependentActions[action] && userState.State != domain.StateIdle {
		logger.Debug("Resetting non-idle state on callback", "user_id", userID, "state", userState.State)
		h.state.ResetKeepChallenge(userID)
	}

	// Handle admin-protected actions
	adminActions := map[string]bool{
		"admin_panel":                true,
		"add_task":                   true,
		"edit_tasks":                 true,
		"edit_task":                  true,
		"edit_task_title":            true,
		"edit_task_description":      true,
		"edit_task_image":            true,
		"delete_task":                true,
		"confirm_delete_task":        true,
		"reorder_tasks":              true,
		"reorder_select":             true,
		"reorder_move":               true,
		"randomize_tasks":            true,
		"edit_challenge_name":        true,
		"edit_challenge_description": true,
		"edit_daily_limit":           true,
		"toggle_hide_future":         true,
		"delete_challenge":           true,
		"confirm_delete_challenge":   true,
	}

	if adminActions[action] {
		isAdmin, _ := h.checkAdminAccess(c)
		if !isAdmin {
			return h.sendError(c, "⚠️ You don't have permission to perform this action.")
		}
	}

	logger.Debug("Switching on callback action", "user_id", userID, "action", action)
	switch action {
	// Start menu actions
	case "create_challenge":
		logger.Info("create_challenge callback triggered", "user_id", userID)
		err := h.handleCreateChallenge(c)
		if err != nil {
			logger.Error("handleCreateChallenge failed", "user_id", userID, "error", err)
		}
		return err
	case "join_challenge":
		logger.Info("join_challenge callback triggered", "user_id", userID)
		err := h.handleJoinChallenge(c)
		if err != nil {
			logger.Error("handleJoinChallenge failed", "user_id", userID, "error", err)
		}
		return err
	case "open_challenge":
		logger.Debug("open_challenge callback", "user_id", userID, "parts", parts)
		if len(parts) > 1 {
			challengeID := parts[1]
			logger.Debug("Opening challenge", "user_id", userID, "challenge_id", challengeID)
			h.state.SetCurrentChallenge(userID, challengeID)
			return h.showMainChallengeView(c, challengeID)
		}

	// Navigation
	case "back_to_main":
		userState, _ := h.state.Get(userID)
		return h.showMainChallengeView(c, userState.CurrentChallenge)
	case "back_to_admin":
		userState, _ := h.state.Get(userID)
		return h.showAdminPanel(c, userState.CurrentChallenge)
	case "back_to_tasks":
		return h.handleEditTasks(c)
	case "exit_challenge":
		h.state.Reset(userID)
		return h.showStartMenu(c)

	// Main view actions
	case "complete_current":
		return h.handleCompleteCurrent(c)
	case "team_progress":
		return h.showTeamProgress(c)
	case "list_all_tasks":
		return h.showAllTasks(c)
	case "share_id":
		return h.handleShareID(c)
	case "settings":
		return h.showSettings(c)
	case "admin_panel":
		userState, _ := h.state.Get(userID)
		return h.showAdminPanel(c, userState.CurrentChallenge)

	// Task detail actions
	case "task_detail":
		if len(parts) > 1 {
			return h.showTaskDetail(c, parts[1])
		}
	case "complete_task":
		if len(parts) > 1 {
			return h.handleCompleteTask(c, parts[1])
		}
	case "uncomplete_task":
		if len(parts) > 1 {
			return h.handleUncompleteTask(c, parts[1])
		}

	// Admin panel actions
	case "add_task":
		return h.handleAddTask(c)
	case "edit_tasks":
		return h.handleEditTasks(c)
	case "edit_task":
		if len(parts) > 1 {
			return h.handleEditTask(c, parts[1])
		}
	case "edit_task_title":
		if len(parts) > 1 {
			return h.handleEditTaskTitle(c, parts[1])
		}
	case "edit_task_description":
		if len(parts) > 1 {
			return h.handleEditTaskDescription(c, parts[1])
		}
	case "edit_task_image":
		if len(parts) > 1 {
			return h.handleEditTaskImage(c, parts[1])
		}
	case "delete_task":
		if len(parts) > 1 {
			return h.handleDeleteTask(c, parts[1])
		}
	case "confirm_delete_task":
		if len(parts) > 1 {
			return h.handleConfirmDeleteTask(c, parts[1])
		}
	case "cancel_delete_task":
		return h.handleEditTasks(c)
	case "edit_challenge_name":
		return h.handleEditChallengeName(c)
	case "edit_challenge_description":
		return h.handleEditChallengeDescription(c)
	case "edit_daily_limit":
		return h.handleEditDailyLimit(c)
	case "toggle_hide_future":
		return h.handleToggleHideFutureTasks(c)
	case "delete_challenge":
		return h.handleDeleteChallenge(c)
	case "confirm_delete_challenge":
		return h.handleConfirmDeleteChallenge(c)
	case "cancel_delete_challenge":
		userState, _ := h.state.Get(userID)
		return h.showAdminPanel(c, userState.CurrentChallenge)

	// Reorder actions
	case "reorder_tasks":
		return h.handleReorderTasks(c)
	case "reorder_select":
		if len(parts) > 1 {
			return h.handleReorderSelect(c, parts[1])
		}
	case "reorder_move":
		if len(parts) > 2 {
			return h.handleReorderMove(c, parts[1], parts[2])
		}
	case "reorder_cancel":
		return h.handleEditTasks(c)
	case "randomize_tasks":
		return h.handleRandomizeTasks(c)

	// Settings actions
	case "toggle_notifications":
		return h.handleToggleNotifications(c)
	case "change_name":
		return h.handleChangeName(c)
	case "change_emoji":
		return h.handleChangeEmoji(c)
	case "sync_time":
		return h.handleSyncTime(c)
	case "leave_challenge":
		return h.handleLeaveChallenge(c)
	case "confirm_leave":
		return h.handleConfirmLeave(c)
	case "cancel_leave":
		return h.showSettings(c)

	// Join flow
	case "start_challenge":
		if len(parts) > 1 {
			challengeID := parts[1]
			h.state.SetCurrentChallenge(userID, challengeID)
			return h.showMainChallengeView(c, challengeID)
		}

	// Emoji selection
	case "select_emoji":
		if len(parts) > 1 {
			return h.handleEmojiSelection(c, parts[1])
		}

	// Cancel and skip
	case "cancel":
		return h.handleCancel(c)
	case "skip":
		return h.handleSkip(c)
	case "skip_daily_limit":
		return h.skipDailyLimit(c)
	case "hide_future_yes":
		return h.processHideFutureTasks(c, true)
	case "hide_future_no":
		return h.processHideFutureTasks(c, false)
	case "skip_creator_sync_time":
		return h.skipCreatorSyncTime(c)
	case "skip_sync_time":
		return h.skipSyncTime(c)

	// No-op (disabled buttons)
	case "noop":
		logger.Debug("noop callback", "user_id", userID)
		return nil
	default:
		logger.Warn("Unhandled callback action", "user_id", userID, "action", action, "data", data)
	}

	logger.Debug("HandleCallback completed without explicit return", "user_id", userID, "action", action)
	return nil
}

// handleEmojiSelection handles emoji selection from keyboard
func (h *Handler) handleEmojiSelection(c tele.Context, emoji string) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)

	switch userState.State {
	case domain.StateAwaitingCreatorEmoji:
		return h.processCreatorEmoji(c, emoji)
	case domain.StateAwaitingParticipantEmoji:
		return h.processParticipantEmoji(c, emoji)
	case domain.StateAwaitingNewEmoji:
		return h.processNewEmoji(c, emoji)
	}

	return nil
}

// handleCancel cancels the current flow
func (h *Handler) handleCancel(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)

	// Check if we're in join flow (has temp data with challenge_id but not yet a participant)
	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	if tempData != nil {
		if _, hasChallenge := tempData["challenge_id"]; hasChallenge {
			// In join flow - just go back to start menu
			h.state.Reset(userID)
			return h.showStartMenu(c)
		}
	}

	// If we have a current challenge, go back to it
	if userState.CurrentChallenge != "" {
		// Verify user is actually a participant of this challenge
		participant, _ := h.participant.GetByChallengeAndUser(userState.CurrentChallenge, userID)
		if participant == nil {
			// Not a participant - clear stale state and go to start menu
			h.state.Reset(userID)
			return h.showStartMenu(c)
		}

		h.state.ResetKeepChallenge(userID)

		// Check if we were in admin flow
		switch userState.State {
		case domain.StateAwaitingTaskTitle,
			domain.StateAwaitingTaskImage,
			domain.StateAwaitingTaskDescription,
			domain.StateAwaitingEditTitle,
			domain.StateAwaitingEditDescription,
			domain.StateAwaitingEditImage,
			domain.StateReorderSelectTask,
			domain.StateReorderSelectPosition,
			domain.StateAwaitingNewChallengeName,
			domain.StateAwaitingNewDailyLimit:
			return h.showAdminPanel(c, userState.CurrentChallenge)
		case domain.StateAwaitingNewName,
			domain.StateAwaitingNewEmoji,
			domain.StateAwaitingSyncTime:
			return h.showSettings(c)
		default:
			return h.showMainChallengeView(c, userState.CurrentChallenge)
		}
	}

	h.state.Reset(userID)
	return h.showStartMenu(c)
}

// handleSkip handles skip actions
func (h *Handler) handleSkip(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)

	switch userState.State {
	case domain.StateAwaitingChallengeDescription:
		return h.skipChallengeDescription(c)
	case domain.StateAwaitingTaskImage:
		return h.skipTaskImage(c)
	case domain.StateAwaitingTaskDescription:
		return h.skipTaskDescription(c)
	}

	return nil
}
