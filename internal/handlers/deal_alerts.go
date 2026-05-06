package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/middleware"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

var priceNotificationService *services.PriceNotificationService

// SetPriceNotificationService wires the price notification service into HTTP handlers at startup.
func SetPriceNotificationService(svc *services.PriceNotificationService) {
	priceNotificationService = svc
}

// DealAlertsCollectionHandler handles POST/GET on /api/deal-alerts.
// @Summary      Create or list deal alerts
// @Description  POST: Create a new deal alert for price drop notification | GET: List authenticated user's deal alerts
// @Tags         deal-alerts
// @Accept       json
// @Produce      json
// @Param        request  body      models.DealAlert  false  "Create request (POST only)"
// @Success      201      {object}  models.DealAlert
// @Success      200      {object}  []models.DealAlert
// @Failure      400      {object}  models.APIError
// @Failure      401      {object}  models.APIError
// @Failure      500      {object}  models.APIError
// @Router       /api/deal-alerts [post]
// @Router       /api/deal-alerts [get]
func DealAlertsCollectionHandler(w http.ResponseWriter, r *http.Request) {
	if priceNotificationService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Price notification service not initialized"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, models.APIError{Error: "Unauthorized"})
		return
	}

	switch r.Method {
	case http.MethodPost:
		gameID, err := strconv.ParseInt(r.URL.Query().Get("game_id"), 10, 64)
		if err != nil || gameID <= 0 {
			writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid game_id"})
			return
		}

		targetPrice, err := strconv.ParseFloat(r.URL.Query().Get("target_price"), 64)
		if err != nil || targetPrice <= 0 {
			writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid target_price"})
			return
		}

		var storeID int64
		if storeIDStr := r.URL.Query().Get("store_id"); storeIDStr != "" {
			storeID, err = strconv.ParseInt(storeIDStr, 10, 64)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid store_id"})
				return
			}
		}

		region := r.URL.Query().Get("region")
		if region == "" {
			region = "IN"
		}

		alert, err := priceNotificationService.CreateDealAlert(r.Context(), userID, gameID, targetPrice, storeID, region)
		if err != nil {
			writeServiceError(w, err, "Failed to create deal alert")
			return
		}
		writeJSON(w, http.StatusCreated, alert)

	case http.MethodGet:
		alerts, err := priceNotificationService.GetUserAlerts(r.Context(), userID)
		if err != nil {
			writeServiceError(w, err, "Failed to list deal alerts")
			return
		}
		writeJSON(w, http.StatusOK, alerts)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
	}
}

// DealAlertItemHandler handles DELETE/PATCH on /api/deal-alerts/{id}.
// @Summary      Update or remove deal alert
// @Description  PATCH: Update target price for deal alert | DELETE: Remove deal alert
// @Tags         deal-alerts
// @Accept       json
// @Produce      json
// @Param        id       path      int64  true  "Deal Alert ID"
// @Param        target_price query number false "New target price (PATCH only)"
// @Success      200
// @Success      204
// @Failure      400  {object}  models.APIError
// @Failure      401  {object}  models.APIError
// @Failure      404  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /api/deal-alerts/{id} [patch]
// @Router       /api/deal-alerts/{id} [delete]
func DealAlertItemHandler(w http.ResponseWriter, r *http.Request) {
	if priceNotificationService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Price notification service not initialized"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, models.APIError{Error: "Unauthorized"})
		return
	}

	const prefix = "/api/deal-alerts/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid deal alert path"})
		return
	}

	alertID, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, prefix), 10, 64)
	if err != nil || alertID <= 0 {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid deal alert id"})
		return
	}

	switch r.Method {
	case http.MethodDelete:
		err := priceNotificationService.DeleteAlert(r.Context(), alertID, userID)
		if err != nil {
			writeServiceError(w, err, "Failed to delete deal alert")
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case http.MethodPatch:
		newTargetPrice, err := strconv.ParseFloat(r.URL.Query().Get("target_price"), 64)
		if err != nil || newTargetPrice <= 0 {
			writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid target_price"})
			return
		}

		err = priceNotificationService.UpdateAlertTarget(r.Context(), alertID, userID, newTargetPrice)
		if err != nil {
			writeServiceError(w, err, "Failed to update deal alert")
			return
		}
		w.WriteHeader(http.StatusOK)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
	}
}

// CheckPriceDropsHandler manually triggers price drop check for all active alerts.
// @Summary      Check price drops
// @Description  Manually triggers price drop check for all active deal alerts (admin only)
// @Tags         deal-alerts
// @Produce      json
// @Success      200  {object}  []models.PriceDropNotification
// @Failure      401  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /api/deal-alerts/check [post]
func CheckPriceDropsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if priceNotificationService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Price notification service not initialized"})
		return
	}

	notifications, err := priceNotificationService.CheckPriceDrops(r.Context())
	if err != nil {
		writeServiceError(w, err, "Failed to check price drops")
		return
	}

	writeJSON(w, http.StatusOK, notifications)
}
