package store

import (
	"database/sql"
	"errors"
	"fmt"

	"kanbanboard/internal/model"
)

// ErrProjectNotFound is returned when a project is not found.
var ErrProjectNotFound = errors.New("project not found")

// CreateProject inserts a new project and returns it with the generated ID and timestamps.
func CreateProject(db *sql.DB, project model.Project) (model.Project, error) {
	err := db.QueryRow(`
		INSERT INTO projects (name, visibility, owner_user_id, owner_team_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`, project.Name, project.Visibility, project.OwnerUserID, project.OwnerTeamID,
	).Scan(&project.ID, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		return model.Project{}, fmt.Errorf("create project: %w", err)
	}
	return project, nil
}

// CreateDefaultColumns inserts the default columns for a new project.
func CreateDefaultColumns(db *sql.DB, projectID string) error {
	defaults := []struct {
		name     string
		position int
	}{
		{"Inbox", 0},
		{"Todo", 1},
		{"In Progress", 2},
		{"Blocked", 3},
		{"Done", 4},
	}

	for _, col := range defaults {
		_, err := db.Exec(
			"INSERT INTO columns (project_id, name, position) VALUES ($1, $2, $3)",
			projectID, col.name, col.position,
		)
		if err != nil {
			return fmt.Errorf("create default column %s: %w", col.name, err)
		}
	}
	return nil
}

// CreateDefaultLabels inserts the default labels for a new project.
func CreateDefaultLabels(db *sql.DB, projectID string) error {
	defaults := []struct {
		name  string
		color string
	}{
		{"task", "#4a90d9"},
		{"bug", "#e53e3e"},
		{"feature", "#38a169"},
		{"chore", "#718096"},
	}

	for _, lbl := range defaults {
		_, err := db.Exec(
			"INSERT INTO labels (project_id, name, color) VALUES ($1, $2, $3)",
			projectID, lbl.name, lbl.color,
		)
		if err != nil {
			return fmt.Errorf("create default label %s: %w", lbl.name, err)
		}
	}
	return nil
}

// ListProjectsForUser returns all projects visible to the given user:
// - Projects owned by the user
// - Projects owned by teams the user belongs to
// - Public projects (visible to everyone)
func ListProjectsForUser(db *sql.DB, userID string) ([]model.Project, error) {
	rows, err := db.Query(`
		SELECT DISTINCT p.id, p.name, p.visibility, p.owner_user_id, p.owner_team_id, p.created_at, p.updated_at
		FROM projects p
		LEFT JOIN team_members tm ON p.owner_team_id = tm.team_id
		WHERE p.owner_user_id = $1
		   OR tm.user_id = $1
		   OR p.visibility = 'public'
		ORDER BY p.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Visibility, &p.OwnerUserID, &p.OwnerTeamID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// GetProject retrieves a project by ID.
func GetProject(db *sql.DB, projectID string) (model.Project, error) {
	var p model.Project
	err := db.QueryRow(`
		SELECT id, name, visibility, owner_user_id, owner_team_id, created_at, updated_at
		FROM projects WHERE id = $1
	`, projectID).Scan(&p.ID, &p.Name, &p.Visibility, &p.OwnerUserID, &p.OwnerTeamID, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Project{}, ErrProjectNotFound
	}
	if err != nil {
		return model.Project{}, fmt.Errorf("get project: %w", err)
	}
	return p, nil
}

// GetColumnsForProject returns all columns for a project, ordered by position.
func GetColumnsForProject(db *sql.DB, projectID string) ([]model.Column, error) {
	rows, err := db.Query(
		"SELECT id, project_id, name, position FROM columns WHERE project_id = $1 ORDER BY position",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}
	defer rows.Close()

	var columns []model.Column
	for rows.Next() {
		var c model.Column
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.Name, &c.Position); err != nil {
			return nil, fmt.Errorf("scan column: %w", err)
		}
		columns = append(columns, c)
	}
	return columns, rows.Err()
}

// GetLabelsForProject returns all labels for a project.
func GetLabelsForProject(db *sql.DB, projectID string) ([]model.Label, error) {
	rows, err := db.Query(
		"SELECT id, project_id, name, color FROM labels WHERE project_id = $1 ORDER BY name",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("get labels: %w", err)
	}
	defer rows.Close()

	var labels []model.Label
	for rows.Next() {
		var l model.Label
		if err := rows.Scan(&l.ID, &l.ProjectID, &l.Name, &l.Color); err != nil {
			return nil, fmt.Errorf("scan label: %w", err)
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

// GetDefaultLabelForProject returns the "task" label for a project.
func GetDefaultLabelForProject(db *sql.DB, projectID string) (model.Label, error) {
	var l model.Label
	err := db.QueryRow(
		"SELECT id, project_id, name, color FROM labels WHERE project_id = $1 AND name = 'task'",
		projectID,
	).Scan(&l.ID, &l.ProjectID, &l.Name, &l.Color)
	if err != nil {
		return model.Label{}, fmt.Errorf("get default label: %w", err)
	}
	return l, nil
}
