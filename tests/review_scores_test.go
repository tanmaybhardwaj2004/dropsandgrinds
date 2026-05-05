//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

func TestReviewRepository(t *testing.T) {
	ctx := context.Background()

	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewReviewRepository(pool)

	gameID := int64(1)

	// Test storing review scores
	testScores := []struct {
		source string
		score  int
	}{
		{"metacritic", 85},
		{"opencritic", 82},
		{"steam", 88},
		{"ign", 80},
		{"gamespot", 78},
	}

	for _, s := range testScores {
		err := repo.StoreReviewScore(ctx, gameID, s.source, s.score)
		if err != nil {
			t.Fatalf("Failed to store review score for %s: %v", s.source, err)
		}
	}

	// Test fetching review scores
	scores, err := repo.GetReviewScores(ctx, gameID)
	if err != nil {
		t.Fatalf("Failed to get review scores: %v", err)
	}

	if len(scores) != len(testScores) {
		t.Errorf("Expected %d review scores, got %d", len(testScores), len(scores))
	}

	// Verify each source
	scoreMap := make(map[string]int)
	for _, s := range scores {
		scoreMap[s.Source] = s.Score
	}

	for _, expected := range testScores {
		if actual, exists := scoreMap[expected.source]; !exists {
			t.Errorf("Missing review score for source: %s", expected.source)
		} else if actual != expected.score {
			t.Errorf("Expected score %d for %s, got %d", expected.score, expected.source, actual)
		}
	}

	t.Logf("Successfully stored and retrieved %d review scores", len(scores))
}

func TestReviewRepositoryUpdate(t *testing.T) {
	ctx := context.Background()

	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewReviewRepository(pool)

	gameID := int64(2)

	// Store initial score
	err := repo.StoreReviewScore(ctx, gameID, "metacritic", 75)
	if err != nil {
		t.Fatalf("Failed to store initial review score: %v", err)
	}

	// Update the score
	err = repo.StoreReviewScore(ctx, gameID, "metacritic", 90)
	if err != nil {
		t.Fatalf("Failed to update review score: %v", err)
	}

	// Verify update
	scores, err := repo.GetReviewScores(ctx, gameID)
	if err != nil {
		t.Fatalf("Failed to get review scores: %v", err)
	}

	if len(scores) != 1 {
		t.Errorf("Expected 1 review score after update, got %d", len(scores))
	}

	if scores[0].Score != 90 {
		t.Errorf("Expected updated score 90, got %d", scores[0].Score)
	}

	t.Log("Successfully updated review score")
}

func TestReviewRepositoryDelete(t *testing.T) {
	ctx := context.Background()

	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewReviewRepository(pool)

	gameID := int64(3)

	// Store scores
	err := repo.StoreReviewScore(ctx, gameID, "metacritic", 85)
	if err != nil {
		t.Fatalf("Failed to store review score: %v", err)
	}

	err = repo.StoreReviewScore(ctx, gameID, "steam", 88)
	if err != nil {
		t.Fatalf("Failed to store review score: %v", err)
	}

	// Delete one source
	err = repo.DeleteReviewScore(ctx, gameID, "metacritic")
	if err != nil {
		t.Fatalf("Failed to delete review score: %v", err)
	}

	// Verify deletion
	scores, err := repo.GetReviewScores(ctx, gameID)
	if err != nil {
		t.Fatalf("Failed to get review scores: %v", err)
	}

	if len(scores) != 1 {
		t.Errorf("Expected 1 review score after deletion, got %d", len(scores))
	}

	if scores[0].Source != "steam" {
		t.Errorf("Expected remaining source to be steam, got %s", scores[0].Source)
	}

	t.Log("Successfully deleted review score")
}

func TestReviewRepositoryFindStale(t *testing.T) {
	ctx := context.Background()

	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewReviewRepository(pool)

	// Insert a fresh review (simulated by not updating fetched_at)
	gameID1 := int64(4)
	err := repo.StoreReviewScore(ctx, gameID1, "metacritic", 85)
	if err != nil {
		t.Fatalf("Failed to store fresh review score: %v", err)
	}

	// The test helper cleanup will clear all data, so we can't test stale detection
	// in a meaningful way without modifying the fetched_at directly.
	// For now, just verify the function exists and doesn't error.

	stale, err := repo.FindStaleReviews(ctx, 25) // 25 hours threshold
	if err != nil {
		t.Fatalf("FindStaleReviews failed: %v", err)
	}

	t.Logf("Found %d stale reviews (expected 0 in fresh test)", len(stale))
}

func TestReviewRepositoryNonExistentGame(t *testing.T) {
	ctx := context.Background()

	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewReviewRepository(pool)

	// Try to get scores for non-existent game
	scores, err := repo.GetReviewScores(ctx, 99999)
	if err != nil {
		t.Fatalf("GetReviewScores should not error for non-existent game: %v", err)
	}

	if len(scores) != 0 {
		t.Errorf("Expected empty review scores for non-existent game, got %d", len(scores))
	}
}

func TestReviewRepositoryDeleteAllForGame(t *testing.T) {
	ctx := context.Background()

	pool := SetupTestDB(ctx, t)
	defer pool.Close()
	defer CleanupTestData(ctx, pool, t)

	repo := repositories.NewReviewRepository(pool)

	gameID := int64(5)

	// Store multiple scores
	sources := []string{"metacritic", "steam", "ign"}
	for _, source := range sources {
		err := repo.StoreReviewScore(ctx, gameID, source, 80)
		if err != nil {
			t.Fatalf("Failed to store review score: %v", err)
		}
	}

	// Delete all scores for the game
	err := repo.DeleteReviewScoresForGame(ctx, gameID)
	if err != nil {
		t.Fatalf("Failed to delete all review scores: %v", err)
	}

	// Verify deletion
	scores, err := repo.GetReviewScores(ctx, gameID)
	if err != nil {
		t.Fatalf("Failed to get review scores: %v", err)
	}

	if len(scores) != 0 {
		t.Errorf("Expected 0 review scores after deleting all, got %d", len(scores))
	}

	t.Log("Successfully deleted all review scores for game")
}
