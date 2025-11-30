package domain

import "time"

// TaskCompletion represents a task completed by a participant
type TaskCompletion struct {
	ID            int64     `db:"id"`
	TaskID        int64     `db:"task_id"`
	ParticipantID int64     `db:"participant_id"`
	CompletedAt   time.Time `db:"completed_at"`
}
