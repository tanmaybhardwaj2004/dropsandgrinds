package handlers

import (
	"fmt"
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
// @Success      200    {object}  models.TokenResponse
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

	// Exchange code for token
	token, err := oauthService.ExchangeGoogleToken(r.Context(), code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Failed to exchange token"})
		return
	}

	// Get user info from Google
	userInfo, err := oauthService.GetGoogleUserInfo(r.Context(), token)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Failed to get user info"})
		return
	}

	// Extract user information
	email, _ := userInfo["email"].(string)
	name, _ := userInfo["name"].(string)

	if email == "" {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Email is required from Google"})
		return
	}

	// Create or update user (this would require user repository - for now return user info)
	response := map[string]interface{}{
		"email": email,
		"name":  name,
		"token": token.AccessToken,
	}

	writeJSON(w, http.StatusOK, response)
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
// @Success      200  {object}  models.TokenResponse
// @Failure      400  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /auth/steam/callback [get]
func SteamCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}

	// Get OpenID response parameters
	openidMode := r.URL.Query().Get("openid.mode")
	openidIdentity := r.URL.Query().Get("openid.identity")

	if openidMode != "id_res" {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid OpenID response"})
		return
	}

	if openidIdentity == "" {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Steam identity not provided"})
		return
	}

	// Extract Steam ID from identity URL
	// Identity URL format: https://steamcommunity.com/openid/id/76561198000000000
	var steamID string
	fmt.Sscanf(openidIdentity, "https://steamcommunity.com/openid/id/%s", &steamID)

	if steamID == "" {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Failed to extract Steam ID"})
		return
	}

	// Validate the OpenID response with Steam API
	// This requires making a POST request to Steam's validation endpoint
	// For now, return the Steam ID
	response := map[string]interface{}{
		"steam_id": steamID,
		"message":  "Steam OAuth successful - user validation pending",
	}

	writeJSON(w, http.StatusOK, response)
}
