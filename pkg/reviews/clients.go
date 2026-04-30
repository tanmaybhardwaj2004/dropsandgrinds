package reviews

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// ReviewSource represents a review source
type ReviewSource string

const (
	SourceMetacritic ReviewSource = "metacritic"
	SourceOpenCritic ReviewSource = "opencritic"
	SourceSteam     ReviewSource = "steam"
	SourceIGN       ReviewSource = "ign"
	SourceGameSpot  ReviewSource = "gamespot"
)

// ReviewScore represents a normalized review score
type ReviewScore struct {
	Source   ReviewSource `json:"source"`
	Score    int          `json:"score"`    // 0-100
	URL      string       `json:"url"`
	FetchedAt time.Time    `json:"fetched_at"`
}

// ReviewClient is the interface for fetching review scores
type ReviewClient interface {
	FetchScore(ctx context.Context, gameID string) (*ReviewScore, error)
}

// MetacriticClient fetches scores from Metacritic
type MetacriticClient struct {
	httpClient *http.Client
}

func NewMetacriticClient() *MetacriticClient {
	return &MetacriticClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *MetacriticClient) FetchScore(ctx context.Context, gameID string) (*ReviewScore, error) {
	// For MVP: simulate Metacritic API call
	// In production: use actual Metacritic API or scrape
	return &ReviewScore{
		Source:   SourceMetacritic,
		Score:    85, // Simulated score
		URL:      fmt.Sprintf("https://www.metacritic.com/game/%s", gameID),
		FetchedAt: time.Now(),
	}, nil
}

// OpenCriticClient fetches scores from OpenCritic
type OpenCriticClient struct {
	httpClient *http.Client
}

func NewOpenCriticClient() *OpenCriticClient {
	return &OpenCriticClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *OpenCriticClient) FetchScore(ctx context.Context, gameID string) (*ReviewScore, error) {
	// For MVP: simulate OpenCritic API call
	// In production: use OpenCritic public API
	return &ReviewScore{
		Source:   SourceOpenCritic,
		Score:    88, // Simulated score
		URL:      fmt.Sprintf("https://opencritic.com/game/%s", gameID),
		FetchedAt: time.Now(),
	}, nil
}

// SteamClient fetches review scores from Steam
type SteamClient struct {
	httpClient *http.Client
	apiKey     string
}

func NewSteamClient(apiKey string) *SteamClient {
	return &SteamClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiKey:     apiKey,
	}
}

func (c *SteamClient) FetchScore(ctx context.Context, gameID string) (*ReviewScore, error) {
	// For MVP: simulate Steam API call
	// In production: use Steam appreviews endpoint
	// Convert gameID to Steam AppID if needed
	appID, err := strconv.ParseInt(gameID, 10, 64)
	if err != nil {
		appID = 1091500 // Default to Cyberpunk 2077
	}
	
	// Simulate Steam review percentage (positive reviews)
	positivePercent := 92
	
	return &ReviewScore{
		Source:   SourceSteam,
		Score:    positivePercent,
		URL:      fmt.Sprintf("https://store.steampowered.com/app/%d", appID),
		FetchedAt: time.Now(),
	}, nil
}

// IGNClient fetches scores from IGN
type IGNClient struct {
	httpClient *http.Client
}

func NewIGNClient() *IGNClient {
	return &IGNClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *IGNClient) FetchScore(ctx context.Context, gameID string) (*ReviewScore, error) {
	// For MVP: simulate IGN API call
	// In production: scrape IGN review pages
	return &ReviewScore{
		Source:   SourceIGN,
		Score:    90, // Simulated score
		URL:      fmt.Sprintf("https://www.ign.com/games/%s", gameID),
		FetchedAt: time.Now(),
	}, nil
}

// GameSpotClient fetches scores from GameSpot
type GameSpotClient struct {
	httpClient *http.Client
	apiKey     string
}

func NewGameSpotClient(apiKey string) *GameSpotClient {
	return &GameSpotClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiKey:     apiKey,
	}
}

func (c *GameSpotClient) FetchScore(ctx context.Context, gameID string) (*ReviewScore, error) {
	// For MVP: simulate GameSpot API call
	// In production: use GameSpot API or scrape
	return &ReviewScore{
		Source:   SourceGameSpot,
		Score:    87, // Simulated score
		URL:      fmt.Sprintf("https://www.gamespot.com/games/%s", gameID),
		FetchedAt: time.Now(),
	}, nil
}

// ReviewAggregator aggregates reviews from multiple sources
type ReviewAggregator struct {
	clients map[ReviewSource]ReviewClient
}

func NewReviewAggregator(steamAPIKey, gameSpotAPIKey string) *ReviewAggregator {
	return &ReviewAggregator{
		clients: map[ReviewSource]ReviewClient{
			SourceMetacritic: NewMetacriticClient(),
			SourceOpenCritic: NewOpenCriticClient(),
			SourceSteam:     NewSteamClient(steamAPIKey),
			SourceIGN:       NewIGNClient(),
			SourceGameSpot:  NewGameSpotClient(gameSpotAPIKey),
		},
	}
}

// FetchAllScores fetches scores from all available sources
func (a *ReviewAggregator) FetchAllScores(ctx context.Context, gameID string) ([]ReviewScore, error) {
	var scores []ReviewScore
	
	for _, client := range a.clients {
		score, err := client.FetchScore(ctx, gameID)
		if err != nil {
			// Log error but continue with other sources
			continue
		}
		scores = append(scores, *score)
	}
	
	return scores, nil
}

// FetchScoreFromSource fetches score from a specific source
func (a *ReviewAggregator) FetchScoreFromSource(ctx context.Context, gameID string, source ReviewSource) (*ReviewScore, error) {
	client, ok := a.clients[source]
	if !ok {
		return nil, fmt.Errorf("unsupported review source: %s", source)
	}
	
	return client.FetchScore(ctx, gameID)
}

// Helper function to fetch HTTP content
func fetchHTTP(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	return io.ReadAll(resp.Body)
}
