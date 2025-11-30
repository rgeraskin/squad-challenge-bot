-- Add daily task limit to challenges and time offset to participants

ALTER TABLE challenges ADD COLUMN daily_task_limit INTEGER DEFAULT 0;
ALTER TABLE participants ADD COLUMN time_offset_minutes INTEGER DEFAULT 0;
