package services

import (
	"context"
	"fmt"
	"time"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

type PriceAggregatorService struct {
	steamAPI      *SteamAPIService
	epicAPI       *EpicGamesAPIService
	catalogRepo   *repositories.EnhancedCatalogRepository
}

func NewPriceAggregatorService(
	steamAPI *SteamAPIService,
	epicAPI *EpicGamesAPIService,
	catalogRepo *repositories.EnhancedCatalogRepository,
) *PriceAggregatorService {
	return &PriceAggregatorService{
		steamAPI:    steamAPI,
		epicAPI:     epicAPI,
		catalogRepo:  catalogRepo,
	}
}

// GetComprehensiveGamePrice fetches prices from all supported stores and returns comparison
func (p *PriceAggregatorService) GetComprehensiveGamePrice(ctx context.Context, externalID string, title string, region string) (*models.PriceComparisonResponse, error) {
	// Try to get game by external ID first
	game, err := p.catalogRepo.GetGameWithPriceComparison(ctx, 0, region)
	if err == nil && game != nil {
		return game, nil
	}

	// If not found, search by title across platforms
	if game == nil {
		games, _, err := p.catalogRepo.SearchEnhancedGames(ctx, title, []string{}, []string{}, []string{}, 0, 0, 0, 0, "", 20, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to search games: %w", err)
		}
		
		if len(games.Games) > 0 {
			// Use first matching game
			game = &games.Games[0]
			game.ID = 0 // Temporary ID for new games
		}
	}

	if game == nil {
		return nil, fmt.Errorf("game not found")
	}

	// Fetch fresh prices from all platforms
	prices, err := p.fetchPricesFromAllPlatforms(ctx, game, region)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}

	// Update game with new price data
	err = p.catalogRepo.UpdateGamePrices(ctx, game.ID, prices)
	if err != nil {
		return nil, fmt.Errorf("failed to update prices: %w", err)
	}

	// Return updated game with price comparison
	return p.catalogRepo.GetGameWithPriceComparison(ctx, game.ID, region)
}

// fetchPricesFromAllPlatforms fetches current prices from all supported platforms
func (p *PriceAggregatorService) fetchPricesFromAllPlatforms(ctx context.Context, game *models.EnhancedGame, region string) ([]models.EnhancedPrice, error) {
	var allPrices []models.EnhancedPrice

	// Fetch from Steam
	if steamPrice, err := p.fetchSteamPrice(ctx, game); err == nil {
		allPrices = append(allPrices, *steamPrice)
	}

	// Fetch from Epic Games
	if epicPrice, err := p.fetchEpicPrice(ctx, game); err == nil {
		allPrices = append(allPrices, *epicPrice)
	}

	// TODO: Add other platforms (Xbox, PlayStation, Nintendo, GreenManGaming, etc.)
	// For now, we'll use existing catalog data for other stores

	// Get prices from existing catalog for other platforms
	existingPrices, err := p.getExistingCatalogPrices(ctx, game.ID, region)
	if err == nil {
		allPrices = append(allPrices, existingPrices...)
	}

	// Find lowest price and calculate discounts
	lowestPrice := findLowestPrice(allPrices)
	for i := range allPrices {
		if allPrices[i].PriceINR == lowestPrice.PriceINR {
			allPrices[i].DiscountPercent = int(((allPrices[i].OriginalPrice - lowestPrice.PriceINR) / allPrices[i].OriginalPrice) * 100)
		}
	}

	return allPrices, nil
}

