package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"kanbanboard/internal/middleware"
	"kanbanboard/internal/model"
	"kanbanboard/internal/store"
)

type createTaskRequest struct {
	Title    string `json:"title"`
	ColumnID string `json:"columnId"`
}

type updateTaskRequest struct {
	Title         *string `json:"title"`
	Description   *string `json:"description"`
	ColumnID      *string `json:"columnId"`
	LabelID       *string `json:"labelId"`
	AssigneeID    *string `json:"assigneeId"`
	Priority      *string `json:"priority"`
	TargetVersion *string `json:"targetVersion"`
	DueDate       *string `json:"dueDate"`
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

// HandleUpdateTask updates a task's fields.
func HandleUpdateTask(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := r.PathValue("taskId")

		// Get existing task
		task, err := store.GetTask(db, taskID)
		if errors.Is(err, store.ErrTaskNotFound) {
			writeError(w, http.StatusNotFound, "Task not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get task")
			return
		}

		var req updateTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Apply updates
		if req.Title != nil {
			task.Title = *req.Title
		}
		if req.Description != nil {
			task.Description = *req.Description
		}
		if req.LabelID != nil {
			if *req.LabelID == "" {
				task.LabelID = nil
			} else {
				task.LabelID = req.LabelID
			}
		}
		if req.AssigneeID != nil {
			if *req.AssigneeID == "" {
				task.AssigneeID = nil
			} else {
				task.AssigneeID = req.AssigneeID
			}
		}
		if req.Priority != nil {
			task.Priority = *req.Priority
		}
		if req.TargetVersion != nil {
			if *req.TargetVersion == "" {
				task.TargetVersion = nil
			} else {
				task.TargetVersion = req.TargetVersion
			}
		}
		if req.DueDate != nil {
			if *req.DueDate == "" {
				task.DueDate = nil
			} else {
				t, err := time.Parse("2006-01-02", *req.DueDate)
				if err != nil {
					writeError(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
					return
				}
				task.DueDate = &t
			}
		}

		// Handle column change (move)
		if req.ColumnID != nil && *req.ColumnID != task.ColumnID {
			if err := store.MoveTask(db, task.ID, *req.ColumnID); err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to move task")
				return
			}
			task.ColumnID = *req.ColumnID
		}

		task, err = store.UpdateTask(db, task)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update task")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}
}

// HandleDeleteTask deletes a task.
func HandleDeleteTask(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := r.PathValue("taskId")

		if err := store.DeleteTask(db, taskID); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete task")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
