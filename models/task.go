package models

import (
	"database/sql"
	"time"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	StatusPending        TaskStatus = "pending"
	StatusInProgress     TaskStatus = "in_progress"
	StatusWaitingForUser TaskStatus = "waiting_for_user"
	StatusCompleted      TaskStatus = "completed"
	StatusCancelled      TaskStatus = "cancelled"
)

// Task represents a task in the system
type Task struct {
	ID          string       `json:"id"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status"`
	CreatedBy   string       `json:"created_by"`
	AssignedTo  string       `json:"assigned_to"`
	IsArchived  bool         `json:"is_archived"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	CompletedAt sql.NullTime `json:"completed_at,omitempty"`
	ArchivedAt  sql.NullTime `json:"archived_at,omitempty"`
}

// IsValidStatus checks if the given status is valid
func IsValidStatus(status string) bool {
	switch TaskStatus(status) {
	case StatusPending, StatusInProgress, StatusWaitingForUser, StatusCompleted, StatusCancelled:
		return true
	default:
		return false
	}
}
