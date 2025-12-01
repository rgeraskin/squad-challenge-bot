package domain

import "time"

// SuperAdmin represents a user with super admin privileges
type SuperAdmin struct {
	ID         int64     `db:"id"`
	TelegramID int64     `db:"telegram_id"`
	CreatedAt  time.Time `db:"created_at"`
}
