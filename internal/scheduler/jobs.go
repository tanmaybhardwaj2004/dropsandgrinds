package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/pkg/cheapshark"
)

// PriceRefreshJob fetches current CheapShark deals and updates matching catalog prices.
func PriceRefreshJob(repo *repositories.CatalogRepository, logger *slog.Logger) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		logger.Info("starting price refresh job")

		client := cheapshark.NewClient()
		deals, err := client.GetDeals(ctx, map[string]string{
			"pageSize": "60",
			"sortBy":   "Deal Rating",
		})
		if err != nil {
			return err
		}

		rate := repo.USDToINR(ctx)
		var matched int
		var skipped int
		for _, deal := range deals {
			gameID, err := repo.FindGameByTitle(ctx, deal.Title)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					skipped++
					continue
				}
				logger.Error("failed to find game for CheapShark deal", "title", deal.Title, "error", err)
				continue
			}

			currentPriceINR := dollarsToINR(float64(deal.SalePrice), rate)
			normalPriceINR := dollarsToINR(float64(deal.NormalPrice), rate)
			if currentPriceINR <= 0 {
				skipped++
				continue
			}

			if err := repo.InsertPrice(ctx, gameID, currentPriceINR, "CheapShark:"+deal.StoreID); err != nil {
				logger.Error("failed to insert CheapShark price", "game_id", gameID, "title", deal.Title, "error", err)
				continue
			}

			discountPercent := int(math.Round(float64(deal.Savings)))
			if normalPriceINR > 0 && discountPercent > 0 {
				if err := repo.UpdateDeal(ctx, gameID, normalPriceINR, discountPercent); err != nil {
					logger.Error("failed to update CheapShark deal", "game_id", gameID, "title", deal.Title, "error", err)
				}
			}
			matched++
		}

		logger.Info("price refresh job completed", "deals_seen", len(deals), "games_updated", matched, "games_skipped", skipped)
		return nil
	}
}

func dollarsToINR(price float64, rate float64) int {
	if price <= 0 {
		return 0
	}
	return int(math.Round(price * rate))
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
