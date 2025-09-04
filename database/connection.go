package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// Connect establishes a connection to the PostgreSQL database
func Connect(databaseURL string) error {
	var err error
	DB, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// RunMigrations executes all SQL migration files
func RunMigrations() error {
	log.Println("Starting database migrations...")

	// Create migrations tracking table
	if err := createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get migration files
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	if len(migrationFiles) == 0 {
		log.Println("No migration files found")
		return nil
	}

	log.Printf("Found %d migration files", len(migrationFiles))

	// Get already applied migrations
	appliedMigrations, err := getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	log.Printf("Already applied %d migrations", len(appliedMigrations))

	// Apply pending migrations
	appliedCount := 0
	for _, filename := range migrationFiles {
		if _, exists := appliedMigrations[filename]; exists {
			log.Printf("Migration %s already applied, skipping", filename)
			continue
		}

		log.Printf("Applying migration: %s", filename)
		start := time.Now()

		if err := applyMigration(filename); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", filename, err)
		}

		duration := time.Since(start)
		log.Printf("Successfully applied migration %s (took %v)", filename, duration)
		appliedCount++
	}

	if appliedCount == 0 {
		log.Println("All migrations are up to date")
	} else {
		log.Printf("Successfully applied %d new migrations", appliedCount)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// createMigrationsTable creates the table for tracking applied migrations
func createMigrationsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		filename VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := DB.Exec(query); err != nil {
		return err
	}

	log.Println("Migrations tracking table ready")
	return nil
}

// getMigrationFiles returns sorted list of migration files
func getMigrationFiles() ([]string, error) {
	var files []string

	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	migrationsDir := filepath.Join(wd, "database", "migrations")

	err = filepath.WalkDir(migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".sql") {
			files = append(files, d.Name())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
}

// getAppliedMigrations returns a map of already applied migrations
func getAppliedMigrations() (map[string]bool, error) {
	applied := make(map[string]bool)

	rows, err := DB.Query("SELECT filename FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		applied[filename] = true
	}

	return applied, nil
}

// applyMigration applies a single migration file
func applyMigration(filename string) error {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	migrationPath := filepath.Join(wd, "database", "migrations", filename)

	// Read migration file
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", filename, err)
	}

	log.Printf("Executing migration SQL from %s", filename)

	// Execute migration
	if _, err := DB.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied
	_, err = DB.Exec("INSERT INTO schema_migrations (filename) VALUES ($1)", filename)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}
