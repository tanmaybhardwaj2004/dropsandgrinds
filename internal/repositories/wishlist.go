package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

type WishlistRepository struct {
	db *pgxpool.Pool
}

func NewWishlistRepository(db *pgxpool.Pool) *WishlistRepository {
	return &WishlistRepository{db: db}
}

func (r *WishlistRepository) CreateWishlistItem(ctx context.Context, userID, gameID int64, targetPriceINR int) (models.WishlistItem, error) {
	query := `
		WITH inserted AS (
			INSERT INTO wishlists (user_id, game_id, target_price_inr)
			VALUES ($1, $2, $3)
			RETURNING id, user_id, game_id, target_price_inr, created_at
		)
		SELECT
			i.id,
			i.user_id,
			i.game_id,
			i.target_price_inr,
			COALESCE(p.price_inr, 0) AS current_price_inr,
			(COALESCE(p.price_inr, 0) > 0 AND COALESCE(p.price_inr, 0) <= i.target_price_inr) AS triggered,
			g.title,
			g.platform,
			g.cover_url,
			i.created_at::text
		FROM inserted i
		JOIN games g ON g.id = i.game_id
		LEFT JOIN LATERAL (
			SELECT price_inr
			FROM prices p
			WHERE p.game_id = i.game_id
			ORDER BY p.fetched_at DESC
			LIMIT 1
		) p ON TRUE
	`

	var item models.WishlistItem
	err := r.db.QueryRow(ctx, query, userID, gameID, targetPriceINR).Scan(
		&item.ID,
		&item.UserID,
		&item.GameID,
		&item.TargetPriceINR,
		&item.CurrentPriceINR,
		&item.Triggered,
		&item.Title,
		&item.Platform,
		&item.CoverURL,
		&item.CreatedAt,
	)
	if err != nil {
		return models.WishlistItem{}, err
	}

	return item, nil
}

func (r *WishlistRepository) ListWishlistItems(ctx context.Context, userID int64, limit, offset int) ([]models.WishlistItem, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM wishlists WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			w.id,
			w.user_id,
			w.game_id,
			w.target_price_inr,
			COALESCE(p.price_inr, 0) AS current_price_inr,
			(COALESCE(p.price_inr, 0) > 0 AND COALESCE(p.price_inr, 0) <= w.target_price_inr) AS triggered,
			g.title,
			g.platform,
			g.cover_url,
			w.created_at::text
		FROM wishlists w
		JOIN games g ON g.id = w.game_id
		LEFT JOIN LATERAL (
			SELECT price_inr
			FROM prices p
			WHERE p.game_id = w.game_id
			ORDER BY p.fetched_at DESC
			LIMIT 1
		) p ON TRUE
		WHERE w.user_id = $1
		ORDER BY w.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]models.WishlistItem, 0, limit)
	for rows.Next() {
		var item models.WishlistItem
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.GameID,
			&item.TargetPriceINR,
			&item.CurrentPriceINR,
			&item.Triggered,
			&item.Title,
			&item.Platform,
			&item.CoverURL,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *WishlistRepository) UpdateWishlistTarget(ctx context.Context, userID, wishlistID int64, targetPriceINR int) (models.WishlistItem, bool, error) {
	query := `
		WITH updated AS (
			UPDATE wishlists
			SET target_price_inr = $3, updated_at = NOW()
			WHERE id = $1 AND user_id = $2
			RETURNING id, user_id, game_id, target_price_inr, created_at
		)
		SELECT
			u.id,
			u.user_id,
			u.game_id,
			u.target_price_inr,
			COALESCE(p.price_inr, 0) AS current_price_inr,
			(COALESCE(p.price_inr, 0) > 0 AND COALESCE(p.price_inr, 0) <= u.target_price_inr) AS triggered,
			g.title,
			g.platform,
			g.cover_url,
			u.created_at::text
		FROM updated u
		JOIN games g ON g.id = u.game_id
		LEFT JOIN LATERAL (
			SELECT price_inr
			FROM prices p
			WHERE p.game_id = u.game_id
			ORDER BY p.fetched_at DESC
			LIMIT 1
		) p ON TRUE
	`

	var item models.WishlistItem
	err := r.db.QueryRow(ctx, query, wishlistID, userID, targetPriceINR).Scan(
		&item.ID,
		&item.UserID,
		&item.GameID,
		&item.TargetPriceINR,
		&item.CurrentPriceINR,
		&item.Triggered,
		&item.Title,
		&item.Platform,
		&item.CoverURL,
		&item.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.WishlistItem{}, false, nil
		}
		return models.WishlistItem{}, false, err
	}

	return item, true, nil
}

func (r *WishlistRepository) DeleteWishlistItem(ctx context.Context, userID, wishlistID int64) (bool, error) {
	result, err := r.db.Exec(ctx, `DELETE FROM wishlists WHERE id = $1 AND user_id = $2`, wishlistID, userID)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

type AlertCandidate struct {
	UserID          int64
	UserEmail       string
	GameID          int64
	GameTitle       string
	TargetPriceINR  int
	CurrentPriceINR int
	IsHistoricalLow bool
}

func (r *WishlistRepository) ListTriggeredAlertCandidates(ctx context.Context) ([]AlertCandidate, error) {
	rows, err := r.db.Query(ctx, `
		SELECT w.user_id, u.email, w.game_id, g.title, w.target_price_inr,
		       COALESCE(p.price_inr, 0), COALESCE(p.is_historical_low, FALSE)
		FROM wishlists w
		JOIN users u ON u.id = w.user_id
		JOIN games g ON g.id = w.game_id
		LEFT JOIN LATERAL (
			SELECT price_inr, is_historical_low
			FROM prices p
			WHERE p.game_id = w.game_id
			ORDER BY fetched_at DESC
			LIMIT 1
		) p ON TRUE
		WHERE u.consent_alerts = TRUE
		  AND COALESCE(p.price_inr, 0) > 0
		  AND (p.price_inr <= w.target_price_inr OR p.is_historical_low = TRUE)
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AlertCandidate
	for rows.Next() {
		var c AlertCandidate
		if err := rows.Scan(&c.UserID, &c.UserEmail, &c.GameID, &c.GameTitle, &c.TargetPriceINR, &c.CurrentPriceINR, &c.IsHistoricalLow); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
