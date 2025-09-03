package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/dushes/simple-task-mcp/config"
	middleware "github.com/dushes/simple-task-mcp/server"
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

	// Create MCP server
	mcpServer, err := createMCPServer()
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Register tools
	if err := registerTools(mcpServer); err != nil {
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

		// Start HTTP server
		log.Printf("Starting MCP HTTP server on port %d", cfg.MCPServerPort)
		log.Printf("Endpoint: http://localhost:%d/mcp", cfg.MCPServerPort)
		log.Println("CORS enabled for cross-origin requests")
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
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
func registerTools(mcpServer *server.MCPServer) error {
	// TODO: Register actual task management tools here

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
