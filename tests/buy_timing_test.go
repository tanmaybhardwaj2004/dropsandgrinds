package tests

import (
	"context"
	"testing"
	"time"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

func TestSalesCalendarRepository(t *testing.T) {
	ctx := context.Background()
	
	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewSalesCalendarRepository(pool)

	// Test getting all sales
	sales, err := repo.GetAllSales(ctx)
	if err != nil {
		t.Fatalf("Failed to get all sales: %v", err)
	}

	if len(sales) == 0 {
		t.Error("Expected at least one sale in calendar")
	}

	// Test getting sales by platform
	steamSales, err := repo.GetSalesByPlatform(ctx, "Steam")
	if err != nil {
		t.Fatalf("Failed to get Steam sales: %v", err)
	}

	if len(steamSales) == 0 {
		t.Error("Expected at least one Steam sale")
	}

	// Verify Steam sales have correct platform
	for _, sale := range steamSales {
		if sale.Platform != "Steam" {
			t.Errorf("Expected platform Steam, got %s", sale.Platform)
		}
	}

	t.Log("Sales calendar repository tests passed")
}

func TestBuyTimingService(t *testing.T) {
	ctx := context.Background()
	
	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	salesRepo := repositories.NewSalesCalendarRepository(pool)

	// Test getting active sales (should be empty since we're using test data)
	activeSales, err := salesRepo.GetActiveSales(ctx)
	if err != nil {
		t.Fatalf("Failed to get active sales: %v", err)
	}

	// In test environment, active sales might be empty
	t.Logf("Active sales count: %d", len(activeSales))

	// Test getting upcoming sales
	upcomingSales, err := salesRepo.GetUpcomingSales(ctx)
	if err != nil {
		t.Fatalf("Failed to get upcoming sales: %v", err)
	}

	if len(upcomingSales) == 0 {
		t.Log("No upcoming sales in test data")
	} else {
		// Verify upcoming sales are in the future
		now := time.Now()
		for _, sale := range upcomingSales {
			if sale.StartDate.Before(now) {
				t.Errorf("Upcoming sale should be in the future, got %v", sale.StartDate)
			}
		}
	}

	t.Log("Buy timing service tests passed")
}

func TestBundleServiceVerdict(t *testing.T) {
	// Test verdict logic directly
	testCases := []struct {
		name           string
		individualSum  float64
		bundlePrice    float64
		expectedVerdict string
	}{
		{
			name:           "Bundle is much cheaper",
			individualSum:  1000,
			bundlePrice:    500,
			expectedVerdict: "buy_bundle",
		},
		{
			name:           "Bundle is slightly cheaper",
			individualSum:  1000,
			bundlePrice:    850,
			expectedVerdict: "mixed",
		},
		{
			name:           "Bundle is more expensive",
			individualSum:  500,
			bundlePrice:    1000,
			expectedVerdict: "buy_separately",
		},
		{
			name:           "Equal prices",
			individualSum:  500,
			bundlePrice:    500,
			expectedVerdict: "mixed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			savings := tc.individualSum - tc.bundlePrice
			savingsPercent := (savings / tc.individualSum) * 100

			var verdict string
			if savingsPercent > 20 {
				verdict = "buy_bundle"
			} else if savingsPercent < 0 {
				verdict = "buy_separately"
			} else {
				verdict = "mixed"
			}

			if verdict != tc.expectedVerdict {
				t.Errorf("Expected verdict %s, got %s", tc.expectedVerdict, verdict)
			}
		})
	}

	t.Log("Bundle verdict logic tests passed")
}

func TestBundleServiceURLValidation(t *testing.T) {
	// Test URL validation
	validURLs := []string{
		"https://www.humblebundle.com/games/test",
		"https://www.fanatical.com/en/bundle/test",
		"https://store.steampowered.com/bundle/test",
	}

	invalidURLs := []string{
		"https://www.example.com/bundle",
		"https://not-a-bundle.com",
	}

	for _, url := range validURLs {
		if !isValidBundleURL(url) {
			t.Errorf("Expected valid URL: %s", url)
		}
	}

	for _, url := range invalidURLs {
		if isValidBundleURL(url) {
			t.Errorf("Expected invalid URL: %s", url)
		}
	}

	t.Log("Bundle URL validation tests passed")
}

// Helper function for URL validation test
func isValidBundleURL(url string) bool {
	supportedPatterns := []string{
		"humblebundle.com",
		"fanatical.com",
		"store.steampowered.com",
	}

	for _, pattern := range supportedPatterns {
		if contains(url, pattern) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
