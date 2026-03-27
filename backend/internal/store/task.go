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

// UpdateTask updates all editable fields of a task.
func UpdateTask(db *sql.DB, task model.Task) (model.Task, error) {
	err := db.QueryRow(`
		UPDATE tasks SET
			title = $1, description = $2, column_id = $3, label_id = $4,
			assignee_id = $5, priority = $6, target_version = $7, due_date = $8,
			updated_at = NOW()
		WHERE id = $9
		RETURNING updated_at
	`, task.Title, task.Description, task.ColumnID, task.LabelID,
		task.AssigneeID, task.Priority, task.TargetVersion, task.DueDate,
		task.ID,
	).Scan(&task.UpdatedAt)
	if err != nil {
		return model.Task{}, fmt.Errorf("update task: %w", err)
	}
	return task, nil
}

// MoveTask moves a task to a column at a specific position.
// It reassigns positions for all tasks in both source and target columns.
func MoveTask(db *sql.DB, taskID, newColumnID string, position int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the current column before moving
	var oldColumnID string
	err = tx.QueryRow("SELECT column_id FROM tasks WHERE id = $1", taskID).Scan(&oldColumnID)
	if err != nil {
		return fmt.Errorf("get current column: %w", err)
	}

	// Move the task to the new column
	_, err = tx.Exec(
		"UPDATE tasks SET column_id = $1, updated_at = NOW() WHERE id = $2",
		newColumnID, taskID,
	)
	if err != nil {
		return fmt.Errorf("move task: %w", err)
	}

	// Reorder the target column with the task at the requested position
	if err := reorderColumn(tx, newColumnID, taskID, position); err != nil {
		return fmt.Errorf("reorder target column: %w", err)
	}

	// If the task moved between columns, reorder the source column too
	if oldColumnID != newColumnID {
		if err := reorderColumn(tx, oldColumnID, "", -1); err != nil {
			return fmt.Errorf("reorder source column: %w", err)
		}
	}

	return tx.Commit()
}

// reorderColumn reassigns sequential positions to all tasks in a column.
// If insertID is non-empty, that task is placed at insertPos; others fill around it.
func reorderColumn(tx *sql.Tx, columnID, insertID string, insertPos int) error {
	rows, err := tx.Query(
		"SELECT id FROM tasks WHERE column_id = $1 ORDER BY position, created_at",
		columnID,
	)
	if err != nil {
		return err
	}

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	rows.Close()

	// If we need to insert a specific task at a position
	if insertID != "" {
		others := make([]string, 0, len(ids))
		for _, id := range ids {
			if id != insertID {
				others = append(others, id)
			}
		}

		ordered := make([]string, 0, len(ids))
		if insertPos >= len(others) {
			ordered = append(others, insertID)
		} else if insertPos <= 0 {
			ordered = append(ordered, insertID)
			ordered = append(ordered, others...)
		} else {
			ordered = append(ordered, others[:insertPos]...)
			ordered = append(ordered, insertID)
			ordered = append(ordered, others[insertPos:]...)
		}
		ids = ordered
	}

	for i, id := range ids {
		_, err := tx.Exec("UPDATE tasks SET position = $1 WHERE id = $2", i, id)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteTask removes a task by ID.
func DeleteTask(db *sql.DB, taskID string) error {
	_, err := db.Exec("DELETE FROM tasks WHERE id = $1", taskID)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}
