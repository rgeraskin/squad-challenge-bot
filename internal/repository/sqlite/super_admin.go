package sqlite

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

// SuperAdminRepo implements SuperAdminRepository for SQLite
type SuperAdminRepo struct {
	db *sqlx.DB
}

func (r *SuperAdminRepo) Create(telegramID int64) error {
	_, err := r.db.Exec(`
		INSERT OR IGNORE INTO super_admins (telegram_id, created_at)
		VALUES (?, ?)
	`, telegramID, time.Now())
	return err
}

func (r *SuperAdminRepo) Delete(telegramID int64) error {
	_, err := r.db.Exec("DELETE FROM super_admins WHERE telegram_id = ?", telegramID)
	return err
}

func (r *SuperAdminRepo) IsSuperAdmin(telegramID int64) (bool, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM super_admins WHERE telegram_id = ?", telegramID)
	return count > 0, err
}

func (r *SuperAdminRepo) GetAll() ([]*domain.SuperAdmin, error) {
	var admins []*domain.SuperAdmin
	err := r.db.Select(&admins, "SELECT * FROM super_admins ORDER BY created_at")
	return admins, err
}

func (r *SuperAdminRepo) Exists(telegramID int64) (bool, error) {
	return r.IsSuperAdmin(telegramID)
}
