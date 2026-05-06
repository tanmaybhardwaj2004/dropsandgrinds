package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

var bundleService *services.BundleService
var buyTimingService *services.BuyTimingService

// SetBundleService sets the bundle service
func SetBundleService(svc *services.BundleService) {
	bundleService = svc
}

// SetBuyTimingService sets the buy timing service
func SetBuyTimingService(svc *services.BuyTimingService) {
	buyTimingService = svc
}

// BundleAnalyzeHandler handles POST /api/bundles/analyze
// @Summary      Analyze bundle
// @Description  Analyzes a bundle URL and provides buy recommendation
// @Tags         bundles
// @Accept       json
// @Produce      json
// @Param        request  body      object{url=string,bundle_price_inr=number}  true  "Bundle analysis request"
// @Success      200      {object}  services.BundleAnalysis
// @Failure      400      {object}  models.APIError
// @Failure      500      {object}  models.APIError
// @Router       /api/bundles/analyze [post]
func BundleAnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if bundleService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Bundle service not initialized"})
		return
	}

	var req struct {
		URL         string  `json:"url"`
		BundlePrice float64 `json:"bundle_price_inr"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid request body"})
		return
	}

	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "URL is required"})
		return
	}
	if req.BundlePrice <= 0 {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Bundle price must be positive"})
		return
	}

	if !bundleService.ValidateBundleURL(req.URL) {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Unsupported bundle URL. Only Humble, Fanatical, and Steam bundles are supported"})
		return
	}

	analysis, err := bundleService.AnalyzeBundle(r.Context(), req.URL, req.BundlePrice)
	if err != nil {
		writeServiceError(w, err, "Failed to analyze bundle")
		return
	}

	writeJSON(w, http.StatusOK, analysis)
}

// BuyTimingHandler handles GET /api/games/{id}/buy-timing
// @Summary      Get buy timing recommendation
// @Description  Returns buy timing recommendation for a game based on sale calendar
// @Tags         games
// @Produce      json
// @Param        id   path      int  true  "Game ID"
// @Success      200  {object}  services.BuyTimingRecommendation
// @Failure      400  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /api/games/{id}/buy-timing [get]
func BuyTimingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if buyTimingService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Buy timing service not initialized"})
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid path"})
		return
	}

	gameID, err := strconv.ParseInt(pathParts[2], 10, 64)
	if err != nil || gameID <= 0 {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid game id"})
		return
	}

	recommendation, err := buyTimingService.GetBuyTiming(r.Context(), gameID)
	if err != nil {
		writeServiceError(w, err, "Failed to get buy timing")
		return
	}

	writeJSON(w, http.StatusOK, recommendation)
}

// ActiveSalesHandler handles GET /api/sales/active
// @Summary      Get active sales
// @Description  Returns currently active sales across all platforms
// @Tags         sales
// @Produce      json
// @Success      200  {object}  []repositories.SaleEvent
// @Failure      500  {object}  models.APIError
// @Router       /api/sales/active [get]
func ActiveSalesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if buyTimingService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Buy timing service not initialized"})
		return
	}

	sales, err := buyTimingService.GetActiveSales(r.Context())
	if err != nil {
		// Return empty array instead of 500 error if database table doesn't exist
		var emptySales []repositories.SaleEvent
		writeJSON(w, http.StatusOK, emptySales)
		return
	}

	writeJSON(w, http.StatusOK, sales)
}

// SalesCalendarHandler handles GET /api/sales/calendar
// @Summary      Get sales calendar
// @Description  Returns full sales calendar
// @Tags         sales
// @Produce      json
// @Success      200  {object}  []repositories.SaleEvent
// @Failure      500  {object}  models.APIError
// @Router       /api/sales/calendar [get]
func SalesCalendarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if buyTimingService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Buy timing service not initialized"})
		return
	}

	calendar, err := buyTimingService.GetSalesCalendar(r.Context())
	if err != nil {
		writeServiceError(w, err, "Failed to get sales calendar")
		return
	}

	writeJSON(w, http.StatusOK, calendar)
}
