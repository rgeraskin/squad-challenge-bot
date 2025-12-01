package handlers

import (
	"github.com/rgeraskin/squad-challenge-bot/internal/logger"
	"github.com/rgeraskin/squad-challenge-bot/internal/repository"
	"github.com/rgeraskin/squad-challenge-bot/internal/service"
	tele "gopkg.in/telebot.v3"
)

// Temp data keys used for state management
const (
	TempKeyObserverMode   = "observer_mode"
	TempKeySuperAdminMode = "super_admin_mode"
	TempKeyChallengeID    = "challenge_id"
	TempKeyTaskID         = "task_id"
)

// Handler holds all bot handlers and services
type Handler struct {
	repo         repository.Repository
	challenge    *service.ChallengeService
	task         *service.TaskService
	participant  *service.ParticipantService
	completion   *service.CompletionService
	state        *service.StateService
	notification *service.NotificationService
	superAdmin   *service.SuperAdminService
	bot          *tele.Bot
}

// NewHandler creates a new Handler
func NewHandler(
	repo repository.Repository,
	challenge *service.ChallengeService,
	task *service.TaskService,
	participant *service.ParticipantService,
	completion *service.CompletionService,
	state *service.StateService,
	notification *service.NotificationService,
	superAdmin *service.SuperAdminService,
	bot *tele.Bot,
) *Handler {
	return &Handler{
		repo:         repo,
		challenge:    challenge,
		task:         task,
		participant:  participant,
		completion:   completion,
		state:        state,
		notification: notification,
		superAdmin:   superAdmin,
		bot:          bot,
	}
}

// sendError sends an error message to the user
func (h *Handler) sendError(c tele.Context, msg string) error {
	logger.Warn("Sending error to user", "user_id", c.Sender().ID, "message", msg)
	return c.Send(msg)
}

// getParticipantAndChallenge gets the current participant and challenge from state
func (h *Handler) getParticipantAndChallenge(c tele.Context) (*service.ChallengeService, string, error) {
	userState, err := h.state.Get(c.Sender().ID)
	if err != nil {
		return nil, "", err
	}
	return h.challenge, userState.CurrentChallenge, nil
}

// isSuperAdmin checks if the current user is a super admin
func (h *Handler) isSuperAdmin(userID int64) bool {
	isSuperAdmin, err := h.superAdmin.IsSuperAdmin(userID)
	if err != nil {
		logger.Warn("Failed to check super admin status", "user_id", userID, "error", err)
	}
	return isSuperAdmin
}

// isInObserverMode checks if the user is in observer mode (super admin viewing a challenge)
func (h *Handler) isInObserverMode(userID int64) bool {
	var tempData map[string]any
	h.state.GetTempData(userID, &tempData)
	if tempData != nil {
		if observerMode, ok := tempData[TempKeyObserverMode].(bool); ok {
			return observerMode
		}
	}
	return false
}

// isInSuperAdminMode checks if the user is in super admin mode
func (h *Handler) isInSuperAdminMode(userID int64) bool {
	var tempData map[string]any
	h.state.GetTempData(userID, &tempData)
	if tempData != nil {
		if superAdminMode, ok := tempData[TempKeySuperAdminMode].(bool); ok {
			return superAdminMode
		}
	}
	return false
}

// getTelegramName returns the user's Telegram username or first name as fallback
func getTelegramName(c tele.Context) string {
	if c.Sender().Username != "" {
		return c.Sender().Username
	}
	return c.Sender().FirstName
}
