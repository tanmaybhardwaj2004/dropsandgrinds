package models

import "time"

type User struct {
	ID            int64     `json:"id" example:"1"`
	Email         string    `json:"email" example:"player@example.com"`
	Username      string    `json:"username" example:"epicgamer"`
	FirstName     string    `json:"first_name" example:"John"`
	LastName      string    `json:"last_name" example:"Doe"`
	IsActive      bool      `json:"is_active" example:"true"`
	EmailVerified bool      `json:"email_verified" example:"true"`
	CreatedAt     time.Time `json:"created_at" example:"2026-01-01T00:00:00Z"`
	UpdatedAt     time.Time `json:"updated_at" example:"2026-01-01T00:00:00Z"`
}

type RegisterRequest struct {
	Username         string `json:"username" binding:"required,min=3,max=50" example:"epicgamer"`
	Email            string `json:"email" binding:"required,email" example:"player@example.com"`
	Password         string `json:"password" binding:"required,min=8" example:"strongpassword123"`
	SteamID          string `json:"steam_id" example:"76561198000000000"`
	ConsentAnalytics bool   `json:"consent_analytics" example:"true"`
	ConsentAlerts    bool   `json:"consent_alerts" example:"true"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"player@example.com"`
	Password string `json:"password" binding:"required" example:"strongpassword123"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5c..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	UserID       int64  `json:"user_id" example:"1"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIs..."`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIs..."`
}

type APIError struct {
	Error string `json:"error" example:"Invalid credentials"`
}

type LibraryImportRequest struct {
	SteamID          string `json:"steam_id" binding:"required" example:"76561198000000000"`
	ConsentAnalytics bool   `json:"consent_analytics" example:"true"`
}

type LibraryListResponse struct {
	GameIDs []int64 `json:"game_ids"`
	Count   int     `json:"count" example:"5"`
}

type LibraryDLCResponse struct {
	DLCs  []Game `json:"dlcs"`
	Count int    `json:"count" example:"3"`
}

type ConsentUpdateRequest struct {
	ConsentAnalytics bool `json:"consent_analytics"`
	ConsentAlerts    bool `json:"consent_alerts"`
}
