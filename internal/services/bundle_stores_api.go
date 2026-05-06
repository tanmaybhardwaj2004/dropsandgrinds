package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

type BundleStoresAPIService struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewBundleStoresAPIService(apiKey string) *BundleStoresAPIService {
	return &BundleStoresAPIService{
		apiKey:  apiKey,
		baseURL: "https://www.greenmangaming.com/api",
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// GreenManGaming API structures
type GreenManGamingResponse struct {
	Success bool                   `json:"success"`
	Data    []GreenManGamingGame `json:"data"`
	Message string                   `json:"message"`
}

type GreenManGamingGame struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	OriginalPrice float64 `json:"original_price"`
	Discount     int     `json:"discount"`
	Currency     string  `json:"currency"`
	Store        string  `json:"store"`
	Image        string  `json:"image"`
	URL          string  `json:"url"`
	ReleaseDate  string  `json:"release_date"`
	Platforms    []string `json:"platforms"`
	Genres       []string `json:"genres"`
	DRM          string  `json:"drm"`
}

// Fanatical API structures
type FanaticalResponse struct {
	Success bool                   `json:"success"`
	Data    []FanaticalGame      `json:"data"`
	Message string                   `json:"message"`
}

type FanaticalGame struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	OriginalPrice float64 `json:"original_price"`
	Discount     int     `json:"discount"`
	Currency     string  `json:"currency"`
	Store        string  `json:"store"`
	Image        string  `json:"image"`
	URL          string  `json:"url"`
	ReleaseDate  string  `json:"release_date"`
	Platforms    []string `json:"platforms"`
	Genres       []string `json:"genres"`
	DRM          string  `json:"drm"`
}

// Humble Bundle API structures
type HumbleBundleResponse struct {
	Success bool                     `json:"success"`
	Data    []HumbleBundleGame      `json:"data"`
	Message string                     `json:"message"`
}

type HumbleBundleGame struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	OriginalPrice float64 `json:"original_price"`
	Discount     int     `json:"discount"`
	Currency     string  `json:"currency"`
	Store        string  `json:"store"`
	Image        string  `json:"image"`
	URL          string  `json:"url"`
	ReleaseDate  string  `json:"release_date"`
	Platforms    []string `json:"platforms"`
	Genres       []string `json:"genres"`
	BundleType   string  `json:"bundle_type"`
}

// GetGreenManGamingGames fetches games from GreenManGaming API
func (b *BundleStoresAPIService) GetGreenManGamingGames(ctx context.Context, query string, limit int) ([]models.EnhancedGame, error) {
	url := fmt.Sprintf("%s/search?query=%s&limit=%d", b.baseURL, query, limit)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	if b.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+b.apiKey)
	}
	
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GreenManGaming API returned status %d", resp.StatusCode)
	}
	
	var response GreenManGamingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Message)
	}
	
	var games []models.EnhancedGame
	for _, game := range response.Data {
		releaseDate, _ := time.Parse("2006-01-02", game.ReleaseDate)
		
		enhancedGame := models.EnhancedGame{
			ID:              generateID(game.ID),
			ExternalID:      game.ID,
			Title:           game.Name,
			Description:     "", // GreenManGaming doesn't provide detailed descriptions in search
			Developer:       "", // Would need separate API call
			Publisher:       "",
			Genres:          game.Genres,
			Platforms:       game.Platforms,
			CoverURL:        game.Image,
			Screenshots:     []string{}, // Would need separate API call
			Trailers:        []string{}, // Would need separate API call
			ReleaseDate:     &releaseDate,
			PriceINR:        convertToINR(game.Price, game.Currency),
			OriginalINR:     int(convertToINR(game.OriginalPrice, game.Currency)),
			DiscountPercent:  game.Discount,
			IsActive:        true,
			LastPriceUpdate:  timePtr(time.Now()),
		}
		
		games = append(games, enhancedGame)
	}
	
	return games, nil
}

