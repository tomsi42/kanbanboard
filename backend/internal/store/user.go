package store

import (
	"database/sql"
	"errors"
	"fmt"

	"kanbanboard/internal/model"
)

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")

// scanUser scans a full user row (id, name, email, username, password_hash, is_admin,
// is_team_manager, is_active, deleted_at, created_at, updated_at).
func scanUser(row interface {
	Scan(dest ...any) error
}, u *model.User) error {
	var emailNS sql.NullString
	err := row.Scan(
		&u.ID, &u.Name, &emailNS, &u.Username,
		&u.PasswordHash, &u.IsAdmin, &u.IsTeamManager, &u.IsActive,
		&u.DeletedAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return err
	}
	u.Email = emailNS.String
	return nil
}

// nullString converts an empty Go string to a NULL sql value; non-empty strings are passed through.
func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// CountUsers returns the total number of users in the database.
func CountUsers(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

// CreateUser inserts a new user and returns it with the generated ID and timestamps.
func CreateUser(db *sql.DB, user model.User) (model.User, error) {
	err := db.QueryRow(`
		INSERT INTO users (name, email, username, password_hash, is_admin, is_team_manager, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`, user.Name, nullString(user.Email), user.Username,
		user.PasswordHash, user.IsAdmin, user.IsTeamManager, user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return model.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// GetUserByLogin retrieves a user by email address or username (case-insensitive).
// Excludes soft-deleted users (prevents login as deleted user).
func GetUserByLogin(db *sql.DB, login string) (model.User, error) {
	var u model.User
	err := scanUser(db.QueryRow(`
		SELECT id, name, email, username, password_hash, is_admin, is_team_manager, is_active, deleted_at, created_at, updated_at
		FROM users
		WHERE deleted_at IS NULL
		  AND (email = $1 OR lower(username) = lower($1))
	`, login), &u)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("get user by login: %w", err)
	}
	return u, nil
}

// GetUserByEmail retrieves a user by email address.
// Excludes soft-deleted users (prevents login as deleted user).
func GetUserByEmail(db *sql.DB, email string) (model.User, error) {
	var u model.User
	err := scanUser(db.QueryRow(`
		SELECT id, name, email, username, password_hash, is_admin, is_team_manager, is_active, deleted_at, created_at, updated_at
		FROM users WHERE email = $1 AND deleted_at IS NULL
	`, email), &u)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

// GetUserByID retrieves a user by ID.
// Includes soft-deleted users (needed for displaying creator/author names).
func GetUserByID(db *sql.DB, id string) (model.User, error) {
	var u model.User
	err := scanUser(db.QueryRow(`
		SELECT id, name, email, username, password_hash, is_admin, is_team_manager, is_active, deleted_at, created_at, updated_at
		FROM users WHERE id = $1
	`, id), &u)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// ListUsers returns all users including soft-deleted (for admin listing).
func ListUsers(db *sql.DB) ([]model.User, error) {
	rows, err := db.Query(`
		SELECT id, name, email, username, password_hash, is_admin, is_team_manager, is_active, deleted_at, created_at, updated_at
		FROM users ORDER BY (deleted_at IS NOT NULL), name
	`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := scanUser(rows, &u); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// BasicUser holds the minimal user fields for listings.
type BasicUser struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Username *string `json:"username,omitempty"`
}

// ListActiveUsersBasic returns active, non-deleted users with only id, name, email, and username.
func ListActiveUsersBasic(db *sql.DB) ([]BasicUser, error) {
	rows, err := db.Query(`
		SELECT id, name, email, username FROM users WHERE is_active = true AND deleted_at IS NULL ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("list active users: %w", err)
	}
	defer rows.Close()

	var users []BasicUser
	for rows.Next() {
		var u BasicUser
		var emailNS sql.NullString
		if err := rows.Scan(&u.ID, &u.Name, &emailNS, &u.Username); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		u.Email = emailNS.String
		users = append(users, u)
	}
	return users, rows.Err()
}

// UpdateUserAdmin updates a user's name, email, username, active status, and roles (admin operation).
func UpdateUserAdmin(db *sql.DB, user model.User) (model.User, error) {
	err := db.QueryRow(`
		UPDATE users SET name = $1, email = $2, username = $3, is_active = $4, is_admin = $5, is_team_manager = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at
	`, user.Name, nullString(user.Email), user.Username,
		user.IsActive, user.IsAdmin, user.IsTeamManager, user.ID,
	).Scan(&user.UpdatedAt)
	if err != nil {
		return model.User{}, fmt.Errorf("update user admin: %w", err)
	}
	return user, nil
}

// UpdateUser updates a user's name, email, and username.
func UpdateUser(db *sql.DB, user model.User) (model.User, error) {
	err := db.QueryRow(`
		UPDATE users SET name = $1, email = $2, username = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`, user.Name, nullString(user.Email), user.Username, user.ID,
	).Scan(&user.UpdatedAt)
	if err != nil {
		return model.User{}, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

// UpdatePassword updates a user's password hash.
func UpdatePassword(db *sql.DB, userID, passwordHash string) error {
	_, err := db.Exec(
		"UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2",
		passwordHash, userID,
	)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

// ResolveNewTeamOwner picks the next owner for a team being transferred.
// Returns the first active, non-deleted member that isn't the excluded user.
// Falls back to fallbackID (typically the admin performing the action).
func ResolveNewTeamOwner(members []model.User, excludeUserID, fallbackID string) string {
	for _, m := range members {
		if m.ID != excludeUserID && m.DeletedAt == nil {
			return m.ID
		}
	}
	return fallbackID
}

// DeleteUserImpact holds the impact data for deleting a user.
type DeleteUserImpact struct {
	ProjectCount int                `json:"projectCount"`
	TaskCount    int                `json:"taskCount"`
	TeamCount    int                `json:"teamCount"`
	Transfers    []TeamTransferInfo `json:"teamTransfers"`
}

// TeamTransferInfo describes a team ownership transfer.
type TeamTransferInfo struct {
	TeamName string `json:"teamName"`
	NewOwner string `json:"newOwner"`
}

// GetDeleteUserImpact calculates the impact of deleting a user without making changes.
func GetDeleteUserImpact(db *sql.DB, userID, adminID string) (DeleteUserImpact, error) {
	var impact DeleteUserImpact

	// Count projects
	rows, err := db.Query("SELECT id FROM projects WHERE owner_user_id = $1", userID)
	if err != nil {
		return impact, fmt.Errorf("list user projects: %w", err)
	}
	defer rows.Close()

	var projectIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return impact, fmt.Errorf("scan project id: %w", err)
		}
		projectIDs = append(projectIDs, id)
	}
	if err := rows.Err(); err != nil {
		return impact, err
	}
	impact.ProjectCount = len(projectIDs)

	// Count tasks across owned projects
	for _, pid := range projectIDs {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM tasks WHERE project_id = $1", pid).Scan(&count); err != nil {
			return impact, fmt.Errorf("count tasks: %w", err)
		}
		impact.TaskCount += count
	}

	// Find teams and plan transfers
	teams, err := ListTeamsForUser(db, userID)
	if err != nil {
		return impact, fmt.Errorf("list teams: %w", err)
	}
	impact.TeamCount = len(teams)

	adminUser, err := GetUserByID(db, adminID)
	if err != nil {
		return impact, fmt.Errorf("get admin: %w", err)
	}

	for _, team := range teams {
		members, err := ListTeamMembers(db, team.ID)
		if err != nil {
			return impact, fmt.Errorf("list team members: %w", err)
		}
		newOwnerID := ResolveNewTeamOwner(members, userID, adminID)
		newOwnerName := adminUser.Name
		for _, m := range members {
			if m.ID == newOwnerID {
				newOwnerName = m.Name
				break
			}
		}
		impact.Transfers = append(impact.Transfers, TeamTransferInfo{
			TeamName: team.Name,
			NewOwner: newOwnerName,
		})
	}

	return impact, nil
}

// DeleteUserCascade performs all steps of user deletion in a single transaction:
// transfer teams, delete owned projects, unassign tasks, remove memberships, soft delete user.
func DeleteUserCascade(db *sql.DB, userID, adminID string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Transfer team ownership
	rows, err := tx.Query("SELECT id FROM teams WHERE owner_id = $1", userID)
	if err != nil {
		return fmt.Errorf("list owned teams: %w", err)
	}
	var teamIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return fmt.Errorf("scan team id: %w", err)
		}
		teamIDs = append(teamIDs, id)
	}
	rows.Close()

	for _, teamID := range teamIDs {
		// Find members for this team
		memberRows, err := tx.Query(`
			SELECT u.id, u.deleted_at FROM users u
			JOIN team_members tm ON u.id = tm.user_id
			WHERE tm.team_id = $1
		`, teamID)
		if err != nil {
			return fmt.Errorf("list team members: %w", err)
		}

		newOwnerID := adminID
		for memberRows.Next() {
			var mID string
			var mDeletedAt *any
			if err := memberRows.Scan(&mID, &mDeletedAt); err != nil {
				memberRows.Close()
				return fmt.Errorf("scan member: %w", err)
			}
			if mID != userID && mDeletedAt == nil {
				newOwnerID = mID
				break
			}
		}
		memberRows.Close()

		_, err = tx.Exec("UPDATE teams SET owner_id = $1, updated_at = NOW() WHERE id = $2", newOwnerID, teamID)
		if err != nil {
			return fmt.Errorf("transfer team: %w", err)
		}
	}

	// 2. Delete owned projects (cascade deletes columns, labels, tasks, comments)
	_, err = tx.Exec("DELETE FROM projects WHERE owner_user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("delete user projects: %w", err)
	}

	// 3. Unassign tasks assigned to user (in other people's projects)
	_, err = tx.Exec("UPDATE tasks SET assignee_id = NULL, updated_at = NOW() WHERE assignee_id = $1", userID)
	if err != nil {
		return fmt.Errorf("unassign tasks: %w", err)
	}

	// 4. Remove from all team memberships
	_, err = tx.Exec("DELETE FROM team_members WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("remove from teams: %w", err)
	}

	// 5. Soft delete the user (clear email and username to free them for reuse)
	_, err = tx.Exec(`
		UPDATE users SET deleted_at = NOW(), email = NULL, username = NULL, password_hash = '', is_active = false, updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}

	return tx.Commit()
}

// SoftDeleteUser marks a user as deleted, clears their email and username, and deactivates them.
// For standalone use in tests. Production code uses DeleteUserCascade.
func SoftDeleteUser(db *sql.DB, userID string) error {
	_, err := db.Exec(`
		UPDATE users SET deleted_at = NOW(), email = NULL, username = NULL, password_hash = '', is_active = false, updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	return nil
}

// UnassignTasksForUser clears the assignee on all tasks assigned to a user.
func UnassignTasksForUser(db *sql.DB, userID string) error {
	_, err := db.Exec(
		"UPDATE tasks SET assignee_id = NULL, updated_at = NOW() WHERE assignee_id = $1",
		userID,
	)
	if err != nil {
		return fmt.Errorf("unassign tasks: %w", err)
	}
	return nil
}

// ListProjectsOwnedByUser returns all projects directly owned by a user.
func ListProjectsOwnedByUser(db *sql.DB, userID string) ([]model.Project, error) {
	rows, err := db.Query(`
		SELECT id, name, visibility, tag, next_task_number, owner_user_id, owner_team_id, created_at, updated_at
		FROM projects WHERE owner_user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list projects owned by user: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Visibility, &p.Tag, &p.NextTaskNumber, &p.OwnerUserID, &p.OwnerTeamID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}
