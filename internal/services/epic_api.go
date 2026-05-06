package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

type EpicGamesAPIService struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewEpicGamesAPIService(apiKey string) *EpicGamesAPIService {
	return &EpicGamesAPIService{
		apiKey:  apiKey,
		baseURL: "https://store.epicgames.com",
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Epic Games GraphQL query structures
type EpicGraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type EpicGraphQLResponse struct {
	Data    EpicGameData `json:"data"`
	Errors []EpicError  `json:"errors"`
}

type EpicGameData struct {
	Catalog struct {
		SearchStore struct {
			Elements []EpicGame `json:"elements"`
		} `json:"searchStore"`
	} `json:"Catalog"`
}

type EpicGame struct {
	ID             string                 `json:"id"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Developer      []EpicDeveloper        `json:"developer"`
	Publisher      []EpicPublisher         `json:"publisher"`
	Genres         []EpicGenre             `json:"genres"`
	ReleaseDate    string                 `json:"releaseDate"`
	Platforms      []string               `json:"platforms"`
	Images         EpicImages              `json:"images"`
	Offer          *EpicOffer             `json:"offer"`
	Rating         string                 `json:"rating"`
	Url            string                 `json:"url"`
	ProductSlug    string                 `json:"productSlug"`
}

type EpicDeveloper struct {
	Name string `json:"name"`
}

type EpicPublisher struct {
	Name string `json:"name"`
}

type EpicGenre struct {
	Name string `json:"name"`
}

type EpicImages struct {
	OfferImageWide []EpicImage `json:"offerImageWide"`
	OfferImageTall []EpicImage `json:"offerImageTall"`
	Thumbnail      []EpicImage `json:"thumbnail"`
}

type EpicImage struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Size   int    `json:"size"`
	MD5    string `json:"md5"`
}

type EpicOffer struct {
	Offers []EpicPrice `json:"offers"`
}

type EpicPrice struct {
	ID             string  `json:"id"`
	Slug           string  `json:"slug"`
	OriginalPrice   float64 `json:"originalPrice"`
	DiscountPrice  float64 `json:"discountPrice"`
	DiscountPercent int     `json:"discountPercentage"`
	Currency       string  `json:"currency"`
	EffectiveDate  string  `json:"effectiveDate"`
	ExpiryDate     string  `json:"expiryDate"`
}

type EpicError struct {
	Message string `json:"message"`
}

// GetGameDetails fetches comprehensive game information from Epic Games
func (e *EpicGamesAPIService) GetGameDetails(ctx context.Context, slug string) (*models.EnhancedGame, error) {
	query := `
		query getGameDetails($slug: String!) {
			Catalog {
				searchStore {
					elements(namespace: "epic/games") {
						id
						title
						description
						developer { name }
						publisher { name }
						genres { name }
						releaseDate
						platforms
						images {
							offerImageWide
							offerImageTall
							thumbnail
						}
						offer {
							offers {
								id
								slug
								originalPrice
								discountPrice
								discountPercentage
								currency
								effectiveDate
								expiryDate
							}
						}
						rating
						url
						productSlug
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"slug": slug,
	}

	request := EpicGraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/graphql", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Epic Games API returned status %d", resp.StatusCode)
	}

	var response EpicGraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", response.Errors)
	}

	if len(response.Data.Catalog.SearchStore.Elements) == 0 {
		return nil, fmt.Errorf("game not found")
	}

	element := response.Data.Catalog.SearchStore.Elements[0]
	game := &models.EnhancedGame{
		ID:              parseEpicID(element.ID),
		ExternalID:      element.ID,
		Title:           element.Title,
		Description:     element.Description,
		Developer:       extractEpicNames(element.Developer),
		Publisher:       extractEpicNames(element.Publisher),
		Genres:          extractEpicGenres(element.Genres),
		Platforms:       element.Platforms,
		CoverURL:        extractBestEpicImage(element.Images.Thumbnail),
		ReleaseDate:     parseEpicDate(element.ReleaseDate),
		IsActive:        true,
		LastPriceUpdate:  &[]time.Time{time.Now()},
	}

	// Extract price information
	if element.Offer != nil && len(element.Offer.Offers) > 0 {
		offer := element.Offer.Offers[0]
		game.PriceINR = convertToINR(offer.DiscountPrice, offer.Currency)
		game.OriginalINR = int(convertToINR(offer.OriginalPrice, offer.Currency))
		game.DiscountPercent = offer.DiscountPercentage
	}

	// Extract screenshots from images
	if len(element.Images.OfferImageWide) > 0 {
		var screenshots []string
		for _, img := range element.Images.OfferImageWide {
			screenshots = append(screenshots, img.URL)
		}
		game.Screenshots = screenshots
	}

	return game, nil
}

// SearchGames searches for games on Epic Games
func (e *EpicGamesAPIService) SearchGames(ctx context.Context, query string, limit int) ([]models.EnhancedGame, error) {
	searchQuery := `
		query searchGames($query: String!, $limit: Int!) {
			Catalog {
				searchStore {
					elements(namespace: "epic/games") {
						id
						title
						description
						developer { name }
						publisher { name }
						genres { name }
						releaseDate
						platforms
						images {
							offerImageWide
							offerImageTall
							thumbnail
						}
						offer {
							offers {
								id
								slug
								originalPrice
								discountPrice
								discountPercentage
								currency
							}
						}
						rating
						url
						productSlug
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"query": query,
		"limit": limit,
	}

	request := EpicGraphQLRequest{
		Query:     searchQuery,
		Variables: variables,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/graphql", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Epic Games API returned status %d", resp.StatusCode)
	}

	var response EpicGraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", response.Errors)
	}

	var games []models.EnhancedGame
	for _, element := range response.Data.Catalog.SearchStore.Elements {
		game := models.EnhancedGame{
			ID:              parseEpicID(element.ID),
			ExternalID:      element.ID,
			Title:           element.Title,
			Description:     element.Description,
			Developer:       extractEpicNames(element.Developer),
			Publisher:       extractEpicNames(element.Publisher),
			Genres:          extractEpicGenres(element.Genres),
			Platforms:       element.Platforms,
			CoverURL:        extractBestEpicImage(element.Images.Thumbnail),
			ReleaseDate:     parseEpicDate(element.ReleaseDate),
			IsActive:        true,
			LastPriceUpdate:  &[]time.Time{time.Now()},
		}

		// Extract price information
		if element.Offer != nil && len(element.Offer.Offers) > 0 {
			offer := element.Offer.Offers[0]
			game.PriceINR = convertToINR(offer.DiscountPrice, offer.Currency)
			game.OriginalINR = int(convertToINR(offer.OriginalPrice, offer.Currency))
			game.DiscountPercent = offer.DiscountPercentage
		}

		// Extract screenshots
		if len(element.Images.OfferImageWide) > 0 {
			var screenshots []string
			for _, img := range element.Images.OfferImageWide {
				screenshots = append(screenshots, img.URL)
			}
			game.Screenshots = screenshots
		}

		games = append(games, game)
	}

	return games, nil
}

// GetFeaturedGames gets featured games from Epic Games
func (e *EpicGamesAPIService) GetFeaturedGames(ctx context.Context, limit int) ([]models.EnhancedGame, error) {
	featuredQuery := `
		query getFeaturedGames($limit: Int!) {
			Catalog {
				searchStore {
					elements(namespace: "epic/games", sortBy: "effectiveDate", sortDir: "DESC", allowPromotions: true) {
						id
						title
						description
						developer { name }
						publisher { name }
						genres { name }
						releaseDate
						platforms
						images {
							offerImageWide
							offerImageTall
							thumbnail
						}
						offer {
							offers {
								id
								slug
								originalPrice
								discountPrice
								discountPercentage
								currency
							}
						}
						rating
						url
						productSlug
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"limit": limit,
	}

	request := EpicGraphQLRequest{
		Query:     featuredQuery,
		Variables: variables,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/graphql", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Epic Games API returned status %d", resp.StatusCode)
	}

	var response EpicGraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", response.Errors)
	}

	var games []models.EnhancedGame
	for _, element := range response.Data.Catalog.SearchStore.Elements {
		game := models.EnhancedGame{
			ID:              parseEpicID(element.ID),
			ExternalID:      element.ID,
			Title:           element.Title,
			Description:     element.Description,
			Developer:       extractEpicNames(element.Developer),
			Publisher:       extractEpicNames(element.Publisher),
			Genres:          extractEpicGenres(element.Genres),
			Platforms:       element.Platforms,
			CoverURL:        extractBestEpicImage(element.Images.Thumbnail),
			ReleaseDate:     parseEpicDate(element.ReleaseDate),
			IsActive:        true,
			LastPriceUpdate:  &[]time.Time{time.Now()},
		}

		// Extract price information
		if element.Offer != nil && len(element.Offer.Offers) > 0 {
			offer := element.Offer.Offers[0]
			game.PriceINR = convertToINR(offer.DiscountPrice, offer.Currency)
			game.OriginalINR = int(convertToINR(offer.OriginalPrice, offer.Currency))
			game.DiscountPercent = offer.DiscountPercentage
		}

		// Extract screenshots
		if len(element.Images.OfferImageWide) > 0 {
			var screenshots []string
			for _, img := range element.Images.OfferImageWide {
				screenshots = append(screenshots, img.URL)
			}
			game.Screenshots = screenshots
		}

		games = append(games, game)
	}

	return games, nil
}

// Helper functions
func parseEpicID(epicID string) int64 {
	// Epic IDs are typically UUID-like strings, extract numeric part for our system
	if len(epicID) > 10 {
		// Simple hash conversion for Epic IDs
		hash := 0
		for _, char := range epicID {
			hash = hash*31 + int(char)
		}
		return int64(hash % 1000000) // Ensure positive ID
	}
	return 1
}

func extractEpicNames(developers []EpicDeveloper) string {
	if len(developers) == 0 {
		return ""
	}
	var names []string
	for _, dev := range developers {
		names = append(names, dev.Name)
	}
	return fmt.Sprintf("%v", names)
}

func extractEpicNames(publishers []EpicPublisher) string {
	if len(publishers) == 0 {
		return ""
	}
	var names []string
	for _, pub := range publishers {
		names = append(names, pub.Name)
	}
	return fmt.Sprintf("%v", names)
}

func extractEpicGenres(genres []EpicGenre) []string {
	if len(genres) == 0 {
		return []string{}
	}
	var result []string
	for _, genre := range genres {
		result = append(result, genre.Name)
	}
	return result
}

func extractBestEpicImage(images []EpicImage) string {
	if len(images) == 0 {
		return ""
	}
	
	// Find the best quality image (largest dimensions)
	bestImage := images[0]
	for _, img := range images {
		if img.Width > bestImage.Width || img.Height > bestImage.Height {
			bestImage = img
		}
	}
	return bestImage.URL
}

func parseEpicDate(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}
	
	// Epic date format: "2020-12-10T00:00:00.000Z"
	if parsed, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return &parsed
	}
	
	// Try alternative formats
	layouts := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02",
		"Jan 2, 2006",
	}
	
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, dateStr); err == nil {
			return &parsed
		}
	}
	
	return nil
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
