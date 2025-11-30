package sqlite

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

// CompletionRepo implements CompletionRepository for SQLite
type CompletionRepo struct {
	db *sqlx.DB
}

func (r *CompletionRepo) Create(completion *domain.TaskCompletion) error {
	completion.CompletedAt = time.Now()

	result, err := r.db.NamedExec(`
		INSERT INTO task_completions (task_id, participant_id, completed_at)
		VALUES (:task_id, :participant_id, :completed_at)
	`, completion)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	completion.ID = id
	return nil
}

func (r *CompletionRepo) Delete(taskID, participantID int64) error {
	_, err := r.db.Exec(`
		DELETE FROM task_completions
		WHERE task_id = ? AND participant_id = ?
	`, taskID, participantID)
	return err
}

func (r *CompletionRepo) GetByTaskID(taskID int64) ([]*domain.TaskCompletion, error) {
	var completions []*domain.TaskCompletion
	err := r.db.Select(&completions, `
		SELECT * FROM task_completions
		WHERE task_id = ?
		ORDER BY completed_at ASC
	`, taskID)
	return completions, err
}

func (r *CompletionRepo) GetByParticipantID(participantID int64) ([]*domain.TaskCompletion, error) {
	var completions []*domain.TaskCompletion
	err := r.db.Select(&completions, `
		SELECT * FROM task_completions
		WHERE participant_id = ?
		ORDER BY completed_at ASC
	`, participantID)
	return completions, err
}

func (r *CompletionRepo) GetByTaskAndParticipant(taskID, participantID int64) (*domain.TaskCompletion, error) {
	var completion domain.TaskCompletion
	err := r.db.Get(&completion, `
		SELECT * FROM task_completions
		WHERE task_id = ? AND participant_id = ?
	`, taskID, participantID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &completion, err
}

func (r *CompletionRepo) CountByParticipantID(participantID int64) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*) FROM task_completions WHERE participant_id = ?
	`, participantID)
	return count, err
}

func (r *CompletionRepo) GetCompletedTaskIDs(participantID int64) ([]int64, error) {
	var ids []int64
	err := r.db.Select(&ids, `
		SELECT task_id FROM task_completions WHERE participant_id = ?
	`, participantID)
	return ids, err
}
