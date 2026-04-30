package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

type AnalyticsRepository struct {
	pool *pgxpool.Pool
}

func NewAnalyticsRepository(pool *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool}
}

func (r *AnalyticsRepository) StoreEvent(ctx context.Context, event models.AnalyticsEvent) error {
	query := `
		INSERT INTO analytics_events (user_id, event_type, event_data, page_url, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	return r.pool.QueryRow(ctx, query,
		event.UserID,
		event.EventType,
		event.EventData,
		event.PageURL,
		event.UserAgent,
		time.Now(),
	).Scan(&event.ID)
}

func (r *AnalyticsRepository) StoreEvents(ctx context.Context, events []models.AnalyticsEvent) error {
	for _, event := range events {
		if err := r.StoreEvent(ctx, event); err != nil {
			return err
		}
	}
	return nil
}
