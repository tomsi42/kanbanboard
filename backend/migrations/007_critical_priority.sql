-- Add 'critical' to the allowed task priority values
ALTER TABLE tasks DROP CONSTRAINT tasks_priority_check;
ALTER TABLE tasks ADD CONSTRAINT tasks_priority_check
    CHECK (priority IN ('none', 'low', 'medium', 'high', 'critical'));
