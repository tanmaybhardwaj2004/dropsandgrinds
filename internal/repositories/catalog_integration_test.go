package repositories

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCatalogRepositoryIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping repository integration tests")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to connect test database: %v", err)
	}
	defer pool.Close()

	if err := resetAndMigrateTestDB(ctx, pool); err != nil {
		t.Fatalf("failed to prepare test database: %v", err)
	}

	repo := NewCatalogRepository(pool, nil)

	t.Run("ListGames", func(t *testing.T) {
		games, total, err := repo.ListGames(ctx, "", "", 20, 0, false, 0)
		if err != nil {
			t.Fatalf("ListGames returned error: %v", err)
		}
		if total <= 0 || len(games) <= 0 {
			t.Fatalf("expected non-empty games list, total=%d len=%d", total, len(games))
		}
	})

	t.Run("GetGameByID", func(t *testing.T) {
		game, found, err := repo.GetGameByID(ctx, 1)
		if err != nil {
			t.Fatalf("GetGameByID returned error: %v", err)
		}
		if !found {
			t.Fatal("expected game id 1 to be found")
		}
		if game.ID != 1 {
			t.Fatalf("expected game id 1, got %d", game.ID)
		}
	})

	t.Run("ListDeals", func(t *testing.T) {
		deals, total, err := repo.ListDeals(ctx, 20, 0)
		if err != nil {
			t.Fatalf("ListDeals returned error: %v", err)
		}
		if total <= 0 || len(deals) <= 0 {
			t.Fatalf("expected non-empty deals list, total=%d len=%d", total, len(deals))
		}
	})

	t.Run("GetPriceHistory", func(t *testing.T) {
		history, err := repo.GetPriceHistory(ctx, 1, 30, 0)
		if err != nil {
			t.Fatalf("GetPriceHistory returned error: %v", err)
		}
		if len(history) <= 0 {
			t.Fatal("expected non-empty price history")
		}
	})

	t.Run("ListGamesIncludesLowestPrice", func(t *testing.T) {
		games, _, err := repo.ListGames(ctx, "", "", 20, 0, false, 0)
		if err != nil {
			t.Fatalf("ListGames returned error: %v", err)
		}
		if len(games) == 0 {
			t.Fatal("expected non-empty games list")
		}
		if games[0].LowestPriceINR <= 0 {
			t.Fatalf("expected lowest_price_inr > 0, got %d", games[0].LowestPriceINR)
		}
	})
}

func resetAndMigrateTestDB(ctx context.Context, pool *pgxpool.Pool) error {
	cleanupSQL := []string{
		"DROP TABLE IF EXISTS deals CASCADE;",
		"DROP TABLE IF EXISTS prices CASCADE;",
		"DROP TABLE IF EXISTS games CASCADE;",
		"DROP TABLE IF EXISTS refresh_tokens CASCADE;",
		"DROP TABLE IF EXISTS users CASCADE;",
		"DROP FUNCTION IF EXISTS update_updated_at() CASCADE;",
	}
	for _, stmt := range cleanupSQL {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return err
		}
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return os.ErrInvalid
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))

	migrationFiles := []string{
		filepath.Join(repoRoot, "migrations", "001_create_users_table.sql"),
		filepath.Join(repoRoot, "migrations", "002_add_auth_columns_and_refresh_tokens.sql"),
		filepath.Join(repoRoot, "migrations", "003_create_games_prices_deals.sql"),
	}

	for _, path := range migrationFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return err
		}
	}

	return nil
}
