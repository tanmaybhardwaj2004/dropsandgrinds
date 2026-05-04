//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

func TestLibraryRepository(t *testing.T) {
	ctx := context.Background()

	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewLibraryRepository(pool)

	userID := int64(1)

	// Test importing owned games
	testAppIDs := []int64{1001, 1002, 1003, 1004, 1005}
	err := repo.ImportOwnedGames(ctx, userID, testAppIDs)
	if err != nil {
		t.Fatalf("Failed to import owned games: %v", err)
	}

	// Test getting owned games
	ownedAppIDs, err := repo.GetOwnedGames(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get owned games: %v", err)
	}

	if len(ownedAppIDs) != len(testAppIDs) {
		t.Errorf("Expected %d owned games, got %d", len(testAppIDs), len(ownedAppIDs))
	}

	// Verify all app IDs were imported
	appIDMap := make(map[int64]bool)
	for _, id := range ownedAppIDs {
		appIDMap[id] = true
	}
	for _, expectedID := range testAppIDs {
		if !appIDMap[expectedID] {
			t.Errorf("Expected app ID %d not found in owned games", expectedID)
		}
	}

	t.Logf("Successfully imported and retrieved %d owned games", len(ownedAppIDs))
}

func TestLibraryRepositoryUpdate(t *testing.T) {
	ctx := context.Background()

	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewLibraryRepository(pool)

	userID := int64(2)

	// Import initial games
	initialApps := []int64{2001, 2002}
	err := repo.ImportOwnedGames(ctx, userID, initialApps)
	if err != nil {
		t.Fatalf("Failed to import initial games: %v", err)
	}

	// Update with new set (should replace old set)
	updatedApps := []int64{2003, 2004, 2005}
	err = repo.ImportOwnedGames(ctx, userID, updatedApps)
	if err != nil {
		t.Fatalf("Failed to update library: %v", err)
	}

	// Verify only new games exist
	ownedAppIDs, err := repo.GetOwnedGames(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get owned games: %v", err)
	}

	if len(ownedAppIDs) != len(updatedApps) {
		t.Errorf("Expected %d owned games after update, got %d", len(updatedApps), len(ownedAppIDs))
	}

	// Verify old games are gone
	appIDMap := make(map[int64]bool)
	for _, id := range ownedAppIDs {
		appIDMap[id] = true
	}
	for _, oldID := range initialApps {
		if appIDMap[oldID] {
			t.Errorf("Old app ID %d should not exist after update", oldID)
		}
	}

	t.Log("Successfully updated library (replaced old games)")
}

func TestLibraryRepositoryLinkGame(t *testing.T) {
	ctx := context.Background()

	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewLibraryRepository(pool)
	catalogRepo := repositories.NewCatalogRepository(pool, nil)

	userID := int64(3)
	steamAppID := int64(3001)

	// Insert test game
	gameID := int64(10)
	err := catalogRepo.InsertGame(ctx, gameID, "Test Game", "https://example.com/cover.jpg", "Action", "2024-01-01")
	if err != nil {
		t.Fatalf("Failed to insert test game: %v", err)
	}

	// Import Steam app ID without game link
	err = repo.ImportOwnedGames(ctx, userID, []int64{steamAppID})
	if err != nil {
		t.Fatalf("Failed to import Steam app: %v", err)
	}

	// Link Steam app to game
	err = repo.LinkSteamAppToGame(ctx, userID, steamAppID, gameID)
	if err != nil {
		t.Fatalf("Failed to link Steam app to game: %v", err)
	}

	// Verify link by getting owned game IDs
	ownedGameIDs, err := repo.GetOwnedGameIDs(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get owned game IDs: %v", err)
	}

	if len(ownedGameIDs) != 1 {
		t.Errorf("Expected 1 owned game ID, got %d", len(ownedGameIDs))
	}

	if ownedGameIDs[0] != gameID {
		t.Errorf("Expected game ID %d, got %d", gameID, ownedGameIDs[0])
	}

	t.Log("Successfully linked Steam app to game")
}

func TestLibraryRepositoryIsGameOwned(t *testing.T) {
	ctx := context.Background()

	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewLibraryRepository(pool)
	catalogRepo := repositories.NewCatalogRepository(pool, nil)

	userID := int64(4)
	gameID := int64(20)

	// Insert test game
	err := catalogRepo.InsertGame(ctx, gameID, "Test Game", "https://example.com/cover.jpg", "Action", "2024-01-01")
	if err != nil {
		t.Fatalf("Failed to insert test game: %v", err)
	}

	// Import Steam app and link to game
	steamAppID := int64(4001)
	err = repo.ImportOwnedGames(ctx, userID, []int64{steamAppID})
	if err != nil {
		t.Fatalf("Failed to import Steam app: %v", err)
	}
	err = repo.LinkSteamAppToGame(ctx, userID, steamAppID, gameID)
	if err != nil {
		t.Fatalf("Failed to link Steam app to game: %v", err)
	}

	// Test is owned
	owned, err := repo.IsGameOwned(ctx, userID, gameID)
	if err != nil {
		t.Fatalf("Failed to check if game is owned: %v", err)
	}

	if !owned {
		t.Error("Expected game to be owned")
	}

	// Test non-owned game
	owned, err = repo.IsGameOwned(ctx, userID, 999)
	if err != nil {
		t.Fatalf("Failed to check if game is owned: %v", err)
	}

	if owned {
		t.Error("Expected game to not be owned")
	}

	t.Log("Successfully tested IsGameOwned")
}

func TestLibraryRepositoryGetCount(t *testing.T) {
	ctx := context.Background()

	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewLibraryRepository(pool)

	userID := int64(5)

	// Import games
	testApps := []int64{5001, 5002, 5003}
	err := repo.ImportOwnedGames(ctx, userID, testApps)
	if err != nil {
		t.Fatalf("Failed to import games: %v", err)
	}

	// Get count
	count, err := repo.GetLibraryCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get library count: %v", err)
	}

	if count != len(testApps) {
		t.Errorf("Expected count %d, got %d", len(testApps), count)
	}

	t.Logf("Library count: %d", count)
}

func TestLibraryRepositoryEmptyLibrary(t *testing.T) {
	ctx := context.Background()

	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewLibraryRepository(pool)

	userID := int64(6)

	// Get owned games for user with no library
	ownedAppIDs, err := repo.GetOwnedGames(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get owned games: %v", err)
	}

	if len(ownedAppIDs) != 0 {
		t.Errorf("Expected empty library, got %d games", len(ownedAppIDs))
	}

	// Get count
	count, err := repo.GetLibraryCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get library count: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	t.Log("Successfully handled empty library")
}
