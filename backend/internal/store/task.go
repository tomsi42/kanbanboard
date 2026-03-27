package store

import (
	"database/sql"
	"errors"
	"fmt"

	"kanbanboard/internal/model"
)

// ErrTaskNotFound is returned when a task is not found.
var ErrTaskNotFound = errors.New("task not found")

// CreateTask inserts a new task. Position is set to the next available in the column.
func CreateTask(db *sql.DB, task model.Task) (model.Task, error) {
	// Get next position in the column
	var maxPos sql.NullInt64
	err := db.QueryRow(
		"SELECT MAX(position) FROM tasks WHERE column_id = $1",
		task.ColumnID,
	).Scan(&maxPos)
	if err != nil {
		return model.Task{}, fmt.Errorf("get max position: %w", err)
	}

	if maxPos.Valid {
		task.Position = int(maxPos.Int64) + 1
	} else {
		task.Position = 0
	}

	err = db.QueryRow(`
		INSERT INTO tasks (project_id, column_id, label_id, assignee_id, creator_id, parent_task_id,
			title, description, priority, target_version, due_date, position)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`, task.ProjectID, task.ColumnID, task.LabelID, task.AssigneeID, task.CreatorID,
		task.ParentTaskID, task.Title, task.Description, task.Priority,
		task.TargetVersion, task.DueDate, task.Position,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return model.Task{}, fmt.Errorf("create task: %w", err)
	}

	return task, nil
}

// ListTasksForProject returns all tasks for a project, ordered by column then position.
func ListTasksForProject(db *sql.DB, projectID string) ([]model.Task, error) {
	rows, err := db.Query(`
		SELECT t.id, t.project_id, t.column_id, t.label_id, t.assignee_id, t.creator_id,
			t.parent_task_id, t.title, t.description, t.priority, t.target_version,
			t.due_date, t.position, t.created_at, t.updated_at
		FROM tasks t
		JOIN columns c ON t.column_id = c.id
		WHERE t.project_id = $1
		ORDER BY c.position, t.position
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.ColumnID, &t.LabelID, &t.AssigneeID,
			&t.CreatorID, &t.ParentTaskID, &t.Title, &t.Description, &t.Priority,
			&t.TargetVersion, &t.DueDate, &t.Position, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// GetTask retrieves a task by ID.
func GetTask(db *sql.DB, taskID string) (model.Task, error) {
	var t model.Task
	err := db.QueryRow(`
		SELECT id, project_id, column_id, label_id, assignee_id, creator_id,
			parent_task_id, title, description, priority, target_version,
			due_date, position, created_at, updated_at
		FROM tasks WHERE id = $1
	`, taskID).Scan(&t.ID, &t.ProjectID, &t.ColumnID, &t.LabelID, &t.AssigneeID,
		&t.CreatorID, &t.ParentTaskID, &t.Title, &t.Description, &t.Priority,
		&t.TargetVersion, &t.DueDate, &t.Position, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Task{}, ErrTaskNotFound
	}
	if err != nil {
		return model.Task{}, fmt.Errorf("get task: %w", err)
	}
	return t, nil
}
