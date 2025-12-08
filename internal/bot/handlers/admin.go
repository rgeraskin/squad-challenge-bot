package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/rgeraskin/squad-challenge-bot/internal/bot/keyboards"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/logger"
	tele "gopkg.in/telebot.v3"
)

// showAdminPanel shows the admin panel
func (h *Handler) showAdminPanel(c tele.Context, challengeID string) error {
	userID := c.Sender().ID

	challenge, err := h.challenge.GetByID(challengeID)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	// Check if we're in observer mode
	isObserverMode := h.isInObserverMode(userID)

	taskCount, _ := h.task.CountByChallengeID(challengeID)
	participantCount, _ := h.participant.CountByChallengeID(challengeID)

	msg := "🔧 <i>Admin Panel</i>\n\n"
	if isObserverMode {
		msg = "🔧 <i>Admin Panel (Super Admin)</i>\n\n"
	}
	msg += fmt.Sprintf("<b>Challenge:</b> %s\n", challenge.Name)
	msg += fmt.Sprintf("<b>Description:</b> %s\n", challenge.Description)
	msg += fmt.Sprintf("<b>Challenge ID:</b> <code>%s</code>\n", challenge.ID)
	msg += fmt.Sprintf("<b>Members:</b> %d/50\n", participantCount)
	msg += fmt.Sprintf("<b>Tasks:</b> %d\n", taskCount)
	if challenge.DailyTaskLimit > 0 {
		msg += fmt.Sprintf("<b>Daily Limit:</b> %d/day\n", challenge.DailyTaskLimit)
	} else {
		msg += "<b>Daily Limit:</b> No daily limit\n"
	}
	if challenge.HideFutureTasks {
		msg += "<b>Mode:</b> Sequential\n"
	} else {
		msg += "<b>Mode:</b> All Visible\n"
	}

	return c.Send(
		msg,
		keyboards.AdminPanel(challenge.DailyTaskLimit, challenge.HideFutureTasks, isObserverMode),
		tele.ModeHTML,
	)
}

// handleEditChallengeName starts editing challenge name
func (h *Handler) handleEditChallengeName(c tele.Context) error {
	userID := c.Sender().ID
	h.state.SetState(userID, domain.StateAwaitingNewChallengeName)
	return c.Send("✏️ What's the new challenge name?", keyboards.CancelOnly())
}

