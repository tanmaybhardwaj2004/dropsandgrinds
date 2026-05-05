//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

func TestSavingsRepository(t *testing.T) {
	ctx := context.Background()

	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewSavingsRepository(pool)
	catalogRepo := repositories.NewCatalogRepository(pool, nil)

	userID := int64(1)

	// Insert test game
	gameID := int64(100)
	err := catalogRepo.InsertGame(ctx, gameID, "Test Game", "https://example.com/cover.jpg", "Action", "2024-01-01")
	if err != nil {
		t.Fatalf("Failed to insert test game: %v", err)
	}

	// Test logging a purchase
	err = repo.LogPurchase(ctx, userID, gameID, "Test Game", 500, 1000)
	if err != nil {
		t.Fatalf("Failed to log purchase: %v", err)
	}

	// Test getting total savings
	totalSavings, err := repo.GetTotalSavings(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get total savings: %v", err)
	}

	if totalSavings != 500 {
		t.Errorf("Expected total savings 500, got %d", totalSavings)
	}

	// Test getting monthly breakdown
	breakdown, err := repo.GetMonthlyBreakdown(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get monthly breakdown: %v", err)
	}

	if len(breakdown) != 1 {
		t.Errorf("Expected 1 month in breakdown, got %d", len(breakdown))
	}

	if breakdown[0].TotalSavings != 500 {
		t.Errorf("Expected monthly savings 500, got %d", breakdown[0].TotalSavings)
	}

	// Test getting purchase history
	purchases, total, err := repo.GetPurchaseHistory(ctx, userID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get purchase history: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 purchase, got %d", total)
	}

	if len(purchases) != 1 {
		t.Errorf("Expected 1 purchase in list, got %d", len(purchases))
	}

	if purchases[0].SavedAmountINR != 500 {
		t.Errorf("Expected saved amount 500, got %d", purchases[0].SavedAmountINR)
	}

	t.Log("Savings repository tests passed")
}

func TestSavingsRepositoryMultiplePurchases(t *testing.T) {
	ctx := context.Background()

	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewSavingsRepository(pool)
	catalogRepo := repositories.NewCatalogRepository(pool, nil)

	userID := int64(2)

	// Insert test games
	gameID1 := int64(200)
	gameID2 := int64(201)
	err := catalogRepo.InsertGame(ctx, gameID1, "Test Game 1", "https://example.com/cover1.jpg", "Action", "2024-01-01")
	if err != nil {
		t.Fatalf("Failed to insert test game 1: %v", err)
	}
	err = catalogRepo.InsertGame(ctx, gameID2, "Test Game 2", "https://example.com/cover2.jpg", "RPG", "2024-01-01")
	if err != nil {
		t.Fatalf("Failed to insert test game 2: %v", err)
	}

	// Log multiple purchases
	err = repo.LogPurchase(ctx, userID, gameID1, "Test Game 1", 300, 600)
	if err != nil {
		t.Fatalf("Failed to log purchase 1: %v", err)
	}
	err = repo.LogPurchase(ctx, userID, gameID2, "Test Game 2", 400, 800)
	if err != nil {
		t.Fatalf("Failed to log purchase 2: %v", err)
	}

	// Test total savings
	totalSavings, err := repo.GetTotalSavings(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get total savings: %v", err)
	}

	if totalSavings != 700 {
		t.Errorf("Expected total savings 700, got %d", totalSavings)
	}

	// Test purchase history pagination
	purchases, total, err := repo.GetPurchaseHistory(ctx, userID, 1, 0)
	if err != nil {
		t.Fatalf("Failed to get purchase history: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 purchases total, got %d", total)
	}

	if len(purchases) != 1 {
		t.Errorf("Expected 1 purchase in page, got %d", len(purchases))
	}

	t.Log("Multiple purchases test passed")
}

func TestSavingsRepositoryEmptyHistory(t *testing.T) {
	ctx := context.Background()

	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewSavingsRepository(pool)

	userID := int64(3)

	// Test total savings for user with no purchases
	totalSavings, err := repo.GetTotalSavings(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get total savings: %v", err)
	}

	if totalSavings != 0 {
		t.Errorf("Expected total savings 0, got %d", totalSavings)
	}

	// Test monthly breakdown for user with no purchases
	breakdown, err := repo.GetMonthlyBreakdown(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get monthly breakdown: %v", err)
	}

	if len(breakdown) != 0 {
		t.Errorf("Expected empty breakdown, got %d months", len(breakdown))
	}

	// Test purchase history for user with no purchases
	purchases, total, err := repo.GetPurchaseHistory(ctx, userID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get purchase history: %v", err)
	}

	if total != 0 {
		t.Errorf("Expected 0 purchases, got %d", total)
	}

	if len(purchases) != 0 {
		t.Errorf("Expected empty purchase list, got %d items", len(purchases))
	}

	t.Log("Empty history test passed")
}
