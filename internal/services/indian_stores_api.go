package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

func timePtrIndian(t time.Time) *time.Time { return &t }

var mockGames = []IndianStoreGame{
	{
		ID:           "instantgaming-cyberpunk",
		Name:         "Cyberpunk 2077",
		Price:        1499.00,
		OriginalPrice: 2999.00,
		Discount:     50,
		Currency:     "INR",
		Store:        "Instant Gaming",
		Region:       "IN",
		IsAvailable:  true,
		Image:        "https://example.com/cyberpunk.jpg",
		URL:          "https://instant-gaming.com/game/cyberpunk-2077",
		ReleaseDate:  "2020-12-10",
		Platforms:    []string{"PC", "PlayStation", "Xbox"},
		Genres:       []string{"RPG", "Action"},
		IndianPayment: true,
	},
	{
		ID:           "gamivo-elden-ring",
		Name:         "Elden Ring",
		Price:        2399.00,
		OriginalPrice: 3999.00,
		Discount:     40,
		Currency:     "INR",
		Store:        "Gamivo",
		Region:       "IN",
		IsAvailable:  true,
		Image:        "https://example.com/elden-ring.jpg",
		URL:          "https://www.gamivo.com/game/elden-ring",
		ReleaseDate:  "2022-02-25",
		Platforms:    []string{"PC", "PlayStation", "Xbox"},
		Genres:       []string{"RPG", "Action"},
		IndianPayment: true,
	},
	{
		ID:           "eneba-gta-v",
		Name:         "Grand Theft Auto V",
		Price:        999.00,
		OriginalPrice: 1999.00,
		Discount:     50,
		Currency:     "INR",
		Store:        "Eneba",
		Region:       "IN",
		IsAvailable:  true,
		Image:        "https://example.com/gta-v.jpg",
		URL:          "https://www.eneba.com/game/grand-theft-auto-v",
		ReleaseDate:  "2022-08-25",
		Platforms:    []string{"PC", "PlayStation", "Xbox"},
		Genres:       []string{"Action", "Open World"},
		IndianPayment: true,
	},
}