// processNewChallengeName processes new challenge name
func (h *Handler) processNewChallengeName(c tele.Context, name string) error {
	userID := c.Sender().ID

	if len(name) == 0 || len(name) > 50 {
		return c.Send("😅 Keep it between 1-50 characters. Try again:", keyboards.CancelOnly())
	}

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	isSuperAdmin := h.isSuperAdmin(userID)
	if err := h.challenge.UpdateName(challengeID, name, userID, isSuperAdmin); err != nil {
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	h.state.ResetKeepChallenge(userID)
	c.Send(fmt.Sprintf("✅ Done! Challenge is now \"%s\"", name))
	return h.showAdminPanel(c, challengeID)
}

// handleEditChallengeDescription starts editing challenge description
func (h *Handler) handleEditChallengeDescription(c tele.Context) error {
	userID := c.Sender().ID
	h.state.SetState(userID, domain.StateAwaitingNewChallengeDescription)
	return c.Send(
		"📝 What's the new description?",
		keyboards.CancelOnly(),
	)
}

// processNewChallengeDescription processes new challenge description
func (h *Handler) processNewChallengeDescription(c tele.Context, description string) error {
	userID := c.Sender().ID

	if len(description) > 500 {
		return c.Send("😅 That's a bit long! Keep it under 500 characters:", keyboards.CancelOnly())
	}

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	isSuperAdmin := h.isSuperAdmin(userID)
	if err := h.challenge.UpdateDescription(challengeID, description, userID, isSuperAdmin); err != nil {
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	h.state.ResetKeepChallenge(userID)
	if description == "" {
		c.Send("✅ Description cleared!")
	} else {
		c.Send("✅ Description updated!")
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
	msg := "🕓 <i>Daily Limit</i>\n\n"
	msg += fmt.Sprintf("Right now: <b>%s</b>\n\n", currentLimit)
	msg += "Pick a number (1-50) or 0 for unlimited"
	return c.Send(msg, keyboards.CancelOnly(), tele.ModeHTML)
}

// processNewDailyLimit processes new daily limit
func (h *Handler) processNewDailyLimit(c tele.Context, input string) error {
	userID := c.Sender().ID

	limit, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || limit < 0 || limit > 50 {
		return c.Send("🤔 Pick a number between 0 and 50 (0 = no limit):", keyboards.CancelOnly())
	}

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	// Check if in super admin mode or observer mode
	isSuperAdminMode := h.isInSuperAdminMode(userID)
	isObserverMode := h.isInObserverMode(userID)

	// Use super admin mode flag or check if user is super admin
	isSuperAdmin := isSuperAdminMode || h.isSuperAdmin(userID)
	if err := h.challenge.UpdateDailyLimit(challengeID, limit, userID, isSuperAdmin); err != nil {
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	// Preserve observer mode if it was set
	if isObserverMode {
		newTempData := map[string]any{TempKeyObserverMode: true}
		h.state.SetStateWithData(userID, domain.StateIdle, newTempData)
	} else {
		h.state.ResetKeepChallenge(userID)
	}

	if limit > 0 {
		c.Send(fmt.Sprintf("✅ Got it! %d tasks/day max", limit))
	} else {
		c.Send("✅ No limits now — go wild! 🚀")
	}

	return h.showAdminPanel(c, challengeID)
}

// handleToggleHideFutureTasks toggles the hide future tasks setting
func (h *Handler) handleToggleHideFutureTasks(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	isSuperAdmin := h.isSuperAdmin(userID)
	newValue, err := h.challenge.ToggleHideFutureTasks(challengeID, userID, isSuperAdmin)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	if newValue {
		c.Send("✅ Sequential mode on — one task at a time! 🔒")
	} else {
		c.Send("✅ All tasks visible now! 👀")
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

	msg := "🚨 Whoa! Delete this challenge?\n\n"
	msg += fmt.Sprintf("\"%s\" will be gone forever.\n\n", challenge.Name)
	msg += "This nukes:\n"
	msg += fmt.Sprintf("• %d tasks\n", taskCount)
	msg += fmt.Sprintf("• %d participants\n", participantCount)
	msg += "• All progress\n\n"
	msg += "⚠️ No take-backs!"

	return c.Send(msg, keyboards.DeleteChallengeConfirm())
}

// handleConfirmDeleteChallenge confirms challenge deletion
func (h *Handler) handleConfirmDeleteChallenge(c tele.Context) error {
	userID := c.Sender().ID

	userState, err := h.state.Get(userID)
	if err != nil {
		logger.Error(
			"handleConfirmDeleteChallenge: failed to get state",
			"user_id",
			userID,
			"error",
			err,
		)
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}
	challengeID := userState.CurrentChallenge
	if challengeID == "" {
		logger.Warn("handleConfirmDeleteChallenge: no current challenge", "user_id", userID)
		return h.sendError(c, "😅 No challenge selected. Please try again.")
	}

	challenge, err := h.challenge.GetByID(challengeID)
	if err != nil {
		logger.Error(
			"handleConfirmDeleteChallenge: failed to get challenge",
			"user_id",
			userID,
			"challenge_id",
			challengeID,
			"error",
			err,
		)
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}
	if challenge == nil {
		logger.Warn(
			"handleConfirmDeleteChallenge: challenge not found",
			"user_id",
			userID,
			"challenge_id",
			challengeID,
		)
		return h.sendError(c, "😅 Challenge not found. It may have already been deleted.")
	}

	// Get participants BEFORE deletion (CASCADE will remove them)
	participantIDs := h.notification.GetParticipantsForDeletion(challengeID)
	challengeName := challenge.Name

	isSuperAdmin := h.isSuperAdmin(userID)
	if err := h.challenge.Delete(challengeID, userID, isSuperAdmin); err != nil {
		logger.Error(
			"handleConfirmDeleteChallenge: failed to delete",
			"user_id",
			userID,
			"challenge_id",
			challengeID,
			"error",
			err,
		)
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	// Notify participants AFTER successful deletion (in background to not block response)
	go h.notification.NotifyChallengeDeletedAsync(
		challengeID,
		challengeName,
		participantIDs,
		userID,
	)

	h.state.Reset(userID)
	c.Send("💨 Poof! Challenge deleted.")
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

	msg := "🔗 <i>Share with friends!</i>\n\n"
	msg += fmt.Sprintf("<b>Challenge ID:</b> <code>%s</code>\n\n", challengeID)
	msg += "Or send this link:\n"
	msg += fmt.Sprintf("<code>t.me/%s?start=%s</code>", botUsername, challengeID)

	kb := keyboards.ShareID(challengeID, botUsername)
	kbJSON, _ := json.Marshal(kb)

	// Use raw API call since telebot doesn't support copy_text buttons natively
	params := map[string]string{
		"chat_id":      fmt.Sprintf("%d", c.Chat().ID),
		"text":         msg,
		"parse_mode":   "HTML",
		"reply_markup": string(kbJSON),
	}
	_, err := h.bot.Raw("sendMessage", params)
	return err
}
