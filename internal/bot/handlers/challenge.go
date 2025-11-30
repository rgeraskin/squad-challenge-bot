package handlers

import (
	"fmt"

	"github.com/rgeraskin/squad-challenge-bot/internal/bot/keyboards"
	"github.com/rgeraskin/squad-challenge-bot/internal/bot/views"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/logger"
	"github.com/rgeraskin/squad-challenge-bot/internal/service"
	tele "gopkg.in/telebot.v3"
)

// showMainChallengeView displays the main challenge view
func (h *Handler) showMainChallengeView(c tele.Context, challengeID string) error {
	userID := c.Sender().ID

	challenge, err := h.challenge.GetByID(challengeID)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	tasks, err := h.task.GetByChallengeID(challengeID)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	participant, err := h.participant.GetByChallengeAndUser(challengeID, userID)
	if err != nil || participant == nil {
		return h.sendError(c, "❌ You're not a participant of this challenge.")
	}

	participants, err := h.participant.GetByChallengeID(challengeID)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	// Get completion data
	completedIDs, _ := h.completion.GetCompletedTaskIDs(participant.ID)
	completedSet := make(map[int64]bool)
	for _, id := range completedIDs {
		completedSet[id] = true
	}

	// Calculate current task for each participant
	participantEmojis := make(map[int64][]string)
	for _, p := range participants {
		currentTaskNum := h.completion.GetCurrentTaskNum(p.ID, tasks)
		if currentTaskNum > 0 {
			// Find task ID for this order number
			for _, t := range tasks {
				if t.OrderNum == currentTaskNum {
					participantEmojis[t.ID] = append(participantEmojis[t.ID], p.Emoji)
					break
				}
			}
		}
	}

	// Build view data
	currentTaskNum := h.completion.GetCurrentTaskNum(participant.ID, tasks)

	data := views.TaskListData{
		ChallengeName:        challenge.Name,
		ChallengeDescription: challenge.Description,
		TotalTasks:           len(tasks),
		CompletedTasks:       len(completedIDs),
		ParticipantCount:     len(participants),
		Tasks:                tasks,
		CompletedTaskIDs:     completedSet,
		ParticipantEmojis:    participantEmojis,
		CurrentUserEmoji:     participant.Emoji,
		CurrentTaskNum:       currentTaskNum,
	}

	text := views.RenderTaskList(data)
	isAdmin := challenge.CreatorID == userID

	// Build task buttons for all tasks
	var taskButtons []keyboards.TaskButton
	for _, task := range tasks {
		taskButtons = append(taskButtons, keyboards.TaskButton{
			ID:          task.ID,
			OrderNum:    task.OrderNum,
			Title:       task.Title,
			IsCompleted: completedSet[task.ID],
			IsCurrent:   task.OrderNum == currentTaskNum,
		})
	}

	kb := keyboards.MainChallengeView(challengeID, currentTaskNum, isAdmin, taskButtons)

	return c.Send(text, kb)
}

// handleCreateChallenge starts the challenge creation flow
func (h *Handler) handleCreateChallenge(c tele.Context) error {
	userID := c.Sender().ID
	logger.Debug("handleCreateChallenge called", "user_id", userID)

	// Check max challenges
	challenges, err := h.challenge.GetByUserID(userID)
	if err != nil {
		logger.Error("Failed to get user challenges in create", "user_id", userID, "error", err)
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}
	logger.Debug("User challenges count", "user_id", userID, "count", len(challenges))
	if len(challenges) >= service.MaxChallengesPerUser {
		logger.Warn("User reached max challenges", "user_id", userID, "count", len(challenges))
		return h.sendError(c, "❌ You've reached the maximum of 10 active challenges.")
	}

	logger.Debug("Setting state to awaiting challenge name", "user_id", userID)
	h.state.SetState(userID, domain.StateAwaitingChallengeName)
	err = c.Send("Enter challenge name:", keyboards.CancelOnly())
	if err != nil {
		logger.Error("Failed to send challenge name prompt", "user_id", userID, "error", err)
	} else {
		logger.Debug("Challenge name prompt sent", "user_id", userID)
	}
	return err
}

