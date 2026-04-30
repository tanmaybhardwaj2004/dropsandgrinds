package steam

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
	// Steam Web API base URL
	baseURL = "https://api.steampowered.com/IPlayerService/GetOwnedGames/v0001/"
	// Timeout for Steam API calls
	apiTimeout = 30 * time.Second
)

// Client is a Steam Web API client
type Client struct {
	apiKey    string
	httpClient *http.Client
}

// NewClient creates a new Steam API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: apiTimeout,
		},
	}
}

// OwnedGame represents a game from Steam GetOwnedGames response
type OwnedGame struct {
	AppID           int64  `json:"appid"`
	Name            string `json:"name"`
	PlaytimeForever int    `json:"playtime_forever"`
	Playtime2Weeks  int    `json:"playtime_2weeks"`
	ImgIconURL      string `json:"img_icon_url"`
	ImgLogoURL      string `json:"img_logo_url"`
	HasCommunityStats bool `json:"has_community_visible_stats"`
}

// GetOwnedGamesResponse is the response from Steam GetOwnedGames API
type GetOwnedGamesResponse struct {
	Response struct {
		GameCount int          `json:"game_count"`
		Games    []OwnedGame `json:"games"`
	} `json:"response"`
}

// GetOwnedGames fetches the list of owned games for a SteamID
// steamID should be the 64-bit Steam ID (e.g., 76561198000000000)
func (c *Client) GetOwnedGames(ctx context.Context, steamID string) ([]OwnedGame, error) {
	// Build request URL
	reqURL := fmt.Sprintf("%s?key=%s&steamid=%s&format=json&include_appinfo=true&include_played_free_games=true",
		baseURL, c.apiKey, steamID)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Steam API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result GetOwnedGamesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Response.Games, nil
}

// ValidateSteamID checks if a SteamID is a valid 64-bit Steam ID
func ValidateSteamID(steamID string) bool {
	// Steam IDs are 64-bit integers starting with 76561197960265728
	// This is a basic validation - actual validation would require more complex logic
	id, err := strconv.ParseInt(steamID, 10, 64)
	if err != nil {
		return false
	}
	
	// Steam64 IDs start at 76561197960265728
	const minSteamID = 76561197960265728
	return id >= minSteamID
}

// ResolveVanityURL converts a Steam vanity URL (custom username) to Steam64 ID
// This is a placeholder - actual implementation would call ISteamUser/ResolveVanityURL
func (c *Client) ResolveVanityURL(ctx context.Context, vanityURL string) (string, error) {
	// Placeholder implementation
	// In production, this would call: https://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001/
	return "", fmt.Errorf("vanity URL resolution not implemented")
}

// GetPlayerSummaries fetches player profiles (for getting SteamID from vanity URL)
// This is a placeholder for future implementation
func (c *Client) GetPlayerSummaries(ctx context.Context, steamIDs []string) (map[string]interface{}, error) {
	// Placeholder implementation
	// In production, this would call: https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/
	return nil, fmt.Errorf("player summaries not implemented")
}
