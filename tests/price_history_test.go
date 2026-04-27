package tests

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

func TestPriceHistoryRepository(t *testing.T) {
	ctx := context.Background()
	
	// Setup test database
	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewCatalogRepository(pool, nil)

	// Insert test game
	gameID := int64(1)
	err := repo.InsertGame(ctx, gameID, "Test Game", "https://example.com/cover.jpg", "Action", "2024-01-01")
	if err != nil {
		t.Fatalf("Failed to insert test game: %v", err)
	}

	// Test inserting price history
	testPrices := []struct {
		price int
		store string
	}{
		{2999, "Steam"},
		{2499, "Steam"},
		{1999, "Steam"},
		{1499, "Steam"},
	}

	for i, p := range testPrices {
		err := repo.InsertPrice(ctx, gameID, p.price, p.store)
		if err != nil {
			t.Fatalf("Failed to insert price %d: %v", i, err)
		}
		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Test fetching price history
	prices, err := repo.GetPriceHistory(ctx, gameID, 10)
	if err != nil {
		t.Fatalf("Failed to get price history: %v", err)
	}

	if len(prices) != len(testPrices) {
		t.Errorf("Expected %d prices, got %d", len(testPrices), len(prices))
	}

	// Verify prices are in descending order (newest first)
	for i := 0; i < len(prices)-1; i++ {
		if prices[i].FetchedAt.Before(prices[i+1].FetchedAt) {
			t.Error("Prices not in descending order by date")
		}
	}

	t.Logf("Successfully retrieved %d price history entries", len(prices))
}

func TestPriceHistoryWithLimit(t *testing.T) {
	ctx := context.Background()
	
	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewCatalogRepository(pool, nil)

	gameID := int64(2)
	err := repo.InsertGame(ctx, gameID, "Test Game 2", "https://example.com/cover2.jpg", "RPG", "2024-01-01")
	if err != nil {
		t.Fatalf("Failed to insert test game: %v", err)
	}

	// Insert more prices than we'll request
	for i := 0; i < 15; i++ {
		err := repo.InsertPrice(ctx, gameID, 1000+i*100, "Steam")
		if err != nil {
			t.Fatalf("Failed to insert price %d: %v", i, err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Request only 10 prices
	prices, err := repo.GetPriceHistory(ctx, gameID, 10)
	if err != nil {
		t.Fatalf("Failed to get price history: %v", err)
	}

	if len(prices) > 10 {
		t.Errorf("Expected at most 10 prices, got %d", len(prices))
	}

	t.Logf("Retrieved %d prices with limit 10", len(prices))
}

func TestPriceHistoryNonExistentGame(t *testing.T) {
	ctx := context.Background()
	
	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewCatalogRepository(pool, nil)

	// Try to get price history for non-existent game
	prices, err := repo.GetPriceHistory(ctx, 99999, 10)
	if err != nil {
		t.Fatalf("GetPriceHistory should not error for non-existent game: %v", err)
	}

	if len(prices) != 0 {
		t.Errorf("Expected empty price history for non-existent game, got %d entries", len(prices))
	}
}
