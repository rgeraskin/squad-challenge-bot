package handlers

import (
	"fmt"

	"github.com/rgeraskin/squad-challenge-bot/internal/bot/keyboards"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/service"
	tele "gopkg.in/telebot.v3"
)

// showSettings shows the settings view
func (h *Handler) showSettings(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	challenge, _ := h.challenge.GetByID(challengeID)
	participant, _ := h.participant.GetByChallengeAndUser(challengeID, userID)

	if participant == nil {
		return h.sendError(c, "❌ You're not a participant of this challenge.")
	}

	isAdmin := challenge.CreatorID == userID

	msg := "⚙️ Settings\n\n"
	msg += fmt.Sprintf("Current challenge: %s\n", challenge.Name)
	msg += fmt.Sprintf("Your emoji: %s\n", participant.Emoji)
	msg += fmt.Sprintf("Your name: %s\n", participant.DisplayName)

	return c.Send(msg, keyboards.Settings(participant.NotifyEnabled, isAdmin))
}

// handleToggleNotifications toggles notifications
func (h *Handler) handleToggleNotifications(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	participant, _ := h.participant.GetByChallengeAndUser(challengeID, userID)
	if participant == nil {
		return h.sendError(c, "❌ You're not a participant of this challenge.")
	}

	newState, err := h.participant.ToggleNotifications(participant.ID)
	if err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	if newState {
		c.Send("🔔 Notifications enabled")
	} else {
		c.Send("🔕 Notifications disabled")
	}

	return h.showSettings(c)
}

// handleChangeName starts changing display name
func (h *Handler) handleChangeName(c tele.Context) error {
	userID := c.Sender().ID
	h.state.SetState(userID, domain.StateAwaitingNewName)
	return c.Send("Enter new display name:", keyboards.CancelOnly())
}

// processNewName processes new display name
func (h *Handler) processNewName(c tele.Context, name string) error {
	userID := c.Sender().ID

	if len(name) == 0 || len(name) > 30 {
		return c.Send("❌ Display name must be 1-30 characters. Try again:", keyboards.CancelOnly())
	}

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	participant, _ := h.participant.GetByChallengeAndUser(challengeID, userID)
	if participant == nil {
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "❌ You're not a participant of this challenge.")
	}

	if err := h.participant.UpdateName(participant.ID, name); err != nil {
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	h.state.ResetKeepChallenge(userID)
	c.Send(fmt.Sprintf("✅ Display name changed to \"%s\"", name))
	return h.showSettings(c)
}

// handleChangeEmoji starts changing emoji
func (h *Handler) handleChangeEmoji(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	usedEmojis, _ := h.participant.GetUsedEmojis(challengeID)

	h.state.SetState(userID, domain.StateAwaitingNewEmoji)
	return c.Send("Choose your new emoji or send your own:", keyboards.EmojiSelector(usedEmojis))
}

// processNewEmoji processes new emoji
func (h *Handler) processNewEmoji(c tele.Context, emoji string) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	participant, _ := h.participant.GetByChallengeAndUser(challengeID, userID)
	if participant == nil {
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "❌ You're not a participant of this challenge.")
	}

	if err := h.participant.UpdateEmoji(participant.ID, emoji, challengeID); err != nil {
		if err == service.ErrEmojiTaken {
			usedEmojis, _ := h.participant.GetUsedEmojis(challengeID)
			return c.Send("❌ This emoji is already taken. Choose another:", keyboards.EmojiSelector(usedEmojis))
		}
		h.state.ResetKeepChallenge(userID)
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	h.state.ResetKeepChallenge(userID)
	c.Send(fmt.Sprintf("✅ Emoji changed to %s", emoji))
	return h.showSettings(c)
}

// handleLeaveChallenge shows leave confirmation
func (h *Handler) handleLeaveChallenge(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	challenge, _ := h.challenge.GetByID(challengeID)

	msg := fmt.Sprintf("⚠️ Are you sure you want to leave \"%s\"?\n\nYour progress will be deleted.", challenge.Name)
	return c.Send(msg, keyboards.LeaveConfirm())
}

// handleConfirmLeave confirms leaving challenge
func (h *Handler) handleConfirmLeave(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challengeID := userState.CurrentChallenge

	participant, _ := h.participant.GetByChallengeAndUser(challengeID, userID)
	if participant == nil {
		return h.sendError(c, "❌ You're not a participant of this challenge.")
	}

	if err := h.participant.Leave(participant.ID); err != nil {
		return h.sendError(c, "⚠️ Something went wrong. Please try again.")
	}

	h.state.Reset(userID)
	c.Send("✅ You left the challenge.")
	return h.showStartMenu(c)
}
