package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/rgeraskin/squad-challenge-bot/internal/bot/keyboards"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	tele "gopkg.in/telebot.v3"
)

// showAdminPanel shows the admin panel
func (h *Handler) showAdminPanel(c tele.Context, challengeID string) error {
	challenge, err := h.challenge.GetByID(challengeID)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	taskCount, _ := h.task.CountByChallengeID(challengeID)
	participantCount, _ := h.participant.CountByChallengeID(challengeID)

	msg := fmt.Sprintf("🔧 Admin Panel - %s\n\n", challenge.Name)
	if challenge.Description != "" {
		msg += fmt.Sprintf("%s\n\n", challenge.Description)
	}
	msg += fmt.Sprintf("Challenge ID: %s\n", challenge.ID)
	msg += fmt.Sprintf("Participants: %d/10\n", participantCount)
	msg += fmt.Sprintf("Tasks: %d\n", taskCount)
	if challenge.DailyTaskLimit > 0 {
		msg += fmt.Sprintf("Daily limit: %d tasks/day\n", challenge.DailyTaskLimit)
	} else {
		msg += "Daily limit: unlimited\n"
	}
	if challenge.HideFutureTasks {
		msg += "Task visibility: Sequential\n"
	} else {
		msg += "Task visibility: All visible\n"
	}

	return c.Send(msg, keyboards.AdminPanel(challenge.DailyTaskLimit, challenge.HideFutureTasks))
}

// handleEditChallengeName starts editing challenge name
func (h *Handler) handleEditChallengeName(c tele.Context) error {
	userID := c.Sender().ID
	h.state.SetState(userID, domain.StateAwaitingNewChallengeName)
	return c.Send("Enter new challenge name:", keyboards.CancelOnly())
}

