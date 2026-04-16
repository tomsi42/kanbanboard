package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"kanbanboard/internal/model"
	"kanbanboard/internal/store"
	"kanbanboard/internal/validate"

	"golang.org/x/crypto/bcrypt"
)

const sessionDuration = 7 * 24 * time.Hour // 7 days
const cookieName = "session_token"

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// HandleLogin authenticates a user and creates a session.
// The login field accepts either a username or an email address.
func HandleLogin(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Login == "" || req.Password == "" {
			writeError(w, http.StatusBadRequest, "Username/email and password are required")
			return
		}

		// Find user by username or email
		user, err := store.GetUserByLogin(db, req.Login)
		if errors.Is(err, store.ErrUserNotFound) {
			writeError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to authenticate")
			return
		}

		// Check active
		if !user.IsActive {
			writeError(w, http.StatusUnauthorized, "Account is deactivated")
			return
		}

		// Check password
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			writeError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		// Create session
		token, err := store.CreateSession(db, user.ID, sessionDuration)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create session")
			return
		}

		// Set cookie
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int(sessionDuration.Seconds()),
		})

		writeJSON(w, http.StatusOK, user)
	}
}

// HandleLogout destroys the current session.
func HandleLogout(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err == nil {
			_ = store.DeleteSession(db, cookie.Value)
		}

		// Clear cookie
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
		})

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// HandleRegister creates a new regular user account when registration is enabled.
func HandleRegister(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enabled, err := store.GetSetting(db, "registration_enabled", "false")
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to check registration status")
			return
		}
		if enabled != "true" {
			writeError(w, http.StatusForbidden, "Registration is not enabled")
			return
		}

		var req registerRequest
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
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: string(hash),
			IsActive:     true,
		}

		user, err = store.CreateUser(db, user)
		if err != nil {
			if store.IsUniqueViolation(err) {
				writeError(w, http.StatusConflict, "Email is already in use")
				return
			}
			writeError(w, http.StatusInternalServerError, "Failed to create account")
			return
		}

		token, err := store.CreateSession(db, user.ID, sessionDuration)
		if err != nil {
			writeJSON(w, http.StatusCreated, user)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int(sessionDuration.Seconds()),
		})

		writeJSON(w, http.StatusCreated, user)
	}
}

// HandleMe returns the currently authenticated user.
func HandleMe(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Not authenticated")
			return
		}

		session, err := store.GetSession(db, cookie.Value)
		if err != nil {
			// Clear invalid cookie
			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   -1,
			})
			writeError(w, http.StatusUnauthorized, "Not authenticated")
			return
		}

		user, err := store.GetUserByID(db, session.UserID)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Not authenticated")
			return
		}

		writeJSON(w, http.StatusOK, user)
	}
}
