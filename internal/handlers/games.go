package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

var gamesService *services.GamesService

// SetGamesService wires the games service into HTTP handlers at startup.
func SetGamesService(svc *services.GamesService) {
	gamesService = svc
}

// GamesListHandler returns paginated games filtered by query and platform.
// @Summary      List games
// @Description  Returns a paginated deal grid with optional search and platform filtering
// @Tags         games
// @Produce      json
// @Param        q         query  string  false  "Search query"
// @Param        platform  query  string  false  "Platform filter"
// @Param        limit     query  int     false  "Page size"  default(20)
// @Param        offset    query  int     false  "Page offset"  default(0)
// @Success      200       {object}  models.GameListResponse
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
	response, err := gamesService.ListGames(r.Context(), services.GameFilter{
		Query:    r.URL.Query().Get("q"),
		Platform: r.URL.Query().Get("platform"),
		Limit:    limit,
		Offset:   offset,
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

// DealsListHandler returns current active deals.
// @Summary      List deals
// @Description  Returns paginated current active deals
// @Tags         deals
// @Produce      json
// @Param        limit   query  int  false  "Page size"  default(20)
// @Param        offset  query  int  false  "Page offset"  default(0)
// @Success      200     {object}  models.DealListResponse
// @Failure      500     {object}  models.APIError
// @Router       /api/deals [get]
func DealsListHandler(w http.ResponseWriter, r *http.Request) {
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
	response, err := gamesService.ListDeals(r.Context(), limit, offset)
	if err != nil {
		writeServiceError(w, err, "Failed to list deals")
		return
	}

	writeJSON(w, http.StatusOK, response)
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
	response, err := gamesService.GetPriceHistory(r.Context(), gameID, limit)
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
