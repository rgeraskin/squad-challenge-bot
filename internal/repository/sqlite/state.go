package sqlite

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

// StateRepo implements StateRepository for SQLite
type StateRepo struct {
	db *sqlx.DB
}

func (r *StateRepo) Get(telegramID int64) (*domain.UserState, error) {
	var state domain.UserState
	// Use COALESCE to handle NULL values for temp_data and current_challenge
	err := r.db.Get(&state, `
		SELECT telegram_id, state,
			COALESCE(temp_data, '') as temp_data,
			COALESCE(current_challenge, '') as current_challenge,
			updated_at
		FROM user_states WHERE telegram_id = ?`, telegramID)
	if err == sql.ErrNoRows {
		// Return default idle state
		return &domain.UserState{
			TelegramID: telegramID,
			State:      domain.StateIdle,
			UpdatedAt:  time.Now(),
		}, nil
	}
	return &state, err
}

func (r *StateRepo) Set(state *domain.UserState) error {
	state.UpdatedAt = time.Now()

	_, err := r.db.NamedExec(`
		INSERT INTO user_states (telegram_id, state, temp_data, current_challenge, updated_at)
		VALUES (:telegram_id, :state, :temp_data, :current_challenge, :updated_at)
		ON CONFLICT(telegram_id) DO UPDATE SET
			state = :state,
			temp_data = :temp_data,
			current_challenge = :current_challenge,
			updated_at = :updated_at
	`, state)
	return err
}

func (r *StateRepo) Reset(telegramID int64) error {
	_, err := r.db.Exec(`
		UPDATE user_states
		SET state = ?, temp_data = NULL, updated_at = ?
		WHERE telegram_id = ?
	`, domain.StateIdle, time.Now(), telegramID)
	return err
}

// ResetByChallenge resets all users who have a specific challenge as their current challenge
func (r *StateRepo) ResetByChallenge(challengeID string) error {
	_, err := r.db.Exec(`
		UPDATE user_states
		SET state = ?, temp_data = NULL, current_challenge = NULL, updated_at = ?
		WHERE current_challenge = ?
	`, domain.StateIdle, time.Now(), challengeID)
	return err
}
