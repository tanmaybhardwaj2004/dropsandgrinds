package scheduler

import (
	"context"
	"log"
	"time"

	"log/slog"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

// MeilisearchSyncJob syncs games from PostgreSQL to Meilisearch
func MeilisearchSyncJob(catalogRepo *repositories.CatalogRepository, meilisearchService *services.MeilisearchService, logger *slog.Logger) func(ctx context.Context) {
	return func(ctx context.Context) {
		if meilisearchService == nil {
			logger.Debug("Meilisearch not configured, skipping sync")
			return
		}

		logger.Info("starting Meilisearch sync")

		// Fetch all games from PostgreSQL
		games, _, err := catalogRepo.ListGames(ctx, "", "", 10000, 0, false, 0)
		if err != nil {
			logger.Error("failed to fetch games for Meilisearch sync", "error", err)
			return
		}

		// Index games in Meilisearch
		if err := meilisearchService.IndexGames(ctx, games); err != nil {
			logger.Error("failed to index games in Meilisearch", "error", err)
			return
		}

		logger.Info("Meilisearch sync completed", "games_indexed", len(games))
	}
}

// StartMeilisearchSync schedules periodic Meilisearch sync jobs
func StartMeilisearchSync(sched *Scheduler, catalogRepo *repositories.CatalogRepository, meilisearchService *services.MeilisearchService, logger *slog.Logger) {
	if meilisearchService == nil {
		log.Println("Meilisearch not configured, skipping sync scheduler")
		return
	}

	// Initial sync
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		MeilisearchSyncJob(catalogRepo, meilisearchService, logger)(ctx)
	}()

	// Schedule periodic sync every hour
	sched.AddJob(Job{
		Name:     "meilisearch-sync",
		Interval: 1 * time.Hour,
		Run: func(ctx context.Context) error {
			MeilisearchSyncJob(catalogRepo, meilisearchService, logger)(ctx)
			return nil
		},
	})
}
