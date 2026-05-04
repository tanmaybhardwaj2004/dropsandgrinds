//go:build integration

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/config"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

// TestDBConfig holds test database configuration
type TestDBConfig struct {
	DatabaseURL string
	RedisURL    string
}

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// Use test database URL from environment or default
	dbURL := "postgres://postgres:postgres@localhost:5432/dropsandgrinds_test?sslmode=disable"

	conn, err := config.ConnectDB(dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	return conn
}

// SetupTestRedis creates a test Redis connection
func SetupTestRedis(t *testing.T) *redis.Client {
	t.Helper()

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use DB 1 for tests
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Logf("Redis not available for tests: %v", err)
		return nil
	}

	return client
}

// TeardownTestDB closes the test database connection
func TeardownTestDB(t *testing.T, db *pgxpool.Pool) {
	t.Helper()
	if db != nil {
		db.Close()
	}
}

// TeardownTestRedis closes the test Redis connection
func TeardownTestRedis(t *testing.T, client *redis.Client) {
	t.Helper()
	if client != nil {
		client.Close()
	}
}

// CreateTestServices creates test services for integration tests
func CreateTestServices(t *testing.T, db *pgxpool.Pool, redis *redis.Client) (*repositories.CatalogRepository, *services.GamesService) {
	t.Helper()

	catalogRepo := repositories.NewCatalogRepository(db, redis)
	gamesService := services.NewGamesService(catalogRepo)

	return catalogRepo, gamesService
}

// CleanupTestData cleans up test data from database
func CleanupTestData(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()

	// Clean up in order of dependencies
	tables := []string{
		"review_scores",
		"clicks",
		"wishlist",
		"deals",
		"prices",
		"refresh_tokens",
		"users",
		"games",
	}

	for _, table := range tables {
		_, err := db.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("Failed to truncate table %s: %v", table, err)
		}
	}
}

// WaitForCondition waits for a condition to be true or timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for condition: %s", message)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}
