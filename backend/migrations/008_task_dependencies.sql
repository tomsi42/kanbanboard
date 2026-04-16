-- Task dependencies: tracks which tasks block other tasks (same project)
CREATE TABLE task_dependencies (
    blocker_task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    blocked_task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    PRIMARY KEY (blocker_task_id, blocked_task_id)
);

CREATE INDEX idx_task_dependencies_blocked ON task_dependencies(blocked_task_id);
