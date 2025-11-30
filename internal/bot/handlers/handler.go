package handlers

import (
	"github.com/rgeraskin/squad-challenge-bot/internal/logger"
	"github.com/rgeraskin/squad-challenge-bot/internal/repository"
	"github.com/rgeraskin/squad-challenge-bot/internal/service"
	tele "gopkg.in/telebot.v3"
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
