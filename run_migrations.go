package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "postgres://postgres:postgres@localhost:5433/dropsandgrinds?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer pool.Close()

	// Ensure users table exists first so we can track migrations if we wanted to
	migrationsDir := "./migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}

	sort.Strings(sqlFiles)

	for _, file := range sqlFiles {
		path := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("Failed to read %s: %v", file, err)
		}

		fmt.Printf("Applying %s...\n", file)
		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			log.Printf("Warning: Failed to apply %s (it might have already been applied or contains errors): %v\n", file, err)
			// Continue to next since we don't have a strict migration state table here
		} else {
			fmt.Printf("Successfully applied %s\n", file)
		}
	}

	fmt.Println("Migration script finished.")
}
