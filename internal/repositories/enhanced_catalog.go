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
	enhancedCacheTTL = 10 * time.Minute
	trendingCacheTTL = 5 * time.Minute
)

type EnhancedCatalogRepository struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewEnhancedCatalogRepository(db *pgxpool.Pool, redis *redis.Client) *EnhancedCatalogRepository {
	return &EnhancedCatalogRepository{db: db, redis: redis}
}

// GetGameWithPriceComparison returns comprehensive game data with price comparison across all stores
func (r *EnhancedCatalogRepository) GetGameWithPriceComparison(ctx context.Context, gameID int64, region string) (*models.PriceComparisonResponse, error) {
	cacheKey := fmt.Sprintf("game_comparison:%d:%s", gameID, region)
	if r.redis != nil {
		cached, err := r.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var result models.PriceComparisonResponse
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				monitoring.RecordCacheHit("game_comparison")
				return &result, nil
			}
		}
		monitoring.RecordCacheMiss("game_comparison")
	}

	// Get game details
	gameQuery := `
		SELECT g.id, g.external_id, g.title, g.slug, g.description, g.release_date, 
		       g.developer, g.publisher, g.genres, g.platforms, g.cover_url, 
		       g.screenshots, g.trailers, g.system_requirements, g.editions, 
		       g.dlc_info, g.rating, g.user_rating, g.is_dlc, g.parent_game_id, 
		       g.region, g.is_active, g.last_price_update
		FROM games g 
		WHERE g.id = $1 AND g.is_active = TRUE
	`
	
	var game models.EnhancedGame
	err := r.db.QueryRow(ctx, gameQuery, gameID).Scan(
		&game.ID, &game.ExternalID, &game.Title, &game.Slug, &game.Description,
		&game.ReleaseDate, &game.Developer, &game.Publisher, &game.Genres,
		&game.Platforms, &game.CoverURL, &game.Screenshots, &game.Trailers,
		&game.SystemRequirements, &game.Editions, &game.DLCInfo, &game.Rating,
		&game.UserRating, &game.IsDLC, &game.ParentGameID, &game.Region,
		&game.IsActive, &game.LastPriceUpdate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Get prices from all stores for this game
	pricesQuery := `
		SELECT p.id, p.game_id, p.store_id, p.external_id, p.price_inr, 
		       p.original_price, p.discount_amount, p.discount_percent, p.region, 
		       p.currency, p.is_available, p.stock_status, p.deal_type, p.updated_at,
		       s.id, s.name, s.slug, s.website_url, s.logo_url
		FROM prices p
		JOIN stores s ON p.store_id = s.id
		WHERE p.game_id = $1 AND p.region = $2 AND s.is_active = TRUE AND p.is_available = TRUE
		ORDER BY p.price_inr ASC
	`
	
	rows, err := r.db.Query(ctx, pricesQuery, gameID, region)
	if err != nil {
		return nil, fmt.Errorf("failed to get prices: %w", err)
	}
	defer rows.Close()

	var prices []models.EnhancedPrice
	var lowestPrice *models.EnhancedPrice
	
	for rows.Next() {
		var price models.EnhancedPrice
		var store models.Store
		err := rows.Scan(
			&price.ID, &price.GameID, &price.StoreID, &price.ExternalID,
			&price.PriceINR, &price.OriginalPrice, &price.DiscountAmount,
			&price.DiscountPercent, &price.Region, &price.Currency,
			&price.IsAvailable, &price.StockStatus, &price.DealType, &price.UpdatedAt,
			&store.ID, &store.Name, &store.Slug, &store.WebsiteURL, &store.LogoURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan price: %w", err)
		}
		
		price.Store = store
		prices = append(prices, price)
		
		if lowestPrice == nil || price.PriceINR < lowestPrice.PriceINR {
			lowestPrice = &price
		}
	}

	// Get price history
	historyQuery := `
		SELECT ph.id, ph.game_id, ph.store_id, ph.external_id, ph.price_inr,
		       ph.original_price, ph.discount_percent, ph.region, ph.currency,
		       ph.is_available, ph.recorded_at, s.id, s.name, s.slug
		FROM price_history ph
		JOIN stores s ON ph.store_id = s.id
		WHERE ph.game_id = $1 AND ph.region = $2
		ORDER BY ph.recorded_at DESC
		LIMIT 30
	`
	
	rows, err = r.db.Query(ctx, historyQuery, gameID, region)
	if err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}
	defer rows.Close()

	var priceHistory []models.PriceHistory
	for rows.Next() {
		var history models.PriceHistory
		var store models.Store
		err := rows.Scan(
			&history.ID, &history.GameID, &history.StoreID, &history.ExternalID,
			&history.PriceINR, &history.OriginalPrice, &history.DiscountPercent,
			&history.Region, &history.Currency, &history.IsAvailable, &history.RecordedAt,
			&store.ID, &store.Name, &store.Slug,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}
		
		history.Store = store
		priceHistory = append(priceHistory, history)
	}

	response := &models.PriceComparisonResponse{
		Game:          game,
		Prices:        prices,
		LowestPrice:   lowestPrice,
		PriceHistory:  priceHistory,
		LastUpdated:   time.Now(),
	}

	// Cache the response
	if r.redis != nil {
		jsonData, _ := json.Marshal(response)
		r.redis.Set(ctx, cacheKey, jsonData, enhancedCacheTTL)
	}

	return response, nil
}

