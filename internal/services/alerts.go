package services

import (
	"context"
	"log/slog"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

// AlertService handles price alert checking and email notifications
type AlertService struct {
	wishlistRepo *repositories.WishlistRepository
	catalogRepo  *repositories.CatalogRepository
	logger       *slog.Logger
}

// NewAlertService creates a new alert service
func NewAlertService(
	wishlistRepo *repositories.WishlistRepository,
	catalogRepo *repositories.CatalogRepository,
	logger *slog.Logger,
) *AlertService {
	return &AlertService{
		wishlistRepo: wishlistRepo,
		catalogRepo:  catalogRepo,
		logger:       logger,
	}
}

// CheckAndTriggerAlerts checks all wishlist items and triggers alerts for matching prices
// This should be called on every price refresh cron run
func (s *AlertService) CheckAndTriggerAlerts(ctx context.Context) error {
	s.logger.Info("Starting price alert check")

	// For MVP: Placeholder implementation
	// In production, this would:
	// 1. Query all wishlist items with consent_alerts = true
	// 2. For each item, check if current price <= target_price_inr
	// 3. Also check if price is at new all-time low
	// 4. Send email notification for triggered alerts
	// 5. Track sent alerts to avoid duplicate notifications

	s.logger.Info("Price alert check completed (MVP - no actual emails sent)")
	return nil
}

// TriggeredAlert represents an alert that was triggered
type TriggeredAlert struct {
	UserID       int64  `json:"user_id"`
	GameID       int64  `json:"game_id"`
	GameTitle    string `json:"game_title"`
	TargetPrice  int    `json:"target_price"`
	CurrentPrice int    `json:"current_price"`
	AlertType    string `json:"alert_type"` // "threshold" or "all_time_low"
}

// SendEmailAlert sends an email alert to a user
// For MVP: This is a placeholder that logs instead of sending actual emails
func (s *AlertService) SendEmailAlert(ctx context.Context, alert TriggeredAlert, userEmail string) error {
	// For MVP: Log the alert instead of sending email
	s.logger.Info("Email alert triggered (MVP - not actually sent)",
		"user_id", alert.UserID,
		"game_id", alert.GameID,
		"game_title", alert.GameTitle,
		"target_price", alert.TargetPrice,
		"current_price", alert.CurrentPrice,
		"alert_type", alert.AlertType,
		"user_email", userEmail,
	)

	// In production, this would use an email service like:
	// - SMTP library
	// - SendGrid API
	// - AWS SES
	// - Mailgun
	// etc.

	return nil
}

// GetUserConsentAlerts checks if a user has consented to email alerts
func (s *AlertService) GetUserConsentAlerts(ctx context.Context, userID int64) (bool, error) {
	// For MVP: Return true (assume consent)
	// In production, this would query the users table for consent_alerts flag
	return true, nil
}

// SetUserConsentAlerts sets a user's consent for email alerts
func (s *AlertService) SetUserConsentAlerts(ctx context.Context, userID int64, consent bool) error {
	// For MVP: Placeholder
	// In production, this would update the users table consent_alerts flag
	s.logger.Info("User consent for alerts updated",
		"user_id", userID,
		"consent", consent,
	)
	return nil
}

// CheckPriceThreshold checks if a price meets the threshold for an alert
func (s *AlertService) CheckPriceThreshold(currentPrice, targetPrice int) bool {
	return currentPrice > 0 && currentPrice <= targetPrice
}

// CheckAllTimeLow checks if a price is at an all-time low
func (s *AlertService) CheckAllTimeLow(ctx context.Context, gameID int64, currentPrice int) (bool, error) {
	// For MVP: Return false
	// In production, this would query the price history to check if current price is the lowest ever
	return false, nil
}
