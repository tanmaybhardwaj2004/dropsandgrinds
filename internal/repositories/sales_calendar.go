package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SaleEvent represents a sale event in the calendar
type SaleEvent struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Platform    string    `json:"platform"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	IsRecurring bool      `json:"is_recurring"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SalesCalendarRepository handles sale calendar operations
type SalesCalendarRepository struct {
	pool *pgxpool.Pool
}

// NewSalesCalendarRepository creates a new sales calendar repository
func NewSalesCalendarRepository(pool *pgxpool.Pool) *SalesCalendarRepository {
	return &SalesCalendarRepository{pool: pool}
}

// GetActiveSales returns all currently active sales
func (r *SalesCalendarRepository) GetActiveSales(ctx context.Context) ([]SaleEvent, error) {
	query := `
		SELECT id, name, platform, start_date, end_date, is_recurring, created_at, updated_at
		FROM sales_calendar
		WHERE start_date <= NOW() AND end_date >= NOW()
		ORDER BY end_date ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sales: %w", err)
	}
	defer rows.Close()

	var sales []SaleEvent
	for rows.Next() {
		var sale SaleEvent
		err := rows.Scan(
			&sale.ID, &sale.Name, &sale.Platform,
			&sale.StartDate, &sale.EndDate, &sale.IsRecurring,
			&sale.CreatedAt, &sale.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sale: %w", err)
		}
		sales = append(sales, sale)
	}

	return sales, nil
}

// GetUpcomingSales returns upcoming sales within the next 90 days
func (r *SalesCalendarRepository) GetUpcomingSales(ctx context.Context) ([]SaleEvent, error) {
	query := `
		SELECT id, name, platform, start_date, end_date, is_recurring, created_at, updated_at
		FROM sales_calendar
		WHERE start_date > NOW() AND start_date <= NOW() + INTERVAL '90 days'
		ORDER BY start_date ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming sales: %w", err)
	}
	defer rows.Close()

	var sales []SaleEvent
	for rows.Next() {
		var sale SaleEvent
		err := rows.Scan(
			&sale.ID, &sale.Name, &sale.Platform,
			&sale.StartDate, &sale.EndDate, &sale.IsRecurring,
			&sale.CreatedAt, &sale.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sale: %w", err)
		}
		sales = append(sales, sale)
	}

	return sales, nil
}

// GetAllSales returns all sales (for calendar view)
func (r *SalesCalendarRepository) GetAllSales(ctx context.Context) ([]SaleEvent, error) {
	query := `
		SELECT id, name, platform, start_date, end_date, is_recurring, created_at, updated_at
		FROM sales_calendar
		ORDER BY start_date ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all sales: %w", err)
	}
	defer rows.Close()

	var sales []SaleEvent
	for rows.Next() {
		var sale SaleEvent
		err := rows.Scan(
			&sale.ID, &sale.Name, &sale.Platform,
			&sale.StartDate, &sale.EndDate, &sale.IsRecurring,
			&sale.CreatedAt, &sale.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sale: %w", err)
		}
		sales = append(sales, sale)
	}

	return sales, nil
}

// GetSalesByPlatform returns sales for a specific platform
func (r *SalesCalendarRepository) GetSalesByPlatform(ctx context.Context, platform string) ([]SaleEvent, error) {
	query := `
		SELECT id, name, platform, start_date, end_date, is_recurring, created_at, updated_at
		FROM sales_calendar
		WHERE platform = $1
		ORDER BY start_date ASC
	`
	rows, err := r.pool.Query(ctx, query, platform)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales by platform: %w", err)
	}
	defer rows.Close()

	var sales []SaleEvent
	for rows.Next() {
		var sale SaleEvent
		err := rows.Scan(
			&sale.ID, &sale.Name, &sale.Platform,
			&sale.StartDate, &sale.EndDate, &sale.IsRecurring,
			&sale.CreatedAt, &sale.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sale: %w", err)
		}
		sales = append(sales, sale)
	}

	return sales, nil
}