// fetchSteamPrice fetches price from Steam API
func (p *PriceAggregatorService) fetchSteamPrice(ctx context.Context, game *models.EnhancedGame) (*models.EnhancedPrice, error) {
	steamAppID := game.ExternalID
	if steamAppID == "" {
		// Try to find Steam game by title
		steamGames, err := p.steamAPI.SearchGames(ctx, game.Title, 5)
		if err != nil {
			return nil, fmt.Errorf("failed to search Steam: %w", err)
		}
		
		for _, steamGame := range steamGames {
			if steamGame.Title == game.Title {
				steamAppID = steamGame.ExternalID
				break
			}
		}
	}

	if steamAppID == "" {
		return nil, fmt.Errorf("Steam game not found")
	}

	steamPrice, err := p.steamAPI.GetGamePrice(ctx, parseInt(steamAppID))
	if err != nil {
		return nil, fmt.Errorf("failed to get Steam price: %w", err)
	}

	return &models.EnhancedPrice{
		GameID:         game.ID,
		StoreID:        1, // Steam store ID
		Store:          models.Store{ID: 1, Name: "Steam", Slug: "steam"},
		ExternalID:     steamAppID,
		PriceINR:       steamPrice.PriceINR,
		OriginalPrice:  steamPrice.OriginalPrice,
		DiscountPercent: steamPrice.DiscountPercent,
		Region:         "IN",
		Currency:       "INR",
		IsAvailable:    steamPrice.IsAvailable,
		StockStatus:    steamPrice.StockStatus,
		DealType:       steamPrice.DealType,
		UpdatedAt:       time.Now(),
	}, nil
}

// fetchEpicPrice fetches price from Epic Games API
func (p *PriceAggregatorService) fetchEpicPrice(ctx context.Context, game *models.EnhancedGame) (*models.EnhancedPrice, error) {
	epicSlug := game.ExternalID
	if epicSlug == "" {
		// Try to find Epic game by title
		epicGames, err := p.epicAPI.SearchGames(ctx, game.Title, 5)
		if err != nil {
			return nil, fmt.Errorf("failed to search Epic Games: %w", err)
		}
		
		for _, epicGame := range epicGames {
			if epicGame.Title == game.Title {
				epicSlug = epicGame.ExternalID
				break
			}
		}
	}

	if epicSlug == "" {
		return nil, fmt.Errorf("Epic Games game not found")
	}

	epicGame, err := p.epicAPI.GetGameDetails(ctx, epicSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get Epic Games price: %w", err)
	}

	return &models.EnhancedPrice{
		GameID:         game.ID,
		StoreID:        2, // Epic Games store ID
		Store:          models.Store{ID: 2, Name: "Epic Games", Slug: "epic"},
		ExternalID:     epicSlug,
		PriceINR:       epicGame.PriceINR,
		OriginalPrice:  float64(epicGame.OriginalINR),
		DiscountPercent: epicGame.DiscountPercent,
		Region:         "IN",
		Currency:       "INR",
		IsAvailable:    true, // Epic Games doesn't provide stock status
		StockStatus:    "available",
		DealType:       "regular",
		UpdatedAt:       time.Now(),
	}, nil
}

// getExistingCatalogPrices gets prices from existing catalog database
func (p *PriceAggregatorService) getExistingCatalogPrices(ctx context.Context, gameID int64, region string) ([]models.EnhancedPrice, error) {
	// This would query the existing catalog repository for other stores
	// For now, return empty slice to be implemented later
	return []models.EnhancedPrice{}, nil
}

// findLowestPrice finds the lowest price among all available prices
func findLowestPrice(prices []models.EnhancedPrice) *models.EnhancedPrice {
	if len(prices) == 0 {
		return nil
	}

	lowest := &prices[0]
	for _, price := range prices {
		if price.IsAvailable && price.PriceINR < lowest.PriceINR {
			lowest = &price
		}
	}
	return lowest
}

// GetTrendingDealsAggregated returns trending deals from all platforms
func (p *PriceAggregatorService) GetTrendingDealsAggregated(ctx context.Context, region string, limit int) (*models.TrendingDealsResponse, error) {
	// Get trending from catalog (aggregated from all platforms)
	return p.catalogRepo.GetTrendingDeals(ctx, region, limit, 0)
}

