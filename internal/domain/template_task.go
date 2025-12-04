package domain

// TemplateTask represents a task within a template
type TemplateTask struct {
	ID          int64  `db:"id"`
	TemplateID  int64  `db:"template_id"`
	OrderNum    int    `db:"order_num"`
	Title       string `db:"title"`
	Description string `db:"description"`
	ImageFileID string `db:"image_file_id"`
}