// SearchEnhancedGames provides advanced search with multi-platform support
func (r *EnhancedCatalogRepository) SearchEnhancedGames(ctx context.Context, query string, platforms []string, genres []string, stores []string, minPrice, maxPrice float64, minDiscount, maxDiscount int, minRating, maxRating float64, sortBy string, limit, offset int) (*models.EnhancedGameListResponse, error) {
	cacheKey := fmt.Sprintf("enhanced_search:%s:%v:%v:%v:%f:%f:%d:%d:%f:%f:%s:%d:%d", 
		query, platforms, genres, stores, minPrice, maxPrice, minDiscount, maxDiscount, minRating, maxRating, sortBy, limit, offset)
	
	if r.redis != nil {
		cached, err := r.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var result models.EnhancedGameListResponse
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				monitoring.RecordCacheHit("enhanced_search")
				return &result, nil
			}
		}
		monitoring.RecordCacheMiss("enhanced_search")
	}

	// Build WHERE clause with advanced filters
	whereClause := "WHERE g.is_active = TRUE"
	args := []interface{}{1}
	argIndex := 1

	if query != "" {
		whereClause += fmt.Sprintf(" AND (g.title ILIKE $%d OR g.description ILIKE $%d)", argIndex+1, argIndex+2)
		args = append(args, "%"+query+"%", "%"+query+"%")
		argIndex += 2
	}

	if len(platforms) > 0 {
		whereClause += fmt.Sprintf(" AND g.platforms && $%d", argIndex)
		args = append(args, platforms)
		argIndex++
	}

	if len(genres) > 0 {
		whereClause += fmt.Sprintf(" AND g.genres && $%d", argIndex)
		args = append(args, genres)
		argIndex++
	}

	if len(stores) > 0 {
		whereClause += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM prices p JOIN stores s ON p.store_id = s.id WHERE p.game_id = g.id AND s.slug = ANY($%d))", argIndex)
		args = append(args, stores)
		argIndex++
	}

	if minPrice > 0 {
		whereClause += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM prices p WHERE p.game_id = g.id AND p.price_inr >= $%d)", argIndex)
		args = append(args, minPrice)
		argIndex++
	}

	if maxPrice > 0 {
		whereClause += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM prices p WHERE p.game_id = g.id AND p.price_inr <= $%d)", argIndex)
		args = append(args, maxPrice)
		argIndex++
	}

	if minDiscount > 0 {
		whereClause += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM prices p WHERE p.game_id = g.id AND p.discount_percent >= $%d)", argIndex)
		args = append(args, minDiscount)
		argIndex++
	}

	if minRating > 0 {
		whereClause += fmt.Sprintf(" AND g.user_rating >= $%d", argIndex)
		args = append(args, minRating)
		argIndex++
	}

	// Build ORDER BY clause based on sort parameter
	orderByClause := "ORDER BY g.last_price_update DESC"
	switch sortBy {
	case "price_low":
		orderByClause = "ORDER BY (SELECT MIN(p.price_inr) FROM prices p WHERE p.game_id = g.id) ASC"
	case "price_high":
		orderByClause = "ORDER BY (SELECT MIN(p.price_inr) FROM prices p WHERE p.game_id = g.id) DESC"
	case "discount":
		orderByClause = "ORDER BY (SELECT MAX(p.discount_percent) FROM prices p WHERE p.game_id = g.id) DESC"
	case "rating":
		orderByClause = "ORDER BY g.user_rating DESC"
	case "release_date":
		orderByClause = "ORDER BY g.release_date DESC"
	case "trending":
		orderByClause = "ORDER BY g.last_price_update DESC, g.user_rating DESC"
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM games g %s", whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count games: %w", err)
	}

	// Get games with pagination
	searchQuery := fmt.Sprintf(`
		SELECT g.id, g.external_id, g.title, g.slug, g.description, g.release_date,
		       g.developer, g.publisher, g.genres, g.platforms, g.cover_url,
		       g.screenshots, g.trailers, g.system_requirements, g.editions,
		       g.dlc_info, g.rating, g.user_rating, g.is_dlc, g.parent_game_id,
		       g.region, g.is_active, g.last_price_update,
		       (SELECT MIN(p.price_inr) FROM prices p WHERE p.game_id = g.id AND p.is_available = TRUE) as min_price
		FROM games g
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderByClause, argIndex+1, argIndex+2)

	args = append(args, limit, offset)
	
	rows, err := r.db.Query(ctx, searchQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search games: %w", err)
	}
	defer rows.Close()

	var games []models.EnhancedGame
	for rows.Next() {
		var game models.EnhancedGame
		var minPrice float64
		err := rows.Scan(
			&game.ID, &game.ExternalID, &game.Title, &game.Slug, &game.Description,
			&game.ReleaseDate, &game.Developer, &game.Publisher, &game.Genres,
			&game.Platforms, &game.CoverURL, &game.Screenshots, &game.Trailers,
			&game.SystemRequirements, &game.Editions, &game.DLCInfo, &game.Rating,
			&game.UserRating, &game.IsDLC, &game.ParentGameID, &game.Region,
			&game.IsActive, &game.LastPriceUpdate, &minPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		
		// Add lowest price to game model
		game.PriceINR = minPrice
		
		games = append(games, game)
	}

	response := &models.EnhancedGameListResponse{
		Games:  games,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	// Cache the response
	if r.redis != nil {
		jsonData, _ := json.Marshal(response)
		r.redis.Set(ctx, cacheKey, jsonData, enhancedCacheTTL)
	}

	return response, nil
}

// GetTrendingDeals returns trending deals based on view counts and conversion rates
func (r *EnhancedCatalogRepository) GetTrendingDeals(ctx context.Context, region string, limit, offset int) (*models.TrendingDealsResponse, error) {
	cacheKey := fmt.Sprintf("trending_deals:%s:%d:%d", region, limit, offset)
	if r.redis != nil {
		cached, err := r.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var result models.TrendingDealsResponse
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				monitoring.RecordCacheHit("trending_deals")
				return &result, nil
			}
		}
		monitoring.RecordCacheMiss("trending_deals")
	}

	query := `
		SELECT td.id, td.game_id, td.store_id, td.trend_score, td.view_count,
		       td.click_count, td.conversion_rate, td.trend_period, td.is_active,
		       td.created_at, td.updated_at,
		       g.id, g.title, g.slug, g.cover_url, g.description,
		       p.price_inr, p.discount_percent, p.original_price,
		       s.id, s.name, s.slug, s.logo_url
		FROM trending_deals td
		JOIN games g ON td.game_id = g.id
		JOIN stores s ON td.store_id = s.id
		JOIN prices p ON td.game_id = p.game_id AND td.store_id = p.store_id
		WHERE td.is_active = TRUE AND td.trend_period = '24h'
		ORDER BY td.trend_score DESC, td.updated_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending deals: %w", err)
	}
	defer rows.Close()

	var deals []models.TrendingDeal
	for rows.Next() {
		var deal models.TrendingDeal
		var game models.EnhancedGame
		var store models.Store
		var price float64
		var discount int
		var originalPrice float64
		
		err := rows.Scan(
			&deal.ID, &deal.GameID, &deal.StoreID, &deal.TrendScore, &deal.ViewCount,
			&deal.ClickCount, &deal.ConversionRate, &deal.TrendPeriod, &deal.IsActive,
			&deal.CreatedAt, &deal.UpdatedAt,
			&game.ID, &game.Title, &game.Slug, &game.CoverURL, &game.Description,
			&price, &discount, &originalPrice,
			&store.ID, &store.Name, &store.Slug, &store.LogoURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trending deal: %w", err)
		}
		
		game.PriceINR = price
		game.OriginalINR = int(originalPrice)
		game.DiscountPercent = discount
		deal.Store = store
		deal.Game = game
		
		deals = append(deals, deal)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM trending_deals WHERE is_active = TRUE AND trend_period = '24h'"
	var total int
	err = r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count trending deals: %w", err)
	}

	response := &models.TrendingDealsResponse{
		Deals: deals,
		Total: total,
		Limit: limit,
		Offset: offset,
	}

	// Cache the response
	if r.redis != nil {
		jsonData, _ := json.Marshal(response)
		r.redis.Set(ctx, cacheKey, jsonData, trendingCacheTTL)
	}

	return response, nil
}

