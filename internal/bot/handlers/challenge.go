package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

	// Add daily progress if limit is set
	if challenge.DailyTaskLimit > 0 {
		limitInfo, err := h.completion.CheckDailyLimit(participant, challenge.DailyTaskLimit)
		if err == nil && limitInfo != nil {
			text += fmt.Sprintf("\n📅 Today: %d/%d completed (Resets in %s)\n",
				limitInfo.Completed, limitInfo.Limit, formatDuration(limitInfo.TimeToReset))
		}
	}

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
	tempData["emoji"] = emoji
	h.state.SetStateWithData(userID, domain.StateAwaitingDailyLimit, tempData)

	msg := "⏱ Daily Task Limit\n\nHow many tasks can each participant complete per day?\n\nEnter a number (1-50) or tap Skip for unlimited."
	return c.Send(msg, keyboards.SkipDailyLimit())
}

// processDailyLimit processes daily limit input during creation
func (h *Handler) processDailyLimit(c tele.Context, input string) error {
	userID := c.Sender().ID

	limit, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || limit < 1 || limit > 50 {
		return c.Send("❌ Please enter a number between 1 and 50:", keyboards.SkipDailyLimit())
	}

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	tempData["daily_limit"] = limit
	h.state.SetStateWithData(userID, domain.StateAwaitingCreatorSyncTime, tempData)

	return h.promptSyncTime(c, true)
}

// skipDailyLimit skips the daily limit step (sets to unlimited)
func (h *Handler) skipDailyLimit(c tele.Context) error {
	userID := c.Sender().ID

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	tempData["daily_limit"] = 0
	h.state.SetStateWithData(userID, domain.StateAwaitingCreatorSyncTime, tempData)

	return h.promptSyncTime(c, true)
}

// promptSyncTime shows the time sync prompt
func (h *Handler) promptSyncTime(c tele.Context, isCreator bool) error {
	serverTime := time.Now().UTC().Format("15:04")
	msg := fmt.Sprintf("🕐 Synchronize Your Clock\n\nPlease enter your current time in 24-hour format (HH:MM).\nExample: 14:30 or 09:15\n\nThis helps us track your daily progress correctly.\n\nCurrent server time: %s", serverTime)
	return c.Send(msg, keyboards.SkipSyncTime(isCreator))
}

// parseTimeInput validates HH:MM format and calculates offset from server time
func parseTimeInput(input string) (offsetMinutes int, err error) {
	input = strings.TrimSpace(input)
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid format")
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil || hours < 0 || hours > 23 {
		return 0, fmt.Errorf("invalid hours")
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil || minutes < 0 || minutes > 59 {
		return 0, fmt.Errorf("invalid minutes")
	}

	// Calculate offset: user_time - server_time in minutes
	serverNow := time.Now().UTC()
	serverMinutes := serverNow.Hour()*60 + serverNow.Minute()
	userMinutes := hours*60 + minutes

	offset := userMinutes - serverMinutes

	// Handle day wrap-around (if offset is too large, user might be in different day)
	if offset > 12*60 {
		offset -= 24 * 60
	} else if offset < -12*60 {
		offset += 24 * 60
	}

	return offset, nil
}

// processCreatorSyncTime processes time sync input during challenge creation
func (h *Handler) processCreatorSyncTime(c tele.Context, input string) error {
	offset, err := parseTimeInput(input)
	if err != nil {
		return c.Send("❌ Invalid time format. Please enter time as HH:MM (e.g., 14:30):", keyboards.SkipSyncTime(true))
	}

	return h.finishChallengeCreation(c, offset)
}

// skipCreatorSyncTime skips time sync during challenge creation (uses server time)
func (h *Handler) skipCreatorSyncTime(c tele.Context) error {
	return h.finishChallengeCreation(c, 0)
}

// finishChallengeCreation creates the challenge with all collected data
func (h *Handler) finishChallengeCreation(c tele.Context, timeOffset int) error {
	userID := c.Sender().ID

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)

	challengeName := tempData["challenge_name"].(string)
	challengeDescription := ""
	if desc, ok := tempData["challenge_description"].(string); ok {
		challengeDescription = desc
	}
	displayName := tempData["display_name"].(string)
	emoji := tempData["emoji"].(string)
	dailyLimit := 0
	if limit, ok := tempData["daily_limit"].(int); ok {
		dailyLimit = limit
	}

	// Create challenge
	challenge, err := h.challenge.Create(challengeName, challengeDescription, userID, dailyLimit)
	if err != nil {
		h.state.Reset(userID)
		if err == service.ErrMaxChallengesReached {
			return h.sendError(c, "❌ You've reached the maximum of 10 active challenges.")
		}
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	// Add creator as first participant with time offset
	_, err = h.participant.Join(challenge.ID, userID, displayName, emoji, timeOffset)
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

	dailyLimitText := "unlimited"
	if challenge.DailyTaskLimit > 0 {
		dailyLimitText = fmt.Sprintf("%d", challenge.DailyTaskLimit)
	}

	msg := fmt.Sprintf("Challenge: %s\n", challenge.Name)
	if challenge.Description != "" {
		msg += fmt.Sprintf("%s\n", challenge.Description)
	}
	msg += fmt.Sprintf("\nTasks: %d\nMembers: %d\nTasks per day limit: %s\n\nEnter your display name:", taskCount, participantCount, dailyLimitText)
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

	// Check if emoji is taken before proceeding
	usedEmojis, _ := h.participant.GetUsedEmojis(challengeID)
	for _, e := range usedEmojis {
		if e == emoji {
			return c.Send("❌ This emoji is already taken. Choose another:", keyboards.EmojiSelector(usedEmojis))
		}
	}

	tempData["emoji"] = emoji
	h.state.SetStateWithData(userID, domain.StateAwaitingSyncTime, tempData)

	return h.promptSyncTime(c, false)
}

// processSyncTime processes time sync input during join or settings
func (h *Handler) processSyncTime(c tele.Context, input string) error {
	userID := c.Sender().ID

	offset, err := parseTimeInput(input)
	if err != nil {
		return c.Send("❌ Invalid time format. Please enter time as HH:MM (e.g., 14:30):", keyboards.SkipSyncTime(false))
	}

	// Check if we're in join flow (has temp data with challenge_id) or settings flow
	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)

	if tempData != nil {
		if _, hasChallenge := tempData["challenge_id"]; hasChallenge {
			// Join flow
			return h.finishJoinChallenge(c, offset)
		}
	}

	// Settings flow - update time offset
	return h.processSettingsSyncTime(c, input)
}

// skipSyncTime skips time sync during join or settings
func (h *Handler) skipSyncTime(c tele.Context) error {
	userID := c.Sender().ID

	// Check if we're in join flow or settings flow
	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)

	if tempData != nil {
		if _, hasChallenge := tempData["challenge_id"]; hasChallenge {
			// Join flow
			return h.finishJoinChallenge(c, 0)
		}
	}

	// Settings flow - keep current offset
	return h.skipSettingsSyncTime(c)
}

// finishJoinChallenge joins the challenge with all collected data
func (h *Handler) finishJoinChallenge(c tele.Context, timeOffset int) error {
	userID := c.Sender().ID

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)

	challengeID := tempData["challenge_id"].(string)
	challengeName := tempData["challenge_name"].(string)
	displayName := tempData["display_name"].(string)
	emoji := tempData["emoji"].(string)

	// Join challenge with time offset
	participant, err := h.participant.Join(challengeID, userID, displayName, emoji, timeOffset)
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
