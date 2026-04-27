package scheduler

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

// PriceRefreshJob fetches current prices from external APIs and updates the database
// This is a skeleton implementation - actual API integration would call CheapShark, Steam, etc.
func PriceRefreshJob(repo *repositories.CatalogRepository, logger *slog.Logger) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		logger.Info("starting price refresh job")

		// Get all game IDs
		gameIDs, err := repo.GetAllGameIDs(ctx)
		if err != nil {
			return err
		}

		logger.Info("fetching prices for games", "count", len(gameIDs))

		// For each game, fetch and update prices
		// In production, this would call external APIs (CheapShark, Steam, etc.)
		for _, gameID := range gameIDs {
			// Simulate API call with random price variation
			// In production: call actual APIs here
			basePrice := 1000 + rand.Intn(2000) // Random price between 1000-3000
			discountPercent := rand.Intn(70)    // Random discount 0-70%

			currentPrice := basePrice * (100 - discountPercent) / 100
			originalPrice := basePrice

			// Insert new price entry
			if err := repo.InsertPrice(ctx, gameID, currentPrice, "Steam"); err != nil {
				logger.Error("failed to insert price", "game_id", gameID, "error", err)
				continue
			}

			// Update deal if there's a discount (> 10%)
			if discountPercent > 10 {
				if err := repo.UpdateDeal(ctx, gameID, originalPrice, discountPercent); err != nil {
					logger.Error("failed to update deal", "game_id", gameID, "error", err)
				}
			}

			// Small delay to avoid overwhelming APIs
			time.Sleep(10 * time.Millisecond)
		}

		logger.Info("price refresh job completed", "games_updated", len(gameIDs))
		return nil
	}
}

// ReviewRefreshJob refreshes review scores for all games with stale data
func ReviewRefreshJob(reviewService *services.ReviewService, logger *slog.Logger) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		logger.Info("starting review refresh job")

		if err := reviewService.RefreshAllStaleReviews(ctx); err != nil {
			logger.Error("failed to refresh stale reviews", "error", err)
			return err
		}

		logger.Info("review refresh job completed")
		return nil
	}
}
