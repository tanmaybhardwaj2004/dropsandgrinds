package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

type CatalogRepository struct {
	db *pgxpool.Pool
}

func NewCatalogRepository(db *pgxpool.Pool) *CatalogRepository {
	return &CatalogRepository{db: db}
}

func (r *CatalogRepository) ListGames(ctx context.Context, query, platform string, limit, offset int) ([]models.Game, int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM games g
		WHERE ($1 = '' OR LOWER(g.title) LIKE '%' || LOWER($1) || '%')
		  AND ($2 = '' OR LOWER(g.platform) = LOWER($2))
	`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, query, platform).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT
			g.id,
			g.title,
			g.platform,
			g.cover_url,
			COALESCE(p.price_inr, 0) AS price_inr,
			COALESCE(p_low.lowest_price_inr, COALESCE(p.price_inr, 0), 0) AS lowest_price_inr,
			(COALESCE(p.price_inr, 0) > 0 AND COALESCE(p.price_inr, 0) = COALESCE(p_low.lowest_price_inr, COALESCE(p.price_inr, 0), 0)) AS is_all_time_low,
			COALESCE(d.original_inr, 0) AS original_inr,
			COALESCE(d.discount_percent, 0) AS discount_percent,
			0 AS review_score
		FROM games g
		LEFT JOIN LATERAL (
			SELECT price_inr
			FROM prices p
			WHERE p.game_id = g.id
			ORDER BY p.fetched_at DESC
			LIMIT 1
		) p ON TRUE
		LEFT JOIN LATERAL (
			SELECT MIN(price_inr) AS lowest_price_inr
			FROM prices p_low
			WHERE p_low.game_id = g.id
		) p_low ON TRUE
		LEFT JOIN LATERAL (
			SELECT original_inr, discount_percent
			FROM deals d
			WHERE d.game_id = g.id AND d.is_active = TRUE
			ORDER BY d.cached_at DESC
			LIMIT 1
		) d ON TRUE
		WHERE ($1 = '' OR LOWER(g.title) LIKE '%' || LOWER($1) || '%')
		  AND ($2 = '' OR LOWER(g.platform) = LOWER($2))
		ORDER BY COALESCE(d.discount_percent, 0) DESC, g.title ASC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, dataQuery, query, platform, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	games := make([]models.Game, 0, limit)
	for rows.Next() {
		var g models.Game
		if err := rows.Scan(
			&g.ID,
			&g.Title,
			&g.Platform,
			&g.CoverURL,
			&g.PriceINR,
			&g.LowestPriceINR,
			&g.IsAllTimeLow,
			&g.OriginalINR,
			&g.DiscountPercent,
			&g.ReviewScore,
		); err != nil {
			return nil, 0, err
		}
		games = append(games, g)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return games, total, nil
}

func (r *CatalogRepository) GetGameByID(ctx context.Context, id int64) (models.Game, bool, error) {
	query := `
		SELECT
			g.id,
			g.title,
			g.platform,
			g.cover_url,
			COALESCE(p.price_inr, 0) AS price_inr,
			COALESCE(p_low.lowest_price_inr, COALESCE(p.price_inr, 0), 0) AS lowest_price_inr,
			(COALESCE(p.price_inr, 0) > 0 AND COALESCE(p.price_inr, 0) = COALESCE(p_low.lowest_price_inr, COALESCE(p.price_inr, 0), 0)) AS is_all_time_low,
			COALESCE(d.original_inr, 0) AS original_inr,
			COALESCE(d.discount_percent, 0) AS discount_percent,
			0 AS review_score
		FROM games g
		LEFT JOIN LATERAL (
			SELECT price_inr
			FROM prices p
			WHERE p.game_id = g.id
			ORDER BY p.fetched_at DESC
			LIMIT 1
		) p ON TRUE
		LEFT JOIN LATERAL (
			SELECT MIN(price_inr) AS lowest_price_inr
			FROM prices p_low
			WHERE p_low.game_id = g.id
		) p_low ON TRUE
		LEFT JOIN LATERAL (
			SELECT original_inr, discount_percent
			FROM deals d
			WHERE d.game_id = g.id AND d.is_active = TRUE
			ORDER BY d.cached_at DESC
			LIMIT 1
		) d ON TRUE
		WHERE g.id = $1
	`

	var g models.Game
	err := r.db.QueryRow(ctx, query, id).Scan(
		&g.ID,
		&g.Title,
		&g.Platform,
		&g.CoverURL,
		&g.PriceINR,
		&g.LowestPriceINR,
		&g.IsAllTimeLow,
		&g.OriginalINR,
		&g.DiscountPercent,
		&g.ReviewScore,
	)
	if err != nil {
		return models.Game{}, false, nil
	}

	return g, true, nil
}

func (r *CatalogRepository) ListDeals(ctx context.Context, limit, offset int) ([]models.Deal, int, error) {
	countQuery := `SELECT COUNT(*) FROM deals d WHERE d.is_active = TRUE`
	var total int
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			g.id,
			g.title,
			g.platform,
			g.cover_url,
			COALESCE(p.price_inr, 0) AS price_inr,
			COALESCE(p_low.lowest_price_inr, COALESCE(p.price_inr, 0), 0) AS lowest_price_inr,
			(COALESCE(p.price_inr, 0) > 0 AND COALESCE(p.price_inr, 0) = COALESCE(p_low.lowest_price_inr, COALESCE(p.price_inr, 0), 0)) AS is_all_time_low,
			d.original_inr,
			d.discount_percent,
			0 AS review_score,
			d.cached_at::text
		FROM deals d
		JOIN games g ON g.id = d.game_id
		LEFT JOIN LATERAL (
			SELECT price_inr
			FROM prices p
			WHERE p.game_id = g.id
			ORDER BY p.fetched_at DESC
			LIMIT 1
		) p ON TRUE
		LEFT JOIN LATERAL (
			SELECT MIN(price_inr) AS lowest_price_inr
			FROM prices p_low
			WHERE p_low.game_id = g.id
		) p_low ON TRUE
		WHERE d.is_active = TRUE
		ORDER BY d.discount_percent DESC, g.title ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	deals := make([]models.Deal, 0, limit)
	for rows.Next() {
		var d models.Deal
		if err := rows.Scan(
			&d.ID,
			&d.Title,
			&d.Platform,
			&d.CoverURL,
			&d.PriceINR,
			&d.LowestPriceINR,
			&d.IsAllTimeLow,
			&d.OriginalINR,
			&d.DiscountPercent,
			&d.ReviewScore,
			&d.CachedAt,
		); err != nil {
			return nil, 0, err
		}
		deals = append(deals, d)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return deals, total, nil
}

func (r *CatalogRepository) GetPriceHistory(ctx context.Context, gameID int64, limit int) ([]models.PriceHistoryPoint, error) {
	query := `
		SELECT price_inr, fetched_at::text
		FROM prices
		WHERE game_id = $1
		ORDER BY fetched_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, query, gameID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]models.PriceHistoryPoint, 0, limit)
	for rows.Next() {
		var p models.PriceHistoryPoint
		if err := rows.Scan(&p.PriceINR, &p.FetchedAt); err != nil {
			return nil, err
		}
		history = append(history, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

// InsertPrice adds a new price entry for a game
func (r *CatalogRepository) InsertPrice(ctx context.Context, gameID int64, priceINR int, store string) error {
	query := `
		INSERT INTO prices (game_id, price_inr, store, fetched_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := r.db.Exec(ctx, query, gameID, priceINR, store)
	return err
}

// UpdateDeal updates or creates a deal entry for a game
func (r *CatalogRepository) UpdateDeal(ctx context.Context, gameID int64, originalINR, discountPercent int) error {
	query := `
		INSERT INTO deals (game_id, original_inr, discount_percent, is_active, cached_at)
		VALUES ($1, $2, $3, TRUE, NOW())
		ON CONFLICT (game_id) DO UPDATE SET
			original_inr = EXCLUDED.original_inr,
			discount_percent = EXCLUDED.discount_percent,
			is_active = TRUE,
			cached_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, gameID, originalINR, discountPercent)
	return err
}

// GetAllGameIDs returns all game IDs for price refresh
func (r *CatalogRepository) GetAllGameIDs(ctx context.Context) ([]int64, error) {
	query := `SELECT id FROM games`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// GetIndiaArbitrage calculates India vs Global pricing with GST
func (r *CatalogRepository) GetIndiaArbitrage(ctx context.Context, gameID int64) (models.IndiaArbitrage, error) {
	// Get current price from database
	query := `
		SELECT COALESCE(p.price_inr, 0) AS current_price
		FROM games g
		LEFT JOIN LATERAL (
			SELECT price_inr
			FROM prices p
			WHERE p.game_id = g.id
			ORDER BY p.fetched_at DESC
			LIMIT 1
		) p ON TRUE
		WHERE g.id = $1
	`

	var currentPrice int
	err := r.db.QueryRow(ctx, query, gameID).Scan(&currentPrice)
	if err != nil {
		return models.IndiaArbitrage{}, err
	}

	// For MVP: use current price as India price, simulate global price
	// In production: fetch actual Steam India and Global prices from API
	steamIndiaPrice := currentPrice
	steamGlobalPrice := currentPrice * 8 // Simulate global price in USD
	usdToINR := 83.0
	steamGlobalINR := int(float64(steamGlobalPrice) * usdToINR)
	gstRate := 0.18
	gstAmount := int(float64(steamGlobalINR) * gstRate)
	totalWithGST := steamGlobalINR + gstAmount

	// Determine cheapest region
	cheapestRegion := "India"
	if steamIndiaPrice > totalWithGST {
		cheapestRegion = "Global"
	}

	// Generate verdict
	verdict := ""
	if cheapestRegion == "India" {
		savings := totalWithGST - steamIndiaPrice
		verdict = fmt.Sprintf("Buy from India - saves ₹%d", savings)
	} else {
		savings := steamIndiaPrice - totalWithGST
		verdict = fmt.Sprintf("Buy from Global - saves ₹%d", savings)
	}

	return models.IndiaArbitrage{
		GameID:           gameID,
		SteamIndiaPrice:  steamIndiaPrice,
		SteamGlobalPrice: steamGlobalPrice,
		SteamGlobalINR:   steamGlobalINR,
		GSTAmount:        gstAmount,
		TotalWithGST:     totalWithGST,
		CheapestRegion:   cheapestRegion,
		Verdict:          verdict,
	}, nil
}
