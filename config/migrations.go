package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read migrations directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, file := range files {
		var applied bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, file).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", file, err)
		}
		if applied {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		migrationCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		tx, err := pool.Begin(migrationCtx)
		if err != nil {
			cancel()
			return fmt.Errorf("begin migration %s: %w", file, err)
		}
		if _, err := tx.Exec(migrationCtx, string(data)); err != nil {
			_ = tx.Rollback(migrationCtx)
			cancel()
			return fmt.Errorf("apply migration %s: %w", file, err)
		}
		if _, err := tx.Exec(migrationCtx, `INSERT INTO schema_migrations (version) VALUES ($1)`, file); err != nil {
			_ = tx.Rollback(migrationCtx)
			cancel()
			return fmt.Errorf("record migration %s: %w", file, err)
		}
		if err := tx.Commit(migrationCtx); err != nil {
			cancel()
			return fmt.Errorf("commit migration %s: %w", file, err)
		}
		cancel()
	}

	return nil
}
