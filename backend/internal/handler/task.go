package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"kanbanboard/internal/middleware"
	"kanbanboard/internal/model"
	"kanbanboard/internal/store"
)

type createTaskRequest struct {
	Title    string `json:"title"`
	ColumnID string `json:"columnId"`
}

// HandleCreateTask creates a new task in a project.
func HandleCreateTask(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())
		projectID := r.PathValue("projectId")

		var req createTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Title == "" {
			writeError(w, http.StatusBadRequest, "Task title is required")
			return
		}

		if req.ColumnID == "" {
			writeError(w, http.StatusBadRequest, "Column ID is required")
			return
		}

		task := model.Task{
			ProjectID:   projectID,
			ColumnID:    req.ColumnID,
			CreatorID:   user.ID,
			Title:       req.Title,
			Description: "",
			Priority:    "none",
		}

		task, err := store.CreateTask(db, task)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create task")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	}
}

// HandleListTasks returns all tasks for a project.
func HandleListTasks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("projectId")

		tasks, err := store.ListTasksForProject(db, projectID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list tasks")
			return
		}

		if tasks == nil {
			tasks = []model.Task{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	}
}
