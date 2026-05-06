package cheapshark

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	DealID      string  `json:"dealID"`
	GameID      string  `json:"gameID"`
	Title       string  `json:"title"`
	StoreID     string  `json:"storeID"`
	StoreName   string  `json:"storeName"`
	SalePrice   Float64 `json:"salePrice"`
	NormalPrice Float64 `json:"normalPrice"`
	IsOnSale    Bool    `json:"isOnSale"`
	Savings     Float64 `json:"savings"`
	Metacritic  Int     `json:"metacriticScore"`
	SteamRating Float64 `json:"steamRatingPercent"`
	Thumb       string  `json:"thumb"`
	ReleaseDate Int64   `json:"releaseDate"`
	LastChange  Int64   `json:"lastChange"`
	DealRating  string  `json:"dealRating"`
}

type Float64 float64

func (f *Float64) UnmarshalJSON(data []byte) error {
	var number float64
	if err := json.Unmarshal(data, &number); err == nil {
		*f = Float64(number)
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	if text == "" {
		*f = 0
		return nil
	}
	parsed, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return err
	}
	*f = Float64(parsed)
	return nil
}

type Bool bool

func (b *Bool) UnmarshalJSON(data []byte) error {
	var value bool
	if err := json.Unmarshal(data, &value); err == nil {
		*b = Bool(value)
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	if text == "" {
		*b = false
		return nil
	}
	parsed, err := strconv.ParseBool(text)
	if err != nil {
		return err
	}
	*b = Bool(parsed)
	return nil
}

type Int int

func (i *Int) UnmarshalJSON(data []byte) error {
	value, err := parseJSONInt(data)
	if err != nil {
		return err
	}
	*i = Int(value)
	return nil
}

type Int64 int64

func (i *Int64) UnmarshalJSON(data []byte) error {
	value, err := parseJSONInt(data)
	if err != nil {
		return err
	}
	*i = Int64(value)
	return nil
}

func parseJSONInt(data []byte) (int64, error) {
	var number int64
	if err := json.Unmarshal(data, &number); err == nil {
		return number, nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return 0, err
	}
	if text == "" {
		return 0, nil
	}
	return strconv.ParseInt(text, 10, 64)
}

// Game represents game details from CheapShark
type Game struct {
	GameID            string  `json:"gameID"`
	Title             string  `json:"title"`
	SteamAppID        string  `json:"steamAppID"`
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