type IndianStoresAPIService struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewIndianStoresAPIService(apiKey string) *IndianStoresAPIService {
	return &IndianStoresAPIService{
		apiKey:  apiKey,
		baseURL: "https://api.indianstores.com", // Mock API endpoint
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Indian store API structures (mock implementation for demonstration)
type IndianStoresResponse struct {
	Success bool                     `json:"success"`
	Data    []IndianStoreGame      `json:"data"`
	Message string                   `json:"message"`
}

type IndianStoreGame struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	OriginalPrice float64 `json:"original_price"`
	Discount     int     `json:"discount"`
	Currency     string  `json:"currency"`
	Store        string  `json:"store"`
	Region       string  `json:"region"`
	IsAvailable  bool    `json:"is_available"`
	Image        string  `json:"image"`
	URL          string  `json:"url"`
	ReleaseDate  string  `json:"release_date"`
	Platforms    []string `json:"platforms"`
	Genres       []string `json:"genres"`
	IndianPayment bool   `json:"indian_payment"` // UPI, PhonePe, etc.
}

// GetIndianStoreGames fetches games from Indian-friendly stores
func (i *IndianStoresAPIService) GetIndianStoreGames(ctx context.Context, query string, limit int) ([]models.EnhancedGame, error) {
	// Filter by query if provided
	var filteredGames []IndianStoreGame
	if query != "" {
		queryLower := strings.ToLower(query)
		for _, game := range mockGames {
			if strings.Contains(strings.ToLower(game.Name), queryLower) {
				filteredGames = append(filteredGames, game)
			}
		}
	} else {
		filteredGames = mockGames
	}

	// Apply limit
	if limit > 0 && len(filteredGames) > limit {
		filteredGames = filteredGames[:limit]
	}

	// Convert to EnhancedGame models
	var games []models.EnhancedGame
	for _, game := range filteredGames {
		releaseDate, _ := time.Parse("2006-01-02", game.ReleaseDate)
		
		enhancedGame := models.EnhancedGame{
			ID:              generateIndianStoreID(game.ID),
			ExternalID:      game.ID,
			Title:           game.Name,
			Description:     fmt.Sprintf("%s - Available on %s with Indian payment options", game.Name, game.Store),
			Developer:       "Various Publishers",
			Publisher:       game.Store,
			Genres:          game.Genres,
			Platforms:       game.Platforms,
			CoverURL:        game.Image,
			Screenshots:     []string{}, // Would need separate API call
			Trailers:        []string{}, // Would need separate API call
			ReleaseDate:     &releaseDate,
			PriceINR:        game.Price,
			OriginalINR:     int(game.OriginalPrice),
			DiscountPercent:  game.Discount,
			IsActive:        game.IsAvailable,
			LastPriceUpdate:  timePtrIndian(time.Now()),
		}
		
		games = append(games, enhancedGame)
	}

	return games, nil
}

// GetGameDetails fetches detailed game information from Indian stores
func (i *IndianStoresAPIService) GetGameDetails(ctx context.Context, gameID string) (*models.EnhancedGame, error) {
	// Mock implementation - find game by ID in mock data
	for _, game := range mockGames {
		if game.ID == gameID {
			releaseDate, _ := time.Parse("2006-01-02", game.ReleaseDate)
			
			return &models.EnhancedGame{
				ID:              generateIndianStoreID(game.ID),
				ExternalID:      game.ID,
				Title:           game.Name,
				Description:     fmt.Sprintf("%s - Available with Indian payment methods like UPI, PhonePe, Paytm, and local bank transfers", game.Name),
				Developer:       "Various Publishers",
				Publisher:       game.Store,
				Genres:          game.Genres,
				Platforms:       game.Platforms,
				CoverURL:        game.Image,
				Screenshots:     []string{}, // Would need separate API call
				Trailers:        []string{}, // Would need separate API call
				ReleaseDate:     &releaseDate,
				PriceINR:        game.Price,
				OriginalINR:     int(game.OriginalPrice),
				DiscountPercent:  game.Discount,
				IsActive:        game.IsAvailable,
				LastPriceUpdate:  timePtrIndian(time.Now()),
			}, nil
		}
	}
	
	return nil, fmt.Errorf("game not found")
}

// GetIndianPaymentOffers returns available Indian payment offers
func (i *IndianStoresAPIService) GetIndianPaymentOffers(ctx context.Context) (*models.IndianOffersResponse, error) {
	// Mock implementation of Indian payment offers
	mockOffers := []models.IndianPaymentOffer{
		{
			StoreID:       1, // Instant Gaming
			Store:         models.Store{ID: 1, Name: "Instant Gaming", Slug: "instantgaming"},
			OfferType:     "upi_discount",
			Provider:      "PhonePe",
			Description:   "10% cashback on UPI payments",
			DiscountPercent: 10,
			MaxDiscountAmount: 100.00,
			MinOrderAmount: 500.00,
			IsActive:      true,
			CreatedAt:     time.Now(),
		},
		{
			StoreID:       2, // Gamivo
			Store:         models.Store{ID: 2, Name: "Gamivo", Slug: "gamivo"},
			OfferType:     "wallet_bonus",
			Provider:      "Paytm",
			Description:   "₹50 wallet bonus on first order",
			DiscountPercent: 5,
			MaxDiscountAmount: 50.00,
			MinOrderAmount: 1000.00,
			IsActive:      true,
			CreatedAt:     time.Now(),
		},
		{
			StoreID:       3, // Eneba
			Store:         models.Store{ID: 3, Name: "Eneba", Slug: "eneba"},
			OfferType:     "card_cashback",
			Provider:      "HDFC Bank",
			Description:   "5% cashback on credit/debit cards",
			DiscountPercent: 5,
			MaxDiscountAmount: 250.00,
			MinOrderAmount: 200.00,
			IsActive:      true,
			CreatedAt:     time.Now(),
		},
	}

	response := &models.IndianOffersResponse{
		Offers: mockOffers,
		Total:  len(mockOffers),
	}

	return response, nil
}

// ValidateIndianPayment validates if a payment method is Indian-friendly
func (i *IndianStoresAPIService) ValidateIndianPayment(ctx context.Context, paymentMethod string) (bool, error) {
	indianMethods := []string{
		"upi", "phonepe", "paytm", "gpay", "mobikwik", "amazonpay", 
		"bhim", "googlepay", "payzapp", "freecharge", "mobikwik",
		"phonepe", "paytm", "gpay", "mobikwik", "amazonpay",
	}

	for _, method := range indianMethods {
		if strings.EqualFold(paymentMethod, method) {
			return true, nil
		}
	}

	return false, fmt.Errorf("payment method %s is not supported for Indian users", paymentMethod)
}

// Helper functions
func generateIndianStoreID(id string) int64 {
	// Generate consistent ID for Indian store games
	hash := 0
	for _, char := range id {
		hash = hash*31 + int(char)
	}
	return int64(2000000 + hash%100000) // Ensure unique range
}

// GetRegionalPricing fetches pricing for different regions
func (i *IndianStoresAPIService) GetRegionalPricing(ctx context.Context, gameID string, regions []string) (map[string]models.RegionalPricing, error) {
	// Mock implementation - in production, this would call real APIs
	pricing := make(map[string]models.RegionalPricing)
	
	for _, region := range regions {
		switch region {
		case "US":
			pricing[region] = models.RegionalPricing{
				Region:     "US",
				Currency:   "USD",
				Price:      59.99, // Base price in USD
				Store:      models.Store{ID: 1, Name: "Steam", Slug: "steam"},
				UpdatedAt:   time.Now(),
			}
		case "EU":
			pricing[region] = models.RegionalPricing{
				Region:     "EU",
				Currency:   "EUR",
				Price:      49.99, // Base price in EUR
				Store:      models.Store{ID: 1, Name: "Steam", Slug: "steam"},
				UpdatedAt:   time.Now(),
			}
		case "UK":
			pricing[region] = models.RegionalPricing{
				Region:     "UK",
				Currency:   "GBP",
				Price:      44.99, // Base price in GBP
				Store:      models.Store{ID: 1, Name: "Steam", Slug: "steam"},
				UpdatedAt:   time.Now(),
			}
		default:
			pricing[region] = models.RegionalPricing{
				Region:     region,
				Currency:   "INR",
				Price:      1499.00, // Base price in INR
				Store:      models.Store{ID: 1, Name: "Steam", Slug: "steam"},
				UpdatedAt:   time.Now(),
			}
		}
	}

	return pricing, nil
}
