package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 12
)

type ServiceError struct {
	StatusCode int
	Message    string
}

func (e *ServiceError) Error() string {
	return e.Message
}

type AuthService struct {
	db              *pgxpool.Pool
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type AuthServiceConfig struct {
	JWTSecret             string
	AccessTokenTTLMinutes int
	RefreshTokenTTLHours  int
}

func NewAuthService(db *pgxpool.Pool, cfg AuthServiceConfig) (*AuthService, error) {
	if db == nil {
		return nil, errors.New("db connection is required")
	}
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return nil, errors.New("jwt secret is required")
	}
	if cfg.AccessTokenTTLMinutes <= 0 {
		cfg.AccessTokenTTLMinutes = 15
	}
	if cfg.RefreshTokenTTLHours <= 0 {
		cfg.RefreshTokenTTLHours = 168
	}

	return &AuthService{
		db:              db,
		jwtSecret:       []byte(cfg.JWTSecret),
		accessTokenTTL:  time.Duration(cfg.AccessTokenTTLMinutes) * time.Minute,
		refreshTokenTTL: time.Duration(cfg.RefreshTokenTTLHours) * time.Hour,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (models.TokenResponse, error) {
	username := strings.TrimSpace(req.Username)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)
	steamID := strings.TrimSpace(req.SteamID)

	if len(username) < 3 || len(username) > 50 {
		return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusBadRequest, Message: "Username must be between 3 and 50 characters"}
	}
	if email == "" || !strings.Contains(email, "@") {
		return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusBadRequest, Message: "Valid email is required"}
	}
	if len(password) < 8 {
		return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusBadRequest, Message: "Password must be at least 8 characters"}
	}

	passwordHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to hash password"}
	}

	var userID int64
	query := `
		INSERT INTO users (username, email, password_hash, steam_id, consent_analytics, consent_alerts)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6)
		RETURNING id
	`
	err = s.db.QueryRow(ctx, query, username, email, string(passwordHashBytes), steamID, req.ConsentAnalytics, req.ConsentAlerts).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusConflict, Message: "Username or email already exists"}
		}
		return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to create user"}
	}

	return s.issueTokenPair(ctx, userID)
}

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (models.TokenResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)
	if email == "" || password == "" {
		return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusBadRequest, Message: "Email and password are required"}
	}

	var userID int64
	var passwordHash string
	err := s.db.QueryRow(ctx, `SELECT id, password_hash FROM users WHERE email = $1`, email).Scan(&userID, &passwordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusUnauthorized, Message: "Invalid credentials"}
		}
		return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to process login"}
	}

	if err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusUnauthorized, Message: "Invalid credentials"}
	}

	return s.issueTokenPair(ctx, userID)
}

func (s *AuthService) Refresh(ctx context.Context, req models.RefreshRequest) (models.TokenResponse, error) {
	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusBadRequest, Message: "Refresh token is required"}
	}

	tokenHash := hashToken(refreshToken)
	var userID int64
	var expiresAt time.Time
	err := s.db.QueryRow(ctx, `
		SELECT user_id, expires_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash).Scan(&userID, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusUnauthorized, Message: "Invalid refresh token"}
		}
		return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to validate refresh token"}
	}

	if expiresAt.Before(time.Now().UTC()) {
		_, _ = s.db.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1`, tokenHash)
		return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusUnauthorized, Message: "Refresh token expired"}
	}

	_, err = s.db.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to rotate refresh token"}
	}

	return s.issueTokenPair(ctx, userID)
}

func (s *AuthService) Logout(ctx context.Context, req models.LogoutRequest) error {
	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		return &ServiceError{StatusCode: http.StatusBadRequest, Message: "Refresh token is required"}
	}

	_, err := s.db.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL`, hashToken(refreshToken))
	if err != nil {
		return &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to revoke refresh token"}
	}
	return nil
}

func (s *AuthService) issueTokenPair(ctx context.Context, userID int64) (models.TokenResponse, error) {
	accessToken, err := s.generateAccessToken(userID)
	if err != nil {
		return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to generate access token"}
	}

	refreshToken, err := generateOpaqueToken()
	if err != nil {
		return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to generate refresh token"}
	}
	refreshExpiry := time.Now().UTC().Add(s.refreshTokenTTL)

	_, err = s.db.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, hashToken(refreshToken), refreshExpiry)
	if err != nil {
		return models.TokenResponse{}, &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to persist refresh token"}
	}

	return models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       userID,
	}, nil
}

func (s *AuthService) generateAccessToken(userID int64) (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub": fmt.Sprintf("%d", userID),
		"iat": now.Unix(),
		"exp": now.Add(s.accessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func generateOpaqueToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
