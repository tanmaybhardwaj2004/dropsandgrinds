package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService handles caching of game prices and API responses using Redis
type CacheService struct {
	client *redis.Client
	logger *slog.Logger
}

// NewCacheService creates a new cache service
func NewCacheService(client *redis.Client, logger *slog.Logger) *CacheService {
	return &CacheService{
		client: client,
		logger: logger,
	}
}

// Cache TTL configurations
const (
	GamePricesTTL       = 5 * time.Minute  // Cache game prices for 5 minutes
	DealsCacheTTL       = 2 * time.Minute  // Cache deals for 2 minutes
	GameDetailsTTL      = 10 * time.Minute // Cache game details for 10 minutes
	SearchResultsTTL    = 3 * time.Minute  // Cache search results for 3 minutes
	IndianOffersTTL     = 30 * time.Minute // Cache Indian offers for 30 minutes
	StoreHealthTTL      = 1 * time.Minute  // Cache store health for 1 minute
)

// Cache keys
const (
	GamePricesKeyPrefix   = "game:prices:"
	GameDetailsKeyPrefix  = "game:details:"
	DealsKeyPrefix        = "deals:"
	SearchResultsKeyPrefix = "search:"
	IndianOffersKeyPrefix = "indian:offers:"
	StoreHealthKeyPrefix  = "health:store:"
)

// GetGamePrices retrieves cached game prices
func (c *CacheService) GetGamePrices(ctx context.Context, gameID int64) ([]byte, error) {
	key := fmt.Sprintf("%s%d", GamePricesKeyPrefix, gameID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cached game prices: %w", err)
	}
	return data, nil
}

// SetGamePrices caches game prices
func (c *CacheService) SetGamePrices(ctx context.Context, gameID int64, data []byte) error {
	key := fmt.Sprintf("%s%d", GamePricesKeyPrefix, gameID)
	if err := c.client.Set(ctx, key, data, GamePricesTTL).Err(); err != nil {
		return fmt.Errorf("failed to cache game prices: %w", err)
	}
	c.logger.Debug("cached game prices", "game_id", gameID, "ttl", GamePricesTTL)
	return nil
}

// GetGameDetails retrieves cached game details
func (c *CacheService) GetGameDetails(ctx context.Context, gameID int64) ([]byte, error) {
	key := fmt.Sprintf("%s%d", GameDetailsKeyPrefix, gameID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cached game details: %w", err)
	}
	return data, nil
}

// SetGameDetails caches game details
func (c *CacheService) SetGameDetails(ctx context.Context, gameID int64, data []byte) error {
	key := fmt.Sprintf("%s%d", GameDetailsKeyPrefix, gameID)
	if err := c.client.Set(ctx, key, data, GameDetailsTTL).Err(); err != nil {
		return fmt.Errorf("failed to cache game details: %w", err)
	}
	c.logger.Debug("cached game details", "game_id", gameID, "ttl", GameDetailsTTL)
	return nil
}

// GetDeals retrieves cached deals
func (c *CacheService) GetDeals(ctx context.Context, params string) ([]byte, error) {
	key := fmt.Sprintf("%s%s", DealsKeyPrefix, params)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cached deals: %w", err)
	}
	return data, nil
}

// SetDeals caches deals
func (c *CacheService) SetDeals(ctx context.Context, params string, data []byte) error {
	key := fmt.Sprintf("%s%s", DealsKeyPrefix, params)
	if err := c.client.Set(ctx, key, data, DealsCacheTTL).Err(); err != nil {
		return fmt.Errorf("failed to cache deals: %w", err)
	}
	c.logger.Debug("cached deals", "params", params, "ttl", DealsCacheTTL)
	return nil
}

// GetSearchResults retrieves cached search results
func (c *CacheService) GetSearchResults(ctx context.Context, query string) ([]byte, error) {
	key := fmt.Sprintf("%s%s", SearchResultsKeyPrefix, query)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cached search results: %w", err)
	}
	return data, nil
}

// SetSearchResults caches search results
func (c *CacheService) SetSearchResults(ctx context.Context, query string, data []byte) error {
	key := fmt.Sprintf("%s%s", SearchResultsKeyPrefix, query)
	if err := c.client.Set(ctx, key, data, SearchResultsTTL).Err(); err != nil {
		return fmt.Errorf("failed to cache search results: %w", err)
	}
	c.logger.Debug("cached search results", "query", query, "ttl", SearchResultsTTL)
	return nil
}

// GetIndianOffers retrieves cached Indian offers
func (c *CacheService) GetIndianOffers(ctx context.Context, filter string) ([]byte, error) {
	key := fmt.Sprintf("%s%s", IndianOffersKeyPrefix, filter)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cached Indian offers: %w", err)
	}
	return data, nil
}

// SetIndianOffers caches Indian offers
func (c *CacheService) SetIndianOffers(ctx context.Context, filter string, data []byte) error {
	key := fmt.Sprintf("%s%s", IndianOffersKeyPrefix, filter)
	if err := c.client.Set(ctx, key, data, IndianOffersTTL).Err(); err != nil {
		return fmt.Errorf("failed to cache Indian offers: %w", err)
	}
	c.logger.Debug("cached Indian offers", "filter", filter, "ttl", IndianOffersTTL)
	return nil
}

// GetStoreHealth retrieves cached store health status
func (c *CacheService) GetStoreHealth(ctx context.Context, store string) ([]byte, error) {
	key := fmt.Sprintf("%s%s", StoreHealthKeyPrefix, store)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cached store health: %w", err)
	}
	return data, nil
}

// SetStoreHealth caches store health status
func (c *CacheService) SetStoreHealth(ctx context.Context, store string, data []byte) error {
	key := fmt.Sprintf("%s%s", StoreHealthKeyPrefix, store)
	if err := c.client.Set(ctx, key, data, StoreHealthTTL).Err(); err != nil {
		return fmt.Errorf("failed to cache store health: %w", err)
	}
	c.logger.Debug("cached store health", "store", store, "ttl", StoreHealthTTL)
	return nil
}

// InvalidateGamePrices invalidates cached prices for a specific game
func (c *CacheService) InvalidateGamePrices(ctx context.Context, gameID int64) error {
	key := fmt.Sprintf("%s%d", GamePricesKeyPrefix, gameID)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate game prices cache: %w", err)
	}
	c.logger.Debug("invalidated game prices cache", "game_id", gameID)
	return nil
}

// InvalidateAllDeals invalidates all cached deals
func (c *CacheService) InvalidateAllDeals(ctx context.Context) error {
	pattern := DealsKeyPrefix + "*"
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan deals cache keys: %w", err)
	}
	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to invalidate deals cache: %w", err)
		}
		c.logger.Debug("invalidated all deals cache", "count", len(keys))
	}
	return nil
}

// GetOrSetJSON is a helper for caching JSON data
func (c *CacheService) GetOrSetJSON(ctx context.Context, key string, ttl time.Duration, fetchFunc func() (interface{}, error)) ([]byte, error) {
	// Try to get from cache
	data, err := c.client.Get(ctx, key).Bytes()
	if err == nil {
		return data, nil
	}
	if err != redis.Nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	// Cache miss - fetch data
	result, err := fetchFunc()
	if err != nil {
		return nil, err
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	// Set in cache
	if err := c.client.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		c.logger.Warn("failed to set cache", "key", key, "error", err)
		// Return data even if cache set fails
		return jsonData, nil
	}

	return jsonData, nil
}
