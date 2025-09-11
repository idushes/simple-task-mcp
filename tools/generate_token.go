package tools

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/dushes/simple-task-mcp/auth"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterGenerateTokenTool registers the generate_token tool
func RegisterGenerateTokenTool(s *server.MCPServer, jwtManager *auth.JWTManager) error {
	tool := mcp.NewTool("generate_token",
		mcp.WithDescription("Generate a new JWT token for an existing user (admin only)"),
		mcp.WithString("user_id",
			mcp.Required(),
			mcp.Description("User ID (UUID) to generate token for"),
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
			return mcp.NewToolResultError("only admins can generate tokens for users"), nil
		}

		// Extract parameters
		userID, err := request.RequireString("user_id")
		if err != nil {
			return mcp.NewToolResultError("user_id is required"), nil
		}

		// Verify user exists
		user, err := GetUserByID(userID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to find user: %v", err)), nil
		}

		// Generate token for the user
		isAdmin := user["is_admin"].(bool)
		newToken, err := jwtManager.GenerateToken(userID, isAdmin)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to generate token: %v", err)), nil
		}

		// Return success with user details and token
		user["token"] = newToken
		result := map[string]interface{}{
			"token":   newToken,
			"success": true,
			"user":    user,
			"message": fmt.Sprintf("Token generated successfully for user '%s'", user["name"]),
		}

		return mcp.NewToolResultStructured(result, fmt.Sprintf("Token generated for %s", user["name"])), nil
	}

	s.AddTool(tool, handler)
	log.Println("generate_token tool registered")
	return nil
}
