package store

import (
	"database/sql"
	"errors"
	"fmt"
)

// ErrSelfDependency is returned when a task is added as its own blocker.
var ErrSelfDependency = errors.New("a task cannot block itself")

// ErrCrossProjectDependency is returned when the blocker is in a different project.
var ErrCrossProjectDependency = errors.New("blocker must be in the same project")

// ErrDependencyCycle is returned when adding a blocker would create a cycle.
var ErrDependencyCycle = errors.New("adding this blocker would create a dependency cycle")

// AddBlocker records that blockerTaskID must be completed before blockedTaskID.
// Both tasks must belong to projectID.
// Rejects self-references and cycles.
func AddBlocker(db *sql.DB, projectID, blockerTaskID, blockedTaskID string) error {
	if blockerTaskID == blockedTaskID {
		return ErrSelfDependency
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Verify both tasks belong to the given project.
	if err := requireSameProject(tx, projectID, blockerTaskID, blockedTaskID); err != nil {
		return err
	}

	// Reject if adding this edge would create a cycle.
	cycle, err := hasCycle(tx, blockerTaskID, blockedTaskID)
	if err != nil {
		return fmt.Errorf("check cycle: %w", err)
	}
	if cycle {
		return ErrDependencyCycle
	}

	_, err = tx.Exec(
		"INSERT INTO task_dependencies (blocker_task_id, blocked_task_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		blockerTaskID, blockedTaskID,
	)
	if err != nil {
		return fmt.Errorf("add blocker: %w", err)
	}

	return tx.Commit()
}

// RemoveBlocker removes the dependency that blockerTaskID blocks blockedTaskID.
// A no-op if the dependency does not exist.
func RemoveBlocker(db *sql.DB, blockerTaskID, blockedTaskID string) error {
	_, err := db.Exec(
		"DELETE FROM task_dependencies WHERE blocker_task_id = $1 AND blocked_task_id = $2",
		blockerTaskID, blockedTaskID,
	)
	if err != nil {
		return fmt.Errorf("remove blocker: %w", err)
	}
	return nil
}

// ListBlockersForProject returns all blocker IDs grouped by blocked task ID
// for every dependency within the given project.
// The returned map is keyed by blocked task ID; values are slices of blocker task IDs.
func ListBlockersForProject(db *sql.DB, projectID string) (map[string][]string, error) {
	rows, err := db.Query(`
		SELECT td.blocker_task_id, td.blocked_task_id
		FROM task_dependencies td
		JOIN tasks t ON td.blocker_task_id = t.id
		WHERE t.project_id = $1
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list blockers for project: %w", err)
	}
	defer rows.Close()

	result := map[string][]string{}
	for rows.Next() {
		var blockerID, blockedID string
		if err := rows.Scan(&blockerID, &blockedID); err != nil {
			return nil, fmt.Errorf("scan blocker row: %w", err)
		}
		result[blockedID] = append(result[blockedID], blockerID)
	}
	return result, rows.Err()
}

// GetBlockersForTask returns the IDs of all tasks that directly block the given task.
func GetBlockersForTask(db *sql.DB, taskID string) ([]string, error) {
	rows, err := db.Query(
		"SELECT blocker_task_id FROM task_dependencies WHERE blocked_task_id = $1",
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("get blockers for task: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan blocker id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// requireSameProject checks that both task IDs belong to projectID.
func requireSameProject(tx *sql.Tx, projectID, taskID1, taskID2 string) error {
	var count int
	err := tx.QueryRow(
		"SELECT COUNT(*) FROM tasks WHERE id IN ($1, $2) AND project_id = $3",
		taskID1, taskID2, projectID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("check project membership: %w", err)
	}
	if count != 2 {
		return ErrCrossProjectDependency
	}
	return nil
}

// hasCycle reports whether adding the edge blockerID → blockedID would form a cycle.
// It performs a depth-first search from blockedID following blocker→blocked edges.
// If blockerID is reachable from blockedID, a cycle would result.
func hasCycle(tx *sql.Tx, blockerID, blockedID string) (bool, error) {
	visited := map[string]bool{}
	queue := []string{blockedID}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if visited[cur] {
			continue
		}
		visited[cur] = true

		if cur == blockerID {
			return true, nil
		}

		rows, err := tx.Query(
			"SELECT blocked_task_id FROM task_dependencies WHERE blocker_task_id = $1",
			cur,
		)
		if err != nil {
			return false, fmt.Errorf("dfs query: %w", err)
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return false, fmt.Errorf("dfs scan: %w", err)
			}
			queue = append(queue, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return false, err
		}
	}

	return false, nil
}
