package reviews

import (
	"context"
	"encoding/json"
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
	SourceSteam      ReviewSource = "steam"
	SourceIGN        ReviewSource = "ign"
	SourceGameSpot   ReviewSource = "gamespot"
)

// ReviewScore represents a normalized review score
type ReviewScore struct {
	Source    ReviewSource `json:"source"`
	Score     int          `json:"score"` // 0-100
	URL       string       `json:"url"`
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
	return nil, fmt.Errorf("metacritic client not configured")
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
	url := fmt.Sprintf("https://api.opencritic.com/api/game/%s", gameID)
	body, err := fetchHTTP(ctx, c.httpClient, url)
	if err != nil {
		return nil, err
	}
	var payload struct {
		TopCriticScore float64 `json:"topCriticScore"`
		MedianScore    float64 `json:"medianScore"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	score := payload.TopCriticScore
	if score <= 0 {
		score = payload.MedianScore
	}
	if score <= 0 {
		return nil, fmt.Errorf("opencritic score unavailable")
	}
	return &ReviewScore{
		Source:    SourceOpenCritic,
		Score:     int(score + 0.5),
		URL:       fmt.Sprintf("https://opencritic.com/game/%s", gameID),
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
	appID, err := strconv.ParseInt(gameID, 10, 64)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://store.steampowered.com/appreviews/%d?json=1&language=all&purchase_type=all", appID)
	body, err := fetchHTTP(ctx, c.httpClient, url)
	if err != nil {
		return nil, err
	}
	var payload struct {
		QuerySummary struct {
			TotalPositive int `json:"total_positive"`
			TotalNegative int `json:"total_negative"`
		} `json:"query_summary"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	total := payload.QuerySummary.TotalPositive + payload.QuerySummary.TotalNegative
	if total == 0 {
		return nil, fmt.Errorf("steam review score unavailable")
	}
	score := int(float64(payload.QuerySummary.TotalPositive)/float64(total)*100 + 0.5)
	return &ReviewScore{
		Source:    SourceSteam,
		Score:     score,
		URL:       fmt.Sprintf("https://store.steampowered.com/app/%d", appID),
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
	return nil, fmt.Errorf("ign client not configured")
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
	return nil, fmt.Errorf("gamespot client not configured")
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
			SourceSteam:      NewSteamClient(steamAPIKey),
			SourceIGN:        NewIGNClient(),
			SourceGameSpot:   NewGameSpotClient(gameSpotAPIKey),
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
		if score != nil {
			scores = append(scores, *score)
		}
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
