package sqlite

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

// ChallengeRepo implements ChallengeRepository for SQLite
type ChallengeRepo struct {
	db *sqlx.DB
}

func (r *ChallengeRepo) Create(challenge *domain.Challenge) error {
	challenge.CreatedAt = time.Now()
	challenge.UpdatedAt = time.Now()

	_, err := r.db.NamedExec(`
		INSERT INTO challenges (id, name, description, creator_id, daily_task_limit, hide_future_tasks, created_at, updated_at)
		VALUES (:id, :name, :description, :creator_id, :daily_task_limit, :hide_future_tasks, :created_at, :updated_at)
	`, challenge)
	return err
}

func (r *ChallengeRepo) GetByID(id string) (*domain.Challenge, error) {
	var challenge domain.Challenge
	err := r.db.Get(&challenge, "SELECT * FROM challenges WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &challenge, err
}

func (r *ChallengeRepo) GetByUserID(telegramID int64) ([]*domain.Challenge, error) {
	var challenges []*domain.Challenge
	err := r.db.Select(&challenges, `
		SELECT DISTINCT c.* FROM challenges c
		LEFT JOIN participants p ON c.id = p.challenge_id
		WHERE p.telegram_id = ? OR c.creator_id = ?
		ORDER BY c.updated_at DESC
	`, telegramID, telegramID)
	return challenges, err
}

func (r *ChallengeRepo) Update(challenge *domain.Challenge) error {
	challenge.UpdatedAt = time.Now()
	_, err := r.db.NamedExec(`
		UPDATE challenges
		SET name = :name, description = :description, daily_task_limit = :daily_task_limit, hide_future_tasks = :hide_future_tasks, updated_at = :updated_at
		WHERE id = :id
	`, challenge)
	return err
}

func (r *ChallengeRepo) UpdateDailyLimit(id string, limit int) error {
	_, err := r.db.Exec(`
		UPDATE challenges
		SET daily_task_limit = ?, updated_at = ?
		WHERE id = ?
	`, limit, time.Now(), id)
	return err
}

func (r *ChallengeRepo) UpdateHideFutureTasks(id string, hide bool) error {
	_, err := r.db.Exec(`
		UPDATE challenges
		SET hide_future_tasks = ?, updated_at = ?
		WHERE id = ?
	`, hide, time.Now(), id)
	return err
}

func (r *ChallengeRepo) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM challenges WHERE id = ?", id)
	return err
}

func (r *ChallengeRepo) Exists(id string) (bool, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM challenges WHERE id = ?", id)
	return count > 0, err
}
