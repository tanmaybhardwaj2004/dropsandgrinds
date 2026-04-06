package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

// Register handles user sign-up
// @Summary      Register a new user
// @Description  Creates a new DropsAndGrinds account, saving GDPR consent and optionally a SteamID
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.RegisterRequest  true  "Registration info"
// @Success      201      {object}  models.TokenResponse
// @Failure      400      {object}  models.APIError
// @Failure      500      {object}  models.APIError
// @Router       /api/auth/register [post]
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Dev A to implement actual DB logic
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(models.APIError{Error: "Not implemented by Dev A yet"})
}

// Login handles user authentication
// @Summary      Log in to account
// @Description  Validates credentials and returns short-lived JWT + refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.LoginRequest   true  "Credentials"
// @Success      200      {object}  models.TokenResponse
// @Failure      401      {object}  models.APIError
// @Router       /api/auth/login [post]
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Dev A to implement bcrypt validation logic
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(models.APIError{Error: "Not implemented by Dev A yet"})
}

// Refresh handles token rotation
// @Summary      Refresh Access Token
// @Description  Validates a refresh token and rotates it for a new JWT pair
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.RefreshRequest   true  "Refresh Token"
// @Success      200      {object}  models.TokenResponse
// @Failure      400      {object}  models.APIError
// @Failure      401      {object}  models.APIError
// @Router       /api/auth/refresh [post]
func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	// Dev A to implement redis validation logic
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(models.APIError{Error: "Not implemented by Dev A yet"})
}

// Logout invalidates the refresh token
// @Summary      Log out
// @Description  Invalidates the current refresh token in backend storage (Redis)
// @Tags         auth
// @Produce      json
// @Success      200      {string}  string "Logged out"
// @Router       /api/auth/logout [post]
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Dev A to implement token revocation logic
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"message": "Not implemented by Dev A yet"})
}
