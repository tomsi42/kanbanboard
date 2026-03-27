-- Add "task" label to all existing projects that don't have one
INSERT INTO labels (project_id, name, color)
SELECT p.id, 'task', '#4a90d9'
FROM projects p
WHERE NOT EXISTS (
    SELECT 1 FROM labels l WHERE l.project_id = p.id AND l.name = 'task'
);
