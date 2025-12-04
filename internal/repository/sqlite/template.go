package sqlite

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

// TemplateRepo implements TemplateRepository for SQLite
type TemplateRepo struct {
	db *sqlx.DB
}

func (r *TemplateRepo) Create(template *domain.Template) error {
	template.CreatedAt = time.Now()

	result, err := r.db.NamedExec(`
		INSERT INTO templates (name, description, daily_task_limit, hide_future_tasks, created_at)
		VALUES (:name, :description, :daily_task_limit, :hide_future_tasks, :created_at)
	`, template)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	template.ID = id
	return nil
}

func (r *TemplateRepo) GetByID(id int64) (*domain.Template, error) {
	var template domain.Template
	err := r.db.Get(&template, "SELECT * FROM templates WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &template, err
}

func (r *TemplateRepo) GetAll() ([]*domain.Template, error) {
	var templates []*domain.Template
	err := r.db.Select(&templates, "SELECT * FROM templates ORDER BY created_at DESC")
	return templates, err
}

func (r *TemplateRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM templates WHERE id = ?", id)
	return err
}

func (r *TemplateRepo) Count() (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM templates")
	return count, err
}

func (r *TemplateRepo) ExistsByName(name string) (bool, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM templates WHERE name = ?", name)
	return count > 0, err
}

func (r *TemplateRepo) UpdateName(id int64, name string) error {
	_, err := r.db.Exec("UPDATE templates SET name = ? WHERE id = ?", name, id)
	return err
}

func (r *TemplateRepo) UpdateDescription(id int64, description string) error {
	_, err := r.db.Exec("UPDATE templates SET description = ? WHERE id = ?", description, id)
	return err
}

func (r *TemplateRepo) UpdateDailyLimit(id int64, limit int) error {
	_, err := r.db.Exec("UPDATE templates SET daily_task_limit = ? WHERE id = ?", limit, id)
	return err
}

func (r *TemplateRepo) UpdateHideFutureTasks(id int64, hide bool) error {
	_, err := r.db.Exec("UPDATE templates SET hide_future_tasks = ? WHERE id = ?", hide, id)
	return err
}
