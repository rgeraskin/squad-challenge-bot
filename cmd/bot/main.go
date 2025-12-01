package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rgeraskin/squad-challenge-bot/internal/bot"
	"github.com/rgeraskin/squad-challenge-bot/internal/config"
	"github.com/rgeraskin/squad-challenge-bot/internal/logger"
	"github.com/rgeraskin/squad-challenge-bot/internal/repository/sqlite"
)

func main() {
	cfg := config.Load()

	// Initialize logger
	logger.Init(cfg.LogLevel)

	if cfg.TelegramBotToken == "" {
		logger.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	// Initialize database
	logger.Info("Initializing database", "path", cfg.DatabasePath)
	repo, err := sqlite.New(cfg.DatabasePath)
	if err != nil {
		logger.Fatal("Failed to initialize database", "error", err)
	}
	defer repo.Close()

	// Initialize bot
	logger.Info("Initializing bot")
	if cfg.SuperAdminID > 0 {
		logger.Info("Super admin ID configured", "telegram_id", cfg.SuperAdminID)
	}
	b, err := bot.New(cfg.TelegramBotToken, repo, cfg.SuperAdminID)
	if err != nil {
		logger.Fatal("Failed to initialize bot", "error", err)
	}

	// Start health check server if port is configured
	if cfg.HealthPort != "" {
		go startHealthServer(cfg.HealthPort)
	}

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("Shutting down...")
		b.Stop()
	}()

	// Start bot
	logger.Info("Bot started", "username", b.Username())
	b.Start()
}

func startHealthServer(port string) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	logger.Info("Health server started", "port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Error("Health server error", "error", err)
	}
}
