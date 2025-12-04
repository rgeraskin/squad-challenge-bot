-- Templates table
-- Stores challenge templates that can be reused
CREATE TABLE IF NOT EXISTS templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    daily_task_limit INTEGER DEFAULT 0,
    hide_future_tasks INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Template tasks table
-- Stores tasks belonging to templates
CREATE TABLE IF NOT EXISTS template_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INTEGER NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
    order_num INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    image_file_id TEXT,
    UNIQUE(template_id, order_num)
);


CREATE INDEX IF NOT EXISTS idx_template_tasks_template ON template_tasks(template_id);
