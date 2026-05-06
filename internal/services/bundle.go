package services

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
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
		currentPrice := 0.0
		steamID := int64(0)
		matches, err := s.catalogRepo.FindGamePricesByTitle(ctx, gameTitle, 1)
		if err == nil && len(matches) > 0 {
			currentPrice = float64(matches[0].PriceINR)
			steamID = matches[0].ID
		}

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
	time.Sleep(1 * time.Second)
	if err := s.allowedByRobots(ctx, url); err != nil {
		return nil, err
	}

	// Detect bundle type from URL
	if strings.Contains(url, "humblebundle.com") {
		return s.extractGamesFromPage(ctx, url)
	}
	if strings.Contains(url, "fanatical.com") {
		return s.extractGamesFromPage(ctx, url)
	}
	if strings.Contains(url, "store.steampowered.com") && strings.Contains(url, "bundle") {
		return s.extractGamesFromPage(ctx, url)
	}

	return nil, fmt.Errorf("unsupported bundle URL")
}

func (s *BundleService) extractGamesFromPage(ctx context.Context, pageURL string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "DropsAndGrindsBot/1.0")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bundle page returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	var games []string
	addTitle := func(raw string) {
		title := cleanBundleTitle(raw)
		if title == "" {
			return
		}
		key := strings.ToLower(title)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		games = append(games, title)
	}
	html := string(body)
	for _, title := range extractJSONLDTitles(html) {
		addTitle(title)
	}
	if len(games) == 0 {
		for _, title := range extractBundleSelectorTitles(html) {
			addTitle(title)
		}
	}
	return games, nil
}

func extractJSONLDTitles(html string) []string {
	scriptRe := regexp.MustCompile(`(?is)<script[^>]+type=["']application/ld\+json["'][^>]*>(.*?)</script>`)
	nameRe := regexp.MustCompile(`(?i)"name"\s*:\s*"([^"]{3,120})"`)
	var titles []string
	for _, script := range scriptRe.FindAllStringSubmatch(html, -1) {
		for _, match := range nameRe.FindAllStringSubmatch(script[1], -1) {
			titles = append(titles, htmlUnescape(match[1]))
		}
	}
	return titles
}

func extractBundleSelectorTitles(html string) []string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)<[^>]+class=["'][^"']*(?:product-title|game-title|bundle-item-title|entity-title|product-name)[^"']*["'][^>]*>(.*?)</[^>]+>`),
		regexp.MustCompile(`(?is)<[^>]+(?:data-title|data-game-name|aria-label)=["']([^"']{3,120})["'][^>]*>`),
	}
	var titles []string
	for _, pattern := range patterns {
		for _, match := range pattern.FindAllStringSubmatch(html, -1) {
			titles = append(titles, htmlUnescape(stripTags(match[1])))
		}
	}
	return titles
}

func stripTags(value string) string {
	tagRe := regexp.MustCompile(`(?is)<[^>]+>`)
	return tagRe.ReplaceAllString(value, " ")
}

func htmlUnescape(value string) string {
	replacer := strings.NewReplacer("&amp;", "&", "&#39;", "'", "&quot;", `"`, "&nbsp;", " ")
	return replacer.Replace(value)
}

func (s *BundleService) allowedByRobots(ctx context.Context, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	robotsURL := parsed.Scheme + "://" + parsed.Host + "/robots.txt"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, robotsURL, nil)
	if err != nil {
		return err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	path := parsed.EscapedPath()
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(strings.ToLower(line))
		if strings.HasPrefix(line, "disallow:") {
			disallowed := strings.TrimSpace(strings.TrimPrefix(line, "disallow:"))
			if disallowed != "" && strings.HasPrefix(strings.ToLower(path), disallowed) {
				return fmt.Errorf("bundle URL disallowed by robots.txt")
			}
		}
	}
	return nil
}

func cleanBundleTitle(title string) string {
	title = strings.TrimSpace(strings.ReplaceAll(title, "\n", " "))
	lower := strings.ToLower(title)
	if strings.Contains(lower, "logo") || strings.Contains(lower, "icon") || strings.Contains(lower, "bundle") || strings.Contains(lower, "cart") {
		return ""
	}
	return title
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
