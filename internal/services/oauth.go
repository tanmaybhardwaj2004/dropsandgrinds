package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthService struct {
	googleConfig *oauth2.Config
	authService  *AuthService
}

func NewOAuthService(googleClientID, googleClientSecret, redirectURL string, authService *AuthService) *OAuthService {
	googleConfig := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}

	return &OAuthService{
		googleConfig: googleConfig,
		authService:  authService,
	}
}

func (s *OAuthService) GoogleAuthURL(state string) string {
	return s.googleConfig.AuthCodeURL(state)
}

func (s *OAuthService) ExchangeGoogleToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.googleConfig.Exchange(ctx, code)
}

func (s *OAuthService) GetGoogleUserInfo(ctx context.Context, token *oauth2.Token) (map[string]interface{}, error) {
	// In a real implementation, you would use the token to fetch user info from Google's API
	// For now, we'll create a placeholder implementation
	// This would typically involve making an HTTP request to https://www.googleapis.com/oauth2/v3/userinfo

	// Placeholder - in production, fetch actual user data from Google
	user := map[string]interface{}{
		"email":      "user@example.com",
		"name":       "Google User",
		"avatar_url": "https://example.com/avatar.png",
	}

	return user, nil
}

func (s *OAuthService) GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
