package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"os"
	"strconv"
	"strings"

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
	candidates, err := s.wishlistRepo.ListTriggeredAlertCandidates(ctx)
	if err != nil {
		return err
	}
	for _, c := range candidates {
		alertType := "threshold"
		if c.IsHistoricalLow {
			alertType = "all_time_low"
		}
		if err := s.SendEmailAlert(ctx, TriggeredAlert{
			UserID:       c.UserID,
			GameID:       c.GameID,
			GameTitle:    c.GameTitle,
			TargetPrice:  c.TargetPriceINR,
			CurrentPrice: c.CurrentPriceINR,
			AlertType:    alertType,
		}, c.UserEmail); err != nil {
			s.logger.Error("failed to send price alert", "user_id", c.UserID, "game_id", c.GameID, "error", err)
		}
	}
	s.logger.Info("Price alert check completed", "candidates", len(candidates))
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
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = username
	}
	if host == "" || port == "" || username == "" || password == "" || from == "" {
		return fmt.Errorf("smtp configuration is incomplete")
	}
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("invalid SMTP_PORT")
	}
	subject := "DropsAndGrinds price alert"
	body := fmt.Sprintf("%s is now ₹%d. Your target was ₹%d.", alert.GameTitle, alert.CurrentPrice, alert.TargetPrice)
	msg := strings.Join([]string{
		"From: " + from,
		"To: " + userEmail,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")
	auth := smtp.PlainAuth("", username, password, host)
	return smtp.SendMail(host+":"+port, auth, from, []string{userEmail}, []byte(msg))
}

// GetUserConsentAlerts checks if a user has consented to email alerts
func (s *AlertService) GetUserConsentAlerts(ctx context.Context, userID int64) (bool, error) {
	return false, fmt.Errorf("consent lookup is handled by alert candidate query")
}

// SetUserConsentAlerts sets a user's consent for email alerts
func (s *AlertService) SetUserConsentAlerts(ctx context.Context, userID int64, consent bool) error {
	return fmt.Errorf("consent updates are handled by user settings")
}

// CheckPriceThreshold checks if a price meets the threshold for an alert
func (s *AlertService) CheckPriceThreshold(currentPrice, targetPrice int) bool {
	return currentPrice > 0 && currentPrice <= targetPrice
}

// CheckAllTimeLow checks if a price is at an all-time low
func (s *AlertService) CheckAllTimeLow(ctx context.Context, gameID int64, currentPrice int) (bool, error) {
	return false, fmt.Errorf("all-time-low checks are stored with price history")
}