// SyncAllPrices updates prices for all games across all platforms
func (p *PriceAggregatorService) SyncAllPrices(ctx context.Context) error {
	// Get all games from catalog
	games, _, err := p.catalogRepo.SearchEnhancedGames(ctx, "", []string{}, []string{}, []string{}, 0, 0, 0, 0, "", 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to get games: %w", err)
	}

	// Update prices for each game
	for _, game := range games.Games {
		prices, err := p.fetchPricesFromAllPlatforms(ctx, &game, "IN")
		if err != nil {
			fmt.Printf("Failed to sync prices for game %s: %v", game.Title, err)
			continue
		}

		err = p.catalogRepo.UpdateGamePrices(ctx, game.ID, prices)
		if err != nil {
			fmt.Printf("Failed to update prices for game %s: %v", game.Title, err)
		}
	}

	return nil
}

// GetPriceHistoryAggregated returns price history from all platforms
func (p *PriceAggregatorService) GetPriceHistoryAggregated(ctx context.Context, gameID int64, region string, days int) ([]models.PriceHistory, error) {
	// Get price comparison which includes history
	comparison, err := p.catalogRepo.GetGameWithPriceComparison(ctx, gameID, region)
	if err != nil {
		return nil, fmt.Errorf("failed to get game comparison: %w", err)
	}

	// Return price history from comparison
	return comparison.PriceHistory, nil
}

// CalculateBestValue calculates best value deals based on price per hour of gameplay
func (p *PriceAggregatorService) CalculateBestValue(ctx context.Context, region string, limit int) (*models.TrendingDealsResponse, error) {
	// Get all games with prices and calculate value scores
	games, _, err := p.catalogRepo.SearchEnhancedGames(ctx, "", []string{}, []string{}, []string{}, 0, 0, 0, 0, "", 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get games: %w", err)
	}

	var trendingDeals []models.TrendingDeal
	
	for _, game := range games.Games {
		if game.PriceINR > 0 {
			// Simple value calculation: lower price = better value
			// In a real implementation, this would consider review scores, playtime, etc.
			valueScore := calculateValueScore(game)
			
			deal := models.TrendingDeal{
				GameID:        game.ID,
				StoreID:       1, // Default to Steam
				Store:         models.Store{ID: 1, Name: "Steam", Slug: "steam"},
				TrendScore:    valueScore,
				ViewCount:     0, // Would be tracked separately
				ClickCount:    0, // Would be tracked separately
				ConversionRate: 0, // Would be calculated from analytics
				TrendPeriod:   "7d",
				IsActive:      true,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			
			// Add game to deal
			deal.Game = *game
			deal.Store.LogoURL = "" // Would be set from store data
			
			trendingDeals = append(trendingDeals, deal)
		}
	}

	response := &models.TrendingDealsResponse{
		Deals: trendingDeals,
		Total: len(trendingDeals),
		Limit: limit,
		Offset: 0,
	}

	return response, nil
}

// calculateValueScore calculates a value score for a game
func calculateValueScore(game *models.EnhancedGame) float64 {
	score := 50.0 // Base score
	
	// Adjust for price (lower is better)
	if game.PriceINR > 0 {
		if game.PriceINR < 500 {
			score += 20 // Very affordable
		} else if game.PriceINR < 1000 {
			score += 10 // Affordable
		} else if game.PriceINR < 2000 {
			score += 5 // Moderate price
		}
	}
	
	// Adjust for discount
	if game.DiscountPercent > 0 {
		score += float64(game.DiscountPercent) / 2 // Higher discount = better value
	}
	
	// Adjust for rating
	if game.UserRating > 0 {
		score += game.UserRating * 2 // Higher rating = better value
	}
	
	// Adjust for age (newer games might have more value)
	if game.ReleaseDate != nil {
		yearsSinceRelease := time.Since(*game.ReleaseDate).Hours() / 24 / 365
		if yearsSinceRelease < 1 {
			score += 15 // New release bonus
		} else if yearsSinceRelease < 2 {
			score += 10 // Recent release bonus
		} else if yearsSinceRelease < 5 {
			score += 5 // Moderately new
		}
	}
	
	return score
}
