//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/pkg/cheapshark"
)

func TestCheapSharkClient_GetDeals(t *testing.T) {
	client := cheapshark.NewClient()
	ctx := context.Background()

	// Test fetching deals without filters
	deals, err := client.GetDeals(ctx, map[string]string{})
	if err != nil {
		t.Fatalf("GetDeals failed: %v", err)
	}

	if len(deals) == 0 {
		t.Log("No deals returned (API might be empty)")
	} else {
		t.Logf("Fetched %d deals", len(deals))
		// Verify deal structure
		for _, deal := range deals {
			if deal.GameID == "" {
				t.Error("Deal has empty GameID")
			}
			if deal.Title == "" {
				t.Error("Deal has empty Title")
			}
		}
	}
}

func TestCheapSharkClient_GetDealsWithFilters(t *testing.T) {
	client := cheapshark.NewClient()
	ctx := context.Background()

	// Test with store filter
	params := map[string]string{
		"storeID": "1", // Steam
	}

	deals, err := client.GetDeals(ctx, params)
	if err != nil {
		t.Fatalf("GetDeals with filters failed: %v", err)
	}

	t.Logf("Fetched %d Steam deals", len(deals))
}

func TestCheapSharkClient_GetGameDetails(t *testing.T) {
	client := cheapshark.NewClient()
	ctx := context.Background()

	// Use a known game ID (e.g., Portal 2)
	gameID := "128"

	game, err := client.GetGameDetails(ctx, gameID)
	if err != nil {
		t.Fatalf("GetGameDetails failed: %v", err)
	}

	if game.GameID != gameID {
		t.Errorf("Expected GameID %s, got %s", gameID, game.GameID)
	}

	if game.Title == "" {
		t.Error("Game has empty Title")
	}

	t.Logf("Game: %s (ID: %s)", game.Title, game.GameID)
}
