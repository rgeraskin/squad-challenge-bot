package domain

import "time"

// Challenge represents a team challenge with tasks
type Challenge struct {
	ID              string    `db:"id"`
	Name            string    `db:"name"`
	Description     string    `db:"description"`
	CreatorID       int64     `db:"creator_id"`
	DailyTaskLimit  int       `db:"daily_task_limit"`  // 0 = unlimited
	HideFutureTasks bool      `db:"hide_future_tasks"` // hide task names after current task
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}