// GetFanaticalGames fetches games from Fanatical API
func (b *BundleStoresAPIService) GetFanaticalGames(ctx context.Context, query string, limit int) ([]models.EnhancedGame, error) {
	url := fmt.Sprintf("%s/api/v2/search?query=%s&limit=%d", b.baseURL, query, limit)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	if b.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+b.apiKey)
	}
	
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Fanatical API returned status %d", resp.StatusCode)
	}
	
	var response FanaticalResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Message)
	}
	
	var games []models.EnhancedGame
	for _, game := range response.Data {
		releaseDate, _ := time.Parse("2006-01-02", game.ReleaseDate)
		
		enhancedGame := models.EnhancedGame{
			ID:              generateID(game.ID),
			ExternalID:      game.ID,
			Title:           game.Name,
			Description:     "", // Fanatical doesn't provide detailed descriptions in search
			Developer:       "", // Would need separate API call
			Publisher:       "",
			Genres:          game.Genres,
			Platforms:       game.Platforms,
			CoverURL:        game.Image,
			Screenshots:     []string{}, // Would need separate API call
			Trailers:        []string{}, // Would need separate API call
			ReleaseDate:     &releaseDate,
			PriceINR:        convertToINR(game.Price, game.Currency),
			OriginalINR:     int(convertToINR(game.OriginalPrice, game.Currency)),
			DiscountPercent:  game.Discount,
			IsActive:        true,
			LastPriceUpdate:  timePtr(time.Now()),
		}
		
		games = append(games, enhancedGame)
	}
	
	return games, nil
}

// GetHumbleBundleGames fetches games from Humble Bundle API
func (b *BundleStoresAPIService) GetHumbleBundleGames(ctx context.Context, query string, limit int) ([]models.EnhancedGame, error) {
	url := fmt.Sprintf("%s/api/v1/search?search=%s&limit=%d", b.baseURL, query, limit)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	if b.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+b.apiKey)
	}
	
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Humble Bundle API returned status %d", resp.StatusCode)
	}
	
	var response HumbleBundleResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Message)
	}
	
	var games []models.EnhancedGame
	for _, game := range response.Data {
		releaseDate, _ := time.Parse("2006-01-02", game.ReleaseDate)
		
		enhancedGame := models.EnhancedGame{
			ID:              generateID(game.ID),
			ExternalID:      game.ID,
			Title:           game.Name,
			Description:     "", // Humble Bundle doesn't provide detailed descriptions in search
			Developer:       "", // Would need separate API call
			Publisher:       "",
			Genres:          game.Genres,
			Platforms:       game.Platforms,
			CoverURL:        game.Image,
			Screenshots:     []string{}, // Would need separate API call
			Trailers:        []string{}, // Would need separate API call
			ReleaseDate:     &releaseDate,
			PriceINR:        convertToINR(game.Price, game.Currency),
			OriginalINR:     int(convertToINR(game.OriginalPrice, game.Currency)),
			DiscountPercent:  game.Discount,
			IsActive:        true,
			LastPriceUpdate:  timePtr(time.Now()),
		}
		
		games = append(games, enhancedGame)
	}
	
	return games, nil
}

