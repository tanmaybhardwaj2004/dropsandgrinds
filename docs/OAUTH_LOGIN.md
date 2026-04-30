# OAuth Login Implementation Guide

## Overview
Implement OAuth 2.0 authentication using Google and Steam providers for seamless user login.

## Supported Providers

### 1. Google OAuth
- User-friendly login with Google account
- Access to user profile information
- Widely adopted and trusted

### 2. Steam OAuth
- Login with Steam account
- Access to Steam library and profile
- Natural fit for gaming platform

## Google OAuth Setup

### 1. Create Google Cloud Project
1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create a new project or select existing
3. Enable Google+ API and Google Identity Platform

### 2. Create OAuth 2.0 Credentials
1. Navigate to APIs & Services → Credentials
2. Create OAuth 2.0 Client ID
3. Application type: Web application
4. Authorized redirect URIs:
   - `https://dropsandgrinds.com/auth/google/callback`
   - `http://localhost:8080/auth/google/callback` (for development)

### 3. Get Credentials
Save the Client ID and Client Secret for configuration.

## Steam OAuth Setup

### 1. Register Application
1. Go to [Steam Web API Key](https://steamcommunity.com/dev/apikey)
2. Register your domain
3. Get API Key

### 2. Steam OpenID
Steam uses OpenID 2.0 for authentication. No additional setup required beyond API key.

## Backend Implementation

### 1. Add OAuth Dependencies
```bash
go get golang.org/x/oauth2
go get golang.org/x/oauth2/google
go get github.com/nicklaw5/helix
```

### 2. Create OAuth Service
```go
package services

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

type OAuthService struct {
    googleConfig *oauth2.Config
    steamAPIKey  string
    userRepo     UserRepository
}

func NewOAuthService(googleClientID, googleClientSecret, steamAPIKey string, userRepo UserRepository) *OAuthService {
    googleConfig := &oauth2.Config{
        ClientID:     googleClientID,
        ClientSecret: googleClientSecret,
        RedirectURL:  "https://dropsandgrinds.com/auth/google/callback",
        Scopes:       []string{"openid", "profile", "email"},
        Endpoint:     google.Endpoint,
    }
    
    return &OAuthService{
        googleConfig: googleConfig,
        steamAPIKey:  steamAPIKey,
        userRepo:     userRepo,
    }
}

func (s *OAuthService) GoogleAuthURL(state string) string {
    return s.googleConfig.AuthCodeURL(state)
}

func (s *OAuthService) ExchangeGoogleToken(ctx context.Context, code string) (*oauth2.Token, error) {
    return s.googleConfig.Exchange(ctx, code)
}

func (s *OAuthService) GetGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*models.User, error) {
    // Implement Google user info retrieval
    // Create or update user in database
    return &models.User{}, nil
}
```

### 3. Create OAuth Handlers
```go
package handlers

var oauthService *services.OAuthService

func SetOAuthService(svc *services.OAuthService) {
    oauthService = svc
}

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
    state := generateState()
    // Store state in session/Redis
    
    url := oauthService.GoogleAuthURL(state)
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    state := r.URL.Query().Get("state")
    
    // Validate state
    
    token, err := oauthService.ExchangeGoogleToken(r.Context(), code)
    if err != nil {
        // Handle error
        return
    }
    
    user, err := oauthService.GetGoogleUserInfo(r.Context(), token)
    if err != nil {
        // Handle error
        return
    }
    
    // Generate JWT and set cookie
    jwtToken := generateJWT(user.ID)
    setAuthCookie(w, jwtToken)
    
    http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func SteamLoginHandler(w http.ResponseWriter, r *http.Request) {
    // Implement Steam OpenID flow
    steamURL := "https://steamcommunity.com/openid/login"
    // Redirect to Steam
}

func SteamCallbackHandler(w http.ResponseWriter, r *http.Request) {
    // Validate Steam response
    // Get user info from Steam API
    // Create/update user
    // Generate JWT
}
```

### 4. Add Routes
```go
mux.HandleFunc("/auth/google", handlers.GoogleLoginHandler)
mux.HandleFunc("/auth/google/callback", handlers.GoogleCallbackHandler)
mux.HandleFunc("/auth/steam", handlers.SteamLoginHandler)
mux.HandleFunc("/auth/steam/callback", handlers.SteamCallbackHandler)
```

## Frontend Implementation

### 1. Add OAuth Buttons
```html
<div class="oauth-buttons">
    <button onclick="loginWithGoogle()" class="btn btn-google">
        <img src="images/google-logo.svg" alt="Google">
        Continue with Google
    </button>
    <button onclick="loginWithSteam()" class="btn btn-steam">
        <img src="images/steam-logo.svg" alt="Steam">
        Sign in with Steam
    </button>
</div>
```

### 2. Add JavaScript
```javascript
function loginWithGoogle() {
    window.location.href = '/auth/google';
}

function loginWithSteam() {
    window.location.href = '/auth/steam';
}
```

## Environment Variables
```
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
STEAM_API_KEY=your-steam-api-key
OAUTH_STATE_SECRET=your-state-secret
```

## Security Considerations

1. **State Parameter**: Use cryptographically random state to prevent CSRF
2. **HTTPS Only**: OAuth requires HTTPS in production
3. **Secret Management**: Store OAuth secrets in AWS Secrets Manager
4. **Token Storage**: Store access tokens securely if needed
5. **Token Refresh**: Implement token refresh for long-lived sessions

## User Data Mapping

### Google User Data
- Email (unique identifier)
- Name (display name)
- Picture (avatar URL)
- Google ID (for linking accounts)

### Steam User Data
- Steam ID (unique identifier)
- Persona name (display name)
- Avatar URL
- Profile URL

## Troubleshooting

### Invalid OAuth State
- Ensure state is generated and validated
- Check state storage (session/Redis)
- Verify state is not expired

### Token Exchange Failed
- Verify redirect URI matches Google Console
- Check Client ID and Secret are correct
- Ensure authorization code is valid

### Steam OpenID Failures
- Verify Steam API key is valid
- Check realm matches registered domain
- Ensure OpenID parameters are correct