// GetIndianPaymentOffers returns available Indian payment offers
func (r *EnhancedCatalogRepository) GetIndianPaymentOffers(ctx context.Context) (*models.IndianOffersResponse, error) {
	cacheKey := "indian_payment_offers"
	if r.redis != nil {
		cached, err := r.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var result models.IndianOffersResponse
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				monitoring.RecordCacheHit("indian_offers")
				return &result, nil
			}
		}
		monitoring.RecordCacheMiss("indian_offers")
	}

	query := `
		SELECT ipo.id, ipo.store_id, ipo.offer_type, ipo.provider, ipo.description,
		       ipo.discount_percent, ipo.max_discount_amount, ipo.min_order_amount,
		       ipo.valid_from, ipo.valid_until, ipo.is_active, ipo.created_at,
		       s.id, s.name, s.slug, s.logo_url
		FROM indian_payment_offers ipo
		JOIN stores s ON ipo.store_id = s.id
		WHERE ipo.is_active = TRUE 
		AND (ipo.valid_until IS NULL OR ipo.valid_until > CURRENT_TIMESTAMP)
		ORDER BY ipo.discount_percent DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get Indian payment offers: %w", err)
	}
	defer rows.Close()

	var offers []models.IndianPaymentOffer
	for rows.Next() {
		var offer models.IndianPaymentOffer
		var store models.Store
		
		err := rows.Scan(
			&offer.ID, &offer.StoreID, &offer.OfferType, &offer.Provider,
			&offer.Description, &offer.DiscountPercent, &offer.MaxDiscountAmount,
			&offer.MinOrderAmount, &offer.ValidFrom, &offer.ValidUntil,
			&offer.IsActive, &offer.CreatedAt,
			&store.ID, &store.Name, &store.Slug, &store.LogoURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan Indian offer: %w", err)
		}
		
		offer.Store = store
		offers = append(offers, offer)
	}

	response := &models.IndianOffersResponse{
		Offers: offers,
		Total:  len(offers),
	}

	// Cache the response
	if r.redis != nil {
		jsonData, _ := json.Marshal(response)
		r.redis.Set(ctx, cacheKey, jsonData, time.Hour)
	}

	return response, nil
}

// UpdateGamePrices updates prices for a game across all stores
func (r *EnhancedCatalogRepository) UpdateGamePrices(ctx context.Context, gameID int64, prices []models.EnhancedPrice) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update current prices
	for _, price := range prices {
		upsertQuery := `
			INSERT INTO prices (game_id, store_id, external_id, price_inr, original_price, 
				discount_amount, discount_percent, region, currency, is_available, 
				stock_status, deal_type, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (game_id, store_id) 
			DO UPDATE SET price_inr = EXCLUDED.price_inr, original_price = EXCLUDED.original_price,
				discount_amount = EXCLUDED.discount_amount, discount_percent = EXCLUDED.discount_percent,
				is_available = EXCLUDED.is_available, stock_status = EXCLUDED.stock_status,
				deal_type = EXCLUDED.deal_type, updated_at = EXCLUDED.updated_at
		`
		
		_, err := tx.Exec(ctx, upsertQuery,
			price.GameID, price.StoreID, price.ExternalID, price.PriceINR,
			price.OriginalPrice, price.DiscountAmount, price.DiscountPercent,
			price.Region, price.Currency, price.IsAvailable, price.StockStatus,
			price.DealType, time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to upsert price: %w", err)
		}

		// Add to price history
		historyQuery := `
			INSERT INTO price_history (game_id, store_id, external_id, price_inr, original_price,
				discount_percent, region, currency, is_available, recorded_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		
		_, err = tx.Exec(ctx, historyQuery,
			price.GameID, price.StoreID, price.ExternalID, price.PriceINR,
			price.OriginalPrice, price.DiscountPercent, price.Region, price.Currency,
			price.IsAvailable, time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert price history: %w", err)
		}
	}

	// Update game's last price update timestamp
	_, err = tx.Exec(ctx, "UPDATE games SET last_price_update = $1 WHERE id = $2", time.Now(), gameID)
	if err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Clear relevant caches
	if r.redis != nil {
		patterns := []string{
			fmt.Sprintf("game_comparison:%d:*", gameID),
			"enhanced_search:*",
			"trending_deals:*",
		}
		for _, pattern := range patterns {
			keys, _ := r.redis.Keys(ctx, pattern).Result()
			if len(keys) > 0 {
				r.redis.Del(ctx, keys...)
			}
		}
	}

	return nil
}
