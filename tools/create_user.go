package tools

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/dushes/simple-task-mcp/auth"
	"github.com/dushes/simple-task-mcp/database"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterCreateUserTool registers the create_user tool
func RegisterCreateUserTool(s *server.MCPServer, jwtManager *auth.JWTManager) error {
	tool := mcp.NewTool("create_user",
		mcp.WithDescription("Create a new user in the system (admin only)"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the user"),
		),
		mcp.WithBoolean("is_admin",
			mcp.Description("Whether the user should have admin privileges (default: false)"),
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

		// Check if user is admin
		if !claims.IsAdmin {
			return mcp.NewToolResultError("only admins can create users"), nil
		}

		// Extract parameters
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}

		// Default is_admin to false if not provided
		isAdmin := request.GetBool("is_admin", false)

		// Create user in database
		var userID string
		query := `
			INSERT INTO users (name, is_admin)
			VALUES ($1, $2)
			RETURNING id
		`

		err = database.DB.QueryRow(query, name, isAdmin).Scan(&userID)
		if err != nil {
			if strings.Contains(err.Error(), "users_name_unique") || strings.Contains(err.Error(), "duplicate key value") {
				return mcp.NewToolResultError(fmt.Sprintf("user with name '%s' already exists", name)), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("failed to create user: %v", err)), nil
		}

		// Generate token for the new user
		newUserToken, err := jwtManager.GenerateToken(userID, isAdmin)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to generate token for new user: %v", err)), nil
		}

		// Return success with user details
		result := map[string]interface{}{
			"success": true,
			"user": map[string]interface{}{
				"id":       userID,
				"name":     name,
				"is_admin": isAdmin,
				"token":    newUserToken,
			},
			"message": fmt.Sprintf("User '%s' created successfully", name),
		}

		return mcp.NewToolResultStructured(result, fmt.Sprintf("User '%s' created successfully", name)), nil
	}

	s.AddTool(tool, handler)
	return nil
}

// GetUserByID retrieves user information by ID
func GetUserByID(userID string) (map[string]interface{}, error) {
	var name string
	var isAdmin bool
	var createdAt, updatedAt sql.NullTime

	query := `
		SELECT name, is_admin, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := database.DB.QueryRow(query, userID).Scan(&name, &isAdmin, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	user := map[string]interface{}{
		"id":       userID,
		"name":     name,
		"is_admin": isAdmin,
	}

	if createdAt.Valid {
		user["created_at"] = createdAt.Time
	}
	if updatedAt.Valid {
		user["updated_at"] = updatedAt.Time
	}

	return user, nil
}
