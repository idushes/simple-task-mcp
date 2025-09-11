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

// CompleteTaskInput represents the input for complete_task tool
type CompleteTaskInput struct {
	ID     string `json:"id"`
	Result string `json:"result,omitempty"`
}

// RegisterCompleteTaskTool registers the complete_task tool
func RegisterCompleteTaskTool(s *server.MCPServer, jwtManager *auth.JWTManager) error {
	completeTaskTool := mcp.NewTool("complete_task",
		mcp.WithDescription("Mark task as completed and optionally add result"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID (UUID)"),
		),
		mcp.WithString("result",
			mcp.Description("Task completion result or notes"),
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
		var input CompleteTaskInput
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

		// Validate UUID format
		if !isValidUUID(input.ID) {
			return mcp.NewToolResultError("invalid task ID format"), nil
		}

		// Get database connection
		db := database.DB

		// Check if task exists and user has permission to complete it
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
			return mcp.NewToolResultError("permission denied: you can only complete tasks you created or are assigned to"), nil
		}

		// Check if task is already archived
		if isArchived {
			return mcp.NewToolResultError("cannot complete archived task"), nil
		}

		// Check if task is already completed
		if currentStatus == string(models.StatusCompleted) {
			return mcp.NewToolResultError("task is already completed"), nil
		}

		// Update task to completed status
		updateQuery := `
			UPDATE tasks 
			SET status = $1, result = $2, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
			WHERE id = $3
			RETURNING updated_at, completed_at`

		var updatedAt, completedAt string
		err = db.QueryRow(updateQuery, models.StatusCompleted, input.Result, input.ID).Scan(&updatedAt, &completedAt)
		if err != nil {
			log.Printf("Error completing task: %v", err)
			return mcp.NewToolResultError("failed to complete task"), nil
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
			CompletedAt  string
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
			"completed_at":     task.CompletedAt,
		}

		if task.Result != nil {
			response["result"] = *task.Result
		}

		return mcp.NewToolResultStructured(response, fmt.Sprintf("Task completed: %s (ID: %s)", task.Description, task.ID)), nil
	}

	s.AddTool(completeTaskTool, handler)
	log.Println("complete_task tool registered")
	return nil
}
