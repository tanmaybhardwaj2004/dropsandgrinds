package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/resend/resend-go/v2"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

// EmailService handles sending email notifications using Resend API
type EmailService struct {
	client    *resend.Client
	fromEmail string
	logger    *slog.Logger
}

// NewEmailService creates a new email service with Resend
func NewEmailService(logger *slog.Logger) *EmailService {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		logger.Warn("RESEND_API_KEY not found, email service will be in demo mode")
	}

	fromEmail := os.Getenv("FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = "noreply@dropsandgrinds.com"
	}

	var client *resend.Client
	if apiKey != "" {
		client = resend.NewClient(apiKey)
	}

	return &EmailService{
		client:    client,
		fromEmail: fromEmail,
		logger:    logger,
	}
}

// SendPriceDropNotification sends an email notification for a price drop
func (e *EmailService) SendPriceDropNotification(ctx context.Context, notification models.PriceDropNotification, userEmail string) error {
	subject := fmt.Sprintf("🎮 Price Drop Alert: Game ID %d is now ₹%.2f!", notification.GameID, notification.CurrentPrice)

	// Create HTML email template
	htmlContent := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Price Drop Alert</title>
		<style>
			body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
			.container { max-width: 600px; margin: 0 auto; padding: 20px; }
			.header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
			.content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
			.price-drop { background: #4CAF50; color: white; padding: 15px; border-radius: 5px; margin: 20px 0; text-align: center; }
			.price-details { display: flex; justify-content: space-between; margin: 20px 0; }
			.cta-button { background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; margin: 20px 0; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>🎮 Price Drop Alert!</h1>
				<p>Great news! A game on your wishlist is now available at a lower price</p>
			</div>
			<div class="content">
				<div class="price-drop">
					<h2>Game ID: %d</h2>
					<p style="font-size: 24px; margin: 10px 0;">Now only ₹%.2f</p>
					<p>You wanted it at ₹%.2f - Save ₹%.2f!</p>
				</div>
				<div class="price-details">
					<div>
						<strong>Target Price:</strong><br>
						₹%.2f
					</div>
					<div>
						<strong>Current Price:</strong><br>
						₹%.2f
					</div>
					<div>
						<strong>Discount:</strong><br>
						%d%% OFF
					</div>
				</div>
				<p>This is the perfect time to buy! The price has dropped below your target price.</p>
				<div style="text-align: center;">
					<a href="https://dropsandgrinds.com/game/%d" class="cta-button">View Deal Now</a>
				</div>
				<hr style="margin: 30px 0; border: none; border-top: 1px solid #ddd;">
				<p style="color: #666; font-size: 14px;">
					This alert was triggered because you set a price alert for this game.<br>
					<a href="https://dropsandgrinds.com/alerts/manage">Manage your alerts</a> | 
					<a href="https://dropsandgrinds.com/unsubscribe?alert=%d">Unsubscribe from this alert</a>
				</p>
			</div>
		</div>
	</body>
	</html>
	`, notification.GameID, notification.CurrentPrice, notification.TargetPrice,
		notification.TargetPrice-notification.CurrentPrice, notification.TargetPrice,
		notification.CurrentPrice, notification.Discount, notification.GameID, notification.AlertID)

	// Create plain text version
	textContent := fmt.Sprintf(`
Price Drop Alert - DropsAndGrinds

Great news! Game ID %d is now available at a lower price.

Current Price: ₹%.2f
Your Target: ₹%.2f
You Save: ₹%.2f (%d%% discount)

This is the perfect time to buy! The price has dropped below your target price.

View the deal here: https://dropsandgrinds.com/game/%d

---
This alert was triggered because you set a price alert for this game.
Manage your alerts: https://dropsandgrinds.com/alerts/manage
Unsubscribe: https://dropsandgrinds.com/unsubscribe?alert=%d
	`, notification.GameID, notification.CurrentPrice, notification.TargetPrice,
		notification.TargetPrice-notification.CurrentPrice, notification.Discount,
		notification.GameID, notification.AlertID)

	// Send email using Resend
	if e.client != nil {
		params := &resend.SendEmailRequest{
			From:    e.fromEmail,
			To:      []string{userEmail},
			Subject: subject,
			Html:    htmlContent,
			Text:    textContent,
		}

		sent, err := e.client.Emails.Send(params)
		if err != nil {
			e.logger.Error("failed to send price drop email via Resend",
				"error", err, "user_email", userEmail, "game_id", notification.GameID)
			return fmt.Errorf("failed to send email: %w", err)
		}

		e.logger.Info("price drop email sent successfully via Resend",
			"user_email", userEmail, "game_id", notification.GameID, "message_id", sent.Id)
		return nil
	}

	// Demo mode - log the email that would be sent
	e.logger.Info("DEMO MODE: Price drop email would be sent",
		"to", userEmail,
		"subject", subject,
		"game_id", notification.GameID,
		"target_price", notification.TargetPrice,
		"current_price", notification.CurrentPrice,
		"discount", notification.Discount,
		"html_length", len(htmlContent))

	return nil
}

// SendWelcomeEmail sends a welcome email to new users
func (e *EmailService) SendWelcomeEmail(ctx context.Context, userEmail, userName string) error {
	subject := "Welcome to DropsAndGrinds - Your Game Deal Tracker! 🎮"

	htmlContent := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Welcome to DropsAndGrinds</title>
		<style>
			body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
			.container { max-width: 600px; margin: 0 auto; padding: 20px; }
			.header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
			.content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
			.feature { background: white; padding: 20px; margin: 15px 0; border-radius: 5px; border-left: 4px solid #667eea; }
			.cta-button { background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; margin: 20px 0; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>🎮 Welcome to DropsAndGrinds!</h1>
				<p>Hi %s, thanks for joining our community of smart gamers!</p>
			</div>
			<div class="content">
				<h2>Start Saving on Games Today</h2>
				<p>DropsAndGrinds helps you never miss a great deal on your favorite games. Here's what you can do:</p>
				
				<div class="feature">
					<h3>🔍 Track Game Prices</h3>
					<p>Monitor prices across Steam, Epic Games, PlayStation, Xbox, Nintendo, and Indian-friendly stores.</p>
				</div>
				
				<div class="feature">
					<h3>🔔 Set Price Alerts</h3>
					<p>Get notified instantly when games drop below your target price.</p>
				</div>
				
				<div class="feature">
					<h3>💰 Find Best Deals</h3>
					<p>Compare prices across multiple platforms and find the best value for your money.</p>
				</div>
				
				<div class="feature">
					<h3>🇮🇳 Indian Payment Support</h3>
					<p>Discover UPI discounts, card cashback, and wallet bonuses for Indian gamers.</p>
				</div>
				
				<div style="text-align: center;">
					<a href="https://dropsandgrinds.com" class="cta-button">Start Finding Deals</a>
				</div>
				
				<hr style="margin: 30px 0; border: none; border-top: 1px solid #ddd;">
				<p style="color: #666; font-size: 14px;">
					Have questions? Visit our <a href="https://dropsandgrinds.com/help">Help Center</a><br>
					Too many emails? <a href="https://dropsandgrinds.com/preferences">Manage your preferences</a>
				</p>
			</div>
		</div>
	</body>
	</html>
	`, userName)

	textContent := fmt.Sprintf(`
Welcome to DropsAndGrinds - Your Game Deal Tracker!

Hi %s,

Thanks for joining DropsAndGrinds! We're excited to help you save money on games.

What you can do:
- Track game prices across multiple platforms
- Set price alerts and get notified of deals
- Find the best value for your money
- Discover Indian payment discounts

Start finding deals: https://dropsandgrinds.com

Questions? Visit our help center: https://dropsandgrinds.com/help

---
Manage preferences: https://dropsandgrinds.com/preferences
	`, userName)

	if e.client != nil {
		params := &resend.SendEmailRequest{
			From:    e.fromEmail,
			To:      []string{userEmail},
			Subject: subject,
			Html:    htmlContent,
			Text:    textContent,
		}

		sent, err := e.client.Emails.Send(params)
		if err != nil {
			e.logger.Error("failed to send welcome email via Resend",
				"error", err, "user_email", userEmail)
			return fmt.Errorf("failed to send welcome email: %w", err)
		}

		e.logger.Info("welcome email sent successfully via Resend",
			"user_email", userEmail, "message_id", sent.Id)
		return nil
	}

	e.logger.Info("DEMO MODE: Welcome email would be sent",
		"to", userEmail, "name", userName)
	return nil
}

// SendBulkNotifications sends multiple price drop notifications in bulk
func (e *EmailService) SendBulkNotifications(ctx context.Context, notifications []models.PriceDropNotification, userEmails map[int64]string) error {
	if e.client == nil {
		e.logger.Info("DEMO MODE: Bulk email notifications would be sent",
			"count", len(notifications))
		return nil
	}

	totalSent := 0
	totalErrors := 0

	for _, notification := range notifications {
		userEmail, exists := userEmails[notification.UserID]
		if !exists {
			e.logger.Warn("user email not found for notification",
				"user_id", notification.UserID, "alert_id", notification.AlertID)
			continue
		}

		err := e.SendPriceDropNotification(ctx, notification, userEmail)
		if err != nil {
			totalErrors++
			e.logger.Error("failed to send bulk notification",
				"error", err, "alert_id", notification.AlertID)
		} else {
			totalSent++
		}
	}

	e.logger.Info("bulk email notifications completed",
		"total_notifications", len(notifications),
		"sent", totalSent,
		"errors", totalErrors)

	return nil
}
