package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"kanbanboard/internal/middleware"
	"kanbanboard/internal/model"
	"kanbanboard/internal/store"
)

type updateProjectRequest struct {
	Name       *string `json:"name"`
	Visibility *string `json:"visibility"`
}

type columnRequest struct {
	Name string `json:"name"`
}

type reorderColumnsRequest struct {
	ColumnIDs []string `json:"columnIds"`
}

type labelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// HandleUpdateProject updates a project's name and visibility.
func HandleUpdateProject(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())
		projectID := r.PathValue("id")

		project, err := store.GetProject(db, projectID)
		if errors.Is(err, store.ErrProjectNotFound) {
			writeError(w, http.StatusNotFound, "Project not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get project")
			return
		}

		if !isProjectOwner(db, project, user) {
			writeError(w, http.StatusForbidden, "Only the project owner can edit settings")
			return
		}

		var req updateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Name != nil {
			if *req.Name == "" {
				writeError(w, http.StatusBadRequest, "Project name is required")
				return
			}
			project.Name = *req.Name
		}
		if req.Visibility != nil {
			if *req.Visibility != "public" && *req.Visibility != "private" {
				writeError(w, http.StatusBadRequest, "Visibility must be 'public' or 'private'")
				return
			}
			project.Visibility = *req.Visibility
		}

		project, err = store.UpdateProject(db, project)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update project")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(project)
	}
}

// HandleCreateColumn adds a column to a project.
func HandleCreateColumn(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")

		var req columnRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "Column name is required")
			return
		}

		col, err := store.CreateColumn(db, model.Column{ProjectID: projectID, Name: req.Name})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create column")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(col)
	}
}

// HandleUpdateColumn renames a column.
func HandleUpdateColumn(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		colID := r.PathValue("colId")

		var req columnRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "Column name is required")
			return
		}

		col, err := store.UpdateColumn(db, model.Column{ID: colID, Name: req.Name})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update column")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(col)
	}
}

// HandleDeleteColumn deletes a column if it has no tasks.
func HandleDeleteColumn(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		colID := r.PathValue("colId")

		count, err := store.CountTasksInColumn(db, colID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to check column")
			return
		}
		if count > 0 {
			writeError(w, http.StatusConflict, "Cannot delete column that contains tasks. Move or delete the tasks first.")
			return
		}

		if err := store.DeleteColumn(db, colID); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete column")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// HandleReorderColumns reorders columns in a project.
func HandleReorderColumns(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")

		var req reorderColumnsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		if len(req.ColumnIDs) == 0 {
			writeError(w, http.StatusBadRequest, "Column IDs are required")
			return
		}

		if err := store.ReorderColumns(db, projectID, req.ColumnIDs); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to reorder columns")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

// HandleCreateLabel adds a label to a project.
func HandleCreateLabel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")

		var req labelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "Label name is required")
			return
		}
		if req.Color == "" {
			req.Color = "#808080"
		}

		label, err := store.CreateLabel(db, model.Label{ProjectID: projectID, Name: req.Name, Color: req.Color})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create label")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(label)
	}
}

// HandleUpdateLabel updates a label's name and color.
func HandleUpdateLabel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		labelID := r.PathValue("labelId")

		var req labelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "Label name is required")
			return
		}
		if req.Color == "" {
			req.Color = "#808080"
		}

		label, err := store.UpdateLabel(db, model.Label{ID: labelID, Name: req.Name, Color: req.Color})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update label")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(label)
	}
}

// HandleDeleteLabel deletes a label if no tasks use it.
func HandleDeleteLabel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		labelID := r.PathValue("labelId")

		count, err := store.CountTasksWithLabel(db, labelID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to check label")
			return
		}
		if count > 0 {
			writeError(w, http.StatusConflict, "Cannot delete label that is used by tasks. Reassign the tasks first.")
			return
		}

		if err := store.DeleteLabel(db, labelID); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete label")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func isProjectOwner(db *sql.DB, project model.Project, user model.User) bool {
	// User owner
	if project.OwnerUserID != nil && *project.OwnerUserID == user.ID {
		return true
	}
	// Team owner
	if project.OwnerTeamID != nil {
		team, err := store.GetTeam(db, *project.OwnerTeamID)
		if err == nil && team.OwnerID == user.ID {
			return true
		}
	}
	return false
}
