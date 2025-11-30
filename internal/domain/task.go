package domain

import "time"

// Task represents a single task within a challenge
type Task struct {
	ID          int64     `db:"id"`
	ChallengeID string    `db:"challenge_id"`
	OrderNum    int       `db:"order_num"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	ImageFileID string    `db:"image_file_id"`
	CreatedAt   time.Time `db:"created_at"`
}
