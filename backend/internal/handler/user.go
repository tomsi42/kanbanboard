package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"kanbanboard/internal/middleware"
	"kanbanboard/internal/store"
	"kanbanboard/internal/validate"

	"golang.org/x/crypto/bcrypt"
)

type updateProfileRequest struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Username *string `json:"username"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// HandleUpdateProfile updates the current user's name, email, and username.
func HandleUpdateProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())

		var req updateProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "Name is required")
			return
		}
		if req.Username != nil && *req.Username != "" {
			if msg := validate.Username(*req.Username); msg != "" {
				writeError(w, http.StatusBadRequest, msg)
				return
			}
		}

		user.Name = req.Name
		user.Email = req.Email
		if req.Username != nil {
			if *req.Username == "" {
				user.Username = nil
			} else {
				user.Username = req.Username
			}
		}

		user, err := store.UpdateUser(db, user)
		if err != nil {
			if store.IsUniqueViolation(err) {
				writeError(w, http.StatusConflict, "Email or username is already in use")
				return
			}
			writeError(w, http.StatusInternalServerError, "Failed to update profile")
			return
		}

		writeJSON(w, http.StatusOK, user)
	}
}

// HandleChangePassword changes the current user's password.
func HandleChangePassword(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())

		var req changePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.CurrentPassword == "" || req.NewPassword == "" {
			writeError(w, http.StatusBadRequest, "Current and new passwords are required")
			return
		}

		// Verify current password
		fullUser, err := store.GetUserByID(db, user.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to verify password")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(fullUser.PasswordHash), []byte(req.CurrentPassword)); err != nil {
			writeError(w, http.StatusUnauthorized, "Current password is incorrect")
			return
		}

		// Validate new password
		if msg := validate.Password(req.NewPassword); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		// Hash and save
		hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		if err := store.UpdatePassword(db, user.ID, string(hash)); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update password")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// HandleListUsersBasic returns a lightweight list of all active users (id, name, email, username).
func HandleListUsersBasic(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := store.ListActiveUsersBasic(db)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list users")
			return
		}

		if users == nil {
			users = []store.BasicUser{}
		}

		writeJSON(w, http.StatusOK, users)
	}
}
