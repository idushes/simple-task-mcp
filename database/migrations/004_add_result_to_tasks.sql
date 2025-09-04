-- Add result field to tasks table
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS result TEXT;
