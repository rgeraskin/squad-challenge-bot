-- Add image_file_id column to template_tasks if it doesn't exist
-- This handles databases created before image support was added
-- The error is ignored in db.go if column already exists
ALTER TABLE template_tasks ADD COLUMN image_file_id TEXT;
