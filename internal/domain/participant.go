package domain

import "time"

// Participant represents a user participating in a challenge
type Participant struct {
	ID            int64     `db:"id"`
	ChallengeID   string    `db:"challenge_id"`
	TelegramID    int64     `db:"telegram_id"`
	DisplayName   string    `db:"display_name"`
	Emoji         string    `db:"emoji"`
	NotifyEnabled bool      `db:"notify_enabled"`
	JoinedAt      time.Time `db:"joined_at"`
}
