// Package model defines the domain entities.
package model

import "time"

// User represents a user account.
type User struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	IsAdmin       bool      `json:"isAdmin"`
	IsTeamManager bool      `json:"isTeamManager"`
	IsActive      bool      `json:"isActive"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// Team represents a group of users that can own projects.
type Team struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Project represents a kanban board.
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Visibility  string    `json:"visibility"`
	OwnerUserID *string   `json:"ownerUserId,omitempty"`
	OwnerTeamID *string   `json:"ownerTeamId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Column represents a board column within a project.
type Column struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
	Position  int    `json:"position"`
}

// Label represents a project-scoped task label.
type Label struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
	Color     string `json:"color"`
}

// Task represents a work item on the board.
type Task struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"projectId"`
	ColumnID      string     `json:"columnId"`
	LabelID       *string    `json:"labelId,omitempty"`
	AssigneeID    *string    `json:"assigneeId,omitempty"`
	CreatorID     string     `json:"creatorId"`
	ParentTaskID  *string    `json:"parentTaskId,omitempty"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Priority      string     `json:"priority"`
	TargetVersion *string    `json:"targetVersion,omitempty"`
	DueDate       *time.Time `json:"dueDate,omitempty"`
	Position      int        `json:"position"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// Comment represents a comment on a task.
type Comment struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"taskId"`
	AuthorID  string    `json:"authorId"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AppSetting represents a key-value application setting.
type AppSetting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
