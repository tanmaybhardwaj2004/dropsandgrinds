package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/monitoring"
)

const (
	dealCacheTTL = 5 * time.Minute
)

type CatalogRepository struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewCatalogRepository(db *pgxpool.Pool, redis *redis.Client) *CatalogRepository {
	return &CatalogRepository{db: db, redis: redis}
}

func (r *CatalogRepository) SearchGames(ctx context.Context, query string, platform string, minPrice, maxPrice float64, minDiscount, maxDiscount int, minReviewScore, maxReviewScore float64, limit, offset int) ([]models.Game, int, error) {
	// Try cache first if Redis is available
	cacheKey := fmt.Sprintf("search:%s:%s:%f:%f:%d:%d:%f:%f:%d:%d", query, platform, minPrice, maxPrice, minDiscount, maxDiscount, minReviewScore, maxReviewScore, limit, offset)
	if r.redis != nil {
		cached, err := r.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedResult struct {
				Games []models.Game `json:"games"`
				Total int           `json:"total"`
			}
			if err := json.Unmarshal([]byte(cached), &cachedResult); err == nil {
				monitoring.RecordCacheHit("catalog_search")
				return cachedResult.Games, cachedResult.Total, nil
			}
		}
		monitoring.RecordCacheMiss("catalog_search")
	}

	// Build WHERE clause with filters
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if query != "" {
		// Use ILIKE for case-insensitive search (will use GIN index with pg_trgm)
		whereClause += fmt.Sprintf(" AND LOWER(g.title) LIKE LOWER($%d)", argIndex)
		args = append(args, "%"+query+"%")
		argIndex++
	}

	if platform != "" {
		whereClause += fmt.Sprintf(" AND LOWER(g.platform) = LOWER($%d)", argIndex)
		args = append(args, platform)
		argIndex++
	}

	if minPrice > 0 {
		whereClause += fmt.Sprintf(" AND COALESCE(p.price_inr, 0) >= $%d", argIndex)
		args = append(args, minPrice)
		argIndex++
	}

	if maxPrice > 0 {
		whereClause += fmt.Sprintf(" AND COALESCE(p.price_inr, 0) <= $%d", argIndex)
		args = append(args, maxPrice)
		argIndex++
	}

	if minDiscount > 0 {
		whereClause += fmt.Sprintf(" AND COALESCE(d.discount_percent, 0) >= $%d", argIndex)
		args = append(args, minDiscount)
		argIndex++
	}

	if maxDiscount > 0 {
		whereClause += fmt.Sprintf(" AND COALESCE(d.discount_percent, 0) <= $%d", argIndex)
		args = append(args, maxDiscount)
		argIndex++
	}

	if minReviewScore > 0 {
		whereClause += fmt.Sprintf(" AND COALESCE(r.avg_score, 0) >= $%d", argIndex)
		args = append(args, minReviewScore)
		argIndex++
	}

	if maxReviewScore > 0 {
		whereClause += fmt.Sprintf(" AND COALESCE(r.avg_score, 0) <= $%d", argIndex)
		args = append(args, maxReviewScore)
		argIndex++
	}

	// Count query
	countQuery := `
		SELECT COUNT(*)
		FROM games g
		LEFT JOIN LATERAL (
			SELECT price_inr
			FROM prices p
			WHERE p.game_id = g.id
			ORDER BY p.fetched_at DESC
			LIMIT 1
		) p ON TRUE
		LEFT JOIN LATERAL (
			SELECT original_inr, discount_percent
			FROM deals d
			WHERE d.game_id = g.id AND d.is_active = TRUE
			ORDER BY d.discount_percent DESC
			LIMIT 1
		) d ON TRUE
		LEFT JOIN LATERAL (
			SELECT AVG(score) AS avg_score
			FROM review_scores r
			WHERE r.game_id = g.id
		) r ON TRUE
		` + whereClause

	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Data query
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
			COALESCE(r.avg_score, 0) AS review_score
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
			ORDER BY d.discount_percent DESC
			LIMIT 1
		) d ON TRUE
		LEFT JOIN LATERAL (
			SELECT AVG(score) AS avg_score
			FROM review_scores r
			WHERE r.game_id = g.id
		) r ON TRUE
		` + whereClause + `
		ORDER BY COALESCE(d.discount_percent, 0) DESC, g.title ASC
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
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

	// Cache the result if Redis is available
	if r.redis != nil {
		cacheResult := struct {
			Games []models.Game `json:"games"`
			Total int           `json:"total"`
		}{
			Games: games,
			Total: total,
		}
		if data, err := json.Marshal(cacheResult); err == nil {
			r.redis.Set(ctx, cacheKey, data, 1*time.Minute) // 1 minute TTL for search
		}
	}

	return games, total, nil
}

