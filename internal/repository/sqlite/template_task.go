package sqlite

import (
	"github.com/jmoiron/sqlx"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
)

// TemplateTaskRepo implements TemplateTaskRepository for SQLite
type TemplateTaskRepo struct {
	db *sqlx.DB
}

func (r *TemplateTaskRepo) Create(task *domain.TemplateTask) error {
	result, err := r.db.NamedExec(`
		INSERT INTO template_tasks (template_id, order_num, title, description, image_file_id)
		VALUES (:template_id, :order_num, :title, :description, :image_file_id)
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

func (r *TemplateTaskRepo) GetByID(id int64) (*domain.TemplateTask, error) {
	var task domain.TemplateTask
	err := r.db.Get(&task, "SELECT * FROM template_tasks WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TemplateTaskRepo) GetByTemplateID(templateID int64) ([]*domain.TemplateTask, error) {
	var tasks []*domain.TemplateTask
	err := r.db.Select(&tasks, `
		SELECT * FROM template_tasks
		WHERE template_id = ?
		ORDER BY order_num ASC
	`, templateID)
	return tasks, err
}

func (r *TemplateTaskRepo) DeleteByTemplateID(templateID int64) error {
	_, err := r.db.Exec("DELETE FROM template_tasks WHERE template_id = ?", templateID)
	return err
}

func (r *TemplateTaskRepo) CountByTemplateID(templateID int64) (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM template_tasks WHERE template_id = ?", templateID)
	return count, err
}

func (r *TemplateTaskRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM template_tasks WHERE id = ?", id)
	return err
}

func (r *TemplateTaskRepo) UpdateTitle(id int64, title string) error {
	_, err := r.db.Exec("UPDATE template_tasks SET title = ? WHERE id = ?", title, id)
	return err
}

func (r *TemplateTaskRepo) UpdateDescription(id int64, description string) error {
	_, err := r.db.Exec("UPDATE template_tasks SET description = ? WHERE id = ?", description, id)
	return err
}

func (r *TemplateTaskRepo) UpdateImage(id int64, imageFileID string) error {
	_, err := r.db.Exec("UPDATE template_tasks SET image_file_id = ? WHERE id = ?", imageFileID, id)
	return err
}

func (r *TemplateTaskRepo) GetMaxOrderNum(templateID int64) (int, error) {
	var maxOrder *int
	err := r.db.Get(&maxOrder, "SELECT MAX(order_num) FROM template_tasks WHERE template_id = ?", templateID)
	if err != nil || maxOrder == nil {
		return 0, err
	}
	return *maxOrder, nil
}

func (r *TemplateTaskRepo) UpdateOrderNum(id int64, orderNum int) error {
	_, err := r.db.Exec("UPDATE template_tasks SET order_num = ? WHERE id = ?", orderNum, id)
	return err
}

func (r *TemplateTaskRepo) UpdateOrderNums(templateID int64, updates map[int64]int) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// First, set all to negative values to avoid unique constraint violations
	for taskID, newOrder := range updates {
		_, err = tx.Exec("UPDATE template_tasks SET order_num = ? WHERE id = ?", -newOrder, taskID)
		if err != nil {
			return err
		}
	}

	// Then set to positive values
	for taskID, newOrder := range updates {
		_, err = tx.Exec("UPDATE template_tasks SET order_num = ? WHERE id = ?", newOrder, taskID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
