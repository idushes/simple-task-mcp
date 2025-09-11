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

// ListCreatedTasksInput represents the input for list_created_tasks tool
type ListCreatedTasksInput struct {
	UserName *string  `json:"user_name,omitempty"`
	Limit    *int     `json:"limit,omitempty"`
	Statuses []string `json:"statuses,omitempty"`
}

// ListCreatedTasksOutput represents the output for list_created_tasks tool
type ListCreatedTasksOutput struct {
	Tasks       []TaskWithUsers `json:"tasks"`
	TotalCount  int             `json:"total_count"`
	LimitUsed   int             `json:"limit_used"`
	CreatedBy   string          `json:"created_by"`
	CreatedByID string          `json:"created_by_id"`
}

// TaskWithUsers represents a task with user information
type TaskWithUsers struct {
	ID           string                       `json:"id"`
	Description  string                       `json:"description"`
	Status       string                       `json:"status"`
	CreatedBy    string                       `json:"created_by"`
	CreatedByID  string                       `json:"created_by_id"`
	AssignedTo   string                       `json:"assigned_to"`
	AssignedToID string                       `json:"assigned_to_id"`
	Result       *string                      `json:"result,omitempty"`
	Comments     []models.TaskCommentWithUser `json:"comments,omitempty"`
	IsArchived   bool                         `json:"is_archived"`
	CreatedAt    string                       `json:"created_at"`
	UpdatedAt    string                       `json:"updated_at"`
	CompletedAt  *string                      `json:"completed_at,omitempty"`
	ArchivedAt   *string                      `json:"archived_at,omitempty"`
}