// handleJoinChallenge starts the join challenge flow
func (h *Handler) handleJoinChallenge(c tele.Context) error {
	userID := c.Sender().ID
	logger.Debug("handleJoinChallenge called", "user_id", userID)

	logger.Debug("Setting state to awaiting challenge ID", "user_id", userID)
	h.state.SetState(userID, domain.StateAwaitingChallengeID)
	err := c.Send("Enter the Challenge ID:", keyboards.CancelOnly())
	if err != nil {
		logger.Error("Failed to send challenge ID prompt", "user_id", userID, "error", err)
	} else {
		logger.Debug("Challenge ID prompt sent", "user_id", userID)
	}
	return err
}

// processChallengeName processes challenge name input during creation
func (h *Handler) processChallengeName(c tele.Context, name string) error {
	userID := c.Sender().ID

	if len(name) == 0 || len(name) > 50 {
		return c.Send("❌ Challenge name must be 1-50 characters. Try again:", keyboards.CancelOnly())
	}

	tempData := map[string]interface{}{
		"challenge_name": name,
	}
	h.state.SetStateWithData(userID, domain.StateAwaitingChallengeDescription, tempData)

	return c.Send("Enter challenge description (or click Skip):", keyboards.SkipCancel())
}

// processChallengeDescription processes challenge description input during creation
func (h *Handler) processChallengeDescription(c tele.Context, description string) error {
	userID := c.Sender().ID

	if len(description) > 500 {
		return c.Send("❌ Description must be 500 characters or less. Try again:", keyboards.SkipCancel())
	}

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	tempData["challenge_description"] = description
	h.state.SetStateWithData(userID, domain.StateAwaitingCreatorName, tempData)

	return c.Send("Enter your display name:", keyboards.CancelOnly())
}

// skipChallengeDescription skips the description step
func (h *Handler) skipChallengeDescription(c tele.Context) error {
	userID := c.Sender().ID

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	tempData["challenge_description"] = ""
	h.state.SetStateWithData(userID, domain.StateAwaitingCreatorName, tempData)

	return c.Send("Enter your display name:", keyboards.CancelOnly())
}

// processCreatorName processes creator name input during creation
func (h *Handler) processCreatorName(c tele.Context, name string) error {
	userID := c.Sender().ID

	if len(name) == 0 || len(name) > 30 {
		return c.Send("❌ Display name must be 1-30 characters. Try again:", keyboards.CancelOnly())
	}

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	tempData["display_name"] = name
	h.state.SetStateWithData(userID, domain.StateAwaitingCreatorEmoji, tempData)

	return c.Send("Choose your emoji or send your own:", keyboards.EmojiSelector(nil))
}

