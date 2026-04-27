package handlers

import (
	"net/http"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/middleware"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

var savingsService *services.SavingsService

// SetSavingsService sets the savings service
func SetSavingsService(svc *services.SavingsService) {
	savingsService = svc
}

// SavingsLogHandler handles POST /api/savings/purchase
// @Summary      Log a purchase
// @Description  Log a purchase with paid and original price to track savings
// @Tags         savings
// @Accept       json
// @Produce      json
// @Param        request  body      services.LogPurchaseRequest  true  "Purchase log request"
// @Success      201
// @Failure      400  {object}  models.APIError
// @Failure      401  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /api/savings/purchase [post]
func SavingsLogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if savingsService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Savings service not initialized"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, models.APIError{Error: "Unauthorized"})
		return
	}

	var req services.LogPurchaseRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid request body"})
		return
	}

	err := savingsService.LogPurchase(r.Context(), userID, req)
	if err != nil {
		writeServiceError(w, err, "Failed to log purchase")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// SavingsGetHandler handles GET /api/savings
// @Summary      Get savings summary
// @Description  Returns total savings, monthly breakdown, and equivalent free games message
// @Tags         savings
// @Produce      json
// @Success      200  {object}  services.SavingsResponse
// @Failure      401  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /api/savings [get]
func SavingsGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if savingsService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Savings service not initialized"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, models.APIError{Error: "Unauthorized"})
		return
	}

	response, err := savingsService.GetSavings(r.Context(), userID)
	if err != nil {
		writeServiceError(w, err, "Failed to get savings")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// SavingsHistoryHandler handles GET /api/savings/history
// @Summary      Get purchase history
// @Description  Returns paginated purchase history for the authenticated user
// @Tags         savings
// @Produce      json
// @Param        limit    query  int  false  "Page size"  default(20)
// @Param        offset   query  int  false  "Page offset"  default(0)
// @Success      200      {object}  map[string]interface{}
// @Failure      401      {object}  models.APIError
// @Failure      500      {object}  models.APIError
// @Router       /api/savings/history [get]
func SavingsHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if savingsService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Savings service not initialized"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, models.APIError{Error: "Unauthorized"})
		return
	}

	limit := parseQueryInt(r.URL.Query().Get("limit"), 20)
	offset := parseQueryInt(r.URL.Query().Get("offset"), 0)

	purchases, total, err := savingsService.GetPurchaseHistory(r.Context(), userID, limit, offset)
	if err != nil {
		writeServiceError(w, err, "Failed to get purchase history")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"purchases": purchases,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}
