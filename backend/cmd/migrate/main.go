package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gridify/backend/internal/config"
	"github.com/jackc/pgx/v5"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run ./cmd/migrate [up|down]")
	}
	direction := os.Args[1]
	if direction != "up" && direction != "down" {
		log.Fatalf("Invalid direction: %s. Expected 'up' or 'down'", direction)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer conn.Close(ctx)

	// Ensure schema_migrations table exists
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(64) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		log.Fatalf("Failed to ensure schema_migrations table: %v", err)
	}

	migrationsDir := findMigrationsDir()
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory %s: %v", migrationsDir, err)
	}

	var targetFiles []string
	suffix := fmt.Sprintf(".%s.sql", direction)
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), suffix) {
			targetFiles = append(targetFiles, f.Name())
		}
	}
	sort.Strings(targetFiles)

	if direction == "down" {
		// Reverse order for down migrations
		for i, j := 0, len(targetFiles)-1; i < j; i, j = i+1, j-1 {
			targetFiles[i], targetFiles[j] = targetFiles[j], targetFiles[i]
		}
	}

	for _, filename := range targetFiles {
		version := strings.Split(filename, "_")[0]
		filePath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read file %s: %v", filename, err)
		}

		if direction == "up" {
			var exists bool
			err := conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
			if err != nil {
				log.Fatalf("Failed to check migration status for %s: %v", version, err)
			}
			if exists {
				fmt.Printf("[SKIP] %s already applied\n", filename)
				continue
			}

			fmt.Printf("[APPLYING] %s ...\n", filename)
			tx, err := conn.Begin(ctx)
			if err != nil {
				log.Fatalf("Failed to begin tx: %v", err)
			}

			if _, err := tx.Exec(ctx, string(content)); err != nil {
				_ = tx.Rollback(ctx)
				log.Fatalf("Failed to execute %s: %v", filename, err)
			}

			if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
				_ = tx.Rollback(ctx)
				log.Fatalf("Failed to record migration %s: %v", version, err)
			}

			if err := tx.Commit(ctx); err != nil {
				log.Fatalf("Failed to commit migration %s: %v", version, err)
			}
			fmt.Printf("[DONE] Applied %s successfully\n", filename)
		} else {
			// Down migration
			var exists bool
			err := conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
			if err != nil {
				log.Fatalf("Failed to check migration status for %s: %v", version, err)
			}
			if !exists {
				fmt.Printf("[SKIP] %s is not applied\n", filename)
				continue
			}

			fmt.Printf("[REVERTING] %s ...\n", filename)
			tx, err := conn.Begin(ctx)
			if err != nil {
				log.Fatalf("Failed to begin tx: %v", err)
			}

			if _, err := tx.Exec(ctx, string(content)); err != nil {
				_ = tx.Rollback(ctx)
				log.Fatalf("Failed to revert %s: %v", filename, err)
			}

			if _, err := tx.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", version); err != nil {
				_ = tx.Rollback(ctx)
				log.Fatalf("Failed to delete migration record %s: %v", version, err)
			}

			if err := tx.Commit(ctx); err != nil {
				log.Fatalf("Failed to commit rollback %s: %v", version, err)
			}
			fmt.Printf("[DONE] Reverted %s successfully\n", filename)
		}
	}
	fmt.Println("All migrations processed successfully.")
}

func findMigrationsDir() string {
	candidates := []string{
		"migrations",
		"../migrations",
		"../../migrations",
	}
	for _, c := range candidates {
		if stat, err := os.Stat(c); err == nil && stat.IsDir() {
			return c
		}
	}
	return "migrations"
}
