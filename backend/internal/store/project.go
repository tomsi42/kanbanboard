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
		LEFT JOIN teams t ON p.owner_team_id = t.id
		WHERE p.owner_user_id = $1
		   OR tm.user_id = $1
		   OR t.owner_id = $1
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

// GetProjectMembers returns users who can work on a project.
// For user-owned projects: just the owner.
// For team-owned projects: team owner + all team members.
func GetProjectMembers(db *sql.DB, project model.Project) ([]model.User, error) {
	if project.OwnerUserID != nil {
		user, err := GetUserByID(db, *project.OwnerUserID)
		if err != nil {
			return nil, err
		}
		return []model.User{user}, nil
	}

	if project.OwnerTeamID != nil {
		// Get team owner + members
		team, err := GetTeam(db, *project.OwnerTeamID)
		if err != nil {
			return nil, err
		}

		members, err := ListTeamMembers(db, *project.OwnerTeamID)
		if err != nil {
			return nil, err
		}

		// Add team owner if not already a member
		ownerIncluded := false
		for _, m := range members {
			if m.ID == team.OwnerID {
				ownerIncluded = true
				break
			}
		}
		if !ownerIncluded {
			owner, err := GetUserByID(db, team.OwnerID)
			if err == nil {
				members = append([]model.User{owner}, members...)
			}
		}

		return members, nil
	}

	return []model.User{}, nil
}

// UpdateProject updates a project's name and visibility.
func UpdateProject(db *sql.DB, project model.Project) (model.Project, error) {
	err := db.QueryRow(`
		UPDATE projects SET name = $1, visibility = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at
	`, project.Name, project.Visibility, project.ID).Scan(&project.UpdatedAt)
	if err != nil {
		return model.Project{}, fmt.Errorf("update project: %w", err)
	}
	return project, nil
}

// CreateColumn adds a new column at the end of the project.
func CreateColumn(db *sql.DB, col model.Column) (model.Column, error) {
	// Get next position
	var maxPos sql.NullInt64
	err := db.QueryRow(
		"SELECT MAX(position) FROM columns WHERE project_id = $1", col.ProjectID,
	).Scan(&maxPos)
	if err != nil {
		return model.Column{}, fmt.Errorf("get max column position: %w", err)
	}
	col.Position = 0
	if maxPos.Valid {
		col.Position = int(maxPos.Int64) + 1
	}

	err = db.QueryRow(`
		INSERT INTO columns (project_id, name, position) VALUES ($1, $2, $3)
		RETURNING id
	`, col.ProjectID, col.Name, col.Position).Scan(&col.ID)
	if err != nil {
		return model.Column{}, fmt.Errorf("create column: %w", err)
	}
	return col, nil
}

// UpdateColumn renames a column.
func UpdateColumn(db *sql.DB, col model.Column) (model.Column, error) {
	_, err := db.Exec("UPDATE columns SET name = $1 WHERE id = $2", col.Name, col.ID)
	if err != nil {
		return model.Column{}, fmt.Errorf("update column: %w", err)
	}
	return col, nil
}

// DeleteColumn removes a column by ID.
func DeleteColumn(db *sql.DB, columnID string) error {
	_, err := db.Exec("DELETE FROM columns WHERE id = $1", columnID)
	if err != nil {
		return fmt.Errorf("delete column: %w", err)
	}
	return nil
}

// ReorderColumns sets column positions from an ordered array of IDs.
func ReorderColumns(db *sql.DB, projectID string, columnIDs []string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	for i, id := range columnIDs {
		_, err := tx.Exec(
			"UPDATE columns SET position = $1 WHERE id = $2 AND project_id = $3",
			i, id, projectID,
		)
		if err != nil {
			return fmt.Errorf("reorder column: %w", err)
		}
	}
	return tx.Commit()
}

// CountTasksInColumn returns the number of tasks in a column.
func CountTasksInColumn(db *sql.DB, columnID string) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM tasks WHERE column_id = $1", columnID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count tasks in column: %w", err)
	}
	return count, nil
}

// CreateLabel adds a new label to a project.
func CreateLabel(db *sql.DB, label model.Label) (model.Label, error) {
	err := db.QueryRow(`
		INSERT INTO labels (project_id, name, color) VALUES ($1, $2, $3)
		RETURNING id
	`, label.ProjectID, label.Name, label.Color).Scan(&label.ID)
	if err != nil {
		return model.Label{}, fmt.Errorf("create label: %w", err)
	}
	return label, nil
}

// UpdateLabel updates a label's name and color.
func UpdateLabel(db *sql.DB, label model.Label) (model.Label, error) {
	_, err := db.Exec("UPDATE labels SET name = $1, color = $2 WHERE id = $3",
		label.Name, label.Color, label.ID)
	if err != nil {
		return model.Label{}, fmt.Errorf("update label: %w", err)
	}
	return label, nil
}

// DeleteLabel removes a label by ID.
func DeleteLabel(db *sql.DB, labelID string) error {
	_, err := db.Exec("DELETE FROM labels WHERE id = $1", labelID)
	if err != nil {
		return fmt.Errorf("delete label: %w", err)
	}
	return nil
}

// CountTasksWithLabel returns the number of tasks using a label.
func CountTasksWithLabel(db *sql.DB, labelID string) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM tasks WHERE label_id = $1", labelID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count tasks with label: %w", err)
	}
	return count, nil
}
