package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

var arbitrageService *services.ArbitrageService

// SetArbitrageService sets the arbitrage service
func SetArbitrageService(svc *services.ArbitrageService) {
	arbitrageService = svc
}

// ArbitrageHandler handles GET /api/games/{id}/arbitrage
// @Summary      Get India vs Global arbitrage comparison
// @Description  Returns India vs Global price comparison with GST breakdown
// @Tags         arbitrage
// @Produce      json
// @Param        id   path  int  true  "Game ID"
// @Success      200  {object}  models.ArbitrageData
// @Failure      400  {object}  models.APIError
// @Failure      404  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /api/games/{id}/arbitrage [get]
func ArbitrageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if arbitrageService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Arbitrage service not initialized"})
		return
	}

	const prefix = "/api/games/"
	const suffix = "/arbitrage"
	if !strings.HasPrefix(r.URL.Path, prefix) || !strings.HasSuffix(r.URL.Path, suffix) {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid arbitrage path"})
		return
	}

	middle := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, prefix), suffix)
	gameID, err := strconv.ParseInt(strings.Trim(middle, "/"), 10, 64)
	if err != nil || gameID <= 0 {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid game id"})
		return
	}

	response, err := arbitrageService.CalculateArbitrage(r.Context(), gameID)
	if err != nil {
		writeServiceError(w, err, "Failed to calculate arbitrage")
		return
	}

	writeJSON(w, http.StatusOK, response)
}
