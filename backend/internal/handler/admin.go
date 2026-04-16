package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"kanbanboard/internal/middleware"
	"kanbanboard/internal/model"
	"kanbanboard/internal/store"
	"kanbanboard/internal/validate"

	"golang.org/x/crypto/bcrypt"
)

type settingsResponse struct {
	RegistrationEnabled bool `json:"registrationEnabled"`
}

type updateSettingsRequest struct {
	RegistrationEnabled *bool `json:"registrationEnabled"`
}

// HandleGetSettings returns current admin-controlled settings.
func HandleGetSettings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enabled, err := store.GetSetting(db, "registration_enabled", "false")
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get settings")
			return
		}
		writeJSON(w, http.StatusOK, settingsResponse{RegistrationEnabled: enabled == "true"})
	}
}

// HandleUpdateSettings updates admin-controlled settings.
func HandleUpdateSettings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req updateSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.RegistrationEnabled != nil {
			val := "false"
			if *req.RegistrationEnabled {
				val = "true"
			}
			if err := store.SetSetting(db, "registration_enabled", val); err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to update settings")
				return
			}
		}

		enabled, err := store.GetSetting(db, "registration_enabled", "false")
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get settings")
			return
		}
		writeJSON(w, http.StatusOK, settingsResponse{RegistrationEnabled: enabled == "true"})
	}
}

type createUserRequest struct {
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Username      string  `json:"username"`
	Password      string  `json:"password"`
	IsAdmin       bool    `json:"isAdmin"`
	IsTeamManager bool    `json:"isTeamManager"`
}

type updateUserAdminRequest struct {
	Name          *string `json:"name"`
	Email         *string `json:"email"`
	Username      *string `json:"username"`
	IsActive      *bool   `json:"isActive"`
	IsAdmin       *bool   `json:"isAdmin"`
	IsTeamManager *bool   `json:"isTeamManager"`
}

type resetPasswordRequest struct {
	Password string `json:"password"`
}

// HandleListUsers returns all users (admin only).
func HandleListUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := store.ListUsers(db)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list users")
			return
		}

		if users == nil {
			users = []model.User{}
		}

		writeJSON(w, http.StatusOK, users)
	}
}

// HandleCreateUser creates a new user (admin only).
func HandleCreateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Name == "" || req.Password == "" {
			writeError(w, http.StatusBadRequest, "Name and password are required")
			return
		}
		if req.Email == "" && req.Username == "" {
			writeError(w, http.StatusBadRequest, "At least one of email or username is required")
			return
		}
		if req.Username != "" {
			if msg := validate.Username(req.Username); msg != "" {
				writeError(w, http.StatusBadRequest, msg)
				return
			}
		}

		if msg := validate.Password(req.Password); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		var username *string
		if req.Username != "" {
			username = &req.Username
		}

		user := model.User{
			Name:          req.Name,
			Email:         req.Email,
			Username:      username,
			PasswordHash:  string(hash),
			IsAdmin:       req.IsAdmin,
			IsTeamManager: req.IsTeamManager,
			IsActive:      true,
		}

		user, err = store.CreateUser(db, user)
		if err != nil {
			if store.IsUniqueViolation(err) {
				writeError(w, http.StatusConflict, "Email or username is already in use")
				return
			}
			writeError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		writeJSON(w, http.StatusCreated, user)
	}
}

// HandleUpdateUserAdmin updates a user's profile and roles (admin only).
func HandleUpdateUserAdmin(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("userId")

		user, err := store.GetUserByID(db, userID)
		if errors.Is(err, store.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get user")
			return
		}

		var req updateUserAdminRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Name != nil {
			user.Name = *req.Name
		}
		if req.Email != nil {
			user.Email = *req.Email
		}
		if req.Username != nil {
			if *req.Username == "" {
				user.Username = nil
			} else {
				if msg := validate.Username(*req.Username); msg != "" {
					writeError(w, http.StatusBadRequest, msg)
					return
				}
				user.Username = req.Username
			}
		}
		if req.IsActive != nil {
			user.IsActive = *req.IsActive
		}
		if req.IsAdmin != nil {
			user.IsAdmin = *req.IsAdmin
		}
		if req.IsTeamManager != nil {
			user.IsTeamManager = *req.IsTeamManager
		}

		user, err = store.UpdateUserAdmin(db, user)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}

		writeJSON(w, http.StatusOK, user)
	}
}

// HandleResetPassword resets a user's password (admin only, no current password needed).
func HandleResetPassword(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("userId")

		var req resetPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Password == "" {
			writeError(w, http.StatusBadRequest, "Password is required")
			return
		}

		if msg := validate.Password(req.Password); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		if err := store.UpdatePassword(db, userID, string(hash)); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to reset password")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// HandleDeleteUserImpact returns the impact of deleting a user.
func HandleDeleteUserImpact(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin, _ := middleware.UserFromContext(r.Context())
		userID := r.PathValue("userId")

		user, err := store.GetUserByID(db, userID)
		if errors.Is(err, store.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get user")
			return
		}
		if user.DeletedAt != nil {
			writeError(w, http.StatusConflict, "User is already deleted")
			return
		}

		impact, err := store.GetDeleteUserImpact(db, userID, admin.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to calculate impact")
			return
		}

		writeJSON(w, http.StatusOK, impact)
	}
}

// HandleDeleteUser soft-deletes a user (admin only).
// All cascade operations run in a single transaction.
func HandleDeleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin, _ := middleware.UserFromContext(r.Context())
		userID := r.PathValue("userId")

		if userID == admin.ID {
			writeError(w, http.StatusConflict, "You cannot delete yourself")
			return
		}

		user, err := store.GetUserByID(db, userID)
		if errors.Is(err, store.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get user")
			return
		}
		if user.DeletedAt != nil {
			writeError(w, http.StatusConflict, "User is already deleted")
			return
		}

		if err := store.DeleteUserCascade(db, userID, admin.ID); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete user")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
