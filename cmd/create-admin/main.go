package main

import (
	"fmt"
	"log"

	"github.com/dushes/simple-task-mcp/auth"
	"github.com/dushes/simple-task-mcp/config"
	"github.com/dushes/simple-task-mcp/database"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create initial admin user
	var userID string
	query := `
		INSERT INTO users (name, is_admin)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
		RETURNING id
	`

	err = database.DB.QueryRow(query, "admin", true).Scan(&userID)
	if err != nil {
		// Check if admin already exists
		checkQuery := `SELECT id FROM users WHERE name = $1 AND is_admin = true`
		err = database.DB.QueryRow(checkQuery, "admin").Scan(&userID)
		if err != nil {
			log.Fatalf("Failed to create or find admin user: %v", err)
		}
		fmt.Println("Admin user already exists")
	} else {
		fmt.Println("Admin user created successfully")
	}

	// Generate JWT token for admin
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	token, err := jwtManager.GenerateToken(userID, true)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	fmt.Println("\n=== Initial Admin Credentials ===")
	fmt.Printf("User ID: %s\n", userID)
	fmt.Printf("Name: admin\n")
	fmt.Printf("Is Admin: true\n")
	fmt.Printf("\nJWT Token:\n%s\n", token)
	fmt.Println("\nUse this token in the 'auth_token' parameter when calling admin-only tools.")
}
