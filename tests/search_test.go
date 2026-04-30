package tests

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

func TestSearchGames(t *testing.T) {
	ctx := context.Background()
	
	// Setup test database
	pool := setupTestDB(t)
	defer pool.Close()
	
	// Setup test Redis
	redisClient := setupTestRedis(t)
	defer redisClient.Close()
	
	// Create repository
	catalogRepo := repositories.NewCatalogRepository(pool, redisClient)
	
	// Insert test games
	insertTestGame(t, pool, models.Game{
		Title:    "Elden Ring",
		Platform: "steam",
		CoverURL: "https://example.com/elden-ring.jpg",
	})
	insertTestGame(t, pool, models.Game{
		Title:    "Hades",
		Platform: "epic",
		CoverURL: "https://example.com/hades.jpg",
	})
	insertTestGame(t, pool, models.Game{
		Title:    "Cyberpunk 2077",
		Platform: "gog",
		CoverURL: "https://example.com/cyberpunk.jpg",
	})
	
	t.Run("Search by title", func(t *testing.T) {
		games, total, err := catalogRepo.SearchGames(ctx, "Elden", "", 0, 0, 0, 0, 0, 0, 10, 0)
		
		require.NoError(t, err)
		assert.Greater(t, total, 0)
		assert.Len(t, games, 1)
		assert.Contains(t, games[0].Title, "Elden")
	})
	
	t.Run("Search with platform filter", func(t *testing.T) {
		games, total, err := catalogRepo.SearchGames(ctx, "", "steam", 0, 0, 0, 0, 0, 0, 10, 0)
		
		require.NoError(t, err)
		assert.Greater(t, total, 0)
		for _, game := range games {
			assert.Equal(t, "steam", game.Platform)
		}
	})
	
	t.Run("Search with price range filter", func(t *testing.T) {
		// Insert test prices
		insertTestPrice(t, pool, 1, 2999)
		insertTestPrice(t, pool, 2, 999)
		
		games, total, err := catalogRepo.SearchGames(ctx, "", "", 500, 2000, 0, 0, 0, 0, 10, 0)
		
		require.NoError(t, err)
		for _, game := range games {
			assert.GreaterOrEqual(t, game.PriceINR, 500.0)
			assert.LessOrEqual(t, game.PriceINR, 2000.0)
		}
		_ = total // Use total to avoid unused variable warning
	})
	
	t.Run("Search with pagination", func(t *testing.T) {
		// Insert more games
		for i := 0; i < 5; i++ {
			insertTestGame(t, pool, models.Game{
				Title:    "Test Game",
				Platform: "steam",
				CoverURL: "https://example.com/test.jpg",
			})
		}
		
		// First page
		games1, total1, err := catalogRepo.SearchGames(ctx, "Test", "", 0, 0, 0, 0, 0, 0, 2, 0)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(games1), 2)
		
		// Second page
		games2, total2, err := catalogRepo.SearchGames(ctx, "Test", "", 0, 0, 0, 0, 0, 0, 2, 2)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(games2), 2)
		assert.Equal(t, total1, total2)
	})
	
	t.Run("Search with empty query returns all games", func(t *testing.T) {
		games, total, err := catalogRepo.SearchGames(ctx, "", "", 0, 0, 0, 0, 0, 0, 10, 0)
		
		require.NoError(t, err)
		assert.Greater(t, total, 0)
		assert.Greater(t, len(games), 0)
	})
	
	t.Run("Search with review score filter", func(t *testing.T) {
		// Insert test review scores
		insertTestReviewScore(t, pool, 1, "metacritic", 95)
		insertTestReviewScore(t, pool, 2, "metacritic", 85)
		
		games, total, err := catalogRepo.SearchGames(ctx, "", "", 0, 0, 0, 0, 90, 100, 10, 0)
		
		require.NoError(t, err)
		for _, game := range games {
			assert.GreaterOrEqual(t, game.ReviewScore, 90.0)
		}
		_ = total
	})
}

func TestSearchGamesCaching(t *testing.T) {
	ctx := context.Background()
	
	pool := setupTestDB(t)
	defer pool.Close()
	
	redisClient := setupTestRedis(t)
	defer redisClient.Close()
	
	catalogRepo := repositories.NewCatalogRepository(pool, redisClient)
	
	// Insert test game
	insertTestGame(t, pool, models.Game{
		Title:    "Test Game",
		Platform: "steam",
		CoverURL: "https://example.com/test.jpg",
	})
	
	t.Run("Search results are cached", func(t *testing.T) {
		// First call - should cache
		games1, total1, err := catalogRepo.SearchGames(ctx, "Test", "", 0, 0, 0, 0, 0, 0, 10, 0)
		require.NoError(t, err)
		
		// Second call - should hit cache
		games2, total2, err := catalogRepo.SearchGames(ctx, "Test", "", 0, 0, 0, 0, 0, 0, 10, 0)
		require.NoError(t, err)
		
		assert.Equal(t, len(games1), len(games2))
		assert.Equal(t, total1, total2)
	})
}

func TestSearchGamesInvalidFilters(t *testing.T) {
	ctx := context.Background()
	
	pool := setupTestDB(t)
	defer pool.Close()
	
	redisClient := setupTestRedis(t)
	defer redisClient.Close()
	
	catalogRepo := repositories.NewCatalogRepository(pool, redisClient)
	
	t.Run("Search with invalid platform returns empty", func(t *testing.T) {
		games, total, err := catalogRepo.SearchGames(ctx, "", "invalid_platform", 0, 0, 0, 0, 0, 0, 10, 0)
		
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, games)
	})
	
	t.Run("Search with impossible price range returns empty", func(t *testing.T) {
		games, total, err := catalogRepo.SearchGames(ctx, "", "", 10000, 100, 0, 0, 0, 0, 10, 0)
		
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, games)
	})
}
