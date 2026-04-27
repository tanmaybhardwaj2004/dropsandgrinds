package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

// BundleGame represents a game in a bundle
type BundleGame struct {
	Title        string  `json:"title"`
	SteamID      int64   `json:"steam_id,omitempty"`
	CurrentPrice float64 `json:"current_price_inr"`
	BundleShare  float64 `json:"bundle_share_inr"`
}

// BundleAnalysis represents the result of bundle analysis
type BundleAnalysis struct {
	BundleURL     string       `json:"bundle_url"`
	BundleName    string       `json:"bundle_name"`
	BundlePrice   float64      `json:"bundle_price_inr"`
	Games         []BundleGame `json:"games"`
	IndividualSum float64      `json:"individual_sum_inr"`
	Verdict       string       `json:"verdict"` // "buy_bundle", "buy_separately", "mixed"
	Savings       float64      `json:"savings_inr"`
	ScrapedAt     time.Time    `json:"scraped_at"`
}

// BundleService handles bundle analysis
type BundleService struct {
	catalogRepo *repositories.CatalogRepository
	logger      *slog.Logger
	httpClient  *http.Client
}

// NewBundleService creates a new bundle service
func NewBundleService(catalogRepo *repositories.CatalogRepository, logger *slog.Logger) *BundleService {
	return &BundleService{
		catalogRepo: catalogRepo,
		logger:      logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AnalyzeBundle analyzes a bundle URL and provides a recommendation
func (s *BundleService) AnalyzeBundle(ctx context.Context, bundleURL string, bundlePrice float64) (*BundleAnalysis, error) {
	// Detect bundle type and extract games
	games, err := s.scrapeBundle(ctx, bundleURL)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape bundle: %w", err)
	}

	// Fetch current prices for each game
	analysis := &BundleAnalysis{
		BundleURL:   bundleURL,
		BundleName:  s.extractBundleName(bundleURL),
		BundlePrice: bundlePrice,
		Games:       make([]BundleGame, 0, len(games)),
		ScrapedAt:   time.Now(),
	}

	individualSum := 0.0
	bundleShare := bundlePrice / float64(len(games))

	for _, gameTitle := range games {
		// Try to find game in catalog by title (using ILIKE for fuzzy match)
		currentPrice := 0.0
		steamID := int64(0)

		// For MVP, we'll need to add a method to search by title
		// For now, skip games not in catalog
		s.logger.Warn("Game title search not fully implemented", "title", gameTitle)

		analysis.Games = append(analysis.Games, BundleGame{
			Title:        gameTitle,
			SteamID:      steamID,
			CurrentPrice: currentPrice,
			BundleShare:  bundleShare,
		})

		individualSum += currentPrice
	}

	analysis.IndividualSum = individualSum
	analysis.Savings = individualSum - bundlePrice

	// Determine verdict
	analysis.Verdict = s.determineVerdict(analysis)

	return analysis, nil
}

// scrapeBundle scrapes a bundle page and extracts game titles
func (s *BundleService) scrapeBundle(ctx context.Context, url string) ([]string, error) {
	// For MVP, implement basic URL pattern matching
	// In production, this would use proper HTML parsing with colly or similar

	// Respect robots.txt and add delay
	time.Sleep(1 * time.Second)

	// Detect bundle type from URL
	if strings.Contains(url, "humblebundle.com") {
		return s.extractHumbleGames(url), nil
	}
	if strings.Contains(url, "fanatical.com") {
		return s.extractFanaticalGames(url), nil
	}
	if strings.Contains(url, "store.steampowered.com") && strings.Contains(url, "bundle") {
		return s.extractSteamGames(url), nil
	}

	return nil, fmt.Errorf("unsupported bundle URL")
}

// extractHumbleGames extracts games from Humble Bundle URL (placeholder)
func (s *BundleService) extractHumbleGames(url string) []string {
	// MVP: Return empty list
	// In production: Parse HTML and extract game titles
	s.logger.Info("Humble bundle scraping not fully implemented", "url", url)
	return []string{}
}

// extractFanaticalGames extracts games from Fanatical URL (placeholder)
func (s *BundleService) extractFanaticalGames(url string) []string {
	// MVP: Return empty list
	// In production: Parse HTML and extract game titles
	s.logger.Info("Fanatical bundle scraping not fully implemented", "url", url)
	return []string{}
}

// extractSteamGames extracts games from Steam Bundle URL (placeholder)
func (s *BundleService) extractSteamGames(url string) []string {
	// MVP: Return empty list
	// In production: Parse HTML and extract game titles
	s.logger.Info("Steam bundle scraping not fully implemented", "url", url)
	return []string{}
}

// extractBundleName extracts a readable name from the URL
func (s *BundleService) extractBundleName(url string) string {
	// Simple extraction from URL path
	parts := strings.Split(url, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" && !strings.Contains(parts[i], ".") {
			// Convert kebab-case to Title Case
			name := strings.ReplaceAll(parts[i], "-", " ")
			name = strings.Title(name)
			return name
		}
	}
	return "Unknown Bundle"
}

// determineVerdict determines whether to buy the bundle or separately
func (s *BundleService) determineVerdict(analysis *BundleAnalysis) string {
	if analysis.IndividualSum == 0 {
		return "mixed" // Can't determine without price data
	}

	savingsPercent := (analysis.Savings / analysis.IndividualSum) * 100

	if savingsPercent > 20 {
		return "buy_bundle"
	}
	if savingsPercent < 0 {
		return "buy_separately"
	}
	return "mixed"
}

// ValidateBundleURL validates if a URL is a supported bundle URL
func (s *BundleService) ValidateBundleURL(url string) bool {
	supportedPatterns := []string{
		"humblebundle.com",
		"fanatical.com",
		"store.steampowered.com",
	}

	for _, pattern := range supportedPatterns {
		if strings.Contains(url, pattern) {
			return true
		}
	}
	return false
}
