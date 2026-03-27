package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"kanbanboard/internal/model"
	"kanbanboard/internal/store"
	"kanbanboard/internal/validate"

	"golang.org/x/crypto/bcrypt"
)

type createUserRequest struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	IsAdmin       bool   `json:"isAdmin"`
	IsTeamManager bool   `json:"isTeamManager"`
}

type updateUserAdminRequest struct {
	Name          *string `json:"name"`
	Email         *string `json:"email"`
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
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

		if req.Name == "" || req.Email == "" || req.Password == "" {
			writeError(w, http.StatusBadRequest, "Name, email, and password are required")
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

		user := model.User{
			Name:          req.Name,
			Email:         req.Email,
			PasswordHash:  string(hash),
			IsAdmin:       req.IsAdmin,
			IsTeamManager: req.IsTeamManager,
			IsActive:      true,
		}

		user, err = store.CreateUser(db, user)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
