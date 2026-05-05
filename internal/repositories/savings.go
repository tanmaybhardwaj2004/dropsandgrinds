package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SavingsRepository handles user savings operations
type SavingsRepository struct {
	pool *pgxpool.Pool
}

// NewSavingsRepository creates a new savings repository
func NewSavingsRepository(pool *pgxpool.Pool) *SavingsRepository {
	return &SavingsRepository{pool: pool}
}

// LogPurchase logs a purchase with savings amount
func (r *SavingsRepository) LogPurchase(ctx context.Context, userID, gameID int64, gameTitle string, paidPriceINR, originalPriceINR int) error {
	query := `
		INSERT INTO user_savings (user_id, game_id, game_title, paid_price_inr, original_price_inr, purchased_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	_, err := r.pool.Exec(ctx, query, userID, gameID, gameTitle, paidPriceINR, originalPriceINR)
	if err != nil {
		return fmt.Errorf("failed to log purchase: %w", err)
	}
	return nil
}

// GetTotalSavings returns the total savings for a user
func (r *SavingsRepository) GetTotalSavings(ctx context.Context, userID int64) (int, error) {
	var totalSavings int
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(saved_amount_inr), 0)
		FROM user_savings
		WHERE user_id = $1
	`, userID).Scan(&totalSavings)
	if err != nil {
		return 0, fmt.Errorf("failed to get total savings: %w", err)
	}
	return totalSavings, nil
}

// GetMonthlyBreakdown returns savings grouped by month for a user
func (r *SavingsRepository) GetMonthlyBreakdown(ctx context.Context, userID int64) ([]MonthlySavings, error) {
	query := `
		SELECT
			TO_CHAR(purchased_at, 'YYYY-MM') AS month,
			SUM(saved_amount_inr) AS total_savings,
			COUNT(*) AS purchase_count
		FROM user_savings
		WHERE user_id = $1
		GROUP BY TO_CHAR(purchased_at, 'YYYY-MM')
		ORDER BY month DESC
		LIMIT 12
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly breakdown: %w", err)
	}
	defer rows.Close()

	var breakdown []MonthlySavings
	for rows.Next() {
		var ms MonthlySavings
		if err := rows.Scan(&ms.Month, &ms.TotalSavings, &ms.PurchaseCount); err != nil {
			return nil, fmt.Errorf("failed to scan monthly savings: %w", err)
		}
		breakdown = append(breakdown, ms)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating monthly breakdown: %w", err)
	}

	return breakdown, nil
}

// GetPurchaseHistory returns purchase history for a user
func (r *SavingsRepository) GetPurchaseHistory(ctx context.Context, userID int64, limit, offset int) ([]PurchaseRecord, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM user_savings WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get purchase count: %w", err)
	}

	query := `
		SELECT
			id,
			game_id,
			game_title,
			paid_price_inr,
			original_price_inr,
			saved_amount_inr,
			purchased_at::text
		FROM user_savings
		WHERE user_id = $1
		ORDER BY purchased_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get purchase history: %w", err)
	}
	defer rows.Close()

	var purchases []PurchaseRecord
	for rows.Next() {
		var p PurchaseRecord
		if err := rows.Scan(
			&p.ID,
			&p.GameID,
			&p.GameTitle,
			&p.PaidPriceINR,
			&p.OriginalPriceINR,
			&p.SavedAmountINR,
			&p.PurchasedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan purchase record: %w", err)
		}
		purchases = append(purchases, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating purchase history: %w", err)
	}

	return purchases, total, nil
}

// MonthlySavings represents savings data for a month
type MonthlySavings struct {
	Month         string `json:"month"`
	TotalSavings  int    `json:"total_savings"`
	PurchaseCount int    `json:"purchase_count"`
}

// PurchaseRecord represents a single purchase record
type PurchaseRecord struct {
	ID               int64  `json:"id"`
	GameID           int64  `json:"game_id"`
	GameTitle        string `json:"game_title"`
	PaidPriceINR     int    `json:"paid_price_inr"`
	OriginalPriceINR int    `json:"original_price_inr"`
	SavedAmountINR   int    `json:"saved_amount_inr"`
	PurchasedAt      string `json:"purchased_at"`
}
