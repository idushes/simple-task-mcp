package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/dushes/simple-task-mcp/auth"
	"github.com/dushes/simple-task-mcp/database"
	"github.com/dushes/simple-task-mcp/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WaitForUserInput represents the input for wait_for_user tool
type WaitForUserInput struct {
	ID      string `json:"id"`
	Comment string `json:"comment"`
}

// RegisterWaitForUserTool registers the wait_for_user tool
func RegisterWaitForUserTool(s *server.MCPServer, jwtManager *auth.JWTManager) error {
	waitForUserTool := mcp.NewTool("wait_for_user",
		mcp.WithDescription("Send task to waiting for user status with comment"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID (UUID)"),
		),
		mcp.WithString("comment",
			mcp.Required(),
			mcp.Description("Comment explaining why task needs user attention"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract JWT token from Authorization header
		authHeader := request.Header.Get("Authorization")
		if authHeader == "" {
			return mcp.NewToolResultError("Authorization header is required"), nil
		}

		// Remove "Bearer " prefix if present
		authToken := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			authToken = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// Validate JWT token
		claims, err := jwtManager.ValidateToken(authToken)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid token: %v", err)), nil
		}

		userID := claims.UserID

		// Parse input
		var input WaitForUserInput
		inputBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			log.Printf("Error marshaling args: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse input: %v", err)), nil
		}
		if err := json.Unmarshal(inputBytes, &input); err != nil {
			log.Printf("Error unmarshaling input: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse input: %v", err)), nil
		}

		// Validate required parameters
		if input.ID == "" {
			return mcp.NewToolResultError("task ID is required"), nil
		}
		if input.Comment == "" {
			return mcp.NewToolResultError("comment is required"), nil
		}

		// Validate UUID format
		if !isValidUUID(input.ID) {
			return mcp.NewToolResultError("invalid task ID format"), nil
		}

		// Get database connection
		db := database.DB

		// Check if task exists and user has permission to modify it
		var currentStatus string
		var isArchived bool
		var createdBy, assignedTo string
		checkQuery := `
			SELECT status, is_archived, created_by, assigned_to 
			FROM tasks 
			WHERE id = $1`

		err = db.QueryRow(checkQuery, input.ID).Scan(&currentStatus, &isArchived, &createdBy, &assignedTo)
		if err != nil {
			if err == sql.ErrNoRows {
				return mcp.NewToolResultError("task not found"), nil
			}
			log.Printf("Error checking task: %v", err)
			return mcp.NewToolResultError("database error"), nil
		}

		// Check if user has permission (must be creator or assignee)
		if createdBy != userID && assignedTo != userID {
			return mcp.NewToolResultError("permission denied: you can only modify tasks you created or are assigned to"), nil
		}

		// Check if task is already archived
		if isArchived {
			return mcp.NewToolResultError("cannot modify archived task"), nil
		}

		// Check if task is already completed
		if currentStatus == string(models.StatusCompleted) {
			return mcp.NewToolResultError("cannot send completed task to waiting"), nil
		}

		// Check if task is already cancelled
		if currentStatus == string(models.StatusCancelled) {
			return mcp.NewToolResultError("cannot send cancelled task to waiting"), nil
		}

		// Start transaction for atomic operation
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Error starting transaction: %v", err)
			return mcp.NewToolResultError("database transaction error"), nil
		}
		defer tx.Rollback()

		// Update task status to waiting_for_user
		updateTaskQuery := `
			UPDATE tasks 
			SET status = $1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $2`

		_, err = tx.Exec(updateTaskQuery, models.StatusWaitingForUser, input.ID)
		if err != nil {
			log.Printf("Error updating task status: %v", err)
			return mcp.NewToolResultError("failed to update task status"), nil
		}

		// Add comment to task_comments table
		addCommentQuery := `
			INSERT INTO task_comments (task_id, created_by, comment)
			VALUES ($1, $2, $3)
			RETURNING id, created_at`

		var commentID string
		var commentCreatedAt string
		err = tx.QueryRow(addCommentQuery, input.ID, userID, input.Comment).Scan(&commentID, &commentCreatedAt)
		if err != nil {
			log.Printf("Error adding comment: %v", err)
			return mcp.NewToolResultError("failed to add comment"), nil
		}

		// Commit transaction
		if err = tx.Commit(); err != nil {
			log.Printf("Error committing transaction: %v", err)
			return mcp.NewToolResultError("failed to commit changes"), nil
		}

		// Get task details with user names for response
		detailQuery := `
			SELECT 
				t.id, t.description, t.status, t.result,
				t.created_by, t.assigned_to,
				t.created_at, t.updated_at, t.completed_at,
				creator.name as creator_name,
				assignee.name as assignee_name
			FROM tasks t
			JOIN users creator ON t.created_by = creator.id
			JOIN users assignee ON t.assigned_to = assignee.id
			WHERE t.id = $1`

		var task struct {
			ID           string
			Description  string
			Status       string
			Result       *string
			CreatedBy    string
			AssignedTo   string
			CreatedAt    string
			UpdatedAt    string
			CompletedAt  *string
			CreatorName  string
			AssigneeName string
		}

		err = db.QueryRow(detailQuery, input.ID).Scan(
			&task.ID, &task.Description, &task.Status, &task.Result,
			&task.CreatedBy, &task.AssignedTo,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
			&task.CreatorName, &task.AssigneeName,
		)
		if err != nil {
			log.Printf("Error getting task details: %v", err)
			return mcp.NewToolResultError("failed to get updated task details"), nil
		}

		// Prepare response
		response := map[string]interface{}{
			"id":               task.ID,
			"description":      task.Description,
			"status":           task.Status,
			"created_by":       task.CreatedBy,
			"created_by_name":  task.CreatorName,
			"assigned_to":      task.AssignedTo,
			"assigned_to_name": task.AssigneeName,
			"created_at":       task.CreatedAt,
			"updated_at":       task.UpdatedAt,
			"comment_added": map[string]interface{}{
				"id":         commentID,
				"comment":    input.Comment,
				"created_at": commentCreatedAt,
			},
		}

		if task.Result != nil {
			response["result"] = *task.Result
		}

		if task.CompletedAt != nil {
			response["completed_at"] = *task.CompletedAt
		}

		return mcp.NewToolResultStructured(response, fmt.Sprintf("Task %s sent to waiting for user with comment", input.ID)), nil
	}

	s.AddTool(waitForUserTool, handler)
	log.Println("wait_for_user tool registered")
	return nil
}

