package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

var reviewService *services.ReviewService

// SetReviewService sets the review service
func SetReviewService(svc *services.ReviewService) {
	reviewService = svc
}

// ReviewHandler handles GET /api/games/{id}/reviews
// @Summary      Get aggregated review score
// @Description  Returns weighted average review score from multiple sources with per-source breakdown
// @Tags         reviews
// @Produce      json
// @Param        id   path  int  true  "Game ID"
// @Success      200  {object}  services.AggregatedReview
// @Failure      400  {object}  models.APIError
// @Failure      404  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /api/games/{id}/reviews [get]
func ReviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if reviewService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Review service not initialized"})
		return
	}

	const prefix = "/api/games/"
	const suffix = "/reviews"
	if !strings.HasPrefix(r.URL.Path, prefix) || !strings.HasSuffix(r.URL.Path, suffix) {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid reviews path"})
		return
	}

	middle := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, prefix), suffix)
	gameID, err := strconv.ParseInt(strings.Trim(middle, "/"), 10, 64)
	if err != nil || gameID <= 0 {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid game id"})
		return
	}

	response, err := reviewService.GetAggregatedReview(r.Context(), gameID)
	if err != nil {
		writeServiceError(w, err, "Failed to fetch review scores")
		return
	}

	writeJSON(w, http.StatusOK, response)
}
