package store

import (
	"database/sql"
	"errors"
	"fmt"

	"kanbanboard/internal/model"
)

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")

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
		INSERT INTO users (name, email, password_hash, is_admin, is_team_manager, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`, user.Name, user.Email, user.PasswordHash, user.IsAdmin, user.IsTeamManager, user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return model.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email address.
func GetUserByEmail(db *sql.DB, email string) (model.User, error) {
	var u model.User
	err := db.QueryRow(`
		SELECT id, name, email, password_hash, is_admin, is_team_manager, is_active, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.IsTeamManager, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

// GetUserByID retrieves a user by ID.
func GetUserByID(db *sql.DB, id string) (model.User, error) {
	var u model.User
	err := db.QueryRow(`
		SELECT id, name, email, password_hash, is_admin, is_team_manager, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.IsTeamManager, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// ListUsers returns all users.
func ListUsers(db *sql.DB) ([]model.User, error) {
	rows, err := db.Query(`
		SELECT id, name, email, password_hash, is_admin, is_team_manager, is_active, created_at, updated_at
		FROM users ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.IsTeamManager, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// UpdateUserAdmin updates a user's name, email, active status, and roles (admin operation).
func UpdateUserAdmin(db *sql.DB, user model.User) (model.User, error) {
	err := db.QueryRow(`
		UPDATE users SET name = $1, email = $2, is_active = $3, is_admin = $4, is_team_manager = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`, user.Name, user.Email, user.IsActive, user.IsAdmin, user.IsTeamManager, user.ID).Scan(&user.UpdatedAt)
	if err != nil {
		return model.User{}, fmt.Errorf("update user admin: %w", err)
	}
	return user, nil
}

// UpdateUser updates a user's name and email.
func UpdateUser(db *sql.DB, user model.User) (model.User, error) {
	err := db.QueryRow(`
		UPDATE users SET name = $1, email = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at
	`, user.Name, user.Email, user.ID).Scan(&user.UpdatedAt)
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
