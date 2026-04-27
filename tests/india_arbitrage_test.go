package tests

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

func TestIndiaArbitrageCalculation(t *testing.T) {
	ctx := context.Background()
	
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

	// Insert current price (India subsidized)
	currentPriceINR := 1499
	err = repo.InsertPrice(ctx, gameID, currentPriceINR, "Steam")
	if err != nil {
		t.Fatalf("Failed to insert current price: %v", err)
	}

	// Calculate arbitrage
	arbitrage, err := repo.GetIndiaArbitrage(ctx, gameID, 29.99) // $29.99 USD base price
	if err != nil {
		t.Fatalf("Failed to get India arbitrage: %v", err)
	}

	// Verify arbitrage calculation
	if arbitrage.BaseUSD != 29.99 {
		t.Errorf("Expected base USD 29.99, got %f", arbitrage.BaseUSD)
	}

	// Base USD to INR at ~83.2 rate
	expectedBaseINR := 29.99 * 83.2
	if arbitrage.BaseINR < expectedBaseINR-10 || arbitrage.BaseINR > expectedBaseINR+10 {
		t.Errorf("Base INR calculation seems off: got %f, expected around %f", arbitrage.BaseINR, expectedBaseINR)
	}

	// GST should be 18%
	expectedGST := arbitrage.BaseINR * 0.18
	if arbitrage.GST < expectedGST-5 || arbitrage.GST > expectedGST+5 {
		t.Errorf("GST calculation seems off: got %f, expected around %f", arbitrage.GST, expectedGST)
	}

	// True global value = base + GST
	expectedGlobalValue := arbitrage.BaseINR + arbitrage.GST
	if arbitrage.TrueGlobalValue < expectedGlobalValue-10 || arbitrage.TrueGlobalValue > expectedGlobalValue+10 {
		t.Errorf("True global value calculation seems off: got %f, expected around %f", arbitrage.TrueGlobalValue, expectedGlobalValue)
	}

	// Savings should be positive for subsidized pricing
	if arbitrage.SavingsINR <= 0 {
		t.Errorf("Expected positive savings for subsidized pricing, got %f", arbitrage.SavingsINR)
	}

	// Savings percentage
	expectedSavingsPercent := (arbitrage.SavingsINR / arbitrage.TrueGlobalValue) * 100
	if arbitrage.SavingsPercent < expectedSavingsPercent-5 || arbitrage.SavingsPercent > expectedSavingsPercent+5 {
		t.Errorf("Savings percent calculation seems off: got %f, expected around %f", arbitrage.SavingsPercent, expectedSavingsPercent)
	}

	t.Logf("Arbitrage: BaseINR=₹%.2f, GST=₹%.2f, TrueValue=₹%.2f, Current=₹%.2f, Savings=₹%.2f (%.1f%%)",
		arbitrage.BaseINR, arbitrage.GST, arbitrage.TrueGlobalValue, arbitrage.CurrentINR, arbitrage.SavingsINR, arbitrage.SavingsPercent)
}

func TestIndiaArbitrageNoSubsidy(t *testing.T) {
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

	// Insert price at or near global value (no subsidy)
	priceINR := 3500 // Approx $42 USD at 83.2 rate with GST
	err = repo.InsertPrice(ctx, gameID, priceINR, "Steam")
	if err != nil {
		t.Fatalf("Failed to insert price: %v", err)
	}

	arbitrage, err := repo.GetIndiaArbitrage(ctx, gameID, 42.0)
	if err != nil {
		t.Fatalf("Failed to get India arbitrage: %v", err)
	}

	// Savings should be minimal or negative for non-subsidized pricing
	if arbitrage.SavingsINR > 500 {
		t.Errorf("Expected minimal savings for non-subsidized pricing, got ₹%.2f", arbitrage.SavingsINR)
	}

	t.Logf("No subsidy case: Savings=₹%.2f (%.1f%%)", arbitrage.SavingsINR, arbitrage.SavingsPercent)
}

func TestIndiaArbitrageNonExistentGame(t *testing.T) {
	ctx := context.Background()
	
	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewCatalogRepository(pool, nil)

	// Try to get arbitrage for non-existent game
	arbitrage, err := repo.GetIndiaArbitrage(ctx, 99999, 29.99)
	if err != nil {
		t.Fatalf("GetIndiaArbitrage should not error for non-existent game: %v", err)
	}

	// Should return zero values
	if arbitrage.CurrentINR != 0 {
		t.Errorf("Expected zero current price for non-existent game, got %f", arbitrage.CurrentINR)
	}
}
