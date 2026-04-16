package store

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"kanbanboard/internal/model"
)

// ErrTokenNotFound is returned when an API token is not found.
var ErrTokenNotFound = errors.New("token not found")

// generateTokenBytes produces 32 cryptographically random bytes encoded as a 64-char hex string.
func generateTokenBytes() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// hashToken returns the hex-encoded SHA-256 hash of a raw token string.
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// CreateApiToken generates a new API token for a user, stores its hash, and returns
// the token metadata together with the raw token string (shown only once).
func CreateApiToken(db *sql.DB, userID, name string) (model.ApiToken, string, error) {
	raw, err := generateTokenBytes()
	if err != nil {
		return model.ApiToken{}, "", err
	}
	hash := hashToken(raw)

	var t model.ApiToken
	err = db.QueryRow(`
		INSERT INTO api_tokens (user_id, name, token_hash)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, name, created_at, last_used_at
	`, userID, name, hash).Scan(&t.ID, &t.UserID, &t.Name, &t.CreatedAt, &t.LastUsedAt)
	if err != nil {
		return model.ApiToken{}, "", fmt.Errorf("create api token: %w", err)
	}

	return t, raw, nil
}

// GetUserByToken looks up a user by a raw Bearer token value.
// Updates last_used_at on successful lookup. Returns ErrTokenNotFound if no match.
func GetUserByToken(db *sql.DB, rawToken string) (model.User, error) {
	hash := hashToken(rawToken)

	var userID string
	var tokenID string
	err := db.QueryRow(
		"SELECT id, user_id FROM api_tokens WHERE token_hash = $1",
		hash,
	).Scan(&tokenID, &userID)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrTokenNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("lookup token: %w", err)
	}

	// Update last_used_at asynchronously (best-effort — don't fail the request)
	_, _ = db.Exec("UPDATE api_tokens SET last_used_at = NOW() WHERE id = $1", tokenID)

	user, err := GetUserByID(db, userID)
	if err != nil {
		return model.User{}, fmt.Errorf("get token user: %w", err)
	}
	return user, nil
}

// ListApiTokens returns all tokens for a user (metadata only, no hashes).
func ListApiTokens(db *sql.DB, userID string) ([]model.ApiToken, error) {
	rows, err := db.Query(`
		SELECT id, user_id, name, created_at, last_used_at
		FROM api_tokens WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list api tokens: %w", err)
	}
	defer rows.Close()

	var tokens []model.ApiToken
	for rows.Next() {
		var t model.ApiToken
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.CreatedAt, &t.LastUsedAt); err != nil {
			return nil, fmt.Errorf("scan api token: %w", err)
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

// DeleteApiToken removes a token by ID, scoped to a specific user (prevents cross-user deletion).
func DeleteApiToken(db *sql.DB, tokenID, userID string) error {
	result, err := db.Exec(
		"DELETE FROM api_tokens WHERE id = $1 AND user_id = $2",
		tokenID, userID,
	)
	if err != nil {
		return fmt.Errorf("delete api token: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrTokenNotFound
	}
	return nil
}
