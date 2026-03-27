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

type createProjectRequest struct {
	Name   string  `json:"name"`
	TeamID *string `json:"teamId"`
}

type projectResponse struct {
	model.Project
	Columns []model.Column `json:"columns"`
	Labels  []model.Label  `json:"labels"`
	Tasks   []model.Task   `json:"tasks"`
}

// HandleCreateProject creates a new project with default columns and labels.
func HandleCreateProject(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())

		var req createProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "Project name is required")
			return
		}

		var project model.Project
		if req.TeamID != nil && *req.TeamID != "" {
			// Team-owned project — verify user owns the team
			team, err := store.GetTeam(db, *req.TeamID)
			if err != nil {
				writeError(w, http.StatusBadRequest, "Team not found")
				return
			}
			if team.OwnerID != user.ID {
				writeError(w, http.StatusForbidden, "You must own the team to create a project for it")
				return
			}
			project = model.Project{
				Name:        req.Name,
				Visibility:  "public",
				OwnerTeamID: req.TeamID,
			}
		} else {
			project = model.Project{
				Name:        req.Name,
				Visibility:  "public",
				OwnerUserID: &user.ID,
			}
		}

		project, err := store.CreateProject(db, project)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create project")
			return
		}

		if err := store.CreateDefaultColumns(db, project.ID); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create default columns")
			return
		}

		if err := store.CreateDefaultLabels(db, project.ID); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create default labels")
			return
		}

		// Return project with columns and labels
		resp, err := buildProjectResponse(db, project)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to load project details")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}
}

// HandleListProjects returns all projects visible to the current user.
func HandleListProjects(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())

		projects, err := store.ListProjectsForUser(db, user.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list projects")
			return
		}

		if projects == nil {
			projects = []model.Project{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
	}
}

// HandleGetProject returns a single project with its columns and labels.
func HandleGetProject(db *sql.DB) http.HandlerFunc {
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

		// Check visibility
		if !canViewProject(db, project, user) {
			writeError(w, http.StatusNotFound, "Project not found")
			return
		}

		resp, err := buildProjectResponse(db, project)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to load project details")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func buildProjectResponse(db *sql.DB, project model.Project) (projectResponse, error) {
	columns, err := store.GetColumnsForProject(db, project.ID)
	if err != nil {
		return projectResponse{}, err
	}
	if columns == nil {
		columns = []model.Column{}
	}

	labels, err := store.GetLabelsForProject(db, project.ID)
	if err != nil {
		return projectResponse{}, err
	}
	if labels == nil {
		labels = []model.Label{}
	}

	tasks, err := store.ListTasksForProject(db, project.ID)
	if err != nil {
		return projectResponse{}, err
	}
	if tasks == nil {
		tasks = []model.Task{}
	}

	return projectResponse{
		Project: project,
		Columns: columns,
		Labels:  labels,
		Tasks:   tasks,
	}, nil
}

// HandleGetProjectMembers returns the members who can work on a project.
func HandleGetProjectMembers(db *sql.DB) http.HandlerFunc {
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

		if !canViewProject(db, project, user) {
			writeError(w, http.StatusNotFound, "Project not found")
			return
		}

		members, err := store.GetProjectMembers(db, project)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get project members")
			return
		}

		// Return basic info only
		type basicMember struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		result := make([]basicMember, len(members))
		for i, m := range members {
			result[i] = basicMember{ID: m.ID, Name: m.Name}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func canViewProject(db *sql.DB, project model.Project, user model.User) bool {
	// User owner can always view
	if project.OwnerUserID != nil && *project.OwnerUserID == user.ID {
		return true
	}

	// Public projects are visible to everyone
	if project.Visibility == "public" {
		return true
	}

	// Team owner or member can view team projects
	if project.OwnerTeamID != nil {
		team, err := store.GetTeam(db, *project.OwnerTeamID)
		if err == nil && team.OwnerID == user.ID {
			return true
		}
		isMember, _ := store.IsTeamMember(db, *project.OwnerTeamID, user.ID)
		return isMember
	}

	return false
}

// canEditProject checks if a user can edit tasks in a project.
func canEditProject(db *sql.DB, project model.Project, user model.User) bool {
	// User owner can edit
	if project.OwnerUserID != nil && *project.OwnerUserID == user.ID {
		return true
	}

	// Team owner or member can edit team projects
	if project.OwnerTeamID != nil {
		team, err := store.GetTeam(db, *project.OwnerTeamID)
		if err == nil && team.OwnerID == user.ID {
			return true
		}
		isMember, _ := store.IsTeamMember(db, *project.OwnerTeamID, user.ID)
		return isMember
	}

	return false
}
