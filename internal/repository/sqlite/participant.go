package sqlite

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

// ParticipantRepo implements ParticipantRepository for SQLite
type ParticipantRepo struct {
	db *sqlx.DB
}

func (r *ParticipantRepo) Create(participant *domain.Participant) error {
	participant.JoinedAt = time.Now()

	result, err := r.db.NamedExec(`
		INSERT INTO participants (challenge_id, telegram_id, display_name, emoji, notify_enabled, joined_at)
		VALUES (:challenge_id, :telegram_id, :display_name, :emoji, :notify_enabled, :joined_at)
	`, participant)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	participant.ID = id
	return nil
}

func (r *ParticipantRepo) GetByID(id int64) (*domain.Participant, error) {
	var participant domain.Participant
	err := r.db.Get(&participant, "SELECT * FROM participants WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &participant, err
}

func (r *ParticipantRepo) GetByChallengeAndUser(challengeID string, telegramID int64) (*domain.Participant, error) {
	var participant domain.Participant
	err := r.db.Get(&participant, `
		SELECT * FROM participants
		WHERE challenge_id = ? AND telegram_id = ?
	`, challengeID, telegramID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &participant, err
}

func (r *ParticipantRepo) GetByChallengeID(challengeID string) ([]*domain.Participant, error) {
	var participants []*domain.Participant
	err := r.db.Select(&participants, `
		SELECT * FROM participants
		WHERE challenge_id = ?
		ORDER BY joined_at ASC
	`, challengeID)
	return participants, err
}

func (r *ParticipantRepo) Update(participant *domain.Participant) error {
	_, err := r.db.NamedExec(`
		UPDATE participants
		SET display_name = :display_name, emoji = :emoji, notify_enabled = :notify_enabled
		WHERE id = :id
	`, participant)
	return err
}

func (r *ParticipantRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM participants WHERE id = ?", id)
	return err
}

func (r *ParticipantRepo) CountByChallengeID(challengeID string) (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM participants WHERE challenge_id = ?", challengeID)
	return count, err
}

func (r *ParticipantRepo) GetUsedEmojis(challengeID string) ([]string, error) {
	var emojis []string
	err := r.db.Select(&emojis, `
		SELECT emoji FROM participants WHERE challenge_id = ?
	`, challengeID)
	return emojis, err
}
