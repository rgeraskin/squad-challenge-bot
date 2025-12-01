package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	TelegramBotToken string
	DatabasePath     string
	LogLevel         string
	HealthPort       string
	SuperAdminID     int64
}

// Load reads configuration from environment variables
func Load() *Config {
	superAdminID := int64(0)
	if id := os.Getenv("SUPER_ADMIN_ID"); id != "" {
		if parsed, err := strconv.ParseInt(id, 10, 64); err == nil {
			superAdminID = parsed
		}
	}

	return &Config{
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		DatabasePath:     getEnv("DATABASE_PATH", "./data/bot.db"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		HealthPort:       getEnv("HEALTH_PORT", ""),
		SuperAdminID:     superAdminID,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
