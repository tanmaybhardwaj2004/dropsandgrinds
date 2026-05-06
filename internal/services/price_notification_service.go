package services

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

// PriceNotificationService handles price drop notifications for wishlist items
type PriceNotificationService struct {
	alertRepo    *repositories.DealAlertRepository
	catalogRepo  *repositories.EnhancedCatalogRepository
	wishlistRepo *repositories.WishlistRepository
	userRepo     *repositories.UserRepository
	emailService *EmailService
	logger       *slog.Logger
}

// NewPriceNotificationService creates a new price notification service
func NewPriceNotificationService(
	alertRepo *repositories.DealAlertRepository,
	catalogRepo *repositories.EnhancedCatalogRepository,
	wishlistRepo *repositories.WishlistRepository,
	userRepo *repositories.UserRepository,
	emailService *EmailService,
	logger *slog.Logger,
) *PriceNotificationService {
	return &PriceNotificationService{
		alertRepo:    alertRepo,
		catalogRepo:  catalogRepo,
		wishlistRepo: wishlistRepo,
		userRepo:     userRepo,
		emailService: emailService,
		logger:       logger,
	}
}

// CheckPriceDrops checks all active deal alerts and triggers notifications if target prices are reached
func (p *PriceNotificationService) CheckPriceDrops(ctx context.Context) ([]models.PriceDropNotification, error) {
	// Get all active deal alerts that haven't been notified yet
	alerts, err := p.alertRepo.GetActiveAlerts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active alerts: %w", err)
	}

	var notifications []models.PriceDropNotification

	for _, alert := range alerts {
		// Get current prices for the game
		compResult, err := p.catalogRepo.GetGameWithPriceComparison(ctx, alert.GameID, alert.Region)
		if err != nil {
			log.Printf("Failed to get prices for game %d: %v", alert.GameID, err)
			continue
		}
		prices := compResult.Prices

		// Check if any price is below the target price
		for _, price := range prices {
			if price.PriceINR <= alert.TargetPrice {
				// Price drop detected
				notification := models.PriceDropNotification{
					AlertID:      alert.ID,
					UserID:       alert.UserID,
					GameID:       alert.GameID,
					TargetPrice:  alert.TargetPrice,
					CurrentPrice: price.PriceINR,
					StoreID:      price.StoreID,
					Discount:     price.DiscountPercent,
				}
				notifications = append(notifications, notification)

				// Send email notification if email service is configured
				if p.emailService != nil && p.userRepo != nil {
					// Get user email from repository
					userEmail, err := p.userRepo.GetUserEmail(ctx, alert.UserID)
					if err != nil {
						p.logger.Error("failed to get user email for notification",
							"error", err, "user_id", alert.UserID, "alert_id", alert.ID)
						continue
					}

					// Send email notification
					err = p.emailService.SendPriceDropNotification(ctx, notification, userEmail)
					if err != nil {
						p.logger.Error("failed to send price drop email",
							"error", err, "user_id", alert.UserID, "alert_id", alert.ID)
					} else {
						p.logger.Info("price drop email sent successfully",
							"user_id", alert.UserID, "alert_id", alert.ID, "user_email", userEmail)
					}
				}
				break // Only notify once per alert, even if multiple stores have good prices
			}
		}
	}

	return notifications, nil
}

// CreateDealAlert creates a new deal alert for a game
func (p *PriceNotificationService) CreateDealAlert(ctx context.Context, userID, gameID int64, targetPrice float64, storeID int64, region string) (*models.DealAlert, error) {
	var sID *int64
	if storeID > 0 {
		sID = &storeID
	}

	alert := &models.DealAlert{
		UserID:      userID,
		GameID:      gameID,
		TargetPrice: targetPrice,
		StoreID:     sID,
		Region:      region,
		Currency:    "INR",
		IsActive:    true,
	}

	createdAlert, err := p.alertRepo.Create(ctx, alert)
	if err != nil {
		return nil, fmt.Errorf("failed to create deal alert: %w", err)
	}

	return createdAlert, nil
}

// GetUserAlerts gets all deal alerts for a user
func (p *PriceNotificationService) GetUserAlerts(ctx context.Context, userID int64) ([]models.DealAlert, error) {
	return p.alertRepo.GetByUserID(ctx, userID)
}

// DeleteAlert deletes a deal alert
func (p *PriceNotificationService) DeleteAlert(ctx context.Context, alertID, userID int64) error {
	return p.alertRepo.Delete(ctx, alertID, userID)
}

// UpdateAlertTarget updates the target price for a deal alert
func (p *PriceNotificationService) UpdateAlertTarget(ctx context.Context, alertID, userID int64, newTargetPrice float64) error {
	return p.alertRepo.UpdateTargetPrice(ctx, alertID, userID, newTargetPrice)
}
