package scheduler

import (
	"context"
	"log/slog"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

// PriceRefreshJob fetches current CheapShark deals and updates matching catalog prices.
func PriceRefreshJob(repo *repositories.CatalogRepository, logger *slog.Logger) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		logger.Info("starting price refresh job")

		updated, err := repo.SyncCheapSharkDeals(ctx, 60)
		if err != nil {
			return err
		}
		logger.Info("price refresh job completed", "games_updated", updated)
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
