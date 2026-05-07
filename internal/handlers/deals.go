package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/middleware"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

var dealsService *services.DealEvaluationService
var clicksRepo *repositories.ClicksRepository

// SetClicksRepository sets the clicks repository
func SetClicksRepository(repo *repositories.ClicksRepository) {
	clicksRepo = repo
}

// SetDealsService sets the deals service
func SetDealsService(svc *services.DealEvaluationService) {
	dealsService = svc
}

// DealsListHandler handles GET /api/deals
// @Summary      List deals
// @Description  Returns paginated current active deals sorted by discount
// @Tags         deals
// @Produce      json
// @Param        limit   query  int  false  "Page size"  default(20)
// @Param        offset  query  int  false  "Page offset"  default(0)
// @Success      200     {object}  object{deals=[]models.Deal,total=int,limit=int,offset=int}
// @Failure      500     {object}  models.APIError
// @Router       /api/deals [get]
func DealsListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}

	if dealsService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Deals service not initialized"})
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Trigger optional live refresh in the background path (repository has Redis gate).
	deals, total, err := dealsService.ListDeals(r.Context(), limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Failed to fetch deals"})
		return
	}

	response := struct {
		Deals  []models.Deal `json:"deals"`
		Total  int           `json:"total"`
		Limit  int           `json:"limit"`
		Offset int           `json:"offset"`
	}{
		Deals:  deals,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	writeJSON(w, http.StatusOK, response)
}

// DealsForYouHandler handles GET /api/deals/for-you
// @Summary      Personalized deals
// @Description  If authenticated, returns deals filtered by user's wishlist and click history. If not authenticated, returns top deals.
// @Tags         deals
// @Produce      json
// @Param        limit   query  int  false  "Page size"  default(20)
// @Param        offset  query  int  false  "Page offset"  default(0)
// @Success      200     {object}  object{deals=[]models.Deal,total=int,limit=int,offset=int}
// @Failure      500     {object}  models.APIError
// @Router       /api/deals/for-you [get]
func DealsForYouHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}

	if dealsService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Deals service not initialized"})
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	var (
		deals []models.Deal
		total int
		err   error
	)
	if ok {
		deals, total, err = dealsService.GetDealsForYou(r.Context(), userID, limit, offset)
	} else {
		deals, total, err = dealsService.ListDeals(r.Context(), limit, offset)
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Failed to fetch deals"})
		return
	}

	response := struct {
		Deals  []models.Deal `json:"deals"`
		Total  int           `json:"total"`
		Limit  int           `json:"limit"`
		Offset int           `json:"offset"`
	}{
		Deals:  deals,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	writeJSON(w, http.StatusOK, response)
}

// GameRedirectHandler handles GET /api/games/{id}/redirect?platform=X
// @Summary      Get store redirect URL
// @Description  Returns the store URL for a game on a specific platform and logs the click
// @Tags         games
// @Produce      json
// @Param        id       path  int     true  "Game ID"
// @Param        platform query string false "Platform" default(steam)
// @Success      200      {object}  object{url=string}
// @Failure      400      {object}  models.APIError
// @Router       /api/games/{id}/redirect [get]
func GameRedirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}

	// Extract game ID from URL
	const prefix = "/api/games/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid games path"})
		return
	}

	pathParts := strings.Split(r.URL.Path[len(prefix):], "/")
	if len(pathParts) < 1 {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid game id"})
		return
	}

	gameID, err := strconv.ParseInt(pathParts[0], 10, 64)
	if err != nil || gameID <= 0 {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid game id"})
		return
	}

	// Get platform from query parameter
	platform := r.URL.Query().Get("platform")
	if platform == "" {
		platform = "steam" // Default to Steam
	}

	// Log click analytics if user has consented
	userID, hasUser := middleware.UserIDFromContext(r.Context())
	if hasUser && clicksRepo != nil {
		// Check if user has consented to analytics
		consent, err := clicksRepo.GetUserConsentAnalytics(r.Context(), userID)
		if err == nil && consent {
			// Log the click
			_ = clicksRepo.LogClick(r.Context(), userID, gameID, platform)
		}
	}

	storeURL, ok, err := dealsService.StoreURL(r.Context(), gameID, platform)
	if err != nil || !ok {
		writeJSON(w, http.StatusNotFound, models.APIError{Error: "Store URL not found"})
		return
	}

	response := struct {
		URL string `json:"url"`
	}{
		URL: storeURL,
	}

	writeJSON(w, http.StatusOK, response)
}
