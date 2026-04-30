package handlers

import (
	"net/http"
	"net/url"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

var oauthService *services.OAuthService

// SetOAuthService wires the OAuth service into HTTP handlers at startup.
func SetOAuthService(svc *services.OAuthService) {
	oauthService = svc
}

// GoogleLoginHandler initiates Google OAuth login flow.
// @Summary      Google login
// @Description  Redirects to Google OAuth consent screen
// @Tags         auth
// @Success      302
// @Router       /auth/google [get]
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if oauthService == nil {
		writeJSON(w, http.StatusServiceUnavailable, models.APIError{Error: "OAuth service not configured"})
		return
	}

	state, err := oauthService.GenerateState()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Failed to generate state"})
		return
	}

	// In production, store state in Redis/Session with expiration
	// For now, we'll use a simple cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	authURL := oauthService.GoogleAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GoogleCallbackHandler handles Google OAuth callback.
// @Summary      Google callback
// @Description  Handles OAuth callback from Google and issues JWT tokens
// @Tags         auth
// @Param        code   query  string  true  "Authorization code"
// @Param        state  query  string  true  "State parameter"
// @Success      200    {object}  models.AuthResponse
// @Failure      400    {object}  models.APIError
// @Failure      500    {object}  models.APIError
// @Router       /auth/google/callback [get]
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if oauthService == nil {
		writeJSON(w, http.StatusServiceUnavailable, models.APIError{Error: "OAuth service not configured"})
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Authorization code is required"})
		return
	}

	// Validate state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != state {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid state parameter"})
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	})

	// TODO: Implement user creation/retrieval and token generation
	// This requires UserRepository which doesn't exist yet
	writeJSON(w, http.StatusNotImplemented, models.APIError{Error: "OAuth callback not fully implemented - requires user repository"})
}

// SteamLoginHandler initiates Steam OpenID login flow.
// @Summary      Steam login
// @Description  Redirects to Steam OpenID authentication
// @Tags         auth
// @Success      302
// @Router       /auth/steam [get]
func SteamLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}

	// Steam OpenID parameters
	steamURL := "https://steamcommunity.com/openid/login"
	params := url.Values{}
	params.Set("openid.ns", "http://specs.openid.net/auth/2.0")
	params.Set("openid.mode", "checkid_setup")
	params.Set("openid.return_to", "http://localhost:8080/auth/steam/callback")
	params.Set("openid.realm", "http://localhost:8080")
	params.Set("openid.identity", "http://specs.openid.net/auth/2.0/identifier_select")
	params.Set("openid.claimed_id", "http://specs.openid.net/auth/2.0/identifier_select")

	redirectURL := steamURL + "?" + params.Encode()
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// SteamCallbackHandler handles Steam OpenID callback.
// @Summary      Steam callback
// @Description  Handles OpenID callback from Steam and issues JWT tokens
// @Tags         auth
// @Success      200  {object}  models.AuthResponse
// @Failure      400  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /auth/steam/callback [get]
func SteamCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}

	// In a real implementation, validate the Steam OpenID response
	// For now, return a placeholder response
	writeJSON(w, http.StatusNotImplemented, models.APIError{Error: "Steam OAuth not fully implemented"})
}
