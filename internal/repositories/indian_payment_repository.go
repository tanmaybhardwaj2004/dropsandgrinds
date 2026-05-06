package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

// IndianPaymentRepository handles Indian payment offers database operations
type IndianPaymentRepository struct {
	pool *pgxpool.Pool
}

// NewIndianPaymentRepository creates a new Indian payment offers repository
func NewIndianPaymentRepository(pool *pgxpool.Pool) *IndianPaymentRepository {
	return &IndianPaymentRepository{pool: pool}
}

// scanOffer scans a single row into an IndianPaymentOffer struct
func scanOffer(row pgx.Row) (models.IndianPaymentOffer, error) {
	var o models.IndianPaymentOffer
	err := row.Scan(
		&o.ID, &o.StoreID, &o.OfferType, &o.Provider, &o.Description,
		&o.DiscountPercent, &o.MaxDiscountAmount, &o.MinOrderAmount,
		&o.ValidFrom, &o.ValidUntil, &o.IsActive, &o.CreatedAt,
	)
	return o, err
}

// scanOffers scans multiple rows into a slice of IndianPaymentOffer structs
func scanOffers(rows pgx.Rows) ([]models.IndianPaymentOffer, error) {
	defer rows.Close()

	var offers []models.IndianPaymentOffer
	for rows.Next() {
		var o models.IndianPaymentOffer
		if err := rows.Scan(
			&o.ID, &o.StoreID, &o.OfferType, &o.Provider, &o.Description,
			&o.DiscountPercent, &o.MaxDiscountAmount, &o.MinOrderAmount,
			&o.ValidFrom, &o.ValidUntil, &o.IsActive, &o.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan Indian payment offer: %w", err)
		}
		offers = append(offers, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating Indian payment offers: %w", err)
	}

	return offers, nil
}

const offerColumns = `id, store_id, offer_type, provider, description, discount_percent,
	max_discount_amount, min_order_amount, valid_from, valid_until, is_active, created_at`

// GetActiveOffers gets all active Indian payment offers
func (r *IndianPaymentRepository) GetActiveOffers(ctx context.Context) ([]models.IndianPaymentOffer, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM indian_payment_offers
		WHERE is_active = TRUE
		  AND (valid_until IS NULL OR valid_until > NOW())
		ORDER BY created_at DESC
	`, offerColumns)

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active Indian payment offers: %w", err)
	}

	return scanOffers(rows)
}

// GetOffersByStore gets Indian payment offers for a specific store
func (r *IndianPaymentRepository) GetOffersByStore(ctx context.Context, storeID int64) ([]models.IndianPaymentOffer, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM indian_payment_offers
		WHERE store_id = $1 AND is_active = TRUE
		  AND (valid_until IS NULL OR valid_until > NOW())
		ORDER BY created_at DESC
	`, offerColumns)

	rows, err := r.pool.Query(ctx, query, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Indian payment offers for store: %w", err)
	}

	return scanOffers(rows)
}

// GetOffersByProvider gets Indian payment offers by payment provider
func (r *IndianPaymentRepository) GetOffersByProvider(ctx context.Context, provider string) ([]models.IndianPaymentOffer, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM indian_payment_offers
		WHERE provider = $1 AND is_active = TRUE
		  AND (valid_until IS NULL OR valid_until > NOW())
		ORDER BY created_at DESC
	`, offerColumns)

	rows, err := r.pool.Query(ctx, query, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get Indian payment offers for provider: %w", err)
	}

	return scanOffers(rows)
}

// CreateOffer creates a new Indian payment offer
func (r *IndianPaymentRepository) CreateOffer(ctx context.Context, offer *models.IndianPaymentOffer) (*models.IndianPaymentOffer, error) {
	query := fmt.Sprintf(`
		INSERT INTO indian_payment_offers (store_id, offer_type, provider, description, discount_percent,
		                                   max_discount_amount, min_order_amount, valid_from, valid_until, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING %s
	`, offerColumns)

	now := time.Now()
	offer.CreatedAt = now

	createdOffer, err := scanOffer(r.pool.QueryRow(ctx, query,
		offer.StoreID,
		offer.OfferType,
		offer.Provider,
		offer.Description,
		offer.DiscountPercent,
		offer.MaxDiscountAmount,
		offer.MinOrderAmount,
		offer.ValidFrom,
		offer.ValidUntil,
		offer.IsActive,
		now,
	))

	if err != nil {
		return nil, fmt.Errorf("failed to create Indian payment offer: %w", err)
	}

	return &createdOffer, nil
}

// UpdateOffer updates an existing Indian payment offer
func (r *IndianPaymentRepository) UpdateOffer(ctx context.Context, offer *models.IndianPaymentOffer) error {
	query := `
		UPDATE indian_payment_offers
		SET store_id = $1, offer_type = $2, provider = $3, description = $4, discount_percent = $5,
		    max_discount_amount = $6, min_order_amount = $7, valid_from = $8, valid_until = $9, is_active = $10
		WHERE id = $11
	`

	result, err := r.pool.Exec(ctx, query,
		offer.StoreID,
		offer.OfferType,
		offer.Provider,
		offer.Description,
		offer.DiscountPercent,
		offer.MaxDiscountAmount,
		offer.MinOrderAmount,
		offer.ValidFrom,
		offer.ValidUntil,
		offer.IsActive,
		offer.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update Indian payment offer: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// DeleteOffer deletes an Indian payment offer
func (r *IndianPaymentRepository) DeleteOffer(ctx context.Context, offerID int64) error {
	query := `DELETE FROM indian_payment_offers WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, offerID)
	if err != nil {
		return fmt.Errorf("failed to delete Indian payment offer: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
