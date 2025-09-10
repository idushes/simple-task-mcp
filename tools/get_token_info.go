package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dushes/simple-task-mcp/auth"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterGetTokenInfoTool registers the get_token_info tool
func RegisterGetTokenInfoTool(s *server.MCPServer, jwtManager *auth.JWTManager) error {
	tool := mcp.NewTool("get_token_info",
		mcp.WithDescription("Get information about the current JWT token"),
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

		// Get user information
		user, err := GetUserByID(claims.UserID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get user information: %v", err)), nil
		}

		// Calculate token expiration time and remaining time
		expiresAt := claims.ExpiresAt.Time
		remainingTime := time.Until(expiresAt)
		issuedAt := claims.IssuedAt.Time

		// Format times in a human-readable format
		expiresAtFormatted := expiresAt.Format(time.RFC1123)
		issuedAtFormatted := issuedAt.Format(time.RFC1123)

		// Format remaining time in days, hours, minutes
		days := int(remainingTime.Hours()) / 24
		hours := int(remainingTime.Hours()) % 24
		minutes := int(remainingTime.Minutes()) % 60
		remainingTimeFormatted := fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)

		// Return token information
		result := map[string]interface{}{
			"success": true,
			"token_info": map[string]interface{}{
				"user_id":        claims.UserID,
				"user_name":      user["name"],
				"is_admin":       claims.IsAdmin,
				"issued_at":      issuedAtFormatted,
				"expires_at":     expiresAtFormatted,
				"remaining_time": remainingTimeFormatted,
			},
			"message": "Token information retrieved successfully",
		}

		return mcp.NewToolResultStructured(result, "Token information retrieved successfully"), nil
	}

	s.AddTool(tool, handler)
	return nil
}