// RegisterListCreatedTasksTool registers the list_created_tasks tool
func RegisterListCreatedTasksTool(s *server.MCPServer, jwtManager *auth.JWTManager) error {
	listCreatedTasksTool := mcp.NewTool("list_created_tasks",
		mcp.WithDescription("Get a list of tasks created by the current user or specified user (admins only)"),
		mcp.WithString("user_name",
			mcp.Description("Username to get tasks for. If not provided, uses current user. Only admins can specify other users."),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of tasks to return (default: 50, max: 1000)"),
		),
		mcp.WithArray("statuses",
			mcp.Description("Array of statuses to filter by. Available statuses: pending, in_progress, waiting_for_user, completed, cancelled. If not provided, returns tasks with all statuses."),
			mcp.Items(map[string]any{"type": "string"}),
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

		currentUserID := claims.UserID
		isAdmin := claims.IsAdmin

		// Parse input
		var input ListCreatedTasksInput
		inputBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			log.Printf("Error marshaling args: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse input: %v", err)), nil
		}
		if err := json.Unmarshal(inputBytes, &input); err != nil {
			log.Printf("Error unmarshaling input: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse input: %v", err)), nil
		}

		// Set default limit
		limit := 50
		if input.Limit != nil {
			if *input.Limit <= 0 {
				return mcp.NewToolResultError("limit must be positive"), nil
			}
			if *input.Limit > 1000 {
				return mcp.NewToolResultError("limit cannot exceed 1000"), nil
			}
			limit = *input.Limit
		}

		// Validate statuses if provided
		if len(input.Statuses) > 0 {
			validStatuses := map[string]bool{
				"pending":          true,
				"in_progress":      true,
				"waiting_for_user": true,
				"completed":        true,
				"cancelled":        true,
			}

			// Check for duplicates
			seenStatuses := make(map[string]bool)

			for _, status := range input.Statuses {
				// Check for empty strings
				if strings.TrimSpace(status) == "" {
					return mcp.NewToolResultError("status cannot be empty"), nil
				}

				// Check for duplicates
				if seenStatuses[status] {
					return mcp.NewToolResultError(fmt.Sprintf("duplicate status: '%s'", status)), nil
				}
				seenStatuses[status] = true

				// Check for valid status
				if !validStatuses[status] {
					return mcp.NewToolResultError(fmt.Sprintf("invalid status: '%s'. Valid statuses are: pending, in_progress, waiting_for_user, completed, cancelled", status)), nil
				}
			}
		}

		// Get database connection
		db := database.DB

		// Determine target user
		var targetUserID, targetUserName string
		if input.UserName != nil && *input.UserName != "" {
			// Check if user can view other users' tasks
			if !isAdmin {
				return mcp.NewToolResultError("Only admins can view tasks created by other users"), nil
			}

			// Find user by name
			err := db.QueryRow("SELECT id, name FROM users WHERE name = $1", *input.UserName).Scan(&targetUserID, &targetUserName)
			if err == sql.ErrNoRows {
				return mcp.NewToolResultError(fmt.Sprintf("User not found: %s", *input.UserName)), nil
			}
			if err != nil {
				log.Printf("Error finding user: %v", err)
				return mcp.NewToolResultError(fmt.Sprintf("Failed to find user: %v", err)), nil
			}
		} else {
			// Use current user
			targetUserID = currentUserID
			err := db.QueryRow("SELECT name FROM users WHERE id = $1", currentUserID).Scan(&targetUserName)
			if err != nil {
				log.Printf("Error getting current user name: %v", err)
				return mcp.NewToolResultError(fmt.Sprintf("Failed to get current user: %v", err)), nil
			}
		}

		// Build status filter for SQL queries
		var statusFilter string
		var countArgs []interface{}
		var queryArgs []interface{}

		countArgs = append(countArgs, targetUserID)
		queryArgs = append(queryArgs, targetUserID)

		if len(input.Statuses) > 0 {
			placeholders := make([]string, len(input.Statuses))
			for i, status := range input.Statuses {
				placeholders[i] = fmt.Sprintf("$%d", i+2)
				countArgs = append(countArgs, status)
			}
			statusFilter = fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ", "))

			// For main query, we need to adjust parameter numbers
			for i, status := range input.Statuses {
				placeholders[i] = fmt.Sprintf("$%d", i+2)
				queryArgs = append(queryArgs, status)
			}
		}

		// Get total count
		var totalCount int
		countQuery := fmt.Sprintf(`
			SELECT COUNT(*) 
			FROM tasks 
			WHERE created_by = $1%s
		`, statusFilter)

		err = db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
		if err != nil {
			log.Printf("Error counting tasks: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to count tasks: %v", err)), nil
		}

		// Get tasks
		limitParamNum := len(queryArgs) + 1
		queryArgs = append(queryArgs, limit)

		query := fmt.Sprintf(`
			SELECT 
				t.id, t.description, t.status, 
				t.created_by, t.assigned_to, t.result,
				t.is_archived, t.created_at, t.updated_at, 
				t.completed_at, t.archived_at,
				creator.name as creator_name,
				assignee.name as assignee_name
			FROM tasks t
			JOIN users creator ON t.created_by = creator.id
			JOIN users assignee ON t.assigned_to = assignee.id
			WHERE t.created_by = $1%s
			ORDER BY t.created_at DESC
			LIMIT $%d`, statusFilter, limitParamNum)

		rows, err := db.Query(query, queryArgs...)
		if err != nil {
			log.Printf("Error querying tasks: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get tasks: %v", err)), nil
		}
		defer rows.Close()

		var tasks []TaskWithUsers
		for rows.Next() {
			var task TaskWithUsers
			var completedAt, archivedAt sql.NullTime
			var result sql.NullString

			err := rows.Scan(
				&task.ID, &task.Description, &task.Status,
				&task.CreatedByID, &task.AssignedToID, &result,
				&task.IsArchived, &task.CreatedAt, &task.UpdatedAt,
				&completedAt, &archivedAt,
				&task.CreatedBy, &task.AssignedTo,
			)
			if err != nil {
				log.Printf("Error scanning task: %v", err)
				continue
			}

			if result.Valid {
				task.Result = &result.String
			}

			if completedAt.Valid {
				completedAtStr := completedAt.Time.Format("2006-01-02T15:04:05Z")
				task.CompletedAt = &completedAtStr
			}

			if archivedAt.Valid {
				archivedAtStr := archivedAt.Time.Format("2006-01-02T15:04:05Z")
				task.ArchivedAt = &archivedAtStr
			}

			// Get comments for the task
			commentsQuery := `
				SELECT 
					tc.id, tc.task_id, tc.created_by, tc.comment, tc.created_at,
					u.name as created_by_name
				FROM task_comments tc
				JOIN users u ON tc.created_by = u.id
				WHERE tc.task_id = $1
				ORDER BY tc.created_at ASC`

			commentsRows, err := db.Query(commentsQuery, task.ID)
			if err != nil {
				log.Printf("Error querying comments for task %s: %v", task.ID, err)
				// Don't fail the whole request if comments can't be retrieved
			} else {
				var comments []models.TaskCommentWithUser
				for commentsRows.Next() {
					var comment models.TaskCommentWithUser
					err := commentsRows.Scan(
						&comment.ID, &comment.TaskID, &comment.CreatedBy,
						&comment.Comment, &comment.CreatedAt, &comment.CreatedByName,
					)
					if err != nil {
						log.Printf("Error scanning comment: %v", err)
						continue
					}
					comments = append(comments, comment)
				}
				task.Comments = comments
				commentsRows.Close()
			}

			tasks = append(tasks, task)
		}

		// Prepare output
		output := ListCreatedTasksOutput{
			Tasks:       tasks,
			TotalCount:  totalCount,
			LimitUsed:   limit,
			CreatedBy:   targetUserName,
			CreatedByID: targetUserID,
		}

		// Return result
		outputJSON, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			log.Printf("Error marshaling output: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal output: %v", err)), nil
		}

		return mcp.NewToolResultText(string(outputJSON)), nil
	}

	s.AddTool(listCreatedTasksTool, handler)
	log.Println("list_created_tasks tool registered")
	return nil
}
