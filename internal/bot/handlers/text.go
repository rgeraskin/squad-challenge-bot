package handlers

import (
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/util"
	tele "gopkg.in/telebot.v3"
)

// HandleText handles text messages (user input in conversation flows)
func (h *Handler) HandleText(c tele.Context) error {
	userID := c.Sender().ID
	text := c.Text()

	userState, err := h.state.Get(userID)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	switch userState.State {
	// Challenge creation
	case domain.StateAwaitingChallengeName:
		return h.processChallengeName(c, text)
	case domain.StateAwaitingChallengeDescription:
		return h.processChallengeDescription(c, text)
	case domain.StateAwaitingCreatorName:
		return h.processCreatorName(c, text)
	case domain.StateAwaitingCreatorEmoji:
		if util.IsValidEmoji(text) {
			return h.processCreatorEmoji(c, text)
		}
		return c.Send("🎨 Just one emoji please!")
	case domain.StateAwaitingDailyLimit:
		return h.processDailyLimit(c, text)
	case domain.StateAwaitingCreatorSyncTime:
		return h.processCreatorSyncTime(c, text)

	// Task management
	case domain.StateAwaitingTaskTitle:
		return h.processTaskTitle(c, text)
	case domain.StateAwaitingTaskDescription:
		return h.processTaskDescription(c, text)
	case domain.StateAwaitingEditTitle:
		return h.processEditTitle(c, text)
	case domain.StateAwaitingEditDescription:
		return h.processEditDescription(c, text)

	// Joining challenge
	case domain.StateAwaitingChallengeID:
		return h.processChallengeID(c, text)
	case domain.StateAwaitingParticipantName:
		return h.processParticipantName(c, text)
	case domain.StateAwaitingParticipantEmoji:
		if util.IsValidEmoji(text) {
			return h.processParticipantEmoji(c, text)
		}
		return c.Send("🎨 Just one emoji please!")
	case domain.StateAwaitingSyncTime:
		return h.processSyncTime(c, text)

	// Admin
	case domain.StateAwaitingNewChallengeName:
		return h.processNewChallengeName(c, text)
	case domain.StateAwaitingNewChallengeDescription:
		return h.processNewChallengeDescription(c, text)
	case domain.StateAwaitingNewDailyLimit:
		return h.processNewDailyLimit(c, text)

	// Settings
	case domain.StateAwaitingNewName:
		return h.processNewName(c, text)
	case domain.StateAwaitingNewEmoji:
		if util.IsValidEmoji(text) {
			return h.processNewEmoji(c, text)
		}
		return c.Send("🎨 Just one emoji please!")

	// Super Admin
	case domain.StateAwaitingSuperAdminID:
		return h.processGrantSuperAdmin(c, text)

	// Template-based challenge creation
	case domain.StateAwaitingTemplateChallengeName:
		return h.processTemplateChallengeName(c, text)
	case domain.StateAwaitingTemplateCreatorName:
		return h.processTemplateCreatorName(c, text)
	case domain.StateAwaitingTemplateCreatorEmoji:
		if util.IsValidEmoji(text) {
			return h.processTemplateCreatorEmoji(c, text)
		}
		return c.Send("🎨 Just one emoji please!")
	case domain.StateAwaitingTemplateCreatorSyncTime:
		return h.processTemplateCreatorSyncTime(c, text)

	// Template admin editing (Super Admin)
	case domain.StateAwaitingNewTemplateName:
		return h.processNewTemplateName(c, text)
	case domain.StateAwaitingNewTemplateDescription:
		return h.processNewTemplateDescription(c, text)
	case domain.StateAwaitingNewTemplateDailyLimit:
		return h.processNewTemplateDailyLimit(c, text)
	case domain.StateAwaitingTplTaskTitle:
		return h.processTplTaskTitle(c, text)
	case domain.StateAwaitingTplTaskDescription:
		if text == "skip" || text == "Skip" {
			return h.skipTplTaskDescription(c)
		}
		return h.processTplTaskDescription(c, text)
	case domain.StateAwaitingTplTaskImage:
		if text == "skip" || text == "Skip" {
			return h.skipTplTaskImage(c)
		}
		return c.Send("📷 Please send an image or type 'skip'")
	case domain.StateAwaitingTplEditTitle:
		return h.processTplEditTitle(c, text)
	case domain.StateAwaitingTplEditDescription:
		return h.processTplEditDescription(c, text)
	case domain.StateAwaitingTplEditImage:
		if text == "remove" || text == "Remove" {
			return h.processTplEditImage(c, "")
		}
		return c.Send("📷 Please send an image or type 'remove'")

	default:
		// Idle state - ignore text or show help
		return nil
	}
}

// HandlePhoto handles photo messages (for task images)
func (h *Handler) HandlePhoto(c tele.Context) error {
	userID := c.Sender().ID

	userState, err := h.state.Get(userID)
	if err != nil {
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	photo := c.Message().Photo
	if photo == nil {
		return nil
	}

	fileID := photo.FileID

	switch userState.State {
	case domain.StateAwaitingTaskImage:
		return h.processTaskImage(c, fileID)
	case domain.StateAwaitingEditImage:
		return h.processEditImage(c, fileID)
	case domain.StateAwaitingTplTaskImage:
		return h.processTplTaskImage(c, fileID)
	case domain.StateAwaitingTplEditImage:
		return h.processTplEditImage(c, fileID)
	default:
		return nil
	}
}
