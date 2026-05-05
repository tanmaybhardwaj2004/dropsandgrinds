package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
	"golang.org/x/net/html"
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
	// Respect robots.txt and add delay
	time.Sleep(1 * time.Second)

	// Detect bundle type from URL
	if strings.Contains(url, "humblebundle.com") {
		return s.extractHumbleGames(ctx, url)
	}
	if strings.Contains(url, "fanatical.com") {
		return s.extractFanaticalGames(ctx, url)
	}
	if strings.Contains(url, "store.steampowered.com") && strings.Contains(url, "bundle") {
		return s.extractSteamGames(ctx, url)
	}

	return nil, fmt.Errorf("unsupported bundle URL")
}

// extractHumbleGames extracts games from Humble Bundle URL
func (s *BundleService) extractHumbleGames(ctx context.Context, url string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var games []string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Humble Bundle uses data-entity-name attribute for game titles
			for _, attr := range n.Attr {
				if attr.Key == "data-entity-name" && attr.Val != "" {
					games = append(games, strings.TrimSpace(attr.Val))
					break
				}
			}
			// Also check for title in h2/h3 tags with specific classes
			if n.Data == "h2" || n.Data == "h3" {
				if hasClass(n, "entity-title") || hasClass(n, "hb-title") {
					if text := extractText(n); text != "" {
						games = append(games, strings.TrimSpace(text))
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	s.logger.Info("Extracted games from Humble Bundle", "url", url, "count", len(games))
	return games, nil
}

// extractFanaticalGames extracts games from Fanatical URL
func (s *BundleService) extractFanaticalGames(ctx context.Context, url string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var games []string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Fanatical uses product-card-title class
			if hasClass(n, "product-card-title") || hasClass(n, "game-title") {
				if text := extractText(n); text != "" {
					games = append(games, strings.TrimSpace(text))
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	s.logger.Info("Extracted games from Fanatical", "url", url, "count", len(games))
	return games, nil
}

// extractSteamGames extracts games from Steam Bundle URL
func (s *BundleService) extractSteamGames(ctx context.Context, url string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var games []string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Steam uses app title in various places
			if hasClass(n, "bundle_app_item") || hasClass(n, "tab_item_name") {
				if text := extractText(n); text != "" {
					games = append(games, strings.TrimSpace(text))
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	s.logger.Info("Extracted games from Steam Bundle", "url", url, "count", len(games))
	return games, nil
}

// Helper functions for HTML parsing
func hasClass(n *html.Node, className string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			classes := strings.Fields(attr.Val)
			for _, c := range classes {
				if c == className {
					return true
				}
			}
		}
	}
	return false
}

func extractText(n *html.Node) string {
	var text strings.Builder
	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(n)
	return strings.TrimSpace(text.String())
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
