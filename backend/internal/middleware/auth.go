package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"kanbanboard/internal/model"
	"kanbanboard/internal/store"
)

type contextKey string

const userContextKey contextKey = "user"

// RequireAuth is middleware that accepts either a Bearer API token or a session cookie.
// Bearer token takes precedence when the Authorization header is present.
// If authentication fails, returns 401.
func RequireAuth(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := resolveUser(db, r)
		if err != nil {
			writeUnauthorized(w)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next(w, r.WithContext(ctx))
	}
}

// resolveUser extracts the authenticated user from the request.
// Checks Bearer token first, then session cookie.
func resolveUser(db *sql.DB, r *http.Request) (model.User, error) {
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		raw := strings.TrimPrefix(auth, "Bearer ")
		user, err := store.GetUserByToken(db, raw)
		if err != nil {
			return model.User{}, err
		}
		if !user.IsActive || user.DeletedAt != nil {
			return model.User{}, errors.New("account inactive")
		}
		return user, nil
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		return model.User{}, errors.New("no credentials")
	}

	session, err := store.GetSession(db, cookie.Value)
	if err != nil {
		return model.User{}, err
	}

	user, err := store.GetUserByID(db, session.UserID)
	if err != nil || !user.IsActive || user.DeletedAt != nil {
		return model.User{}, errors.New("invalid session")
	}
	return user, nil
}

// RequireAdmin wraps RequireAuth and additionally checks that the user is an admin.
func RequireAdmin(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return RequireAuth(db, func(w http.ResponseWriter, r *http.Request) {
		user, _ := UserFromContext(r.Context())
		if !user.IsAdmin {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Admin access required"})
			return
		}
		next(w, r)
	})
}

// UserFromContext retrieves the authenticated user from the request context.
func UserFromContext(ctx context.Context) (model.User, bool) {
	user, ok := ctx.Value(userContextKey).(model.User)
	return user, ok
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": "Not authenticated"})
}