// GetGameDetails fetches detailed game information from bundle stores
func (b *BundleStoresAPIService) GetGameDetails(ctx context.Context, store string, gameID string) (*models.EnhancedGame, error) {
	var url string
	switch store {
	case "greenmangaming":
		url = fmt.Sprintf("%s/api/v1/games/%s", b.baseURL, gameID)
	case "fanatical":
		url = fmt.Sprintf("%s/api/v2/games/%s", b.baseURL, gameID)
	case "humble":
		url = fmt.Sprintf("%s/api/v1/games/%s", b.baseURL, gameID)
	default:
		return nil, fmt.Errorf("unsupported store: %s", store)
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	if b.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+b.apiKey)
	}
	
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s API returned status %d", strings.Title(store), resp.StatusCode)
	}
	
	// Parse response based on store
	switch store {
	case "greenmangaming":
		var response GreenManGamingResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode GreenManGaming response: %w", err)
		}
		
		if !response.Success || len(response.Data) == 0 {
			return nil, fmt.Errorf("game not found")
		}
		
		game := response.Data[0]
		releaseDate, _ := time.Parse("2006-01-02", game.ReleaseDate)
		
		return &models.EnhancedGame{
			ID:              generateID(game.ID),
			ExternalID:      game.ID,
			Title:           game.Name,
			Description:     "", // Would need separate API call
			Developer:       "", // Would need separate API call
			Publisher:       "", // Would need separate API call
			Genres:          game.Genres,
			Platforms:       game.Platforms,
			CoverURL:        game.Image,
			Screenshots:     []string{}, // Would need separate API call
			Trailers:        []string{}, // Would need separate API call
			ReleaseDate:     &releaseDate,
			PriceINR:        convertToINR(game.Price, game.Currency),
			OriginalINR:     int(convertToINR(game.OriginalPrice, game.Currency)),
			DiscountPercent:  game.Discount,
			IsActive:        true,
			LastPriceUpdate:  timePtr(time.Now()),
		}, nil
		
	case "fanatical":
		var response FanaticalResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode Fanatical response: %w", err)
		}
		
		if !response.Success || len(response.Data) == 0 {
			return nil, fmt.Errorf("game not found")
		}
		
		game := response.Data[0]
		releaseDate, _ := time.Parse("2006-01-02", game.ReleaseDate)
		
		return &models.EnhancedGame{
			ID:              generateID(game.ID),
			ExternalID:      game.ID,
			Title:           game.Name,
			Description:     "", // Would need separate API call
			Developer:       "", // Would need separate API call
			Publisher:       "", // Would need separate API call
			Genres:          game.Genres,
			Platforms:       game.Platforms,
			CoverURL:        game.Image,
			Screenshots:     []string{}, // Would need separate API call
			Trailers:        []string{}, // Would need separate API call
			ReleaseDate:     &releaseDate,
			PriceINR:        convertToINR(game.Price, game.Currency),
			OriginalINR:     int(convertToINR(game.OriginalPrice, game.Currency)),
			DiscountPercent:  game.Discount,
			IsActive:        true,
			LastPriceUpdate:  timePtr(time.Now()),
		}, nil
		
	case "humble":
		var response HumbleBundleResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode Humble Bundle response: %w", err)
		}
		
		if !response.Success || len(response.Data) == 0 {
			return nil, fmt.Errorf("game not found")
		}
		
		game := response.Data[0]
		releaseDate, _ := time.Parse("2006-01-02", game.ReleaseDate)
		
		return &models.EnhancedGame{
			ID:              generateID(game.ID),
			ExternalID:      game.ID,
			Title:           game.Name,
			Description:     "", // Would need separate API call
			Developer:       "", // Would need separate API call
			Publisher:       "", // Would need separate API call
			Genres:          game.Genres,
			Platforms:       game.Platforms,
			CoverURL:        game.Image,
			Screenshots:     []string{}, // Would need separate API call
			Trailers:        []string{}, // Would need separate API call
			ReleaseDate:     &releaseDate,
			PriceINR:        convertToINR(game.Price, game.Currency),
			OriginalINR:     int(convertToINR(game.OriginalPrice, game.Currency)),
			DiscountPercent:  game.Discount,
			IsActive:        true,
			LastPriceUpdate:  timePtr(time.Now()),
		}, nil
	}

	return nil, fmt.Errorf("unsupported store: %s", store)
}

// Helper functions
func generateID(id string) int64 {
	// Simple hash generation for consistent IDs
	hash := 0
	for _, char := range id {
		hash = hash*31 + int(char)
	}
	return int64(hash % 1000000) + 1000000 // Ensure positive ID and avoid conflicts
}

func convertToINR(price float64, currency string) float64 {
	// Simple conversion rates (in production, this should use real-time rates)
	rates := map[string]float64{
		"USD": 83.0,  // 1 USD = 83 INR
		"EUR": 89.0,  // 1 EUR = 89 INR
		"GBP": 105.0, // 1 GBP = 105 INR
	}
	
	if rate, exists := rates[currency]; exists {
		return price * rate
	}
	
	return price // Default to INR if currency not found
}
