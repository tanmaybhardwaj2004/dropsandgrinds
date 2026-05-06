package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

// DealAlertRepository handles deal alerts database operations
type DealAlertRepository struct {
	pool *pgxpool.Pool
}

// NewDealAlertRepository creates a new deal alerts repository
func NewDealAlertRepository(pool *pgxpool.Pool) *DealAlertRepository {
	return &DealAlertRepository{pool: pool}
}

// Create creates a new deal alert
func (r *DealAlertRepository) Create(ctx context.Context, alert *models.DealAlert) (*models.DealAlert, error) {
	query := `
		INSERT INTO deal_alerts (user_id, game_id, target_price, store_id, region, currency, is_active, notification_sent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, game_id, target_price, store_id, region, currency, is_active, notification_sent, created_at, triggered_at
	`

	now := time.Now()
	alert.CreatedAt = now

	var createdAlert models.DealAlert
	err := r.pool.QueryRow(ctx, query,
		alert.UserID,
		alert.GameID,
		alert.TargetPrice,
		alert.StoreID,
		alert.Region,
		alert.Currency,
		alert.IsActive,
		alert.NotificationSent,
		now,
	).Scan(
		&createdAlert.ID,
		&createdAlert.UserID,
		&createdAlert.GameID,
		&createdAlert.TargetPrice,
		&createdAlert.StoreID,
		&createdAlert.Region,
		&createdAlert.Currency,
		&createdAlert.IsActive,
		&createdAlert.NotificationSent,
		&createdAlert.CreatedAt,
		&createdAlert.TriggeredAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create deal alert: %w", err)
	}

	return &createdAlert, nil
}

// GetByUserID gets all deal alerts for a user
func (r *DealAlertRepository) GetByUserID(ctx context.Context, userID int64) ([]models.DealAlert, error) {
	query := `
		SELECT id, user_id, game_id, target_price, store_id, region, currency, is_active, notification_sent, created_at, triggered_at
		FROM deal_alerts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deal alerts for user: %w", err)
	}
	defer rows.Close()

	var alerts []models.DealAlert
	for rows.Next() {
		var a models.DealAlert
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.GameID, &a.TargetPrice, &a.StoreID,
			&a.Region, &a.Currency, &a.IsActive, &a.NotificationSent,
			&a.CreatedAt, &a.TriggeredAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan deal alert: %w", err)
		}
		alerts = append(alerts, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deal alerts: %w", err)
	}

	return alerts, nil
}

// GetActiveAlerts gets all active deal alerts that haven't been notified yet
func (r *DealAlertRepository) GetActiveAlerts(ctx context.Context) ([]models.DealAlert, error) {
	query := `
		SELECT id, user_id, game_id, target_price, store_id, region, currency, is_active, notification_sent, created_at, triggered_at
		FROM deal_alerts
		WHERE is_active = TRUE AND notification_sent = FALSE
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active deal alerts: %w", err)
	}
	defer rows.Close()

	var alerts []models.DealAlert
	for rows.Next() {
		var a models.DealAlert
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.GameID, &a.TargetPrice, &a.StoreID,
			&a.Region, &a.Currency, &a.IsActive, &a.NotificationSent,
			&a.CreatedAt, &a.TriggeredAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan active deal alert: %w", err)
		}
		alerts = append(alerts, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active deal alerts: %w", err)
	}

	return alerts, nil
}

// MarkAlertTriggered marks a deal alert as triggered and notification sent
func (r *DealAlertRepository) MarkAlertTriggered(ctx context.Context, alertID int64) error {
	query := `
		UPDATE deal_alerts
		SET notification_sent = TRUE, triggered_at = $1
		WHERE id = $2
	`

	now := time.Now()
	_, err := r.pool.Exec(ctx, query, now, alertID)
	if err != nil {
		return fmt.Errorf("failed to mark alert as triggered: %w", err)
	}

	return nil
}

// Delete deletes a deal alert
func (r *DealAlertRepository) Delete(ctx context.Context, alertID, userID int64) error {
	query := `
		DELETE FROM deal_alerts
		WHERE id = $1 AND user_id = $2
	`

	result, err := r.pool.Exec(ctx, query, alertID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete deal alert: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// UpdateTargetPrice updates the target price for a deal alert
func (r *DealAlertRepository) UpdateTargetPrice(ctx context.Context, alertID, userID int64, newTargetPrice float64) error {
	query := `
		UPDATE deal_alerts
		SET target_price = $1, notification_sent = FALSE
		WHERE id = $2 AND user_id = $3
	`

	result, err := r.pool.Exec(ctx, query, newTargetPrice, alertID, userID)
	if err != nil {
		return fmt.Errorf("failed to update target price: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// GetByGameID gets all deal alerts for a specific game
func (r *DealAlertRepository) GetByGameID(ctx context.Context, gameID int64) ([]models.DealAlert, error) {
	query := `
		SELECT id, user_id, game_id, target_price, store_id, region, currency, is_active, notification_sent, created_at, triggered_at
		FROM deal_alerts
		WHERE game_id = $1 AND is_active = TRUE
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deal alerts for game: %w", err)
	}
	defer rows.Close()

	var alerts []models.DealAlert
	for rows.Next() {
		var a models.DealAlert
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.GameID, &a.TargetPrice, &a.StoreID,
			&a.Region, &a.Currency, &a.IsActive, &a.NotificationSent,
			&a.CreatedAt, &a.TriggeredAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan game deal alert: %w", err)
		}
		alerts = append(alerts, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating game deal alerts: %w", err)
	}

	return alerts, nil
}
