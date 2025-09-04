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

// GetNextTaskInput represents the input for get_next_task tool
type GetNextTaskInput struct {
	Statuses []string `json:"statuses"`
}

// GetNextTaskOutput represents the output for get_next_task tool
type GetNextTaskOutput struct {
	ID           string  `json:"id"`
	Description  string  `json:"description"`
	Status       string  `json:"status"`
	CreatedBy    string  `json:"created_by"`
	CreatedByID  string  `json:"created_by_id"`
	AssignedTo   string  `json:"assigned_to"`
	AssignedToID string  `json:"assigned_to_id"`
	Result       *string `json:"result,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	CompletedAt  *string `json:"completed_at,omitempty"`
}

// RegisterGetNextTaskTool registers the get_next_task tool
func RegisterGetNextTaskTool(s *server.MCPServer, jwtManager *auth.JWTManager) error {
	getNextTaskTool := mcp.NewTool("get_next_task",
		mcp.WithDescription("Get one task where the current user is either creator or assignee, filtered by status"),
		mcp.WithArray("statuses",
			mcp.Description("Array of statuses to filter by. Available statuses: pending, in_progress, waiting_for_user, completed, cancelled. If not provided, defaults to [\"pending\"]"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract JWT token from Authorization header
		authHeader := request.Header.Get("Authorization")
		if authHeader == "" {
			return mcp.NewToolResultError("Authorization header is required"), nil
		}

		// Remove "Bearer " prefix if present
		token := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}

		// Validate JWT token
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid token: %v", err)), nil
		}

		userID := claims.UserID

		// Parse input
		var input GetNextTaskInput
		inputBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			log.Printf("Error marshaling args: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse input: %v", err)), nil
		}
		if err := json.Unmarshal(inputBytes, &input); err != nil {
			log.Printf("Error unmarshaling input: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse input: %v", err)), nil
		}

		// Use default value if statuses not provided
		if len(input.Statuses) == 0 {
			input.Statuses = []string{"pending"}
		}

		// Validate statuses
		validStatuses := map[string]bool{
			"pending":          true,
			"in_progress":      true,
			"waiting_for_user": true,
			"completed":        true,
			"cancelled":        true,
		}
		for _, status := range input.Statuses {
			if !validStatuses[status] {
				return mcp.NewToolResultError(fmt.Sprintf("invalid status: %s", status)), nil
			}
		}

		// Get database connection
		db := database.DB

		// Build query
		placeholders := make([]string, len(input.Statuses))
		queryArgs := make([]interface{}, 0, len(input.Statuses)+1)
		queryArgs = append(queryArgs, userID)

		for i, status := range input.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", i+2)
			queryArgs = append(queryArgs, status)
		}

		query := fmt.Sprintf(`
			SELECT 
				t.id, t.description, t.status, 
				t.created_by, t.assigned_to, t.result,
				t.created_at, t.updated_at, t.completed_at,
				creator.name as creator_name,
				assignee.name as assignee_name
			FROM tasks t
			JOIN users creator ON t.created_by = creator.id
			JOIN users assignee ON t.assigned_to = assignee.id
			WHERE t.is_archived = false
				AND t.status IN (%s)
				AND (t.created_by = $1 OR t.assigned_to = $1)
			ORDER BY t.created_at ASC
			LIMIT 1
		`, strings.Join(placeholders, ", "))

		// Execute query
		var task models.Task
		var creatorName, assigneeName string
		var completedAt sql.NullTime
		var result sql.NullString
		var statusStr string

		err = db.QueryRow(query, queryArgs...).Scan(
			&task.ID, &task.Description, &statusStr,
			&task.CreatedBy, &task.AssignedTo, &result,
			&task.CreatedAt, &task.UpdatedAt, &completedAt,
			&creatorName, &assigneeName,
		)
		task.Status = models.TaskStatus(statusStr)

		if err == sql.ErrNoRows {
			// No tasks found - return null
			return mcp.NewToolResultText("null"), nil
		}

		if err != nil {
			log.Printf("Error querying task: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get task: %v", err)), nil
		}

		// Prepare output
		output := GetNextTaskOutput{
			ID:           task.ID,
			Description:  task.Description,
			Status:       string(task.Status),
			CreatedBy:    creatorName,
			CreatedByID:  task.CreatedBy,
			AssignedTo:   assigneeName,
			AssignedToID: task.AssignedTo,
			CreatedAt:    task.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:    task.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}

		if result.Valid {
			output.Result = &result.String
		}

		if completedAt.Valid {
			completedAtStr := completedAt.Time.Format("2006-01-02T15:04:05Z")
			output.CompletedAt = &completedAtStr
		}

		// Return result
		outputJSON, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			log.Printf("Error marshaling output: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal output: %v", err)), nil
		}

		return mcp.NewToolResultText(string(outputJSON)), nil
	}

	s.AddTool(getNextTaskTool, handler)
	return nil
}
