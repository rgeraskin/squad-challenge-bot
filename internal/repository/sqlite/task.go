package sqlite

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

// TaskRepo implements TaskRepository for SQLite
type TaskRepo struct {
	db *sqlx.DB
}

func (r *TaskRepo) Create(task *domain.Task) error {
	task.CreatedAt = time.Now()

	result, err := r.db.NamedExec(`
		INSERT INTO tasks (challenge_id, order_num, title, description, image_file_id, created_at)
		VALUES (:challenge_id, :order_num, :title, :description, :image_file_id, :created_at)
	`, task)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	task.ID = id
	return nil
}

func (r *TaskRepo) GetByID(id int64) (*domain.Task, error) {
	var task domain.Task
	err := r.db.Get(&task, "SELECT * FROM tasks WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &task, err
}

func (r *TaskRepo) GetByChallengeID(challengeID string) ([]*domain.Task, error) {
	var tasks []*domain.Task
	err := r.db.Select(&tasks, `
		SELECT * FROM tasks
		WHERE challenge_id = ?
		ORDER BY order_num ASC
	`, challengeID)
	return tasks, err
}

func (r *TaskRepo) Update(task *domain.Task) error {
	_, err := r.db.NamedExec(`
		UPDATE tasks
		SET title = :title, description = :description, image_file_id = :image_file_id, order_num = :order_num
		WHERE id = :id
	`, task)
	return err
}

func (r *TaskRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

func (r *TaskRepo) GetMaxOrderNum(challengeID string) (int, error) {
	var maxOrder sql.NullInt64
	err := r.db.Get(&maxOrder, `
		SELECT MAX(order_num) FROM tasks WHERE challenge_id = ?
	`, challengeID)
	if err != nil {
		return 0, err
	}
	if !maxOrder.Valid {
		return 0, nil
	}
	return int(maxOrder.Int64), nil
}

func (r *TaskRepo) UpdateOrderNums(challengeID string, updates map[int64]int) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// First, set all to negative values to avoid unique constraint violations
	for taskID, newOrder := range updates {
		_, err = tx.Exec("UPDATE tasks SET order_num = ? WHERE id = ?", -newOrder, taskID)
		if err != nil {
			return err
		}
	}

	// Then set to positive values
	for taskID, newOrder := range updates {
		_, err = tx.Exec("UPDATE tasks SET order_num = ? WHERE id = ?", newOrder, taskID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *TaskRepo) CountByChallengeID(challengeID string) (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM tasks WHERE challenge_id = ?", challengeID)
	return count, err
}
