package tools

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/dushes/simple-task-mcp/auth"
	"github.com/dushes/simple-task-mcp/database"
	"github.com/dushes/simple-task-mcp/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterListUsersTool registers the list_users tool with the MCP server
func RegisterListUsersTool(mcpServer *server.MCPServer, jwtManager *auth.JWTManager) error {
	// Create the tool
	listUsersTool := mcp.NewTool("list_users",
		mcp.WithDescription("List all users in the system"),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of users to return (default: 100, max: 1000)"),
			mcp.DefaultNumber(100),
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
		_, err := jwtManager.ValidateToken(authToken)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid token: %v", err)), nil
		}

		// Get limit parameter, default to 100
		limit := int(request.GetFloat("limit", 100))
		if limit < 1 {
			limit = 1
		}
		if limit > 1000 {
			limit = 1000
		}

		// Query users with limit
		rows, err := database.DB.Query(`
			SELECT id, name, description, is_admin, created_at, updated_at
			FROM users
			ORDER BY name ASC
			LIMIT $1
		`, limit)
		if err != nil {
			log.Printf("Error querying users: %v", err)
			return mcp.NewToolResultError("Failed to query users"), nil
		}
		defer rows.Close()

		// Parse results
		var users []map[string]interface{}
		for rows.Next() {
			var user models.User
			var description sql.NullString
			err := rows.Scan(&user.ID, &user.Name, &description, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)
			if err != nil {
				log.Printf("Error scanning user row: %v", err)
				return mcp.NewToolResultError("Failed to parse user data"), nil
			}

			userMap := map[string]interface{}{
				"id":         user.ID,
				"name":       user.Name,
				"is_admin":   user.IsAdmin,
				"created_at": user.CreatedAt.Format("2006-01-02T15:04:05Z"),
				"updated_at": user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			}

			if description.Valid {
				userMap["description"] = description.String
			}

			users = append(users, userMap)
		}

		if err = rows.Err(); err != nil {
			log.Printf("Error iterating user rows: %v", err)
			return mcp.NewToolResultError("Failed to iterate users"), nil
		}

		// Return user list
		result := map[string]interface{}{
			"users": users,
			"count": len(users),
			"limit": limit,
		}

		return mcp.NewToolResultStructured(result, fmt.Sprintf("Found %d users", len(users))), nil
	}

	mcpServer.AddTool(listUsersTool, handler)
	log.Println("list_users tool registered")
	return nil
}
