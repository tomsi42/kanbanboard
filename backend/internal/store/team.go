package store

import (
	"database/sql"
	"errors"
	"fmt"

	"kanbanboard/internal/model"
)

// ErrTeamNotFound is returned when a team is not found.
var ErrTeamNotFound = errors.New("team not found")

// IsTeamMember checks if a user is a member of a team.
func IsTeamMember(db *sql.DB, teamID, userID string) (bool, error) {
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2)",
		teamID, userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check team membership: %w", err)
	}
	return exists, nil
}

// CreateTeam inserts a new team.
func CreateTeam(db *sql.DB, team model.Team) (model.Team, error) {
	err := db.QueryRow(`
		INSERT INTO teams (name, owner_id) VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`, team.Name, team.OwnerID).Scan(&team.ID, &team.CreatedAt, &team.UpdatedAt)
	if err != nil {
		return model.Team{}, fmt.Errorf("create team: %w", err)
	}
	return team, nil
}

// ListTeamsForUser returns all teams owned by a user.
func ListTeamsForUser(db *sql.DB, userID string) ([]model.Team, error) {
	rows, err := db.Query(`
		SELECT id, name, owner_id, created_at, updated_at
		FROM teams WHERE owner_id = $1 ORDER BY name
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}
	defer rows.Close()

	var teams []model.Team
	for rows.Next() {
		var t model.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.OwnerID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan team: %w", err)
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

// GetTeam retrieves a team by ID.
func GetTeam(db *sql.DB, teamID string) (model.Team, error) {
	var t model.Team
	err := db.QueryRow(`
		SELECT id, name, owner_id, created_at, updated_at
		FROM teams WHERE id = $1
	`, teamID).Scan(&t.ID, &t.Name, &t.OwnerID, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Team{}, ErrTeamNotFound
	}
	if err != nil {
		return model.Team{}, fmt.Errorf("get team: %w", err)
	}
	return t, nil
}

// UpdateTeam updates a team's name.
func UpdateTeam(db *sql.DB, team model.Team) (model.Team, error) {
	err := db.QueryRow(`
		UPDATE teams SET name = $1, updated_at = NOW() WHERE id = $2
		RETURNING updated_at
	`, team.Name, team.ID).Scan(&team.UpdatedAt)
	if err != nil {
		return model.Team{}, fmt.Errorf("update team: %w", err)
	}
	return team, nil
}

// DeleteTeam removes a team by ID.
func DeleteTeam(db *sql.DB, teamID string) error {
	_, err := db.Exec("DELETE FROM teams WHERE id = $1", teamID)
	if err != nil {
		return fmt.Errorf("delete team: %w", err)
	}
	return nil
}

// ListTeamMembers returns all members of a team.
func ListTeamMembers(db *sql.DB, teamID string) ([]model.User, error) {
	rows, err := db.Query(`
		SELECT u.id, u.name, u.email, u.password_hash, u.is_admin, u.is_team_manager, u.is_active, u.created_at, u.updated_at
		FROM users u
		JOIN team_members tm ON u.id = tm.user_id
		WHERE tm.team_id = $1
		ORDER BY u.name
	`, teamID)
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.IsTeamManager, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// AddTeamMember adds a user to a team.
func AddTeamMember(db *sql.DB, teamID, userID string) error {
	_, err := db.Exec(
		"INSERT INTO team_members (team_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		teamID, userID,
	)
	if err != nil {
		return fmt.Errorf("add team member: %w", err)
	}
	return nil
}

// RemoveTeamMember removes a user from a team.
func RemoveTeamMember(db *sql.DB, teamID, userID string) error {
	_, err := db.Exec("DELETE FROM team_members WHERE team_id = $1 AND user_id = $2", teamID, userID)
	if err != nil {
		return fmt.Errorf("remove team member: %w", err)
	}
	return nil
}

// CountProjectsForTeam returns the number of projects owned by a team.
func CountProjectsForTeam(db *sql.DB, teamID string) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM projects WHERE owner_team_id = $1", teamID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count team projects: %w", err)
	}
	return count, nil
}
