package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"kanbanboard/internal/model"
	"kanbanboard/internal/store"
)

type contextKey string

const userContextKey contextKey = "user"

// RequireAuth is middleware that checks for a valid session cookie.
// If valid, the user is added to the request context.
// If invalid or missing, returns 401.
func RequireAuth(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			writeUnauthorized(w)
			return
		}

		session, err := store.GetSession(db, cookie.Value)
		if err != nil {
			writeUnauthorized(w)
			return
		}

		user, err := store.GetUserByID(db, session.UserID)
		if err != nil || !user.IsActive {
			writeUnauthorized(w)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next(w, r.WithContext(ctx))
	}
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
