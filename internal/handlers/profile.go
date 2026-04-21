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
