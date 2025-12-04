package bot

import (
	"time"

	"github.com/rgeraskin/squad-challenge-bot/internal/bot/handlers"
	"github.com/rgeraskin/squad-challenge-bot/internal/logger"
	"github.com/rgeraskin/squad-challenge-bot/internal/repository"
	"github.com/rgeraskin/squad-challenge-bot/internal/service"
	tele "gopkg.in/telebot.v3"
)

// Bot wraps the telebot instance and handlers
type Bot struct {
	bot      *tele.Bot
	handlers *handlers.Handler
}

// New creates a new bot instance
func New(token string, repo repository.Repository, superAdminID int64) (*Bot, error) {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	// Initialize services
	challengeSvc := service.NewChallengeService(repo)
	taskSvc := service.NewTaskService(repo)
	participantSvc := service.NewParticipantService(repo)
	completionSvc := service.NewCompletionService(repo)
	stateSvc := service.NewStateService(repo)
	notifySvc := service.NewNotificationService(repo, b)
	superAdminSvc := service.NewSuperAdminService(repo)
	templateSvc := service.NewTemplateService(repo)

	// Seed super admin from environment
	if superAdminID > 0 {
		if err := superAdminSvc.SeedFromEnv(superAdminID); err != nil {
			logger.Warn("Failed to seed super admin", "error", err)
		} else {
			logger.Info("Super admin seeded", "telegram_id", superAdminID)
		}
	}

	// Initialize handlers
	h := handlers.NewHandler(
		repo,
		challengeSvc,
		taskSvc,
		participantSvc,
		completionSvc,
		stateSvc,
		notifySvc,
		superAdminSvc,
		templateSvc,
		b,
	)

	bot := &Bot{
		bot:      b,
		handlers: h,
	}

	bot.registerHandlers()

	return bot, nil
}

func (b *Bot) registerHandlers() {
	logger.Info("Registering bot handlers")

	// Command handlers
	b.bot.Handle("/start", b.handlers.HandleStart)
	logger.Debug("Registered /start handler")

	// Text message handler (for user input in conversation flows)
	b.bot.Handle(tele.OnText, b.handlers.HandleText)
	logger.Debug("Registered OnText handler")

	// Photo handler (for task images)
	b.bot.Handle(tele.OnPhoto, b.handlers.HandlePhoto)
	logger.Debug("Registered OnPhoto handler")

	// Callback query handler
	b.bot.Handle(tele.OnCallback, b.handlers.HandleCallback)
	logger.Debug("Registered OnCallback handler")

	logger.Info("All handlers registered successfully")
}

// Start starts the bot
func (b *Bot) Start() {
	logger.Info("Bot polling started")
	b.bot.Start()
}

// Stop stops the bot
func (b *Bot) Stop() {
	b.bot.Stop()
}

// Username returns the bot's username
func (b *Bot) Username() string {
	return b.bot.Me.Username
}
