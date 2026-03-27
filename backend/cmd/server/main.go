package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"kanbanboard/internal/handler"
	"kanbanboard/internal/middleware"
	"kanbanboard/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Connect to database
	db, err := store.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to database")

	// Run migrations
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := store.RunMigrations(db, migrationsDir); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations complete")

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/v1/health", handleHealth(db))
	mux.HandleFunc("GET /api/v1/setup/status", handler.HandleSetupStatus(db))
	mux.HandleFunc("POST /api/v1/setup", handler.HandleSetup(db))
	mux.HandleFunc("GET /api/v1/app/title", handler.HandleAppTitle(db))
	mux.HandleFunc("POST /api/v1/auth/login", handler.HandleLogin(db))
	mux.HandleFunc("POST /api/v1/auth/logout", handler.HandleLogout(db))
	mux.HandleFunc("GET /api/v1/auth/me", handler.HandleMe(db))

	// Auth middleware helper
	auth := func(h http.HandlerFunc) http.HandlerFunc { return middleware.RequireAuth(db, h) }

	// Admin routes (admin only)
	admin := func(h http.HandlerFunc) http.HandlerFunc { return middleware.RequireAdmin(db, h) }
	mux.HandleFunc("GET /api/v1/admin/users", admin(handler.HandleListUsers(db)))
	mux.HandleFunc("POST /api/v1/admin/users", admin(handler.HandleCreateUser(db)))
	mux.HandleFunc("PUT /api/v1/admin/users/{userId}", admin(handler.HandleUpdateUserAdmin(db)))
	mux.HandleFunc("PUT /api/v1/admin/users/{userId}/password", admin(handler.HandleResetPassword(db)))

	// User routes (auth required)
	mux.HandleFunc("PUT /api/v1/users/me", auth(handler.HandleUpdateProfile(db)))
	mux.HandleFunc("PUT /api/v1/users/me/password", auth(handler.HandleChangePassword(db)))

	// Project routes (auth required)
	mux.HandleFunc("POST /api/v1/projects", auth(handler.HandleCreateProject(db)))
	mux.HandleFunc("GET /api/v1/projects", auth(handler.HandleListProjects(db)))
	mux.HandleFunc("GET /api/v1/projects/{id}", auth(handler.HandleGetProject(db)))
	mux.HandleFunc("PUT /api/v1/projects/{id}", auth(handler.HandleUpdateProject(db)))
	mux.HandleFunc("POST /api/v1/projects/{id}/columns", auth(handler.HandleCreateColumn(db)))
	mux.HandleFunc("PUT /api/v1/projects/{id}/columns/reorder", auth(handler.HandleReorderColumns(db)))
	mux.HandleFunc("PUT /api/v1/projects/{id}/columns/{colId}", auth(handler.HandleUpdateColumn(db)))
	mux.HandleFunc("DELETE /api/v1/projects/{id}/columns/{colId}", auth(handler.HandleDeleteColumn(db)))
	mux.HandleFunc("POST /api/v1/projects/{id}/labels", auth(handler.HandleCreateLabel(db)))
	mux.HandleFunc("PUT /api/v1/projects/{id}/labels/{labelId}", auth(handler.HandleUpdateLabel(db)))
	mux.HandleFunc("DELETE /api/v1/projects/{id}/labels/{labelId}", auth(handler.HandleDeleteLabel(db)))
	mux.HandleFunc("POST /api/v1/projects/{projectId}/tasks", auth(handler.HandleCreateTask(db)))
	mux.HandleFunc("GET /api/v1/projects/{projectId}/tasks", auth(handler.HandleListTasks(db)))
	mux.HandleFunc("PUT /api/v1/projects/{projectId}/tasks/{taskId}", auth(handler.HandleUpdateTask(db)))
	mux.HandleFunc("PUT /api/v1/projects/{projectId}/tasks/{taskId}/move", auth(handler.HandleMoveTask(db)))
	mux.HandleFunc("DELETE /api/v1/projects/{projectId}/tasks/{taskId}", auth(handler.HandleDeleteTask(db)))
	mux.HandleFunc("GET /api/v1/projects/{projectId}/tasks/{taskId}/comments", auth(handler.HandleListComments(db)))
	mux.HandleFunc("POST /api/v1/projects/{projectId}/tasks/{taskId}/comments", auth(handler.HandleCreateComment(db)))
	mux.HandleFunc("PUT /api/v1/projects/{projectId}/tasks/{taskId}/comments/{commentId}", auth(handler.HandleUpdateComment(db)))
	mux.HandleFunc("DELETE /api/v1/projects/{projectId}/tasks/{taskId}/comments/{commentId}", auth(handler.HandleDeleteComment(db)))

	// Serve static frontend files
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "../../frontend/dist"
	}
	absStatic, _ := filepath.Abs(staticDir)

	// Serve static files, fall back to index.html for SPA routing
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(absStatic, r.URL.Path)

		// Check if the file exists
		if _, err := os.Stat(path); err == nil {
			http.ServeFile(w, r, path)
			return
		}

		// Fall back to index.html for SPA routing
		http.ServeFile(w, r, filepath.Join(absStatic, "index.html"))
	})

	log.Printf("Starting server on :%s", port)
	log.Printf("Serving static files from %s", absStatic)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleHealth(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "ok"
		if err := db.Ping(); err != nil {
			dbStatus = "error"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"database": dbStatus,
		})
	}
}
