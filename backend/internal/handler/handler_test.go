package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"kanbanboard/internal/handler"
	"kanbanboard/internal/middleware"
	"kanbanboard/internal/model"
	"kanbanboard/internal/store"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// --- Test helpers ---

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping handler integration test (requires database)")
	}

	dbName := os.Getenv("TEST_DB_NAME")
	if dbName == "" {
		dbName = "kanbanboard_test"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "kanban"),
		getEnv("DB_PASSWORD", "kanban"),
		dbName,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("ping test database: %v (is PostgreSQL running?)", err)
	}

	_, thisFile, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations")
	if err := store.RunMigrations(db, migrationsDir); err != nil {
		db.Close()
		t.Fatalf("run migrations: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func cleanTables(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec("TRUNCATE task_dependencies, comments, tasks, labels, columns, projects, team_members, teams, sessions, users, app_settings CASCADE")
	if err != nil {
		t.Fatalf("clean tables: %v", err)
	}
}

func seedUser(t *testing.T, db *sql.DB, name, email string, isAdmin, isTeamManager bool) model.User {
	t.Helper()
	user, err := store.CreateUser(db, model.User{
		Name:          name,
		Email:         email,
		PasswordHash:  "$2a$10$testhashtesthasttesthashtest",
		IsAdmin:       isAdmin,
		IsTeamManager: isTeamManager,
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("seed user %s: %v", name, err)
	}
	return user
}

func createSession(t *testing.T, db *sql.DB, userID string) string {
	t.Helper()
	token, err := store.CreateSession(db, userID, 24*time.Hour)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	return token
}

var tagCounter int

func seedProject(t *testing.T, db *sql.DB, name string, ownerUserID *string, ownerTeamID *string) model.Project {
	t.Helper()
	tagCounter++
	tag := fmt.Sprintf("H%c%c", 'A'+((tagCounter-1)/26)%26, 'A'+(tagCounter-1)%26)
	project, err := store.CreateProject(db, model.Project{
		Name:        name,
		Tag:         tag,
		Visibility:  "private",
		OwnerUserID: ownerUserID,
		OwnerTeamID: ownerTeamID,
	})
	if err != nil {
		t.Fatalf("seed project %s: %v", name, err)
	}
	if err := store.CreateDefaultColumns(db, project.ID); err != nil {
		t.Fatalf("seed project columns: %v", err)
	}
	if err := store.CreateDefaultLabels(db, project.ID); err != nil {
		t.Fatalf("seed project labels: %v", err)
	}
	return project
}

func setupMux(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	auth := func(h http.HandlerFunc) http.HandlerFunc { return middleware.RequireAuth(db, h) }
	admin := func(h http.HandlerFunc) http.HandlerFunc { return middleware.RequireAdmin(db, h) }

	mux.HandleFunc("POST /api/v1/auth/login", handler.HandleLogin(db))
	mux.HandleFunc("GET /api/v1/auth/me", handler.HandleMe(db))

	mux.HandleFunc("GET /api/v1/admin/users", admin(handler.HandleListUsers(db)))
	mux.HandleFunc("POST /api/v1/admin/users", admin(handler.HandleCreateUser(db)))
	mux.HandleFunc("DELETE /api/v1/admin/users/{userId}", admin(handler.HandleDeleteUser(db)))

	mux.HandleFunc("POST /api/v1/projects", auth(handler.HandleCreateProject(db)))
	mux.HandleFunc("GET /api/v1/projects/{id}", auth(handler.HandleGetProject(db)))
	mux.HandleFunc("DELETE /api/v1/projects/{id}", auth(handler.HandleDeleteProject(db)))
	mux.HandleFunc("POST /api/v1/projects/{id}/columns", auth(handler.HandleCreateColumn(db)))
	mux.HandleFunc("DELETE /api/v1/projects/{id}/labels/{labelId}", auth(handler.HandleDeleteLabel(db)))

	mux.HandleFunc("GET /api/v1/search/tasks", auth(handler.HandleSearchTasks(db)))
	mux.HandleFunc("POST /api/v1/projects/{projectId}/tasks", auth(handler.HandleCreateTask(db)))
	mux.HandleFunc("PUT /api/v1/projects/{projectId}/tasks/{taskId}", auth(handler.HandleUpdateTask(db)))

	return mux
}

func authRequest(method, url string, body any, token string) *http.Request {
	var bodyReader *bytes.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(data)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	}
	return req
}

func doRequest(mux *http.ServeMux, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

// --- Authorization tests ---

func TestAuthRoute_requiresSession(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	mux := setupMux(db)

	req := authRequest("GET", "/api/v1/admin/users", nil, "")
	rr := doRequest(mux, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestAdminRoute_requiresAdminRole(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	mux := setupMux(db)

	user := seedUser(t, db, "Regular", "regular@test.com", false, false)
	token := createSession(t, db, user.ID)

	req := authRequest("GET", "/api/v1/admin/users", nil, token)
	rr := doRequest(mux, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rr.Code)
	}
}

func TestColumnCreate_requiresEditPermission(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	mux := setupMux(db)

	owner := seedUser(t, db, "Owner", "owner@test.com", false, false)
	outsider := seedUser(t, db, "Outsider", "outsider@test.com", false, false)
	project := seedProject(t, db, "Owner Project", &owner.ID, nil)

	token := createSession(t, db, outsider.ID)
	req := authRequest("POST", "/api/v1/projects/"+project.ID+"/columns", map[string]string{"name": "Hacked"}, token)
	rr := doRequest(mux, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rr.Code)
	}
}

func TestLabelDelete_requiresEditPermission(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	mux := setupMux(db)

	owner := seedUser(t, db, "Owner", "owner@test.com", false, false)
	outsider := seedUser(t, db, "Outsider", "outsider@test.com", false, false)
	project := seedProject(t, db, "Owner Project", &owner.ID, nil)

	labels, _ := store.GetLabelsForProject(db, project.ID)
	labelID := labels[0].ID

	token := createSession(t, db, outsider.ID)
	req := authRequest("DELETE", "/api/v1/projects/"+project.ID+"/labels/"+labelID, nil, token)
	rr := doRequest(mux, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rr.Code)
	}
}

// --- Validation tests ---

func TestUpdateTask_invalidPriority(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	mux := setupMux(db)

	user := seedUser(t, db, "Alice", "alice@test.com", false, false)
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := store.GetColumnsForProject(db, project.ID)

	task, err := store.CreateTask(db, model.Task{
		ProjectID: project.ID,
		ColumnID:  columns[0].ID,
		CreatorID: user.ID,
		Title:     "Test Task",
		Priority:  "none",
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	token := createSession(t, db, user.ID)
	req := authRequest("PUT", "/api/v1/projects/"+project.ID+"/tasks/"+task.ID,
		map[string]string{"priority": "urgent"}, token)
	rr := doRequest(mux, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestCreateProject_invalidTag(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	mux := setupMux(db)

	user := seedUser(t, db, "Alice", "alice@test.com", false, false)
	token := createSession(t, db, user.ID)

	req := authRequest("POST", "/api/v1/projects",
		map[string]string{"name": "Test", "tag": "toolong"}, token)
	rr := doRequest(mux, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestCreateProject_duplicateTag(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	mux := setupMux(db)

	user := seedUser(t, db, "Alice", "alice@test.com", false, false)
	seedProject(t, db, "First", &user.ID, nil)

	// Get the tag that was auto-generated for "First"
	projects, _ := store.ListProjectsForUser(db, user.ID)
	existingTag := projects[0].Tag

	token := createSession(t, db, user.ID)
	req := authRequest("POST", "/api/v1/projects",
		map[string]string{"name": "Second", "tag": existingTag}, token)
	rr := doRequest(mux, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rr.Code)
	}
}

func TestCreateUser_duplicateEmail(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	mux := setupMux(db)

	admin := seedUser(t, db, "Admin", "admin@test.com", true, false)
	seedUser(t, db, "Existing", "taken@test.com", false, false)

	token := createSession(t, db, admin.ID)
	req := authRequest("POST", "/api/v1/admin/users",
		map[string]any{
			"name":     "New User",
			"email":    "taken@test.com",
			"password": "password1",
		}, token)
	rr := doRequest(mux, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rr.Code)
	}
}

// --- Delete project tests ---

func TestDeleteProject_ownerCanDelete(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	mux := setupMux(db)

	owner := seedUser(t, db, "Owner", "owner@test.com", false, false)
	project := seedProject(t, db, "To Delete", &owner.ID, nil)

	token := createSession(t, db, owner.ID)
	req := authRequest("DELETE", "/api/v1/projects/"+project.ID, nil, token)
	rr := doRequest(mux, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rr.Code)
	}

	// Verify project is gone
	req = authRequest("GET", "/api/v1/projects/"+project.ID, nil, token)
	rr = doRequest(mux, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("project should be gone, got status %d", rr.Code)
	}
}

func TestDeleteProject_nonOwnerForbidden(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	mux := setupMux(db)

	owner := seedUser(t, db, "Owner", "owner@test.com", false, false)
	outsider := seedUser(t, db, "Outsider", "outsider@test.com", false, false)
	project := seedProject(t, db, "Protected", &owner.ID, nil)

	token := createSession(t, db, outsider.ID)
	req := authRequest("DELETE", "/api/v1/projects/"+project.ID, nil, token)
	rr := doRequest(mux, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rr.Code)
	}
}

// --- Delete user tests ---

func TestDeleteUser_adminCanDelete(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	mux := setupMux(db)

	admin := seedUser(t, db, "Admin", "admin@test.com", true, false)
	victim := seedUser(t, db, "Victim", "victim@test.com", false, false)

	token := createSession(t, db, admin.ID)
	req := authRequest("DELETE", "/api/v1/admin/users/"+victim.ID, nil, token)
	rr := doRequest(mux, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rr.Code)
	}

	// Verify user is soft-deleted (still fetchable by ID)
	deleted, err := store.GetUserByID(db, victim.ID)
	if err != nil {
		t.Fatalf("get deleted user: %v", err)
	}
	if deleted.DeletedAt == nil {
		t.Error("expected deleted_at to be set")
	}
}

func TestDeleteUser_cannotDeleteSelf(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	mux := setupMux(db)

	admin := seedUser(t, db, "Admin", "admin@test.com", true, false)

	token := createSession(t, db, admin.ID)
	req := authRequest("DELETE", "/api/v1/admin/users/"+admin.ID, nil, token)
	rr := doRequest(mux, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rr.Code)
	}
}

func TestDeleteUser_nonAdminForbidden(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	mux := setupMux(db)

	regular := seedUser(t, db, "Regular", "regular@test.com", false, false)
	victim := seedUser(t, db, "Victim", "victim@test.com", false, false)

	token := createSession(t, db, regular.ID)
	req := authRequest("DELETE", "/api/v1/admin/users/"+victim.ID, nil, token)
	rr := doRequest(mux, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rr.Code)
	}
}

// --- Search tests ---

func TestSearchTasks_requiresAuth(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	mux := setupMux(db)

	req := authRequest("GET", "/api/v1/search/tasks?q=test", nil, "")
	rr := doRequest(mux, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestSearchTasks_returnsResults(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	mux := setupMux(db)

	user := seedUser(t, db, "Alice", "alice@test.com", false, false)
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := store.GetColumnsForProject(db, project.ID)
	store.CreateTask(db, model.Task{
		ProjectID: project.ID,
		ColumnID:  columns[0].ID,
		CreatorID: user.ID,
		Title:     "Fix the login bug",
		Priority:  "none",
	})

	token := createSession(t, db, user.ID)
	req := authRequest("GET", "/api/v1/search/tasks?q=login", nil, token)
	rr := doRequest(mux, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}

	var results []map[string]any
	json.NewDecoder(rr.Body).Decode(&results)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0]["title"] != "Fix the login bug" {
		t.Errorf("title = %v, want %q", results[0]["title"], "Fix the login bug")
	}
}
