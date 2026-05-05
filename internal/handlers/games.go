package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

var gamesService *services.GamesService
var meilisearchService *services.MeilisearchService

// SetGamesService wires the games service into HTTP handlers at startup.
func SetGamesService(svc *services.GamesService) {
	gamesService = svc
}

// SetMeilisearchService wires the Meilisearch service into HTTP handlers at startup.
func SetMeilisearchService(svc *services.MeilisearchService) {
	meilisearchService = svc
}

// SearchGamesHandler returns games matching search criteria with filters.
// @Summary      Search games
// @Description  Full-text search with filters for platform, price range, discount, and review score. Uses Meilisearch if configured, otherwise falls back to PostgreSQL.
// @Tags         games
// @Produce      json
// @Param        q                query  string   false  "Search query"
// @Param        platform         query  string   false  "Platform filter (steam, epic, gog)"
// @Param        min_price        query  number   false  "Minimum price in INR"
// @Param        max_price        query  number   false  "Maximum price in INR"
// @Param        min_discount     query  int      false  "Minimum discount percentage"
// @Param        max_discount     query  int      false  "Maximum discount percentage"
// @Param        min_review_score query  number   false  "Minimum review score (0-100)"
// @Param        max_review_score query  number   false  "Maximum review score (0-100)"
// @Param        limit            query  int      false  "Page size"  default(30)
// @Param        offset           query  int      false  "Page offset"  default(0)
// @Success      200              {object}  models.GameListResponse
// @Router       /api/games/search [get]
func SearchGamesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}

	query := r.URL.Query().Get("q")
	platform := r.URL.Query().Get("platform")
	minPrice := parseQueryFloat(r.URL.Query().Get("min_price"), 0)
	maxPrice := parseQueryFloat(r.URL.Query().Get("max_price"), 0)
	minDiscount := parseQueryInt(r.URL.Query().Get("min_discount"), 0)
	maxDiscount := parseQueryInt(r.URL.Query().Get("max_discount"), 0)
	minReviewScore := parseQueryFloat(r.URL.Query().Get("min_review_score"), 0)
	maxReviewScore := parseQueryFloat(r.URL.Query().Get("max_review_score"), 0)
	limit := parseQueryInt(r.URL.Query().Get("limit"), 30)
	offset := parseQueryInt(r.URL.Query().Get("offset"), 0)

	var games []models.Game
	var total int
	var err error

	// Use Meilisearch if available, otherwise fall back to PostgreSQL
	if meilisearchService != nil {
		// Build Meilisearch filter string
		filters := buildMeilisearchFilters(platform, minPrice, maxPrice, minDiscount, maxDiscount, minReviewScore, maxReviewScore)
		games, total, err = meilisearchService.SearchGames(r.Context(), query, filters, limit, offset)
	} else {
		games, total, err = gamesService.SearchGames(r.Context(), query, platform, minPrice, maxPrice, minDiscount, maxDiscount, minReviewScore, maxReviewScore, limit, offset)
	}

	if err != nil {
		writeServiceError(w, err, "Failed to search games")
		return
	}

	response := models.GameListResponse{
		Games:  games,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	writeJSON(w, http.StatusOK, response)
}

