package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"kanbanboard/internal/middleware"
	"kanbanboard/internal/model"
	"kanbanboard/internal/store"
	"kanbanboard/internal/validate"
)

type createTaskRequest struct {
	Title        string  `json:"title"`
	ColumnID     string  `json:"columnId"`
	ParentTaskID *string `json:"parentTaskId"`
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

		if _, ok := checkEditPermission(db, w, projectID, user); !ok {
			return
		}

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

		// Enforce maximum nesting depth of 2 (task → subtask → sub-subtask)
		if req.ParentTaskID != nil {
			depth, err := store.GetTaskDepth(db, *req.ParentTaskID)
			if errors.Is(err, store.ErrTaskNotFound) {
				writeError(w, http.StatusBadRequest, "Parent task not found")
				return
			}
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to check task depth")
				return
			}
			if depth >= 2 {
				writeError(w, http.StatusBadRequest, "Sub-subtasks cannot have children (maximum nesting depth is 2)")
				return
			}
		}

		// Set default label if none provided
		var labelID *string
		if defaultLabel, labelErr := store.GetDefaultLabelForProject(db, projectID); labelErr == nil {
			labelID = &defaultLabel.ID
		}

		task := model.Task{
			ProjectID:    projectID,
			ColumnID:     req.ColumnID,
			CreatorID:    user.ID,
			LabelID:      labelID,
			ParentTaskID: req.ParentTaskID,
			Title:        req.Title,
			Description:  "",
			Priority:     "none",
		}

		task, err := store.CreateTask(db, task)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create task")
			return
		}

		writeJSON(w, http.StatusCreated, task)
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

		writeJSON(w, http.StatusOK, tasks)
	}
}

// HandleUpdateTask updates a task's fields.
func HandleUpdateTask(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())
		projectID := r.PathValue("projectId")
		taskID := r.PathValue("taskId")

		if _, ok := checkEditPermission(db, w, projectID, user); !ok {
			return
		}

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

		// Apply field updates
		if err := applyTaskUpdates(&task, req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Handle column change (move to end of new column)
		if req.ColumnID != nil && *req.ColumnID != task.ColumnID {
			if err := store.MoveTask(db, task.ID, *req.ColumnID, 9999); err != nil {
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

		writeJSON(w, http.StatusOK, task)
	}
}

// HandleSearchTasks searches for tasks across all visible projects.
func HandleSearchTasks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())
		query := r.URL.Query().Get("q")

		if query == "" {
			writeJSON(w, http.StatusOK, []store.SearchResult{})
			return
		}

		results, err := store.SearchTasks(db, user.ID, query)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to search tasks")
			return
		}

		if results == nil {
			results = []store.SearchResult{}
		}

		writeJSON(w, http.StatusOK, results)
	}
}

// applyTaskUpdates applies optional field updates from the request to the task.
// Returns an error if any field value is invalid.
func applyTaskUpdates(task *model.Task, req updateTaskRequest) error {
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
		if msg := validate.Priority(*req.Priority); msg != "" {
			return fmt.Errorf("%s", msg)
		}
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
				return fmt.Errorf("Invalid date format (use YYYY-MM-DD)")
			}
			task.DueDate = &t
		}
	}
	return nil
}

type moveTaskRequest struct {
	ColumnID string `json:"columnId"`
	Position int    `json:"position"`
}

// HandleMoveTask moves a task to a new column and/or position.
func HandleMoveTask(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())
		projectID := r.PathValue("projectId")
		taskID := r.PathValue("taskId")

		if _, ok := checkEditPermission(db, w, projectID, user); !ok {
			return
		}

		var req moveTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.ColumnID == "" {
			writeError(w, http.StatusBadRequest, "Column ID is required")
			return
		}

		if err := store.MoveTask(db, taskID, req.ColumnID, req.Position); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to move task")
			return
		}

		task, err := store.GetTask(db, taskID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get task")
			return
		}

		writeJSON(w, http.StatusOK, task)
	}
}

// HandleAddBlocker adds a blocker dependency to a task.
// POST /api/v1/projects/{projectId}/tasks/{taskId}/blockers
// Body: {"blockerTaskId": "<uuid>"}
func HandleAddBlocker(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())
		projectID := r.PathValue("projectId")
		taskID := r.PathValue("taskId")

		if _, ok := checkEditPermission(db, w, projectID, user); !ok {
			return
		}

		var req struct {
			BlockerTaskID string `json:"blockerTaskId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		if req.BlockerTaskID == "" {
			writeError(w, http.StatusBadRequest, "blockerTaskId is required")
			return
		}

		err := store.AddBlocker(db, projectID, req.BlockerTaskID, taskID)
		if errors.Is(err, store.ErrSelfDependency) {
			writeError(w, http.StatusBadRequest, "A task cannot block itself")
			return
		}
		if errors.Is(err, store.ErrCrossProjectDependency) {
			writeError(w, http.StatusBadRequest, "Blocker must be in the same project")
			return
		}
		if errors.Is(err, store.ErrDependencyCycle) {
			writeError(w, http.StatusBadRequest, "Adding this blocker would create a dependency cycle")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to add blocker")
			return
		}

		task, err := store.GetTask(db, taskID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get task")
			return
		}
		writeJSON(w, http.StatusOK, task)
	}
}

// HandleRemoveBlocker removes a blocker dependency from a task.
// DELETE /api/v1/projects/{projectId}/tasks/{taskId}/blockers/{blockerId}
func HandleRemoveBlocker(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())
		projectID := r.PathValue("projectId")
		taskID := r.PathValue("taskId")
		blockerID := r.PathValue("blockerId")

		if _, ok := checkEditPermission(db, w, projectID, user); !ok {
			return
		}

		if err := store.RemoveBlocker(db, blockerID, taskID); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to remove blocker")
			return
		}

		task, err := store.GetTask(db, taskID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get task")
			return
		}
		writeJSON(w, http.StatusOK, task)
	}
}

// HandleDeleteTask deletes a task.
func HandleDeleteTask(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())
		projectID := r.PathValue("projectId")
		taskID := r.PathValue("taskId")

		if _, ok := checkEditPermission(db, w, projectID, user); !ok {
			return
		}

		if err := store.DeleteTask(db, taskID); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete task")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