func (r *CatalogRepository) ListGames(ctx context.Context, query, platform string, limit, offset int, excludeOwned bool, userID int64) ([]models.Game, int, error) {
	// Try cache first if Redis is available (skip cache for personalized queries)
	cacheKey := ""
	if r.redis != nil && !excludeOwned {
		cacheKey = fmt.Sprintf("games:%s:%s:%d:%d", query, platform, limit, offset)
		cached, err := r.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedResult struct {
				Games []models.Game `json:"games"`
				Total int           `json:"total"`
			}
			if err := json.Unmarshal([]byte(cached), &cachedResult); err == nil {
				monitoring.RecordCacheHit("games_list")
				return cachedResult.Games, cachedResult.Total, nil
			}
		}
		monitoring.RecordCacheMiss("games_list")
	}

	countQuery := `
		SELECT COUNT(*)
		FROM games g
		WHERE ($1 = '' OR LOWER(g.title) LIKE '%' || LOWER($1) || '%')
		  AND ($2 = '' OR LOWER(g.platform) = LOWER($2))
	`

	// Add exclude_owned filter
	countArgs := []interface{}{query, platform}
	if excludeOwned && userID > 0 {
		countQuery += ` AND g.id NOT IN (SELECT game_id FROM user_steam_library WHERE user_id = $3 AND game_id IS NOT NULL)`
		countArgs = append(countArgs, userID)
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
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
	`

	// Add exclude_owned filter to data query
	dataArgs := []interface{}{query, platform}
	if excludeOwned && userID > 0 {
		dataQuery += ` AND g.id NOT IN (SELECT game_id FROM user_steam_library WHERE user_id = $3 AND game_id IS NOT NULL)`
		dataArgs = append(dataArgs, userID)
	}

	dataQuery += ` ORDER BY COALESCE(d.discount_percent, 0) DESC, g.title ASC LIMIT $` + fmt.Sprintf("%d", len(dataArgs)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(dataArgs)+2)
	dataArgs = append(dataArgs, limit, offset)

	rows, err := r.db.Query(ctx, dataQuery, dataArgs...)
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

	// Cache the result if Redis is available
	if r.redis != nil {
		cacheKey := fmt.Sprintf("games:%s:%s:%d:%d", query, platform, limit, offset)
		cacheResult := struct {
			Games []models.Game `json:"games"`
			Total int           `json:"total"`
		}{
			Games: games,
			Total: total,
		}
		if data, err := json.Marshal(cacheResult); err == nil {
			r.redis.Set(ctx, cacheKey, data, dealCacheTTL)
		}
	}

	return games, total, nil
}

func (r *CatalogRepository) GetGameByID(ctx context.Context, id int64) (models.Game, bool, error) {
	// Try cache first if Redis is available
	if r.redis != nil {
		cacheKey := fmt.Sprintf("game:%d", id)
		cached, err := r.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var game models.Game
			if err := json.Unmarshal([]byte(cached), &game); err == nil {
				monitoring.RecordCacheHit("game_detail")
				return game, true, nil
			}
		}
		monitoring.RecordCacheMiss("game_detail")
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

	// Cache the result if Redis is available
	if r.redis != nil {
		cacheKey := fmt.Sprintf("game:%d", id)
		if data, err := json.Marshal(g); err == nil {
			r.redis.Set(ctx, cacheKey, data, dealCacheTTL)
		}
	}

	return g, true, nil
}

func (r *CatalogRepository) ListDeals(ctx context.Context, limit, offset int) ([]models.Deal, int, error) {
	// Try cache first if Redis is available
	if r.redis != nil {
		cacheKey := fmt.Sprintf("deals:%d:%d", limit, offset)
		cached, err := r.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedResult struct {
				Deals []models.Deal `json:"deals"`
				Total int           `json:"total"`
			}
			if err := json.Unmarshal([]byte(cached), &cachedResult); err == nil {
				monitoring.RecordCacheHit("deals_list")
				return cachedResult.Deals, cachedResult.Total, nil
			}
		}
		monitoring.RecordCacheMiss("deals_list")
	}

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

	// Cache the result if Redis is available
	if r.redis != nil {
		cacheKey := fmt.Sprintf("deals:%d:%d", limit, offset)
		cacheResult := struct {
			Deals []models.Deal `json:"deals"`
			Total int           `json:"total"`
		}{
			Deals: deals,
			Total: total,
		}
		if data, err := json.Marshal(cacheResult); err == nil {
			r.redis.Set(ctx, cacheKey, data, dealCacheTTL)
		}
	}

	return deals, total, nil
}

func (r *CatalogRepository) GetPriceHistory(ctx context.Context, gameID int64, limit, offset int) ([]models.PriceHistoryPoint, error) {
	query := `
		SELECT price_inr, fetched_at::text
		FROM prices
		WHERE game_id = $1
		ORDER BY fetched_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, gameID, limit, offset)
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
	rows, err := r.db.Query(ctx, "SELECT id FROM games")
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
	return ids, nil
}

// FindGameByTitle searches for a game by exact title match
func (r *CatalogRepository) FindGameByTitle(ctx context.Context, title string) (int64, error) {
	var gameID int64
	err := r.db.QueryRow(ctx, "SELECT id FROM games WHERE title = $1 LIMIT 1", title).Scan(&gameID)
	if err != nil {
		return 0, err
	}
	return gameID, nil
}

// GetIndiaArbitrage calculates India vs Global pricing with GST
func (r *CatalogRepository) GetIndiaArbitrage(ctx context.Context, gameID int64) (models.ArbitrageData, error) {
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
		return models.ArbitrageData{}, err
	}

	// For MVP: use current price as India price, simulate global price
	// In production: fetch actual Steam India and Global prices from API
	steamIndiaPrice := float64(currentPrice)
	steamGlobalPrice := float64(currentPrice) * 8 // Simulate global price in USD
	usdToINR := 83.0
	steamGlobalINR := steamGlobalPrice * usdToINR
	gstRate := 0.18
	gstAmount := steamGlobalINR * gstRate
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
		verdict = fmt.Sprintf("Buy from India - saves ₹%.0f", savings)
	} else {
		savings := steamIndiaPrice - totalWithGST
		verdict = fmt.Sprintf("Buy from Global - saves ₹%.0f", savings)
	}

	return models.ArbitrageData{
		IndiaBaseINR:   steamIndiaPrice,
		IndiaGSTINR:    steamIndiaPrice * gstRate,
		IndiaTotalINR:  steamIndiaPrice + (steamIndiaPrice * gstRate),
		GlobalBaseINR:  steamGlobalINR,
		GlobalGSTINR:   gstAmount,
		GlobalTotalINR: totalWithGST,
		CheapestRegion: cheapestRegion,
		Verdict:        verdict,
	}, nil
}
