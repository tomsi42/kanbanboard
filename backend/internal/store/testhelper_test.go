package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"

	"kanbanboard/internal/model"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// tagCounter generates unique test tags.
var tagCounter atomic.Int32

// testDB connects to the test database, runs migrations, and returns the connection.
// Skips the test when running with -short (unit tests only).
func testDB(t *testing.T) *sql.DB {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test (requires database)")
	}

	dbName := os.Getenv("TEST_DB_NAME")
	if dbName == "" {
		dbName = "kanbanboard_test"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnvOrDefault("DB_HOST", "localhost"),
		getEnvOrDefault("DB_PORT", "5432"),
		getEnvOrDefault("DB_USER", "kanban"),
		getEnvOrDefault("DB_PASSWORD", "kanban"),
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

	// Run migrations
	_, thisFile, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations")
	if err := RunMigrations(db, migrationsDir); err != nil {
		db.Close()
		t.Fatalf("run migrations: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// cleanTables truncates all application tables.
func cleanTables(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec("TRUNCATE task_dependencies, comments, tasks, labels, columns, projects, team_members, teams, sessions, users, app_settings CASCADE")
	if err != nil {
		t.Fatalf("clean tables: %v", err)
	}
}

// seedUser creates a test user and returns it.
func seedUser(t *testing.T, db *sql.DB, name, email string) model.User {
	t.Helper()
	user, err := CreateUser(db, model.User{
		Name:         name,
		Email:        email,
		PasswordHash: "$2a$10$testhashtesthasttesthashtest", // not a real bcrypt hash, but valid for DB storage
		IsAdmin:      false,
		IsTeamManager: false,
		IsActive:     true,
	})
	if err != nil {
		t.Fatalf("seed user %s: %v", name, err)
	}
	return user
}

// seedTeam creates a test team and returns it.
func seedTeam(t *testing.T, db *sql.DB, name, ownerID string) model.Team {
	t.Helper()
	team, err := CreateTeam(db, model.Team{Name: name, OwnerID: ownerID})
	if err != nil {
		t.Fatalf("seed team %s: %v", name, err)
	}
	return team
}

// seedProject creates a test project with default columns and labels, then returns it.
// Generates a unique tag automatically.
func seedProject(t *testing.T, db *sql.DB, name string, ownerUserID *string, ownerTeamID *string) model.Project {
	t.Helper()
	n := tagCounter.Add(1)
	tag := fmt.Sprintf("T%c%c", 'A'+((n-1)/26)%26, 'A'+(n-1)%26)
	project, err := CreateProject(db, model.Project{
		Name:        name,
		Tag:         tag,
		Visibility:  "private",
		OwnerUserID: ownerUserID,
		OwnerTeamID: ownerTeamID,
	})
	if err != nil {
		t.Fatalf("seed project %s: %v", name, err)
	}
	if err := CreateDefaultColumns(db, project.ID); err != nil {
		t.Fatalf("seed project columns: %v", err)
	}
	if err := CreateDefaultLabels(db, project.ID); err != nil {
		t.Fatalf("seed project labels: %v", err)
	}
	return project
}

// seedTask creates a test task and returns it.
func seedTask(t *testing.T, db *sql.DB, projectID, columnID, creatorID, title string) model.Task {
	t.Helper()
	task, err := CreateTask(db, model.Task{
		ProjectID: projectID,
		ColumnID:  columnID,
		CreatorID: creatorID,
		Title:     title,
		Priority:  "none",
	})
	if err != nil {
		t.Fatalf("seed task %s: %v", title, err)
	}
	return task
}

func strPtr(s string) *string { return &s }
