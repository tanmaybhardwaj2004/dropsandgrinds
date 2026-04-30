package handlers

import (
	"net/http"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

var analyticsRepo *repositories.AnalyticsRepository

// SetAnalyticsRepository wires the analytics repository into HTTP handlers at startup.
func SetAnalyticsRepository(repo *repositories.AnalyticsRepository) {
	analyticsRepo = repo
}

// AnalyticsEventsHandler receives and stores analytics events from the frontend.
// @Summary      Analytics events
// @Description  Receives analytics events from the frontend for tracking
// @Tags         analytics
// @Accept       json
// @Param        request  body  models.AnalyticsEventsRequest  true  "Analytics events"
// @Success      200
// @Failure      400  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /api/analytics/events [post]
func AnalyticsEventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if analyticsRepo == nil {
		writeJSON(w, http.StatusServiceUnavailable, models.APIError{Error: "Analytics repository not initialized"})
		return
	}

	var req models.AnalyticsEventsRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid request body"})
		return
	}

	// Get user_id from context if authenticated
	var userID int64
	if uid := r.Context().Value("user_id"); uid != nil {
		userID = uid.(int64)
	}

	// Attach user ID to events
	for i := range req.Events {
		req.Events[i].UserID = userID
	}

	if err := analyticsRepo.StoreEvents(r.Context(), req.Events); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Failed to store events"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
