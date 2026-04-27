package cheapshark

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL        = "https://www.cheapshark.com/api/1.0"
	dealsEndpoint  = "/deals"
	gameEndpoint   = "/games"
	requestTimeout = 30 * time.Second
)

// Deal represents a game deal from CheapShark
type Deal struct {
	GameID       string  `json:"gameID"`
	Title        string  `json:"title"`
	StoreID      string  `json:"storeID"`
	StoreName    string  `json:"storeName"`
	SalePrice    float64 `json:"salePrice"`
	NormalPrice  float64 `json:"normalPrice"`
	IsOnSale     bool    `json:"isOnSale"`
	Savings      float64 `json:"savings"`
	Metacritic   int     `json:"metacriticScore"`
	SteamRating  float64 `json:"steamRatingPercent"`
	Thumb        string  `json:"thumb"`
	ReleaseDate  int64   `json:"releaseDate"`
	LastChange   int64   `json:"lastChange"`
	DealRating   string  `json:"dealRating"`
}

// Game represents game details from CheapShark
type Game struct {
	GameID      string  `json:"gameID"`
	Title       string  `json:"title"`
	SteamAppID  string  `json:"steamAppID"`
	CheapestPriceEver float64 `json:"cheapestPriceEver"`
}

// Client represents the CheapShark API client
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new CheapShark API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
		baseURL: baseURL,
	}
}

// GetDeals fetches current deals with optional filters
func (c *Client) GetDeals(ctx context.Context, params map[string]string) ([]Deal, error) {
	url := c.baseURL + dealsEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deals: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var deals []Deal
	if err := json.NewDecoder(resp.Body).Decode(&deals); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return deals, nil
}

// GetGameDetails fetches details for a specific game
func (c *Client) GetGameDetails(ctx context.Context, gameID string) (*Game, error) {
	url := fmt.Sprintf("%s%s?id=%s", c.baseURL, gameEndpoint, gameID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch game details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var game Game
	if err := json.NewDecoder(resp.Body).Decode(&game); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &game, nil
}
