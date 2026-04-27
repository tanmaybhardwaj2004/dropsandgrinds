package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ClicksRepository handles click analytics operations
type ClicksRepository struct {
	pool *pgxpool.Pool
}

// NewClicksRepository creates a new clicks repository
func NewClicksRepository(pool *pgxpool.Pool) *ClicksRepository {
	return &ClicksRepository{pool: pool}
}

// LogClick logs a click event for a game
func (r *ClicksRepository) LogClick(ctx context.Context, userID, gameID int64, platform string) error {
	query := `
		INSERT INTO clicks (user_id, game_id, platform, clicked_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := r.pool.Exec(ctx, query, userID, gameID, platform)
	if err != nil {
		return fmt.Errorf("failed to log click: %w", err)
	}
	return nil
}

// GetUserConsentAnalytics checks if a user has consented to analytics tracking
func (r *ClicksRepository) GetUserConsentAnalytics(ctx context.Context, userID int64) (bool, error) {
	var consent bool
	err := r.pool.QueryRow(ctx, `
		SELECT consent_analytics FROM users WHERE id = $1
	`, userID).Scan(&consent)
	if err != nil {
		return false, fmt.Errorf("failed to get user consent: %w", err)
	}
	return consent, nil
}

// SetUserConsentAnalytics sets a user's consent for analytics tracking
func (r *ClicksRepository) SetUserConsentAnalytics(ctx context.Context, userID int64, consent bool) error {
	query := `
		UPDATE users SET consent_analytics = $2 WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, userID, consent)
	if err != nil {
		return fmt.Errorf("failed to set user consent: %w", err)
	}
	return nil
}
