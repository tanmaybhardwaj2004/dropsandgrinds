package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

var authService *services.AuthService

// SetAuthService wires the auth service into HTTP handlers at startup.
func SetAuthService(svc *services.AuthService) {
	authService = svc
}

func decodeJSONBody(r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeServiceError(w http.ResponseWriter, err error, fallback string) {
	var serviceErr *services.ServiceError
	if errors.As(err, &serviceErr) {
		writeJSON(w, serviceErr.StatusCode, models.APIError{Error: serviceErr.Message})
		return
	}
	writeJSON(w, http.StatusInternalServerError, models.APIError{Error: fallback})
}

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
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if authService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Auth service not initialized"})
		return
	}

	var req models.RegisterRequest
	if err := decodeJSONBody(r, &req); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Request body is required"})
			return
		}
		if _, ok := err.(*json.SyntaxError); ok {
			writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Malformed JSON"})
			return
		}
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) {
			writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid request fields"})
			return
		}
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid request body"})
		return
	}

	response, err := authService.Register(r.Context(), req)
	if err != nil {
		writeServiceError(w, err, "Failed to register user")
		return
	}

	writeJSON(w, http.StatusCreated, response)
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
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if authService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Auth service not initialized"})
		return
	}

	var req models.LoginRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid request body"})
		return
	}

	response, err := authService.Login(r.Context(), req)
	if err != nil {
		writeServiceError(w, err, "Failed to login")
		return
	}

	writeJSON(w, http.StatusOK, response)
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
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if authService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Auth service not initialized"})
		return
	}

	var req models.RefreshRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid request body"})
		return
	}

	response, err := authService.Refresh(r.Context(), req)
	if err != nil {
		writeServiceError(w, err, "Failed to refresh token")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// Logout invalidates the refresh token
// @Summary      Log out
// @Description  Invalidates the current refresh token in backend storage (Redis)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.LogoutRequest   true  "Refresh Token"
// @Success      200      {string}  string "Logged out"
// @Failure      400      {object}  models.APIError
// @Failure      500      {object}  models.APIError
// @Router       /api/auth/logout [post]
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if authService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Auth service not initialized"})
		return
	}

	var req models.LogoutRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid request body"})
		return
	}

	if err := authService.Logout(r.Context(), req); err != nil {
		writeServiceError(w, err, "Failed to logout")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Logged out"})
}
