package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"kanbanboard/internal/middleware"
	"kanbanboard/internal/model"
	"kanbanboard/internal/store"
)

type createTokenRequest struct {
	Name string `json:"name"`
}

type createTokenResponse struct {
	Token string `json:"token"` // raw token — shown once only
	ID    string `json:"id"`
	Name  string `json:"name"`
}

// HandleCreateTokenForUser creates an API token for a specific user (admin only).
// POST /api/v1/admin/users/{userId}/tokens
func HandleCreateTokenForUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("userId")

		var req createTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "Token name is required")
			return
		}

		t, raw, err := store.CreateApiToken(db, userID, req.Name)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create token")
			return
		}

		writeJSON(w, http.StatusCreated, createTokenResponse{Token: raw, ID: t.ID, Name: t.Name})
	}
}

// HandleListTokensForUser lists API tokens for a specific user (admin only).
// GET /api/v1/admin/users/{userId}/tokens
func HandleListTokensForUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("userId")

		tokens, err := store.ListApiTokens(db, userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list tokens")
			return
		}
		if tokens == nil {
			tokens = []model.ApiToken{}
		}
		writeJSON(w, http.StatusOK, tokens)
	}
}

// HandleDeleteTokenForUser revokes an API token (admin can delete any; users own only).
// DELETE /api/v1/admin/users/{userId}/tokens/{tokenId}
func HandleDeleteTokenForUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("userId")
		tokenID := r.PathValue("tokenId")

		err := store.DeleteApiToken(db, tokenID, userID)
		if errors.Is(err, store.ErrTokenNotFound) {
			writeError(w, http.StatusNotFound, "Token not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete token")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// HandleCreateOwnToken creates an API token for the authenticated user.
// POST /api/v1/users/me/tokens
func HandleCreateOwnToken(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())

		var req createTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "Token name is required")
			return
		}

		t, raw, err := store.CreateApiToken(db, user.ID, req.Name)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create token")
			return
		}

		writeJSON(w, http.StatusCreated, createTokenResponse{Token: raw, ID: t.ID, Name: t.Name})
	}
}

// HandleListOwnTokens lists API tokens for the authenticated user.
// GET /api/v1/users/me/tokens
func HandleListOwnTokens(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())

		tokens, err := store.ListApiTokens(db, user.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list tokens")
			return
		}
		if tokens == nil {
			tokens = []model.ApiToken{}
		}
		writeJSON(w, http.StatusOK, tokens)
	}
}

// HandleDeleteOwnToken revokes one of the authenticated user's own tokens.
// DELETE /api/v1/users/me/tokens/{tokenId}
func HandleDeleteOwnToken(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())
		tokenID := r.PathValue("tokenId")

		err := store.DeleteApiToken(db, tokenID, user.ID)
		if errors.Is(err, store.ErrTokenNotFound) {
			writeError(w, http.StatusNotFound, "Token not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete token")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
