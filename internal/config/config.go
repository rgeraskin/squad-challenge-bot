package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	TelegramBotToken string
	DatabasePath     string
	LogLevel         string
	HealthPort       string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		DatabasePath:     getEnv("DATABASE_PATH", "./data/bot.db"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		HealthPort:       getEnv("HEALTH_PORT", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
