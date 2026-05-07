package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthService struct {
	googleConfig *oauth2.Config
	authService  *AuthService
	client       *http.Client
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
		client:       &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *OAuthService) GoogleAuthURL(state string) string {
	return s.googleConfig.AuthCodeURL(state)
}

func (s *OAuthService) ExchangeGoogleToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.googleConfig.Exchange(ctx, code)
}

func (s *OAuthService) GetGoogleUserInfo(ctx context.Context, token *oauth2.Token) (map[string]interface{}, error) {
	// Fetch user info from Google's userinfo endpoint
	resp, err := s.client.Get("https://www.googleapis.com/oauth2/v3/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from Google: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return userInfo, nil
}

func (s *OAuthService) GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
