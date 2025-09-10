package models

import (
	"time"
)

// TaskComment represents a comment on a task
type TaskComment struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	CreatedBy string    `json:"created_by"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskCommentWithUser represents a comment with user information
type TaskCommentWithUser struct {
	ID            string    `json:"id"`
	TaskID        string    `json:"task_id"`
	CreatedBy     string    `json:"created_by"`
	CreatedByName string    `json:"created_by_name"`
	Comment       string    `json:"comment"`
	CreatedAt     time.Time `json:"created_at"`
}

