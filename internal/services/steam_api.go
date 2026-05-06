package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

type SteamAPIService struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewSteamAPIService(apiKey string) *SteamAPIService {
	return &SteamAPIService{
		apiKey:  apiKey,
		baseURL: "https://store.steampowered.com/api",
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type SteamAppResponse struct {
	Success bool         `json:"success"`
	Data    SteamAppData `json:"data"`
}

type SteamAppData struct {
	Type                string                `json:"type"`
	Name                string                `json:"name"`
	SteamAppID          int                   `json:"steam_appid"`
	RequiredAge         int                   `json:"required_age"`
	IsFree              bool                  `json:"is_free"`
	ControllerSupport   []string              `json:"controller_support"`
	Description         string                `json:"description"`
	AboutTheGame        string                `json:"about_the_game"`
	SupportedLanguages  []string              `json:"supported_languages"`
	Reviews             string                `json:"reviews"`
	DetailedDescription string                `json:"detailed_description"`
	AboutTheGame        string                `json:"about_the_game"`
	ShortDescription    string                `json:"short_description"`
	Fullgame            *SteamFullgame        `json:"fullgame"`
	LinuxRequirements   *SteamRequirements    `json:"linux_requirements"`
	MacRequirements     *SteamRequirements    `json:"mac_requirements"`
	WindowsRequirements *SteamRequirements    `json:"windows_requirements"`
	HeaderImage         string                `json:"header_image"`
	Website             string                `json:"website"`
	PcRequirements      *SteamPCRequirements  `json:"pc_requirements"`
	LegalNotice         string                `json:"legal_notice"`
	Developers          []string              `json:"developers"`
	Publishers          []string              `json:"publishers"`
	PriceOverview       *SteamPriceOverview   `json:"price_overview"`
	Platforms           []string              `json:"platforms"`
	Metacritic          *SteamMetacritic      `json:"metacritic"`
	Recommendations     *SteamRecommendations `json:"recommendations"`
	IsFreeVRAvailable   bool                  `json:"is_free_vr_available"`
	ReleaseDate         *SteamReleaseDate     `json:"release_date"`
	SupportEmail        string                `json:"support_email"`
	Background          string                `json:"background"`
	ContentDescriptors  []string              `json:"content_descriptors"`
}

type SteamFullgame struct {
	AppID           int    `json:"appid"`
	Title           string `json:"title"`
	Genre           string `json:"genre"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	Featured        bool   `json:"featured"`
	WorkshopStatus  string `json:"workshop_status"`
	ReleaseDate     string `json:"release_date"`
	ComingSoon      bool   `json:"coming_soon"`
	IsDiscount      bool   `json:"is_discount"`
	DiscountPercent int    `json:"discount_percent"`
	OriginalPrice   int    `json:"original_price"`
	FinalPrice      int    `json:"final_price"`
	Currency        string `json:"currency"`
	IsFree          bool   `json:"is_free"`
	IsComingSoon    bool   `json:"is_coming_soon"`
	HasDemo         bool   `json:"has_demo"`
	HasController   bool   `json:"has_controller_support"`
	HasVR           bool   `json:"has_vr_support"`
	HasAchievements bool   `json:"has_achievements"`
	HasCloud        bool   `json:"has_cloud_saves"`
	HasCaptions     bool   `json:"has_captions"`
	HasCommentary   bool   `json:"has_commentary"`
	HasStats        bool   `json:"has_stats"`
	HasLeaderboard  bool   `json:"has_leaderboard"`
	HasTradingCards bool   `json:"has_trading_cards"`
	HasMarket       bool   `json:"has_market"`
	HasWorkshop     bool   `json:"has_workshop"`
	HasDLC          bool   `json:"has_dlc"`
	IsEarlyAccess   bool   `json:"is_early_access"`
	IsPrePurchase   bool   `json:"is_pre_purchase"`
	IsFreeWeekend   bool   `json:"is_free_weekend"`
}

type SteamRequirements struct {
	Minimum     string `json:"minimum"`
	Recommended string `json:"recommended"`
}

type SteamPCRequirements struct {
	Minimum     *SteamRequirements `json:"minimum"`
	Recommended *SteamRequirements `json:"recommended"`
}

type SteamPriceOverview struct {
	Currency        string       `json:"currency"`
	Initial         int          `json:"initial"`
	Final           int          `json:"final"`
	DiscountPercent int          `json:"discount_percent"`
	Individual      []SteamPrice `json:"individual"`
	Volume          []SteamPrice `json:"volume"`
}

type SteamPrice struct {
	PackageID  int    `json:"packageid"`
	BundleID   int    `json:"bundleid"`
	SubID      int    `json:"subid"`
	PriceInINR int    `json:"price_inr"`
	Discount   int    `json:"discount_percent"`
	Currency   string `json:"currency"`
}

type SteamMetacritic struct {
	Score   int    `json:"score"`
	URL     string `json:"url"`
	Summary string `json:"summary"`
}

type SteamRecommendations struct {
	Total int `json:"total"`
	Data  []struct {
		AppID int    `json:"id"`
		Name  string `json:"name"`
	} `json:"data"`
}

type SteamReleaseDate struct {
	Date       string `json:"date"`
	ComingSoon bool   `json:"coming_soon"`
}

// GetGameDetails fetches comprehensive game information from Steam
func (s *SteamAPIService) GetGameDetails(ctx context.Context, appID int) (*models.EnhancedGame, error) {
	url := fmt.Sprintf("%s/appdetails?appids=%d&l=english&cc=IN", s.baseURL, appID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Steam API returned status %d", resp.StatusCode)
	}

	var response []SteamAppResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response) == 0 || !response[0].Success {
		return nil, fmt.Errorf("game not found or API error")
	}

	data := response[0].Data
	game := &models.EnhancedGame{
		ID:              int64(appID),
		ExternalID:      strconv.Itoa(appID),
		Title:           data.Name,
		Description:     data.Description,
		Developer:       fmt.Sprintf("%v", data.Developers),
		Publisher:       fmt.Sprintf("%v", data.Publishers),
		CoverURL:        data.HeaderImage,
		Rating:          data.Reviews,
		Platforms:       data.Platforms,
		ReleaseDate:     parseSteamReleaseDate(data.ReleaseDate),
		IsActive:        true,
		LastPriceUpdate: &[]time.Time{time.Now()},
	}

	// Parse genres and system requirements
	if data.DetailedDescription != "" {
		// Extract genres from description (simplified approach)
		game.Genres = extractGenres(data.DetailedDescription)
	}

	if data.WindowsRequirements != nil {
		game.SystemRequirements = &models.SystemReqs{
			OS:        "Windows",
			Processor: data.WindowsRequirements.Recommended,
			Memory:    data.WindowsRequirements.Recommended,
			Graphics:  data.WindowsRequirements.Recommended,
			Storage:   data.WindowsRequirements.Recommended,
		}
	}

	// Extract screenshots (Steam doesn't provide screenshots in appdetails API)
	game.Screenshots = []string{}
	game.Trailers = []string{}

	// Handle editions and DLC
	if data.Fullgame != nil {
		game.Editions = []models.Edition{
			{
				ID:          int64(data.Fullgame.AppID),
				Name:        data.Fullgame.Title,
				Description: data.Fullgame.Description,
				PriceINR:    data.Fullgame.FinalPrice,
				IsDLC:       false,
				Features:    []string{"Base Game"},
				ReleaseDate: parseSteamReleaseDate(data.Fullgame.ReleaseDate),
			},
		}
	}

	return game, nil
}

// GetGamePrice fetches current price information from Steam
func (s *SteamAPIService) GetGamePrice(ctx context.Context, appID int) (*models.EnhancedPrice, error) {
	url := fmt.Sprintf("%s/appdetails?appids=%d&l=english&cc=IN&filters=price_overview", s.baseURL, appID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Steam API returned status %d", resp.StatusCode)
	}

	var response []SteamAppResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response) == 0 || !response[0].Success || response[0].PriceOverview == nil {
		return nil, fmt.Errorf("price information not available")
	}

	priceData := response[0].PriceOverview
	price := &models.EnhancedPrice{
		PriceINR:        float64(priceData.Final),
		OriginalPrice:   float64(priceData.Initial),
		DiscountPercent: priceData.DiscountPercent,
		Currency:        priceData.Currency,
		Region:          "IN",
		IsAvailable:     true,
		StockStatus:     "in_stock",
		DealType:        "regular",
		UpdatedAt:       time.Now(),
	}

	return price, nil
}

// SearchGames searches for games on Steam
func (s *SteamAPIService) SearchGames(ctx context.Context, query string, limit int) ([]models.EnhancedGame, error) {
	url := fmt.Sprintf("%s/storesearch/?term=%s&l=english&cc=IN&category=9980&supportedlang=english&page=1&count=%d",
		s.baseURL, query, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Steam API returned status %d", resp.StatusCode)
	}

	var response struct {
		Items []struct {
			ID            int    `json:"id"`
			Name          string `json:"name"`
			Description   string `json:"short_description"`
			HeaderImage   string `json:"tiny_image"`
			ReleaseDate   string `json:"release_date"`
			IsFree        bool   `json:"is_free"`
			PriceOverview *struct {
				Final           int    `json:"final"`
				Initial         int    `json:"initial"`
				DiscountPercent int    `json:"discount_percent"`
				Currency        string `json:"currency"`
			} `json:"price_overview"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var games []models.EnhancedGame
	for _, item := range response.Items {
		game := models.EnhancedGame{
			ID:              int64(item.ID),
			ExternalID:      strconv.Itoa(item.ID),
			Title:           item.Name,
			Description:     item.Description,
			CoverURL:        item.HeaderImage,
			ReleaseDate:     parseSteamReleaseDate(item.ReleaseDate),
			IsActive:        true,
			LastPriceUpdate: &[]time.Time{time.Now()},
		}

		if item.PriceOverview != nil {
			game.PriceINR = float64(item.PriceOverview.Final)
			game.OriginalINR = item.PriceOverview.Initial
			game.DiscountPercent = item.PriceOverview.DiscountPercent
		}

		games = append(games, game)
	}

	return games, nil
}

// GetFeaturedGames gets featured games from Steam
func (s *SteamAPIService) GetFeaturedGames(ctx context.Context) ([]models.EnhancedGame, error) {
	url := fmt.Sprintf("%s/featured?cc=IN&l=english", s.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Steam API returned status %d", resp.StatusCode)
	}

	var response struct {
		Featured struct {
			Items []struct {
				AppID         int    `json:"id"`
				Name          string `json:"name"`
				Description   string `json:"header_description"`
				HeaderImage   string `json:"header_image"`
				ReleaseDate   string `json:"release_date"`
				PriceOverview *struct {
					Final           int    `json:"final"`
					Initial         int    `json:"initial"`
					DiscountPercent int    `json:"discount_percent"`
					Currency        string `json:"currency"`
				} `json:"price_overview"`
			} `json:"items"`
		} `json:"featured"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var games []models.EnhancedGame
	for _, item := range response.Featured.Items {
		game := models.EnhancedGame{
			ID:              int64(item.AppID),
			ExternalID:      strconv.Itoa(item.AppID),
			Title:           item.Name,
			Description:     item.Description,
			CoverURL:        item.HeaderImage,
			ReleaseDate:     parseSteamReleaseDate(item.ReleaseDate),
			IsActive:        true,
			LastPriceUpdate: &[]time.Time{time.Now()},
		}

		if item.PriceOverview != nil {
			game.PriceINR = float64(item.PriceOverview.Final)
			game.OriginalINR = item.PriceOverview.Initial
			game.DiscountPercent = item.PriceOverview.DiscountPercent
		}

		games = append(games, game)
	}

	return games, nil
}

// Helper functions
func parseSteamReleaseDate(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}

	// Steam date format: "15 Dec, 2020"
	layout := "2 Jan, 2006"
	if parsed, err := time.Parse(layout, dateStr); err == nil {
		return &parsed
	}

	// Try alternative formats
	layouts := []string{
		"2006-01-02",
		"Jan 2, 2006",
		"January 2, 2006",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, dateStr); err == nil {
			return &parsed
		}
	}

	return nil
}

func extractGenres(description string) []string {
	// Simple genre extraction based on common game genres
	genreKeywords := map[string][]string{
		"action":     {"action", "shooter", "fighting", "platformer"},
		"rpg":        {"rpg", "role-playing", "fantasy", "story-driven"},
		"adventure":  {"adventure", "puzzle", "mystery", "exploration"},
		"strategy":   {"strategy", "tactical", "turn-based", "simulation"},
		"sports":     {"sports", "racing", "football", "basketball"},
		"simulation": {"simulation", "management", "building", "tycoon"},
		"racing":     {"racing", "driving", "motorcycle"},
		"indie":      {"indie", "independent", "pixel art"},
	}

	var foundGenres []string
	descLower := strings.ToLower(description)

	for genre, keywords := range genreKeywords {
		for _, keyword := range keywords {
			if strings.Contains(descLower, keyword) {
				foundGenres = append(foundGenres, genre)
				break
			}
		}
	}

	// Remove duplicates
	unique := make(map[string]bool)
	var result []string
	for _, genre := range foundGenres {
		if !unique[genre] {
			unique[genre] = true
			result = append(result, genre)
		}
	}

	return result
}
