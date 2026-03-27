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

type commentRequest struct {
	Text string `json:"text"`
}

// HandleListComments returns all comments for a task.
func HandleListComments(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := r.PathValue("taskId")

		comments, err := store.ListCommentsForTask(db, taskID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list comments")
			return
		}

		if comments == nil {
			comments = []store.CommentWithAuthor{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comments)
	}
}

// HandleCreateComment creates a new comment on a task.
func HandleCreateComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())
		projectID := r.PathValue("projectId")
		taskID := r.PathValue("taskId")

		if _, ok := checkEditPermission(db, w, projectID, user); !ok {
			return
		}

		var req commentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Text == "" {
			writeError(w, http.StatusBadRequest, "Comment text is required")
			return
		}

		comment := model.Comment{
			TaskID:   taskID,
			AuthorID: user.ID,
			Text:     req.Text,
		}

		comment, err := store.CreateComment(db, comment)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create comment")
			return
		}

		// Return with author name
		resp := store.CommentWithAuthor{
			Comment:    comment,
			AuthorName: user.Name,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}
}

// HandleUpdateComment updates a comment's text. Only the author can edit.
func HandleUpdateComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())
		commentID := r.PathValue("commentId")

		// Check ownership
		existing, err := store.GetComment(db, commentID)
		if errors.Is(err, store.ErrCommentNotFound) {
			writeError(w, http.StatusNotFound, "Comment not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get comment")
			return
		}
		if existing.AuthorID != user.ID {
			writeError(w, http.StatusForbidden, "You can only edit your own comments")
			return
		}

		var req commentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Text == "" {
			writeError(w, http.StatusBadRequest, "Comment text is required")
			return
		}

		comment, err := store.UpdateComment(db, commentID, req.Text)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update comment")
			return
		}

		resp := store.CommentWithAuthor{
			Comment:    comment,
			AuthorName: user.Name,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// HandleDeleteComment deletes a comment. Only the author can delete.
func HandleDeleteComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserFromContext(r.Context())
		commentID := r.PathValue("commentId")

		existing, err := store.GetComment(db, commentID)
		if errors.Is(err, store.ErrCommentNotFound) {
			writeError(w, http.StatusNotFound, "Comment not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get comment")
			return
		}
		if existing.AuthorID != user.ID {
			writeError(w, http.StatusForbidden, "You can only delete your own comments")
			return
		}

		if err := store.DeleteComment(db, commentID); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete comment")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
