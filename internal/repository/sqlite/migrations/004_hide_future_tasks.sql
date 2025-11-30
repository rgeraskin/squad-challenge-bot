-- Add hide future tasks setting to challenges

ALTER TABLE challenges ADD COLUMN hide_future_tasks INTEGER DEFAULT 0;
