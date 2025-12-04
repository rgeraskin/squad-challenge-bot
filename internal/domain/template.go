package domain

import "time"

// Template represents a challenge template
type Template struct {
	ID              int64     `db:"id"`
	Name            string    `db:"name"`
	Description     string    `db:"description"`
	DailyTaskLimit  int       `db:"daily_task_limit"`  // 0 = unlimited
	HideFutureTasks bool      `db:"hide_future_tasks"` // hide task names after current task
	CreatedAt       time.Time `db:"created_at"`
}