// buildMeilisearchFilters constructs a Meilisearch filter string from query parameters
func buildMeilisearchFilters(platform string, minPrice, maxPrice float64, minDiscount, maxDiscount int, minReviewScore, maxReviewScore float64) string {
	var filters []string

	if platform != "" {
		filters = append(filters, "platform = "+platform)
	}
	if minPrice > 0 {
		filters = append(filters, "price_inr >= "+formatFloat(minPrice))
	}
	if maxPrice > 0 {
		filters = append(filters, "price_inr <= "+formatFloat(maxPrice))
	}
	if minDiscount > 0 {
		filters = append(filters, "discount_percent >= "+strconv.Itoa(minDiscount))
	}
	if maxDiscount > 0 {
		filters = append(filters, "discount_percent <= "+strconv.Itoa(maxDiscount))
	}
	if minReviewScore > 0 {
		filters = append(filters, "review_score >= "+formatFloat(minReviewScore))
	}
	if maxReviewScore > 0 {
		filters = append(filters, "review_score <= "+formatFloat(maxReviewScore))
	}

	if len(filters) == 0 {
		return ""
	}
	return strings.Join(filters, " AND ")
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// GamesListHandler returns paginated games filtered by query and platform.
// @Summary      List games
// @Description  Returns a paginated deal grid with optional search and platform filtering. Can exclude owned games.
// @Tags         games
// @Produce      json
// @Param        q              query  string  false  "Search query"
// @Param        platform       query  string  false  "Platform filter"
// @Param        limit          query  int     false  "Page size"  default(20)
// @Param        offset        query  int     false  "Page offset"  default(0)
// @Param        exclude_owned  query  bool    false  "Exclude owned games"  default(false)
// @Success      200            {object}  models.GameListResponse
// @Router       /api/games [get]
func GamesListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if gamesService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Games service not initialized"})
		return
	}

	limit := parseQueryInt(r.URL.Query().Get("limit"), 20)
	offset := parseQueryInt(r.URL.Query().Get("offset"), 0)
	excludeOwned := r.URL.Query().Get("exclude_owned") == "true"

	// Get user_id from context if authenticated
	var userID int64
	if uid := r.Context().Value("user_id"); uid != nil {
		userID = uid.(int64)
	}

	response, err := gamesService.ListGames(r.Context(), services.GameFilter{
		Query:        r.URL.Query().Get("q"),
		Platform:     r.URL.Query().Get("platform"),
		Limit:        limit,
		Offset:       offset,
		ExcludeOwned: excludeOwned,
		UserID:       userID,
	})
	if err != nil {
		writeServiceError(w, err, "Failed to list games")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// GameDetailHandler returns a single game by ID.
// @Summary      Get game
// @Description  Returns a single game record by its ID
// @Tags         games
// @Produce      json
// @Param        id   path  int  true  "Game ID"
// @Success      200  {object}  models.Game
// @Failure      400  {object}  models.APIError
// @Failure      404  {object}  models.APIError
// @Router       /api/games/{id} [get]
func GameDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if strings.HasSuffix(r.URL.Path, "/buy-advice") {
		BuyAdviceHandler(w, r)
		return
	}
	if gamesService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Games service not initialized"})
		return
	}

	const prefix = "/api/games/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		writeJSON(w, http.StatusNotFound, models.APIError{Error: "Game not found"})
		return
	}

	id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, prefix), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid game id"})
		return
	}

	game, ok, err := gamesService.GetGameByID(r.Context(), id)
	if err != nil {
		writeServiceError(w, err, "Failed to get game")
		return
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, models.APIError{Error: "Game not found"})
		return
	}

	// Fetch arbitrage data if service is available
	// Note: This requires GetArbitrageService to be exposed or arbitrageService to be package-level
	// For now, skip arbitrage in game detail - it's available via /api/games/{id}/arbitrage endpoint
	// This is more efficient as it's only fetched when needed

	writeJSON(w, http.StatusOK, game)
}

func parseQueryInt(value string, fallback int) int {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseQueryFloat(value string, fallback float64) float64 {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

// PriceHistoryHandler returns price history for a game.
// @Summary      Price history
// @Description  Returns historical INR prices for a game ID
// @Tags         prices
// @Produce      json
// @Param        game_id  path  int  true  "Game ID"
// @Param        limit    query int  false "History points" default(30)
// @Success      200      {object}  models.PriceHistoryResponse
// @Failure      400      {object}  models.APIError
// @Failure      500      {object}  models.APIError
// @Router       /api/prices/{game_id}/history [get]
func PriceHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if gamesService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Games service not initialized"})
		return
	}

	const prefix = "/api/prices/"
	const suffix = "/history"
	if !strings.HasPrefix(r.URL.Path, prefix) || !strings.HasSuffix(r.URL.Path, suffix) {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid prices path"})
		return
	}

	middle := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, prefix), suffix)
	gameID, err := strconv.ParseInt(strings.Trim(middle, "/"), 10, 64)
	if err != nil || gameID <= 0 {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid game id"})
		return
	}

	limit := parseQueryInt(r.URL.Query().Get("limit"), 30)
	offset := parseQueryInt(r.URL.Query().Get("offset"), 0)
	response, err := gamesService.GetPriceHistory(r.Context(), gameID, limit, offset)
	if err != nil {
		writeServiceError(w, err, "Failed to fetch price history")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// IndiaArbitrageHandler returns India vs Global pricing comparison with GST.
// @Summary      India arbitrage
// @Description  Compares Steam India price vs Global price with GST breakdown
// @Tags         prices
// @Produce      json
// @Param        game_id  path  int  true  "Game ID"
// @Success      200      {object}  models.IndiaArbitrage
// @Failure      400      {object}  models.APIError
// @Failure      500      {object}  models.APIError
// @Router       /api/prices/{game_id}/india [get]
func IndiaArbitrageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if gamesService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Games service not initialized"})
		return
	}

	const prefix = "/api/prices/"
	const suffix = "/india"
	if !strings.HasPrefix(r.URL.Path, prefix) || !strings.HasSuffix(r.URL.Path, suffix) {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid prices path"})
		return
	}

	middle := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, prefix), suffix)
	gameID, err := strconv.ParseInt(strings.Trim(middle, "/"), 10, 64)
	if err != nil || gameID <= 0 {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid game id"})
		return
	}

	response, err := gamesService.GetIndiaArbitrage(r.Context(), gameID)
	if err != nil {
		writeServiceError(w, err, "Failed to fetch India arbitrage data")
		return
	}

	writeJSON(w, http.StatusOK, response)
}