// processNewChallengeName processes new challenge name
func (h *Handler) processNewChallengeName(c tele.Context, name string) error {
	userID := c.Sender().ID

	if len(name) == 0 || len(name) > 50 {
		return c.Send("❌ Challenge name must be 1-50 characters. Try again:", keyboards.CancelOnly())
	}

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	if err := h.challenge.UpdateName(challengeID, name, userID); err != nil {
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	h.state.ResetKeepChallenge(userID)
	c.Send(fmt.Sprintf("✅ Challenge renamed to \"%s\"", name))
	return h.showAdminPanel(c, challengeID)
}

// handleEditChallengeDescription starts editing challenge description
func (h *Handler) handleEditChallengeDescription(c tele.Context) error {
	userID := c.Sender().ID
	h.state.SetState(userID, domain.StateAwaitingNewChallengeDescription)
	return c.Send("Enter new challenge description (or send empty message to clear):", keyboards.CancelOnly())
}

// processNewChallengeDescription processes new challenge description
func (h *Handler) processNewChallengeDescription(c tele.Context, description string) error {
	userID := c.Sender().ID

	if len(description) > 500 {
		return c.Send("❌ Description must be 500 characters or less. Try again:", keyboards.CancelOnly())
	}

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	if err := h.challenge.UpdateDescription(challengeID, description, userID); err != nil {
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	h.state.ResetKeepChallenge(userID)
	if description == "" {
		c.Send("✅ Challenge description cleared")
	} else {
		c.Send("✅ Challenge description updated")
	}
	return h.showAdminPanel(c, challengeID)
}

// handleEditDailyLimit starts editing daily limit
func (h *Handler) handleEditDailyLimit(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	challenge, _ := h.challenge.GetByID(challengeID)

	var currentLimit string
	if challenge.DailyTaskLimit > 0 {
		currentLimit = fmt.Sprintf("%d tasks/day", challenge.DailyTaskLimit)
	} else {
		currentLimit = "unlimited"
	}

	h.state.SetState(userID, domain.StateAwaitingNewDailyLimit)
	msg := fmt.Sprintf("⏱ Edit Daily Limit\n\nCurrent limit: %s\n\nEnter a new limit (1-50) or 0 for unlimited:", currentLimit)
	return c.Send(msg, keyboards.CancelOnly())
}

// processNewDailyLimit processes new daily limit
func (h *Handler) processNewDailyLimit(c tele.Context, input string) error {
	userID := c.Sender().ID

	limit, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || limit < 0 || limit > 50 {
		return c.Send("❌ Please enter a number between 0 and 50 (0 = unlimited):", keyboards.CancelOnly())
	}

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	if err := h.challenge.UpdateDailyLimit(challengeID, limit, userID); err != nil {
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	h.state.ResetKeepChallenge(userID)
	if limit > 0 {
		c.Send(fmt.Sprintf("✅ Daily limit set to %d tasks/day", limit))
	} else {
		c.Send("✅ Daily limit removed (unlimited)")
	}
	return h.showAdminPanel(c, challengeID)
}

// handleToggleHideFutureTasks toggles the hide future tasks setting
func (h *Handler) handleToggleHideFutureTasks(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	newValue, err := h.challenge.ToggleHideFutureTasks(challengeID, userID)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	if newValue {
		c.Send("✅ Task visibility set to Sequential (future tasks hidden)")
	} else {
		c.Send("✅ Task visibility set to All Visible")
	}
	return h.showAdminPanel(c, challengeID)
}

// handleDeleteChallenge shows delete challenge confirmation
func (h *Handler) handleDeleteChallenge(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	challenge, _ := h.challenge.GetByID(challengeID)
	taskCount, _ := h.task.CountByChallengeID(challengeID)
	participantCount, _ := h.participant.CountByChallengeID(challengeID)

	msg := "⚠️ DELETE CHALLENGE?\n\n"
	msg += fmt.Sprintf("\"%s\" will be permanently deleted.\n\n", challenge.Name)
	msg += "This will remove:\n"
	msg += fmt.Sprintf("• All %d tasks\n", taskCount)
	msg += fmt.Sprintf("• All %d participants\n", participantCount)
	msg += "• All progress data\n\n"
	msg += "This action cannot be undone!"

	return c.Send(msg, keyboards.DeleteChallengeConfirm())
}

// handleConfirmDeleteChallenge confirms challenge deletion
func (h *Handler) handleConfirmDeleteChallenge(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	challenge, _ := h.challenge.GetByID(challengeID)

	// Notify participants before deletion
	go h.notification.NotifyChallengeDeleted(challengeID, challenge.Name, userID)

	if err := h.challenge.Delete(challengeID, userID); err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	h.state.Reset(userID)
	c.Send("✅ Challenge deleted.")
	return h.showStartMenu(c)
}

// checkAdminAccess checks if user is admin for current challenge
func (h *Handler) checkAdminAccess(c tele.Context) (bool, error) {
	userID := c.Sender().ID

	userState, err := h.state.Get(userID)
	if err != nil {
		return false, err
	}

	if userState.CurrentChallenge == "" {
		return false, nil
	}

	isAdmin, err := h.challenge.IsAdmin(userState.CurrentChallenge, userID)
	if err != nil {
		return false, err
	}

	return isAdmin, nil
}

// handleShareID shows the share ID view with copy-to-clipboard buttons
func (h *Handler) handleShareID(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	botUsername := h.bot.Me.Username

	msg := "📋 Share Challenge\n\n"
	msg += fmt.Sprintf("Challenge ID: `%s`\n\n", challengeID)
	msg += "Or share this link:\n"
	msg += fmt.Sprintf("`t.me/%s?start=%s`", botUsername, challengeID)

	kb := keyboards.ShareID(challengeID, botUsername)
	kbJSON, _ := json.Marshal(kb)

	// Use raw API call since telebot doesn't support copy_text buttons natively
	params := map[string]string{
		"chat_id":      fmt.Sprintf("%d", c.Chat().ID),
		"text":         msg,
		"parse_mode":   "Markdown",
		"reply_markup": string(kbJSON),
	}
	_, err := h.bot.Raw("sendMessage", params)
	return err
}
