package tools

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/dushes/simple-task-mcp/auth"
	"github.com/dushes/simple-task-mcp/database"
	"github.com/dushes/simple-task-mcp/models"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// uuidRegex is used to validate UUID format
var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// isValidUUID checks if a string is a valid UUID
func isValidUUID(s string) bool {
	return uuidRegex.MatchString(s)
}

// generateUUID generates a new UUID string
func generateUUID() string {
	return uuid.New().String()
}

// RegisterCreateTaskTool registers the create_task tool
func RegisterCreateTaskTool(s *server.MCPServer, jwtManager *auth.JWTManager) error {
	createTaskTool := mcp.NewTool("create_task",
		mcp.WithDescription("Create a new task and assign it to a user"),
		mcp.WithString("description",
			mcp.Required(),
			mcp.Description("Task description"),
		),
		mcp.WithString("assigned_to",
			mcp.Required(),
			mcp.Description("Username to assign the task to"),
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

		// Extract parameters
		description, err := request.RequireString("description")
		if err != nil {
			return mcp.NewToolResultError("description is required"), nil
		}

		assignedToUsername, err := request.RequireString("assigned_to")
		if err != nil {
			return mcp.NewToolResultError("assigned_to is required"), nil
		}

		// Validate UUID format for creator
		if !isValidUUID(claims.UserID) {
			return mcp.NewToolResultError("invalid user ID in token"), nil
		}

		// Get assigned_to user ID and validate existence
		var assignedToID string
		err = database.DB.QueryRow("SELECT id FROM users WHERE name = $1", assignedToUsername).Scan(&assignedToID)
		if err != nil {
			if err == sql.ErrNoRows {
				return mcp.NewToolResultError(fmt.Sprintf("user '%s' does not exist", assignedToUsername)), nil
			}
			log.Printf("Error finding user by name: %v", err)
			return mcp.NewToolResultError("database error"), nil
		}

		// Get creator username
		var creatorName string
		err = database.DB.QueryRow("SELECT name FROM users WHERE id = $1", claims.UserID).Scan(&creatorName)
		if err != nil {
			log.Printf("Error getting creator name: %v", err)
			return mcp.NewToolResultError("database error"), nil
		}

		// Create the task
		taskID := generateUUID()
		task := models.Task{
			ID:          taskID,
			Description: description,
			Status:      models.StatusPending,
			CreatedBy:   claims.UserID,
			AssignedTo:  assignedToID,
			IsArchived:  false,
		}

		// Insert into database
		query := `
			INSERT INTO tasks (id, description, status, created_by, assigned_to, is_archived)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING created_at, updated_at`

		err = database.DB.QueryRow(query,
			task.ID,
			task.Description,
			task.Status,
			task.CreatedBy,
			task.AssignedTo,
			task.IsArchived,
		).Scan(&task.CreatedAt, &task.UpdatedAt)

		if err != nil {
			log.Printf("Error creating task: %v", err)
			return mcp.NewToolResultError("failed to create task"), nil
		}

		// Return the created task with usernames
		result := map[string]interface{}{
			"id":               task.ID,
			"description":      task.Description,
			"status":           string(task.Status),
			"created_by":       task.CreatedBy,
			"created_by_name":  creatorName,
			"assigned_to":      task.AssignedTo,
			"assigned_to_name": assignedToUsername,
			"created_at":       task.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updated_at":       task.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}

		return mcp.NewToolResultStructured(result, fmt.Sprintf("Task created successfully with ID: %s", task.ID)), nil
	}

	s.AddTool(createTaskTool, handler)
	log.Println("create_task tool registered")
	return nil
}
