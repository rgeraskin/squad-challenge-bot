package handlers

import (
	"fmt"

	"github.com/rgeraskin/squad-challenge-bot/internal/bot/keyboards"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/logger"
	"github.com/rgeraskin/squad-challenge-bot/internal/service"
	tele "gopkg.in/telebot.v3"
)

// HandleStart handles the /start command
func (h *Handler) HandleStart(c tele.Context) error {
	userID := c.Sender().ID
	logger.Debug("HandleStart called", "user_id", userID, "username", c.Sender().Username)

	// Reset state on /start
	h.state.Reset(userID)
	logger.Debug("State reset for user", "user_id", userID)

	// Check for deep link parameter
	payload := c.Message().Payload
	logger.Debug("Start payload", "user_id", userID, "payload", payload)
	if payload != "" {
		// Deep link: t.me/bot?start=CHALLENGE_ID
		logger.Info("Deep link detected", "user_id", userID, "payload", payload)
		return h.handleDeepLink(c, payload)
	}

	logger.Debug("Showing start menu", "user_id", userID)
	return h.showStartMenu(c)
}

// handleDeepLink handles deep link joins
func (h *Handler) handleDeepLink(c tele.Context, challengeID string) error {
	userID := c.Sender().ID

	// Check if challenge exists
	challenge, err := h.challenge.GetByID(challengeID)
	if err != nil {
		if err == service.ErrChallengeNotFound {
			return h.sendError(c, "❌ Challenge not found. Check the ID and try again.")
		}
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	// Check if user is already a member
	participant, err := h.participant.GetByChallengeAndUser(challengeID, userID)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	if participant != nil {
		// Already a member, go to main view
		h.state.SetCurrentChallenge(userID, challengeID)
		return h.showMainChallengeView(c, challengeID)
	}

	// Check if can join
	if err := h.challenge.CanJoin(challengeID, userID); err != nil {
		switch err {
		case service.ErrChallengeFull:
			return h.sendError(c, "❌ This challenge is full (10/10 participants).")
		case service.ErrAlreadyMember:
			h.state.SetCurrentChallenge(userID, challengeID)
			return h.showMainChallengeView(c, challengeID)
		default:
			return h.sendError(c, "⚠️ Something went wrong. Please try again.")
		}
	}

	// Get info for the join flow
	taskCount, _ := h.task.CountByChallengeID(challengeID)
	participantCount, _ := h.participant.CountByChallengeID(challengeID)

	// Start join flow (skip challenge ID input)
	tempData := map[string]interface{}{
		"challenge_id":   challengeID,
		"challenge_name": challenge.Name,
	}
	h.state.SetStateWithData(userID, domain.StateAwaitingParticipantName, tempData)

	dailyLimitText := "unlimited"
	if challenge.DailyTaskLimit > 0 {
		dailyLimitText = fmt.Sprintf("%d", challenge.DailyTaskLimit)
	}

	msg := fmt.Sprintf("Challenge: %s\n\nTasks: %d\nMembers: %d\nTasks per day limit: %s\n\nEnter your display name:", challenge.Name, taskCount, participantCount, dailyLimitText)

	return c.Send(msg, keyboards.CancelOnly())
}

// showStartMenu shows the start menu with user's challenges
func (h *Handler) showStartMenu(c tele.Context) error {
	userID := c.Sender().ID
	logger.Debug("showStartMenu called", "user_id", userID)

	challenges, err := h.challenge.GetByUserID(userID)
	if err != nil {
		logger.Error("Failed to get user challenges", "user_id", userID, "error", err)
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}
	logger.Debug("User challenges loaded", "user_id", userID, "count", len(challenges))

	// Get task counts and completion counts for each challenge
	taskCounts := make(map[string]int)
	completedCounts := make(map[string]int)

	for _, ch := range challenges {
		count, _ := h.task.CountByChallengeID(ch.ID)
		taskCounts[ch.ID] = count

		participant, _ := h.participant.GetByChallengeAndUser(ch.ID, userID)
		if participant != nil {
			completed, _ := h.completion.CountByParticipantID(participant.ID)
			completedCounts[ch.ID] = completed
		}
	}

	text := "Welcome to SquadChallengeBot!\n\n"
	if len(challenges) > 0 {
		text += "Your challenges:"
	} else {
		text += "You don't have any challenges yet.\nCreate one or join an existing challenge!"
	}

	kb := keyboards.StartMenu(challenges, taskCounts, completedCounts)
	logger.Debug("Sending start menu", "user_id", userID, "challenges_count", len(challenges), "has_keyboard", kb != nil)
	err = c.Send(text, kb)
	if err != nil {
		logger.Error("Failed to send start menu", "user_id", userID, "error", err)
	} else {
		logger.Debug("Start menu sent successfully", "user_id", userID)
	}
	return err
}
