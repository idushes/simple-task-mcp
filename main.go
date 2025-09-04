package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dushes/simple-task-mcp/auth"
	"github.com/dushes/simple-task-mcp/config"
	"github.com/dushes/simple-task-mcp/database"
	middleware "github.com/dushes/simple-task-mcp/server"
	"github.com/dushes/simple-task-mcp/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Parse command line flags
	var transport string
	flag.StringVar(&transport, "t", "http", "Transport type (stdio or http)")
	flag.StringVar(&transport, "transport", "http", "Transport type (stdio or http)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Configure logging
	configureLogging(cfg.LogLevel)

	// Connect to database
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	// Create MCP server
	mcpServer, err := createMCPServer()
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Register tools
	if err := registerTools(mcpServer, jwtManager); err != nil {
		log.Fatalf("Failed to register tools: %v", err)
	}

	// Start server with selected transport
	if transport == "http" {
		// Create HTTP server with SSE support
		streamableServer := server.NewStreamableHTTPServer(mcpServer,
			server.WithStateLess(true), // Make server stateless for easier testing
		)

		// Create HTTP mux with CORS middleware
		mux := http.NewServeMux()
		mux.Handle("/mcp", middleware.CORSMiddleware(streamableServer))

		// Create custom HTTP server
		httpServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.MCPServerPort),
			Handler: mux,
		}

		// Set up graceful shutdown
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		// Start HTTP server in a goroutine
		go func() {
			log.Printf("Starting MCP HTTP server on port %d", cfg.MCPServerPort)
			log.Printf("Endpoint: http://localhost:%d/mcp", cfg.MCPServerPort)
			log.Println("CORS enabled for cross-origin requests")
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server error: %v", err)
			}
		}()

		// Wait for shutdown signal
		<-shutdown
		log.Println("Shutting down server...")

		// Create a deadline for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Gracefully shutdown the server
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}

		log.Println("Server stopped")
	} else {
		// Start stdio server
		log.Println("Starting MCP server with stdio transport")
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}

// createMCPServer creates and configures the MCP server
func createMCPServer() (*server.MCPServer, error) {
	mcpServer := server.NewMCPServer(
		"simple-task-mcp",
		"0.1.0",
		server.WithPromptCapabilities(false),
		server.WithResourceCapabilities(false, false),
		server.WithToolCapabilities(true),
	)

	return mcpServer, nil
}

// registerTools registers all available tools with the server
func registerTools(mcpServer *server.MCPServer, jwtManager *auth.JWTManager) error {
	// Register create_user tool (admin only)
	if err := tools.RegisterCreateUserTool(mcpServer, jwtManager); err != nil {
		return fmt.Errorf("failed to register create_user tool: %w", err)
	}

	// Register create_task tool
	if err := tools.RegisterCreateTaskTool(mcpServer, jwtManager); err != nil {
		return fmt.Errorf("failed to register create_task tool: %w", err)
	}

	log.Println("All tools registered successfully")
	return nil
}

// configureLogging sets up logging based on the log level
func configureLogging(level string) {
	// For simplicity, we'll just use standard log package
	// In a production app, you might want to use a more sophisticated logging library
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// You can add more sophisticated log level handling here if needed
	switch level {
	case "debug":
		log.Println("Log level: DEBUG")
	case "info":
		log.Println("Log level: INFO")
	case "warn":
		log.Println("Log level: WARN")
	case "error":
		log.Println("Log level: ERROR")
	default:
		log.Printf("Unknown log level: %s, defaulting to INFO", level)
	}
}