// processCreatorEmoji processes creator emoji input during creation
func (h *Handler) processCreatorEmoji(c tele.Context, emoji string) error {
	userID := c.Sender().ID

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)

	challengeName := tempData["challenge_name"].(string)
	challengeDescription := ""
	if desc, ok := tempData["challenge_description"].(string); ok {
		challengeDescription = desc
	}
	displayName := tempData["display_name"].(string)

	// Create challenge
	challenge, err := h.challenge.Create(challengeName, challengeDescription, userID)
	if err != nil {
		h.state.Reset(userID)
		if err == service.ErrMaxChallengesReached {
			return h.sendError(c, "❌ You've reached the maximum of 10 active challenges.")
		}
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	// Add creator as first participant
	_, err = h.participant.Join(challenge.ID, userID, displayName, emoji)
	if err != nil {
		h.state.Reset(userID)
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	// Set current challenge and reset state
	h.state.SetCurrentChallenge(userID, challenge.ID)
	h.state.ResetKeepChallenge(userID)

	msg := fmt.Sprintf("✅ Challenge \"%s\" created!\n\nYou are the admin of this challenge.\nNow add tasks to your challenge.", challengeName)
	c.Send(msg)

	// Show admin panel
	return h.showAdminPanel(c, challenge.ID)
}

// processChallengeID processes challenge ID input during join
func (h *Handler) processChallengeID(c tele.Context, id string) error {
	userID := c.Sender().ID

	// Validate ID format
	if len(id) != 8 {
		return c.Send("❌ Challenge not found. Check the ID and try again.", keyboards.CancelOnly())
	}

	// Check if challenge exists
	challenge, err := h.challenge.GetByID(id)
	if err != nil {
		if err == service.ErrChallengeNotFound {
			return c.Send("❌ Challenge not found. Check the ID and try again.", keyboards.CancelOnly())
		}
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	// Check if can join
	if err := h.challenge.CanJoin(id, userID); err != nil {
		switch err {
		case service.ErrChallengeFull:
			h.state.Reset(userID)
			return h.sendError(c, "❌ This challenge is full (10/10 participants).")
		case service.ErrAlreadyMember:
			h.state.Reset(userID)
			return c.Send("ℹ️ You're already participating in this challenge.")
		default:
			return h.sendError(c, "⚠️ Something went wrong. Please try again.")
		}
	}

	taskCount, _ := h.task.CountByChallengeID(id)
	participantCount, _ := h.participant.CountByChallengeID(id)

	tempData := map[string]interface{}{
		"challenge_id":   id,
		"challenge_name": challenge.Name,
	}
	h.state.SetStateWithData(userID, domain.StateAwaitingParticipantName, tempData)

	msg := fmt.Sprintf("Challenge: %s\n", challenge.Name)
	if challenge.Description != "" {
		msg += fmt.Sprintf("%s\n", challenge.Description)
	}
	msg += fmt.Sprintf("(%d tasks, %d members)\n\nEnter your display name:", taskCount, participantCount)
	return c.Send(msg, keyboards.CancelOnly())
}

// processParticipantName processes participant name input during join
func (h *Handler) processParticipantName(c tele.Context, name string) error {
	userID := c.Sender().ID

	if len(name) == 0 || len(name) > 30 {
		return c.Send("❌ Display name must be 1-30 characters. Try again:", keyboards.CancelOnly())
	}

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	tempData["display_name"] = name
	h.state.SetStateWithData(userID, domain.StateAwaitingParticipantEmoji, tempData)

	// Get used emojis
	challengeID := tempData["challenge_id"].(string)
	usedEmojis, _ := h.participant.GetUsedEmojis(challengeID)

	return c.Send("Choose your emoji or send your own:", keyboards.EmojiSelector(usedEmojis))
}

// processParticipantEmoji processes participant emoji input during join
func (h *Handler) processParticipantEmoji(c tele.Context, emoji string) error {
	userID := c.Sender().ID

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)

	challengeID := tempData["challenge_id"].(string)
	challengeName := tempData["challenge_name"].(string)
	displayName := tempData["display_name"].(string)

	// Join challenge
	participant, err := h.participant.Join(challengeID, userID, displayName, emoji)
	if err != nil {
		if err == service.ErrEmojiTaken {
			usedEmojis, _ := h.participant.GetUsedEmojis(challengeID)
			return c.Send("❌ This emoji is already taken. Choose another:", keyboards.EmojiSelector(usedEmojis))
		}
		h.state.Reset(userID)
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	// Set current challenge and reset state
	h.state.SetCurrentChallenge(userID, challengeID)
	h.state.ResetKeepChallenge(userID)

	// Notify others
	go h.notification.NotifyJoin(challengeID, participant.Emoji, participant.DisplayName, userID)

	msg := fmt.Sprintf("\n🎯 CHALLENGE ACCEPTED! 🎯\n\nWelcome to \"%s\", %s!", challengeName, displayName)
	return c.Send(msg, keyboards.JoinWelcome(challengeID))
}
