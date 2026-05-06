package handlers

import (
	"net/http"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/middleware"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

// MeHandler returns the authenticated user identifier.
// @Summary      Current user
// @Description  Returns the current authenticated user ID from the JWT token
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  models.APIError
// @Router       /api/me [get]
func MeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, models.APIError{Error: "Unauthorized"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"status":  "authenticated",
	})
}

// ConsentHandler updates the authenticated user's consent flags.
// @Summary      Update consent
// @Description  Updates analytics and alert consent flags for the authenticated user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  models.ConsentUpdateRequest  true  "Consent flags"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  models.APIError
// @Failure      401      {object}  models.APIError
// @Failure      500      {object}  models.APIError
// @Router       /api/me/consent [post]
func ConsentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if dbPool == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Database not initialized"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, models.APIError{Error: "Unauthorized"})
		return
	}

	var req models.ConsentUpdateRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid request body"})
		return
	}

	_, err := dbPool.Exec(r.Context(), `
		UPDATE users
		SET consent_analytics = $2, consent_alerts = $3
		WHERE id = $1
	`, userID, req.ConsentAnalytics, req.ConsentAlerts)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Failed to update consent"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
